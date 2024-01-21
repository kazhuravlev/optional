package optional

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
)

// Scan implements the Scanner interface.
func (v *Val[T]) Scan(value any) error {
	if scanner, ok := any(&v.value).(sql.Scanner); ok {
		if err := scanner.Scan(value); err != nil {
			v.value, v.hasVal = *new(T), false

			return fmt.Errorf("scan value: %w", err)
		}

		v.hasVal = true
		return nil
	}

	if value == nil {
		v.value, v.hasVal = *new(T), false
		return nil
	}

	val, ok := value.(T)
	if !ok {
		return errors.New("unexpected value type")
	}

	v.hasVal = true
	v.value = val

	return nil
}

// Value implements the driver Valuer interface.
func (v Val[T]) Value() (driver.Value, error) {
	if !v.hasVal {
		return nil, nil
	}

	if v, ok := any(v.value).(driver.Valuer); ok {
		res, err := v.Value()
		if err != nil {
			return nil, fmt.Errorf("get driver value: %w", err)
		}

		return res, nil
	}

	return v.value, nil
}
