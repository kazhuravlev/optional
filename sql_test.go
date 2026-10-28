package optional

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBadValuer = errors.New("boom")

type badValuer struct{}

func (badValuer) Value() (driver.Value, error) {
	return nil, errBadValuer
}

// nonSubsetValuer returns a type outside the driver.Value subset.
type nonSubsetValuer struct{}

func (nonSubsetValuer) Value() (driver.Value, error) {
	return int(1), nil
}

// nilTolerantScanner mirrors scanners such as github.com/google/uuid.UUID
// that accept NULL and silently keep their zero value.
type nilTolerantScanner struct {
	S string
}

func (s *nilTolerantScanner) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		return nil
	case string:
		s.S = v
	case []byte:
		s.S = string(v)
	default:
		return errors.New("unsupported source")
	}

	return nil
}

// destScanner mirrors scanners such as pq.GenericArray that write through a
// destination prepared on the value before scanning.
type destScanner struct {
	dst *[]byte
}

func (s *destScanner) Scan(src any) error {
	if s.dst == nil {
		return errors.New("no destination")
	}

	b, ok := src.([]byte)
	if !ok {
		return errors.New("want bytes")
	}

	*s.dst = bytes.Clone(b)

	return nil
}

type myString string

func TestScan(t *testing.T) {
	t.Parallel()

	t.Run("int64", func(t *testing.T) {
		type data = Val[int64]
		table := []struct {
			in     any
			exp    data
			expErr bool
		}{
			0: {
				in: nil,
				exp: data{
					hasVal: false,
					value:  0,
				},
				expErr: false,
			},
			1: {
				in: []byte("asdasdasd"),
				exp: data{
					hasVal: false,
					value:  0,
				},
				expErr: true,
			},
			2: {
				in: int64(0),
				exp: data{
					hasVal: true,
					value:  0,
				},
				expErr: false,
			},
			3: {
				in: int64(42),
				exp: data{
					hasVal: true,
					value:  42,
				},
				expErr: false,
			},
			4: {
				in: int64(42),
				exp: data{
					hasVal: true,
					value:  42,
				},
				expErr: false,
			},
		}

		for i := range table {
			row := table[i]
			t.Run("", func(t *testing.T) {
				t.Parallel()

				var res data
				if err := res.Scan(row.in); row.expErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Equal(t, row.exp, res)
			})
		}
	})

	t.Run("sql_Scanner", func(t *testing.T) {
		// this type is used to test scenario with Val[T].value.Scan
		type data = Val[sql.NullBool]
		table := []struct {
			in  any
			exp data
		}{
			0: {
				in: nil,
				exp: data{
					hasVal: false, // NULL is always an empty container, even when T can scan nil itself.
					value:  sql.NullBool{Bool: false, Valid: false},
				},
			},
			1: {
				in: true,
				exp: data{
					hasVal: true,
					value:  sql.NullBool{Bool: true, Valid: true},
				},
			},
			2: {
				in: false,
				exp: data{
					hasVal: true,
					value:  sql.NullBool{Bool: false, Valid: true},
				},
			},
			3: {
				in: []byte("false"),
				exp: data{
					hasVal: true,
					value:  sql.NullBool{Bool: false, Valid: true},
				},
			},
		}

		for i := range table {
			row := table[i]
			t.Run("", func(t *testing.T) {
				t.Parallel()

				var res data
				err := res.Scan(row.in)
				require.NoError(t, err)
				assert.Equal(t, row.exp, res)
			})
		}
	})

	t.Run("sql_Scanner_bad_scenario", func(t *testing.T) {
		t.Parallel()

		var val Val[sql.NullBool]
		err := val.Scan("not-valid-data")
		require.Error(t, err)
	})

	t.Run("sql_Scanner_bad_scenario_resets_existing_value", func(t *testing.T) {
		t.Parallel()

		val := New(sql.NullBool{Bool: true, Valid: true})

		err := val.Scan("not-valid-data")
		require.Error(t, err)
		assert.False(t, val.hasVal)
		assert.Equal(t, sql.NullBool{}, val.value)
	})
}

func scanInto[T any](t *testing.T, in any) (Val[T], error) {
	t.Helper()

	var v Val[T]
	err := v.Scan(in)

	return v, err
}

// TestScanConversions checks that Scan applies the same conversions database/sql
// applies to plain destinations: drivers only ever return int64, float64, bool,
// []byte, string and time.Time.
func TestScanConversions(t *testing.T) {
	t.Parallel()

	t.Run("int_from_int64", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[int](t, int64(7))
		require.NoError(t, err)
		assert.Equal(t, New(7), v)
	})

	t.Run("int32_from_int64", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[int32](t, int64(7))
		require.NoError(t, err)
		assert.Equal(t, New(int32(7)), v)
	})

	t.Run("uint8_from_int64", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[uint8](t, int64(200))
		require.NoError(t, err)
		assert.Equal(t, New(uint8(200)), v)
	})

	t.Run("uint8_out_of_range", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[uint8](t, int64(300))
		require.Error(t, err)
		assert.Equal(t, Empty[uint8](), v)
	})

	t.Run("int_from_string", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[int](t, "15")
		require.NoError(t, err)
		assert.Equal(t, New(15), v)
	})

	t.Run("float32_from_float64", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[float32](t, float64(1.5))
		require.NoError(t, err)
		assert.Equal(t, New(float32(1.5)), v)
	})

	t.Run("string_from_bytes", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[string](t, []byte("hi"))
		require.NoError(t, err)
		assert.Equal(t, New("hi"), v)
	})

	t.Run("string_from_int64", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[string](t, int64(5))
		require.NoError(t, err)
		assert.Equal(t, New("5"), v)
	})

	t.Run("named_string_from_bytes", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[myString](t, []byte("x"))
		require.NoError(t, err)
		assert.Equal(t, New(myString("x")), v)
	})

	t.Run("named_string_from_string", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[myString](t, "x")
		require.NoError(t, err)
		assert.Equal(t, New(myString("x")), v)
	})

	t.Run("bool_from_int64", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[bool](t, int64(1))
		require.NoError(t, err)
		assert.Equal(t, New(true), v)
	})

	t.Run("bool_from_bytes", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[bool](t, []byte("false"))
		require.NoError(t, err)
		assert.Equal(t, New(false), v)
	})

	t.Run("bool_invalid", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[bool](t, int64(7))
		require.Error(t, err)
		assert.Equal(t, Empty[bool](), v)
	})

	t.Run("bytes_are_copied", func(t *testing.T) {
		t.Parallel()

		in := []byte("abc")
		v, err := scanInto[[]byte](t, in)
		require.NoError(t, err)

		in[0] = 'X' // drivers reuse their buffers between rows
		assert.Equal(t, New([]byte("abc")), v)
	})

	t.Run("raw_bytes_are_copied", func(t *testing.T) {
		t.Parallel()

		in := []byte("abc")
		v, err := scanInto[sql.RawBytes](t, in)
		require.NoError(t, err)

		in[0] = 'X'
		assert.Equal(t, New(sql.RawBytes("abc")), v)
	})

	t.Run("raw_message_from_bytes", func(t *testing.T) {
		t.Parallel()

		in := []byte(`{"a":1}`)
		v, err := scanInto[json.RawMessage](t, in)
		require.NoError(t, err)

		in[0] = 'X'
		assert.Equal(t, New(json.RawMessage(`{"a":1}`)), v)
	})

	t.Run("time", func(t *testing.T) {
		t.Parallel()

		ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		v, err := scanInto[time.Time](t, ts)
		require.NoError(t, err)
		assert.Equal(t, New(ts), v)
	})

	t.Run("any_keeps_driver_value", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[any](t, int64(1))
		require.NoError(t, err)
		assert.Equal(t, New[any](int64(1)), v)

		in := []byte("abc")
		v, err = scanInto[any](t, in)
		require.NoError(t, err)
		in[0] = 'X'
		assert.Equal(t, New[any]([]byte("abc")), v)
	})

	t.Run("pointer_element", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[*int](t, int64(7))
		require.NoError(t, err)
		require.True(t, v.HasVal())
		require.NotNil(t, v.Val())
		assert.Equal(t, 7, *v.Val())
	})

	t.Run("scanner_type", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[nilTolerantScanner](t, "x")
		require.NoError(t, err)
		assert.Equal(t, New(nilTolerantScanner{S: "x"}), v)
	})

	t.Run("scanner_type_scans_in_place", func(t *testing.T) {
		t.Parallel()

		var got []byte
		v := New(destScanner{dst: &got})

		require.NoError(t, v.Scan([]byte("x")))
		assert.True(t, v.HasVal())
		assert.Equal(t, []byte("x"), got, "state prepared on the value must survive Scan")

		var fresh Val[destScanner]
		err := fresh.Scan([]byte("x"))
		require.ErrorContains(t, err, "no destination")
		assert.Equal(t, Empty[destScanner](), fresh)
	})

	t.Run("scanner_type_error", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[nilTolerantScanner](t, int64(1))
		require.Error(t, err)
		require.ErrorContains(t, err, "scan value")
		assert.Equal(t, Empty[nilTolerantScanner](), v)
	})

	t.Run("unconvertible", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[struct{ A int }](t, int64(1))
		require.Error(t, err)
		require.ErrorContains(t, err, "scan value")
		assert.Equal(t, Empty[struct{ A int }](), v)
	})
}

func TestScanNullIsAlwaysEmpty(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[int](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[int](), v)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[string](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[string](), v)
	})

	t.Run("bytes", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[[]byte](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[[]byte](), v)
	})

	t.Run("any", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[any](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[any](), v)
	})

	t.Run("pointer", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[*int](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[*int](), v)
	})

	t.Run("nil_tolerant_scanner", func(t *testing.T) {
		t.Parallel()

		// The inner scanner would happily accept nil and report "present with
		// zero value"; the container must not trust it and report NULL as empty.
		v, err := scanInto[nilTolerantScanner](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[nilTolerantScanner](), v)
	})

	t.Run("sql_null_string", func(t *testing.T) {
		t.Parallel()

		v, err := scanInto[sql.NullString](t, nil)
		require.NoError(t, err)
		assert.Equal(t, Empty[sql.NullString](), v)
	})

	t.Run("overwrites_existing_value", func(t *testing.T) {
		t.Parallel()

		v := New(42)
		require.NoError(t, v.Scan(nil))
		assert.Equal(t, Empty[int](), v)
	})
}

func TestScanErrorResetsExistingValue(t *testing.T) {
	t.Parallel()

	v := New(5)
	err := v.Scan([]byte("abc"))
	require.Error(t, err)
	require.ErrorContains(t, err, "scan value")
	assert.Equal(t, Empty[int](), v)
}

func TestValue(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		val, err := Empty[int64]().Value()
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("not_empty", func(t *testing.T) {
		t.Parallel()

		val, err := New[int64](42).Value()
		require.NoError(t, err)
		assert.Equal(t, int64(42), val)
	})

	t.Run("not_empty_valuer", func(t *testing.T) {
		t.Parallel()

		val, err := New[sql.NullBool](sql.NullBool{
			Bool:  true,
			Valid: true,
		}).Value()
		require.NoError(t, err)
		assert.Equal(t, true, val)
	})

	t.Run("valuer_error", func(t *testing.T) {
		t.Parallel()

		val, err := New[badValuer](badValuer{}).Value()
		require.Error(t, err)
		assert.Nil(t, val)
		require.ErrorContains(t, err, "get driver value")
		require.ErrorIs(t, err, errBadValuer)
	})
}

// TestValueConversions checks that Value always returns a member of the
// driver.Value subset, otherwise database/sql rejects the argument with
// "non-Value type ... returned from Value".
func TestValueConversions(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	seven := 7

	table := []struct {
		name string
		in   driver.Valuer
		exp  driver.Value
	}{
		{name: "int", in: New(42), exp: int64(42)},
		{name: "int32", in: New(int32(42)), exp: int64(42)},
		{name: "uint8", in: New(uint8(200)), exp: int64(200)},
		{name: "uint64", in: New(uint64(1)), exp: int64(1)},
		{name: "float32", in: New(float32(1.5)), exp: float64(1.5)},
		{name: "float64", in: New(1.5), exp: 1.5},
		{name: "bool", in: New(true), exp: true},
		{name: "string", in: New("x"), exp: "x"},
		{name: "named_string", in: New(myString("x")), exp: "x"},
		{name: "bytes", in: New([]byte("x")), exp: []byte("x")},
		{name: "raw_message", in: New(json.RawMessage(`{}`)), exp: []byte(`{}`)},
		{name: "time", in: New(ts), exp: ts},
		{name: "duration", in: New(time.Second), exp: int64(time.Second)},
		{name: "nil_pointer", in: New[*int](nil), exp: nil},
		{name: "pointer", in: New(&seven), exp: int64(7)},
		{name: "valuer_invalid", in: New(sql.NullBool{}), exp: nil},
		{name: "valuer_nil_pointer_receiver", in: New[*sql.NullBool](nil), exp: nil},
		{name: "any", in: New[any]("x"), exp: "x"},
	}

	for i := range table {
		row := table[i]
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()

			val, err := row.in.Value()
			require.NoError(t, err)
			assert.True(t, driver.IsValue(val), "%T is not a driver.Value", val)
			assert.Equal(t, row.exp, val)
		})
	}

	t.Run("valuer_result_is_normalised", func(t *testing.T) {
		t.Parallel()

		val, err := New(nonSubsetValuer{}).Value()
		require.NoError(t, err)
		assert.Equal(t, int64(1), val)
	})

	// Values the default converter rejects are handed over unchanged, so a
	// permissive driver (pgx) can encode them; a strict driver rejects them
	// with its own error, exactly as it did before.
	passthrough := []struct {
		name string
		in   driver.Valuer
		exp  any
	}{
		{name: "slice", in: New([]string{"a", "b"}), exp: []string{"a", "b"}},
		{name: "map", in: New(map[string]int{"a": 1}), exp: map[string]int{"a": 1}},
		{name: "struct", in: New(struct{ A int }{A: 1}), exp: struct{ A int }{A: 1}},
		{name: "array", in: New([2]byte{1, 2}), exp: [2]byte{1, 2}},
		{name: "uint64_high_bit", in: New(uint64(1 << 63)), exp: uint64(1 << 63)},
	}

	for i := range passthrough {
		row := passthrough[i]
		t.Run("passthrough_"+row.name, func(t *testing.T) {
			t.Parallel()

			val, err := row.in.Value()
			require.NoError(t, err)
			assert.Equal(t, row.exp, val)
		})
	}
}

// ---- integration with database/sql through a stub driver ----

var (
	stubDatasets sync.Map // query -> [][]driver.Value
	stubExecArgs sync.Map // query -> []driver.Value
)

type stubDriver struct {
	permissive bool
}

func (d stubDriver) Open(string) (driver.Conn, error) {
	if d.permissive {
		return stubPermissiveConn{}, nil
	}

	return stubConn{}, nil
}

type stubConn struct{}

func (stubConn) Prepare(query string) (driver.Stmt, error) { return stubStmt{query: query}, nil }
func (stubConn) Close() error                              { return nil }
func (stubConn) Begin() (driver.Tx, error)                 { return nil, errors.New("not supported") }

// stubPermissiveConn accepts any argument, like pgx stdlib, and resolves
// driver.Valuer itself inside Exec.
type stubPermissiveConn struct {
	stubConn
}

func (stubPermissiveConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (stubPermissiveConn) Prepare(query string) (driver.Stmt, error) {
	return stubStmt{query: query, resolveValuers: true}, nil
}

type stubStmt struct {
	query          string
	resolveValuers bool
}

func (stubStmt) Close() error  { return nil }
func (stubStmt) NumInput() int { return -1 }

func (s stubStmt) Exec(args []driver.Value) (driver.Result, error) {
	if s.resolveValuers {
		for i, arg := range args {
			valuer, ok := arg.(driver.Valuer)
			if !ok {
				continue
			}

			val, err := valuer.Value()
			if err != nil {
				return nil, err
			}

			args[i] = val
		}
	}

	stubExecArgs.Store(s.query, args)

	return driver.RowsAffected(1), nil
}

func (s stubStmt) Query([]driver.Value) (driver.Rows, error) {
	rows, ok := stubDatasets.Load(s.query)
	if !ok {
		return nil, errors.New("unknown dataset: " + s.query)
	}

	return &stubRows{rows: rows.([][]driver.Value), buf: make([]byte, 0, 64)}, nil //nolint:forcetypeassert
}

// stubRows hands out []byte values backed by one reusable buffer, exactly as
// real drivers do.
type stubRows struct {
	rows [][]driver.Value
	next int
	buf  []byte
}

func (*stubRows) Columns() []string { return []string{"c"} }
func (*stubRows) Close() error      { return nil }

func (r *stubRows) Next(dest []driver.Value) error {
	if r.next >= len(r.rows) {
		return io.EOF
	}

	for i, val := range r.rows[r.next] {
		if b, ok := val.([]byte); ok {
			r.buf = append(r.buf[:0], b...)
			dest[i] = r.buf

			continue
		}

		dest[i] = val
	}

	r.next++

	return nil
}

func init() { //nolint:gochecknoinits
	sql.Register("optional-stub", stubDriver{permissive: false})
	sql.Register("optional-stub-permissive", stubDriver{permissive: true})
}

func TestDatabaseSQLIntegration(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("optional-stub", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	scanRows := func(t *testing.T, query string, dests ...any) {
		t.Helper()

		rows, err := db.Query(query)
		require.NoError(t, err)

		defer func() { _ = rows.Close() }()

		for _, dest := range dests {
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(dest))
		}

		require.False(t, rows.Next())
		require.NoError(t, rows.Err())
	}

	t.Run("exec_arguments_are_converted", func(t *testing.T) {
		t.Parallel()

		ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		_, err := db.Exec("exec_args",
			New(42), New(myString("x")), Empty[int](), New(ts),
			New(json.RawMessage(`{}`)), New(sql.NullBool{Bool: true, Valid: true}))
		require.NoError(t, err)

		got, ok := stubExecArgs.Load("exec_args")
		require.True(t, ok)
		assert.Equal(t, []driver.Value{int64(42), "x", nil, ts, []byte(`{}`), true}, got)
	})

	t.Run("exec_strict_driver_rejects_non_subset_type", func(t *testing.T) {
		t.Parallel()

		_, err := db.Exec("exec_bad", New([]string{"a"}))
		require.Error(t, err)
		require.ErrorContains(t, err, "non-Value type []string")
	})

	t.Run("exec_permissive_driver_receives_non_subset_type", func(t *testing.T) {
		t.Parallel()

		pdb, err := sql.Open("optional-stub-permissive", "")
		require.NoError(t, err)
		t.Cleanup(func() { _ = pdb.Close() })

		_, err = pdb.Exec("exec_permissive", New([]string{"a", "b"}), New(42), Empty[int]())
		require.NoError(t, err)

		got, ok := stubExecArgs.Load("exec_permissive")
		require.True(t, ok)
		assert.Equal(t, []driver.Value{[]string{"a", "b"}, int64(42), nil}, got)
	})

	t.Run("scan_converts_driver_values", func(t *testing.T) {
		t.Parallel()

		stubDatasets.Store("scan_conv", [][]driver.Value{{int64(7)}, {[]byte("hello")}, {nil}, {nil}})

		var (
			vi  Val[int]
			vs  Val[string]
			vn  Val[int]
			vsc Val[nilTolerantScanner]
		)
		scanRows(t, "scan_conv", &vi, &vs, &vn, &vsc)

		assert.Equal(t, New(7), vi)
		assert.Equal(t, New("hello"), vs)
		assert.Equal(t, Empty[int](), vn)
		assert.Equal(t, Empty[nilTolerantScanner](), vsc)
	})

	t.Run("scan_bytes_do_not_alias_driver_buffer", func(t *testing.T) {
		t.Parallel()

		stubDatasets.Store("scan_bytes", [][]driver.Value{{[]byte("first")}, {[]byte("second")}})

		var first, second Val[[]byte]
		scanRows(t, "scan_bytes", &first, &second)

		assert.Equal(t, New([]byte("first")), first)
		assert.Equal(t, New([]byte("second")), second)
	})

	t.Run("scan_conversion_error", func(t *testing.T) {
		t.Parallel()

		stubDatasets.Store("scan_bad", [][]driver.Value{{[]byte("abc")}})

		rows, err := db.Query("scan_bad")
		require.NoError(t, err)

		defer func() { _ = rows.Close() }()

		require.True(t, rows.Next())

		var v Val[int]
		err = rows.Scan(&v)
		require.Error(t, err)
		require.ErrorContains(t, err, "scan value")
		assert.Equal(t, Empty[int](), v)
		require.NoError(t, rows.Err())
	})
}
