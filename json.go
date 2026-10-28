package optional

import (
	"encoding/json"
	"errors"
	"fmt"
)

const jsonNull = "null"

var errEmptyJSONInput = errors.New("empty json input")

// UnmarshalJSON implements json.Unmarshaler.
//
// JSON null (and a missing field, which encoding/json never reports) produce
// an empty container with a zero value, even if the container held a value
// before. Any other input is decoded into the stored value in place, with the
// usual encoding/json semantics for maps and structs (existing entries and
// fields that are not mentioned in the input are kept).
func (v *Val[T]) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return errEmptyJSONInput
	}

	if string(buf) == jsonNull {
		v.Reset()

		return nil
	}

	if err := json.Unmarshal(buf, &v.value); err != nil {
		return fmt.Errorf("unmarshal optional value: %w", err)
	}

	v.hasVal = true

	return nil
}

// MarshalJSON implements json.Marshaler.
//
// An empty container is encoded as null. To drop the field instead use the
// `omitzero` tag option (Go 1.24+); `omitempty` has no effect on struct
// types such as Val.
func (v Val[T]) MarshalJSON() ([]byte, error) {
	if !v.hasVal {
		return []byte(jsonNull), nil
	}

	res, err := json.Marshal(v.value)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	return res, nil
}
