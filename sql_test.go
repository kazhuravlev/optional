package optional

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

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
					hasVal: true, // scanner of underlay type return no error. that means that input was successful parsed.
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
		var val Val[sql.NullBool]
		err := val.Scan("not-valid-data")
		require.Error(t, err)
	})
}

func TestValue(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		val, err := Empty[int64]().Value()
		require.NoError(t, err)
		assert.Equal(t, nil, val)
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
}
