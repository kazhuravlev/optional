package optional

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// UnmarshalJSON implements json.Unmarshaler.
func (v *Val[T]) UnmarshalJSON(buf []byte) error {
	switch len(buf) {
	case 0:
		return errors.New("empty json input")
	case 4:
		if bytes.Equal(buf, []byte("null")) {
			v.hasVal = false

			return nil
		}
	}

	if err := json.Unmarshal(buf, &v.value); err != nil {
		return fmt.Errorf("unmarshal optional value: %w", err)
	}

	v.hasVal = true

	return nil
}

// MarshalJSON implements json.Marshaler.
func (v Val[T]) MarshalJSON() ([]byte, error) {
	if !v.hasVal {
		return []byte("null"), nil
	}

	res, err := json.Marshal(v.value)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	return res, nil
}
