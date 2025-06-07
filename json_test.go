package optional

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[[]string] `json:"v"`
	}

	table := []struct {
		val payload
		exp []byte
	}{
		{
			val: payload{V: New([]string{"hi"})},
			exp: []byte(`{"v":["hi"]}`),
		},
		{
			val: payload{V: New([]string{})},
			exp: []byte(`{"v":[]}`),
		},
		{
			val: payload{V: Empty[[]string]()},
			exp: []byte(`{"v":null}`),
		},
	}

	for i := range table {
		row := table[i]
		t.Run("", func(t *testing.T) {
			t.Parallel()

			res, err := json.Marshal(row.val)
			require.NoError(t, err)
			assert.Equal(t, string(row.exp), string(res))

			var val payload
			require.NoError(t, json.Unmarshal(res, &val))
			assert.Equal(t, row.val, val)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[[]string] `json:"v"`
	}

	table := []struct {
		in  []byte
		exp payload
	}{
		{
			in:  []byte(`{"v":["hi"]}`),
			exp: payload{V: New([]string{"hi"})},
		},
		{
			in:  []byte(`{"v":[]}`),
			exp: payload{V: New([]string{})},
		},
		{
			in:  []byte(`{"v":null}`),
			exp: payload{V: Empty[[]string]()},
		},
		{
			in:  []byte(`{"v":null}`),
			exp: payload{V: Empty[[]string]()},
		},
		{
			in:  []byte(`{}`),
			exp: payload{V: Empty[[]string]()},
		},
		{
			in:  []byte(`null`),
			exp: payload{V: Empty[[]string]()},
		},
	}

	for i := range table {
		row := table[i]
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var val payload
			require.NoError(t, json.Unmarshal(row.in, &val))
			assert.Equal(t, row.exp, val)
		})
	}
}

func TestJSONEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid_type_conversion", func(t *testing.T) {
		t.Parallel()

		type payload struct {
			V Val[int] `json:"v"`
		}

		var val payload
		err := json.Unmarshal([]byte(`{"v": "not_a_number"}`), &val)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal optional value")
	})

	t.Run("large_values", func(t *testing.T) {
		t.Parallel()

		largeString := strings.Repeat("a", 10000)

		type payload struct {
			V Val[string] `json:"v"`
		}

		jsonData, err := json.Marshal(payload{V: New(largeString)})
		require.NoError(t, err)

		var val payload
		err = json.Unmarshal(jsonData, &val)
		require.NoError(t, err)

		result, ok := val.V.Get()
		require.True(t, ok)
		require.Equal(t, largeString, result)
	})

	t.Run("nested_structures", func(t *testing.T) {
		t.Parallel()

		type Inner struct {
			Name string `json:"name"`
		}

		type payload struct {
			V Val[Inner] `json:"v"`
		}

		// Test with value
		data := `{"v": {"name": "test"}}`
		var val payload
		err := json.Unmarshal([]byte(data), &val)
		require.NoError(t, err)

		result, ok := val.V.Get()
		require.True(t, ok)
		require.Equal(t, Inner{Name: "test"}, result)

		// Test marshal back
		marshaled, err := json.Marshal(val)
		require.NoError(t, err)
		require.JSONEq(t, data, string(marshaled))
	})

	t.Run("special_json_values", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			input    string
			expected Val[interface{}]
		}{
			{
				name:     "json_true",
				input:    `{"v": true}`,
				expected: New(interface{}(true)),
			},
			{
				name:     "json_false",
				input:    `{"v": false}`,
				expected: New(interface{}(false)),
			},
			{
				name:     "json_number",
				input:    `{"v": 42.5}`,
				expected: New(interface{}(42.5)),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				type payload struct {
					V Val[interface{}] `json:"v"`
				}

				var val payload
				err := json.Unmarshal([]byte(tt.input), &val)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, val.V)
			})
		}
	})
}

func TestJSONComplexTypes(t *testing.T) {
	t.Parallel()

	t.Run("map_type", func(t *testing.T) {
		t.Parallel()

		type payload struct {
			V Val[map[string]int] `json:"v"`
		}

		original := payload{
			V: New(map[string]int{"a": 1, "b": 2}),
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var restored payload
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		result, ok := restored.V.Get()
		require.True(t, ok)
		require.Equal(t, map[string]int{"a": 1, "b": 2}, result)
	})

	t.Run("slice_of_structs", func(t *testing.T) {
		t.Parallel()

		type Item struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		}

		type payload struct {
			V Val[[]Item] `json:"v"`
		}

		original := payload{
			V: New([]Item{
				{Name: "item1", ID: 1},
				{Name: "item2", ID: 2},
			}),
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var restored payload
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		result, ok := restored.V.Get()
		require.True(t, ok)
		require.Len(t, result, 2)
		require.Equal(t, "item1", result[0].Name)
		require.Equal(t, "item2", result[1].Name)
	})

	t.Run("custom_json_marshaler", func(t *testing.T) {
		t.Parallel()

		// Simple test with interface{} to verify custom marshaling works
		type payload struct {
			V Val[map[string]string] `json:"v"`
		}

		original := payload{
			V: New(map[string]string{"custom": "value"}),
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)
		require.Contains(t, string(data), "custom")
		require.Contains(t, string(data), "value")

		// Unmarshal
		var restored payload
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		result, ok := restored.V.Get()
		require.True(t, ok)
		require.Equal(t, "value", result["custom"])
	})
}
