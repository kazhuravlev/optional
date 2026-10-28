package optional

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
)

// Scan implements the sql.Scanner interface.
//
// SQL NULL always produces an empty container, even when T implements
// sql.Scanner itself. A T that implements sql.Scanner scans in place, so
// state prepared on the value beforehand (destination pointers, codecs,
// locations) stays available to it. Any other value is converted with the
// same rules database/sql applies to plain destinations (int64 to any integer
// type, []byte to string, string to numbers, and so on), and []byte input is
// copied so the container never aliases a buffer owned by the driver. On
// error the container is reset to empty.
func (v *Val[T]) Scan(value any) error {
	if value == nil {
		v.Reset()

		return nil
	}

	if scanner, ok := any(&v.value).(sql.Scanner); ok {
		if err := scanner.Scan(value); err != nil {
			v.Reset()

			return fmt.Errorf("scan value: %w", err)
		}

		v.hasVal = true

		return nil
	}

	var res sql.Null[T]
	if err := res.Scan(value); err != nil {
		v.Reset()

		return fmt.Errorf("scan value: %w", err)
	}

	// database/sql hands driver memory to sql.RawBytes without copying.
	if raw, ok := any(res.V).(sql.RawBytes); ok {
		if cloned, ok := any(sql.RawBytes(bytes.Clone(raw))).(T); ok {
			res.V = cloned
		}
	}

	v.value, v.hasVal = res.V, true

	return nil
}

// Value implements the driver.Valuer interface.
//
// An empty container is NULL. A T that implements driver.Valuer is asked for
// its value first. The result is then normalised to the driver.Value subset
// where driver.DefaultParameterConverter knows how (any integer kind,
// including time.Duration, becomes int64; named string types become string;
// and so on). Anything the converter rejects, such as slices other than
// []byte, maps or structs, is handed to the driver unchanged: permissive
// drivers such as pgx encode those themselves, strict drivers report their
// own error.
func (v Val[T]) Value() (driver.Value, error) {
	if !v.hasVal {
		return nil, nil //nolint:nilnil // nil driver.Value is SQL NULL.
	}

	val := any(v.value)
	if valuer, ok := val.(driver.Valuer); ok {
		res, err := callValuer(valuer)
		if err != nil {
			return nil, fmt.Errorf("get driver value: %w", err)
		}

		val = res
	}

	if res, err := driver.DefaultParameterConverter.ConvertValue(val); err == nil {
		return res, nil
	}

	return val, nil
}

// callValuer mirrors database/sql: a nil pointer whose element type
// implements driver.Valuer with a value receiver must become NULL instead of
// panicking inside Value.
func callValuer(valuer driver.Valuer) (driver.Value, error) {
	rv := reflect.ValueOf(valuer)
	if rv.Kind() == reflect.Pointer && rv.IsNil() &&
		rv.Type().Elem().Implements(reflect.TypeFor[driver.Valuer]()) {
		return nil, nil //nolint:nilnil // nil driver.Value is SQL NULL.
	}

	return valuer.Value() //nolint:wrapcheck // wrapped by the caller.
}
