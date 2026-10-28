package optional

import (
	"testing"
	"time"

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var val payload
			err := yaml.Unmarshal([]byte(tt.yaml), &val)
			require.Error(t, err)
		})
	}
}

func TestYAMLMarshalTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
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

func TestYAMLOmitEmpty(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[string] `yaml:"v,omitempty"`
		N Val[int]    `yaml:"num,omitempty"`
	}

	tests := []struct {
		name string
		val  payload
		exp  string
	}{
		{
			name: "values_present",
			val:  payload{V: New("hello"), N: New(0)},
			exp:  "v: hello\nnum: 0\n",
		},
		{
			name: "present_zero_value_is_kept",
			val:  payload{V: New(""), N: Empty[int]()},
			exp:  "v: \"\"\n",
		},
		{
			name: "empty_containers_are_omitted",
			val:  payload{V: Empty[string](), N: Empty[int]()},
			exp:  "{}\n",
		},
		{
			name: "zero_struct",
			val:  payload{},
			exp:  "{}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := yaml.Marshal(tt.val)
			require.NoError(t, err)
			assert.Equal(t, tt.exp, string(res))
		})
	}
}

func TestYAMLStringsThatLookLikeNull(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[string] `yaml:"v"`
	}

	for _, s := range []string{"null", "~", "Null", "NULL", ""} {
		t.Run(s, func(t *testing.T) {
			t.Parallel()

			res, err := yaml.Marshal(payload{V: New(s)})
			require.NoError(t, err)

			var back payload
			require.NoError(t, yaml.Unmarshal(res, &back))
			assert.Equal(t, payload{V: New(s)}, back, "yaml: %q", res)
		})
	}
}

func TestYAMLNullSpellings(t *testing.T) {
	t.Parallel()

	type strPayload struct {
		V Val[string] `yaml:"v"`
	}
	type structPayload struct {
		V Val[struct {
			A int `yaml:"a"`
		}] `yaml:"v"`
	}

	for _, in := range []string{"v: null", "v: ~", "v:", "v: Null", "v: NULL", "v: !!null"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			var s strPayload
			require.NoError(t, yaml.Unmarshal([]byte(in), &s))
			assert.Equal(t, strPayload{}, s)

			var st structPayload
			require.NoError(t, yaml.Unmarshal([]byte(in), &st))
			assert.Equal(t, structPayload{}, st)
		})
	}
}

// TestYAMLNullDoesNotClearPresetValue documents yaml.v3 behaviour: it never
// passes a null node to an Unmarshaler, so decoding null over a container that
// already holds a value leaves it untouched. A *Val[T] field is set to nil.
func TestYAMLNullDoesNotClearPresetValue(t *testing.T) {
	t.Parallel()

	t.Run("value_field_is_kept", func(t *testing.T) {
		t.Parallel()

		type payload struct {
			V Val[string] `yaml:"v"`
		}

		val := payload{V: New("default")}
		require.NoError(t, yaml.Unmarshal([]byte("v: null"), &val))
		assert.Equal(t, payload{V: New("default")}, val)
	})

	t.Run("pointer_field_is_cleared", func(t *testing.T) {
		t.Parallel()

		type payload struct {
			V *Val[string] `yaml:"v"`
		}

		def := New("default")
		val := payload{V: &def}
		require.NoError(t, yaml.Unmarshal([]byte("v: null"), &val))
		assert.Nil(t, val.V)
	})
}

// TestYAMLCollectionsOfVal documents yaml.v3 behaviour for null elements:
// they are dropped from slices of non-nillable element types (the same
// happens to []string), kept as nil in slices of pointers, and kept as empty
// containers in map values.
func TestYAMLCollectionsOfVal(t *testing.T) {
	t.Parallel()

	t.Run("slice_drops_null_elements", func(t *testing.T) {
		t.Parallel()

		var got []Val[int]
		require.NoError(t, yaml.Unmarshal([]byte("- 1\n- null\n- 3\n"), &got))
		assert.Equal(t, []Val[int]{New(1), New(3)}, got)
	})

	t.Run("slice_of_pointers_keeps_positions", func(t *testing.T) {
		t.Parallel()

		var got []*Val[int]
		require.NoError(t, yaml.Unmarshal([]byte("- 1\n- null\n- 3\n"), &got))
		require.Len(t, got, 3)
		assert.Equal(t, New(1), *got[0])
		assert.Nil(t, got[1])
		assert.Equal(t, New(3), *got[2])
	})

	t.Run("map_keeps_empty_values", func(t *testing.T) {
		t.Parallel()

		var got map[string]Val[int]
		require.NoError(t, yaml.Unmarshal([]byte("a: 1\nb: null\n"), &got))
		assert.Equal(t, map[string]Val[int]{"a": New(1), "b": Empty[int]()}, got)
	})

	t.Run("marshal", func(t *testing.T) {
		t.Parallel()

		res, err := yaml.Marshal([]Val[int]{New(1), Empty[int](), New(3)})
		require.NoError(t, err)
		assert.Equal(t, "- 1\n- null\n- 3\n", string(res))
	})
}

// TestUnmarshalYAMLNullNodeDirect exercises the null branch, which yaml.v3
// only reaches through a direct call (or a different YAML decoder).
func TestUnmarshalYAMLNullNodeDirect(t *testing.T) {
	t.Parallel()

	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("- null\n- ~\n- \"null\"\n- &x null\n- *x\n- ''\n- !!null\n"), &doc))
	nodes := doc.Content[0].Content
	require.Len(t, nodes, 7)

	exp := []Val[string]{
		Empty[string](),
		Empty[string](),
		New("null"),
		Empty[string](),
		Empty[string](),
		New(""),
		Empty[string](),
	}

	for i := range nodes {
		node := nodes[i]
		t.Run(node.Value, func(t *testing.T) {
			t.Parallel()

			v := New("default")
			require.NoError(t, v.UnmarshalYAML(node))
			assert.Equal(t, exp[i], v)
		})
	}

	t.Run("zero_node", func(t *testing.T) {
		t.Parallel()

		v := New("default")
		require.NoError(t, v.UnmarshalYAML(&yaml.Node{}))
		assert.Equal(t, Empty[string](), v)
	})
}

func TestYAMLUnmarshalErrorIsWrapped(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[int] `yaml:"v"`
	}

	var val payload
	err := yaml.Unmarshal([]byte("v: [1]"), &val)
	require.Error(t, err)
	require.ErrorContains(t, err, "unmarshal yaml value")
}

func TestYAMLNestedTypes(t *testing.T) {
	t.Parallel()

	type inner struct {
		Name string `yaml:"name"`
	}
	type payload struct {
		S Val[inner]          `yaml:"s"`
		M Val[map[string]int] `yaml:"m"`
		T Val[time.Time]      `yaml:"t"`
	}

	original := payload{
		S: New(inner{Name: "x"}),
		M: New(map[string]int{"a": 1}),
		T: New(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)),
	}

	res, err := yaml.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, "s:\n    name: x\nm:\n    a: 1\nt: 2024-01-02T03:04:05Z\n", string(res))

	var back payload
	require.NoError(t, yaml.Unmarshal(res, &back))
	assert.Equal(t, original, back)
}
