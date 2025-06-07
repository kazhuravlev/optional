package optional

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYAMLMarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[string] `yaml:"v"`
	}

	tests := []struct {
		name string
		val  payload
		exp  string
	}{
		{
			name: "with_value",
			val:  payload{V: New("hello")},
			exp:  "v: hello\n",
		},
		{
			name: "with_empty_string",
			val:  payload{V: New("")},
			exp:  "v: \"\"\n",
		},
		{
			name: "empty_optional",
			val:  payload{V: Empty[string]()},
			exp:  "v: null\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := yaml.Marshal(tt.val)
			require.NoError(t, err)
			assert.Equal(t, tt.exp, string(res))
		})
	}
}

func TestYAMLMarshalComplex(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[[]string] `yaml:"v"`
	}

	tests := []struct {
		name string
		val  payload
		exp  string
	}{
		{
			name: "slice_with_values",
			val:  payload{V: New([]string{"hello", "world"})},
			exp:  "v:\n    - hello\n    - world\n",
		},
		{
			name: "empty_slice",
			val:  payload{V: New([]string{})},
			exp:  "v: []\n",
		},
		{
			name: "nil_slice",
			val:  payload{V: New([]string(nil))},
			exp:  "v: []\n",
		},
		{
			name: "nil_slice",
			val:  payload{V: Empty[[]string]()},
			exp:  "v: null\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := yaml.Marshal(tt.val)
			require.NoError(t, err)
			assert.Equal(t, tt.exp, string(res))
		})
	}
}

func TestYAMLUnmarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[string] `yaml:"v"`
	}

	tests := []struct {
		name string
		yaml string
		exp  payload
	}{
		{
			name: "with_value",
			yaml: "v: hello",
			exp:  payload{V: New("hello")},
		},
		{
			name: "with_empty_string",
			yaml: "v: \"\"",
			exp:  payload{V: New("")},
		},
		{
			name: "null_value",
			yaml: "v: null",
			exp:  payload{V: Empty[string]()},
		},
		{
			name: "missing_field",
			yaml: "other: value",
			exp:  payload{V: Empty[string]()},
		},
		{
			name: "empty_document",
			yaml: "",
			exp:  payload{V: Empty[string]()},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var val payload
			err := yaml.Unmarshal([]byte(tt.yaml), &val)
			require.NoError(t, err)
			assert.Equal(t, tt.exp, val)
		})
	}
}

func TestYAMLUnmarshalComplex(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[[]string] `yaml:"v"`
	}

	tests := []struct {
		name string
		yaml string
		exp  payload
	}{
		{
			name: "slice_with_values",
			yaml: "v:\n  - hello\n  - world",
			exp:  payload{V: New([]string{"hello", "world"})},
		},
		{
			name: "empty_slice",
			yaml: "v: []",
			exp:  payload{V: New([]string{})},
		},
		{
			name: "null_value",
			yaml: "v: null",
			exp:  payload{V: Empty[[]string]()},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var val payload
			err := yaml.Unmarshal([]byte(tt.yaml), &val)
			require.NoError(t, err)
			assert.Equal(t, tt.exp, val)
		})
	}
}

func TestYAMLRoundtrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		StringVal   Val[string]   `yaml:"string_val"`
		IntVal      Val[int]      `yaml:"int_val"`
		SliceVal    Val[[]string] `yaml:"slice_val"`
		EmptyString Val[string]   `yaml:"empty_string"`
		EmptyInt    Val[int]      `yaml:"empty_int"`
		EmptySlice  Val[[]string] `yaml:"empty_slice"`
	}

	original := payload{
		StringVal:   New("test"),
		IntVal:      New(42),
		SliceVal:    New([]string{"a", "b", "c"}),
		EmptyString: Empty[string](),
		EmptyInt:    Empty[int](),
		EmptySlice:  Empty[[]string](),
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var restored payload
	err = yaml.Unmarshal(yamlData, &restored)
	require.NoError(t, err)

	// Verify roundtrip
	assert.Equal(t, original, restored)
}

func TestYAMLUnmarshalErrors(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[int] `yaml:"v"`
	}

	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "invalid_type",
			yaml: "v: not_a_number",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var val payload
			err := yaml.Unmarshal([]byte(tt.yaml), &val)
			assert.Error(t, err)
		})
	}
}

func TestYAMLMarshalTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  interface{}
		exp  string
	}{
		{
			name: "int",
			val: struct {
				V Val[int] `yaml:"v"`
			}{V: New(42)},
			exp: "v: 42\n",
		},
		{
			name: "bool_true",
			val: struct {
				V Val[bool] `yaml:"v"`
			}{V: New(true)},
			exp: "v: true\n",
		},
		{
			name: "bool_false",
			val: struct {
				V Val[bool] `yaml:"v"`
			}{V: New(false)},
			exp: "v: false\n",
		},
		{
			name: "float",
			val: struct {
				V Val[float64] `yaml:"v"`
			}{V: New(3.14)},
			exp: "v: 3.14\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := yaml.Marshal(tt.val)
			require.NoError(t, err)
			assert.Equal(t, tt.exp, string(res))
		})
	}
}

func TestYAMLUnmarshalTypes(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		var val struct {
			V Val[int] `yaml:"v"`
		}
		err := yaml.Unmarshal([]byte("v: 42"), &val)
		require.NoError(t, err)
		assert.Equal(t, New(42), val.V)
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		var val struct {
			V Val[bool] `yaml:"v"`
		}
		err := yaml.Unmarshal([]byte("v: true"), &val)
		require.NoError(t, err)
		assert.Equal(t, New(true), val.V)
	})

	t.Run("float", func(t *testing.T) {
		t.Parallel()

		var val struct {
			V Val[float64] `yaml:"v"`
		}
		err := yaml.Unmarshal([]byte("v: 3.14"), &val)
		require.NoError(t, err)
		assert.Equal(t, New(3.14), val.V)
	})
}
