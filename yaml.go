package optional

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements the interface for unmarshalling yaml.
func (v *Val[T]) UnmarshalYAML(bb []byte) error {
	if len(bb) == 0 {
		v.hasVal = false
		v.value = *new(T)

		return nil
	}

	if err := yaml.Unmarshal(bb, &v.value); err != nil {
		return fmt.Errorf("unmarshal yaml value: %w", err)
	}

	v.hasVal = true

	return nil
}

// MarshalYAML implements the interface for marshaling yaml.
func (v Val[T]) MarshalYAML() ([]byte, error) {
	if !v.hasVal {
		return []byte("null"), nil
	}

	res, err := yaml.Marshal(v.Val)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml value: %w", err)
	}

	return res, nil
}
