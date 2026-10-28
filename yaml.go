package optional

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// A null node (`null`, `~`, `Null`, an empty value) produces an empty container.
// Be aware that yaml.v3 itself never passes null nodes to an Unmarshaler: when
// decoding with yaml.Unmarshal such a node leaves the destination untouched, so
// a Val that already held a value keeps it. Decode into a zero struct, or use
// *Val[T] (which yaml.v3 sets to nil), when "null clears the value" matters.
func (v *Val[T]) UnmarshalYAML(node *yaml.Node) error {
	if node.ShortTag() == "!!null" {
		v.Reset()

		return nil
	}

	if err := node.Decode(&v.value); err != nil {
		return fmt.Errorf("unmarshal yaml value: %w", err)
	}

	v.hasVal = true

	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// An empty container is encoded as null, or omitted entirely with the
// `omitempty` tag option (see IsZero).
func (v Val[T]) MarshalYAML() (any, error) {
	if !v.hasVal {
		return nil, nil //nolint:nilnil // nil is how yaml encoders are told to emit null.
	}

	return v.value, nil
}
