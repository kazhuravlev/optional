package optional

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

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
		require.Error(t, err)
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
			expected Val[any]
		}{
			{
				name:     "json_true",
				input:    `{"v": true}`,
				expected: New(any(true)),
			},
			{
				name:     "json_false",
				input:    `{"v": false}`,
				expected: New(any(false)),
			},
			{
				name:     "json_number",
				input:    `{"v": 42.5}`,
				expected: New(any(42.5)),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				type payload struct {
					V Val[any] `json:"v"`
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

		// Simple test with any to verify custom marshaling works
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

func TestJSONMarshalError(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[chan int] `json:"v"`
	}

	_, err := json.Marshal(payload{V: New(make(chan int))})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal json")
}

func TestUnmarshalJSONEmptyInput(t *testing.T) {
	t.Parallel()

	var v Val[int]
	err := v.UnmarshalJSON(nil)
	require.EqualError(t, err, "empty json input")
}

func TestUnmarshalJSONNullResetsValue(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		v := New(42)
		require.NoError(t, json.Unmarshal([]byte(`null`), &v))
		assert.Equal(t, Empty[int](), v)
	})

	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		v := New([]string{"a"})
		require.NoError(t, json.Unmarshal([]byte(`null`), &v))
		assert.Equal(t, Empty[[]string](), v)
	})

	t.Run("map", func(t *testing.T) {
		t.Parallel()

		v := New(map[string]int{"a": 1})
		require.NoError(t, json.Unmarshal([]byte(`null`), &v))
		assert.Equal(t, Empty[map[string]int](), v)
	})

	t.Run("struct_field", func(t *testing.T) {
		t.Parallel()

		type inner struct {
			Name string `json:"name"`
		}
		type payload struct {
			V Val[inner] `json:"v"`
		}

		val := payload{V: New(inner{Name: "x"})}
		require.NoError(t, json.Unmarshal([]byte(`{"v":null}`), &val))
		assert.Equal(t, payload{V: Empty[inner]()}, val)
	})

	t.Run("direct_call", func(t *testing.T) {
		t.Parallel()

		v := New("x")
		require.NoError(t, v.UnmarshalJSON([]byte("null")))
		assert.Equal(t, Empty[string](), v)
	})
}

func TestUnmarshalJSONDecoderStream(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[int] `json:"v"`
	}

	// One destination reused for a stream of messages, as with json.Decoder loops.
	dec := json.NewDecoder(strings.NewReader(`{"v":1} {"v":null} {"v":2} {}`))

	var val payload

	require.NoError(t, dec.Decode(&val))
	assert.Equal(t, New(1), val.V)

	require.NoError(t, dec.Decode(&val))
	assert.Equal(t, Empty[int](), val.V, "null must clear the previous value")

	require.NoError(t, dec.Decode(&val))
	assert.Equal(t, New(2), val.V)

	require.NoError(t, dec.Decode(&val))
	assert.Equal(t, New(2), val.V, "a missing field is never reported by encoding/json, so the value is kept")
}

func TestUnmarshalJSONMergesIntoExistingValue(t *testing.T) {
	t.Parallel()

	// Same semantics encoding/json has for plain struct and map fields: keys
	// and fields that are not mentioned in the input are kept.
	type inner struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	v := New(inner{A: 1, B: 2})
	require.NoError(t, json.Unmarshal([]byte(`{"a":5}`), &v))
	assert.Equal(t, New(inner{A: 5, B: 2}), v)

	m := New(map[string]int{"a": 1})
	require.NoError(t, json.Unmarshal([]byte(`{"b":2}`), &m))
	assert.Equal(t, New(map[string]int{"a": 1, "b": 2}), m)
}

func TestMarshalJSONOmitEmptyHasNoEffect(t *testing.T) {
	t.Parallel()

	// encoding/json never treats a struct as empty; use omitzero (Go 1.24+).
	type payload struct {
		V Val[int] `json:"v,omitempty"`
	}

	res, err := json.Marshal(payload{})
	require.NoError(t, err)
	assert.Equal(t, `{"v":null}`, string(res))
}

func TestJSONPointerToVal(t *testing.T) {
	t.Parallel()

	type payload struct {
		V *Val[int] `json:"v"`
	}

	t.Run("null_sets_nil_pointer", func(t *testing.T) {
		t.Parallel()

		val := payload{V: &Val[int]{hasVal: true, value: 1}}
		require.NoError(t, json.Unmarshal([]byte(`{"v":null}`), &val))
		assert.Nil(t, val.V)
	})

	t.Run("value_allocates", func(t *testing.T) {
		t.Parallel()

		var val payload
		require.NoError(t, json.Unmarshal([]byte(`{"v":1}`), &val))
		require.NotNil(t, val.V)
		assert.Equal(t, New(1), *val.V)
	})

	t.Run("marshal", func(t *testing.T) {
		t.Parallel()

		res, err := json.Marshal(payload{})
		require.NoError(t, err)
		assert.Equal(t, `{"v":null}`, string(res))

		one := New(1)
		res, err = json.Marshal(payload{V: &one})
		require.NoError(t, err)
		assert.Equal(t, `{"v":1}`, string(res))
	})
}

// TestJSONNilSliceAndNilPointerDecodeAsEmpty documents a limitation: a present
// nil slice or nil pointer is encoded as null, which decodes as an empty
// container. "Present but null" is not representable in JSON with Val.
func TestJSONNilSliceAndNilPointerDecodeAsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("nil_slice", func(t *testing.T) {
		t.Parallel()

		res, err := json.Marshal(New([]string(nil)))
		require.NoError(t, err)
		assert.Equal(t, `null`, string(res))

		var back Val[[]string]
		require.NoError(t, json.Unmarshal(res, &back))
		assert.Equal(t, Empty[[]string](), back)
	})

	t.Run("nil_pointer", func(t *testing.T) {
		t.Parallel()

		res, err := json.Marshal(New[*int](nil))
		require.NoError(t, err)
		assert.Equal(t, `null`, string(res))

		var back Val[*int]
		require.NoError(t, json.Unmarshal(res, &back))
		assert.Equal(t, Empty[*int](), back)
	})
}

func TestJSONCollectionsOfVal(t *testing.T) {
	t.Parallel()

	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		var got []Val[int]
		require.NoError(t, json.Unmarshal([]byte(`[1,null,3]`), &got))
		assert.Equal(t, []Val[int]{New(1), Empty[int](), New(3)}, got)

		res, err := json.Marshal(got)
		require.NoError(t, err)
		assert.Equal(t, `[1,null,3]`, string(res))
	})

	t.Run("map", func(t *testing.T) {
		t.Parallel()

		var got map[string]Val[int]
		require.NoError(t, json.Unmarshal([]byte(`{"a":1,"b":null}`), &got))
		assert.Equal(t, map[string]Val[int]{"a": New(1), "b": Empty[int]()}, got)

		res, err := json.Marshal(got)
		require.NoError(t, err)
		assert.Equal(t, `{"a":1,"b":null}`, string(res))
	})
}

func TestJSONTextMarshalerValue(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[time.Time] `json:"v"`
	}

	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	res, err := json.Marshal(payload{V: New(ts)})
	require.NoError(t, err)
	assert.JSONEq(t, `{"v":"2024-01-02T03:04:05Z"}`, string(res))

	var back payload
	require.NoError(t, json.Unmarshal(res, &back))
	require.True(t, back.V.HasVal())
	assert.True(t, ts.Equal(back.V.Val()))
}

// celsius has its own JSON representation, with the usual receiver mix.
type celsius float64

func (c celsius) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatFloat(float64(c), 'f', -1, 64) + "C")
}

func (c *celsius) UnmarshalJSON(buf []byte) error {
	var s string
	if err := json.Unmarshal(buf, &s); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(strings.TrimSuffix(s, "C"), 64)
	if err != nil {
		return err
	}

	*c = celsius(f)

	return nil
}

func TestJSONCustomMarshalerValue(t *testing.T) {
	t.Parallel()

	type payload struct {
		V Val[celsius] `json:"v"`
	}

	t.Run("roundtrip", func(t *testing.T) {
		t.Parallel()

		res, err := json.Marshal(payload{V: New(celsius(20.5))})
		require.NoError(t, err)
		assert.JSONEq(t, `{"v":"20.5C"}`, string(res))

		var back payload
		require.NoError(t, json.Unmarshal(res, &back))
		assert.Equal(t, payload{V: New(celsius(20.5))}, back)
	})

	t.Run("null_bypasses_custom_unmarshaler", func(t *testing.T) {
		t.Parallel()

		val := payload{V: New(celsius(1))}
		require.NoError(t, json.Unmarshal([]byte(`{"v":null}`), &val))
		assert.Equal(t, payload{V: Empty[celsius]()}, val)
	})

	t.Run("inner_error_is_wrapped", func(t *testing.T) {
		t.Parallel()

		var val payload
		err := json.Unmarshal([]byte(`{"v":"abc"}`), &val)
		require.Error(t, err)
		require.ErrorContains(t, err, "unmarshal optional value")
		require.ErrorIs(t, err, strconv.ErrSyntax)
	})
}

func TestMarshalJSONZeroValueIsNull(t *testing.T) {
	t.Parallel()

	var v Val[int]
	res, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, `null`, string(res))
}
