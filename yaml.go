package optional

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (v *Val[T]) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.Value == "null" {
		v.hasVal = false
		v.value = *new(T)
		return nil
	}

	if err := node.Decode(&v.value); err != nil {
		return fmt.Errorf("unmarshal yaml value: %w", err)
	}

	v.hasVal = true
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
func (v Val[T]) MarshalYAML() (interface{}, error) {
	if !v.hasVal {
		return nil, nil
	}

	return v.value, nil
}
