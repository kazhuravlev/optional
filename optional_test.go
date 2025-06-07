package optional //nolint:testpackage // required, because accessing to private fields

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyValues(t *testing.T) {
	t.Parallel()

	table := []struct {
		name      string
		jsonBytes []byte
		exp       Val[[]string]
	}{
		{
			name:      "empty_input",
			jsonBytes: []byte("{}"),
			exp: Val[[]string]{
				hasVal: false,
				value:  nil,
			},
		},
		{
			name:      "null_value",
			jsonBytes: []byte(`{"v": null}`),
			exp: Val[[]string]{
				hasVal: false,
				value:  nil,
			},
		},
	}

	for i := range table {
		row := table[i]
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()

			var obj struct {
				V Val[[]string] `json:"v"`
			}
			assert.NoError(t, json.Unmarshal(row.jsonBytes, &obj))
			assert.Equal(t, row.exp, obj.V)
		})
	}
}

func TestNewFromPointer(t *testing.T) {
	t.Parallel()

	val := new(string)
	*val = "value"

	opt := NewFromPointer(val)
	assert.True(t, opt.HasVal())
	assert.Equal(t, "value", opt.Val())

	val = nil
	opt = NewFromPointer(val)
	assert.False(t, opt.HasVal())
	assert.Equal(t, "", opt.Val())
}

func TestNotEmptyValues(t *testing.T) {
	t.Parallel()

	table := []struct {
		jsonBytes []byte
		exp       Val[[]string]
	}{
		{
			jsonBytes: []byte(`{"v": []}`),
			exp: Val[[]string]{
				hasVal: true,
				value:  []string{},
			},
		},
		{
			jsonBytes: []byte(`{"v": ["hi"]}`),
			exp: Val[[]string]{
				hasVal: true,
				value:  []string{"hi"},
			},
		},
	}

	for i := range table {
		row := table[i]
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var obj struct {
				V Val[[]string] `json:"v"`
			}
			assert.NoError(t, json.Unmarshal(row.jsonBytes, &obj))
			assert.Equal(t, row.exp, obj.V)
		})
	}
}

func TestNewEmpty(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Val[string]{
		hasVal: true,
		value:  "hello",
	}, New("hello"))

	assert.Equal(t, Val[int]{
		hasVal: true,
		value:  0,
	}, New(0))

	assert.Equal(t, Val[string]{
		hasVal: false,
		value:  "",
	}, Empty[string]())

	assert.Equal(t, Val[int]{
		hasVal: false,
		value:  0,
	}, Empty[int]())
}

func TestSet(t *testing.T) {
	t.Parallel()

	fOk := func(val Val[int]) {
		t.Helper()

		t.Run("", func(t *testing.T) {
			require.Equal(t, false, val.HasVal())

			val.Set(42)
			require.Equal(t, true, val.HasVal())
			require.Equal(t, 42, val.Val())

			val.Set(-42)
			require.Equal(t, true, val.HasVal())
			require.Equal(t, -42, val.Val())
		})
	}

	fOk(Empty[int]())

	var val Val[int]
	fOk(val)
}

func TestReset(t *testing.T) {
	t.Parallel()

	val := New(42)

	require.Equal(t, true, val.HasVal())
	require.Equal(t, 42, val.Val())

	val.Reset()
	require.Equal(t, false, val.HasVal())
	require.Equal(t, 0, val.Val())
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("with_value", func(t *testing.T) {
		t.Parallel()

		val := New("hello")
		result, ok := val.Get()
		require.True(t, ok)
		require.Equal(t, "hello", result)
	})

	t.Run("without_value", func(t *testing.T) {
		t.Parallel()

		val := Empty[string]()
		result, ok := val.Get()
		require.False(t, ok)
		require.Equal(t, "", result) // zero value
	})

	t.Run("zero_value_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New(0) // zero value but has flag
		result, ok := val.Get()
		require.True(t, ok)
		require.Equal(t, 0, result)
	})

	t.Run("empty_string_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New("") // empty string but has flag
		result, ok := val.Get()
		require.True(t, ok)
		require.Equal(t, "", result)
	})

	t.Run("nil_slice_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New([]string(nil)) // nil slice but has flag
		result, ok := val.Get()
		require.True(t, ok)
		require.Nil(t, result)
	})
}

func TestValDefault(t *testing.T) {
	t.Parallel()

	t.Run("with_value_string", func(t *testing.T) {
		t.Parallel()

		val := New("hello")
		require.Equal(t, "hello", val.ValDefault("default"))
		val.Reset()
		require.Equal(t, "default", val.ValDefault("default"))
	})

	t.Run("zero_value_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New(0)
		require.Equal(t, 0, val.ValDefault(42))
		val.Reset()
		require.Equal(t, 42, val.ValDefault(42))
	})

	t.Run("empty_string_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New("")
		result := val.ValDefault("default")
		require.Equal(t, "", result) // should return actual empty string, not default
	})

	t.Run("nil_slice_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New([]string(nil))
		result := val.ValDefault([]string{"default"})
		require.Nil(t, result) // should return actual nil, not default
	})

	t.Run("complex_types", func(t *testing.T) {
		t.Parallel()

		type CustomStruct struct {
			Name string
			ID   int
		}

		val := Empty[CustomStruct]()
		defaultVal := CustomStruct{Name: "default", ID: 999}
		result := val.ValDefault(defaultVal)
		require.Equal(t, defaultVal, result)
	})
}

func TestAsPointer(t *testing.T) {
	t.Parallel()

	t.Run("with_value", func(t *testing.T) {
		t.Parallel()

		val := New("hello")
		ptr := val.AsPointer()
		require.NotNil(t, ptr)
		require.Equal(t, "hello", *ptr)
	})

	t.Run("without_value", func(t *testing.T) {
		t.Parallel()

		val := Empty[string]()
		ptr := val.AsPointer()
		require.Nil(t, ptr)
	})

	t.Run("zero_value_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New(0)
		ptr := val.AsPointer()
		require.NotNil(t, ptr)
		require.Equal(t, 0, *ptr)
	})

	t.Run("empty_string_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New("")
		ptr := val.AsPointer()
		require.NotNil(t, ptr)
		require.Equal(t, "", *ptr)
	})

	t.Run("nil_slice_with_flag", func(t *testing.T) {
		t.Parallel()

		val := New([]string(nil))
		ptr := val.AsPointer()
		require.NotNil(t, ptr)
		require.Nil(t, *ptr)
	})

	t.Run("pointer_safety", func(t *testing.T) {
		t.Parallel()

		val := New("hello")
		ptr := val.AsPointer()
		*ptr = "modified"

		// Since Val is passed by value, modifying through pointer affects
		// the copy used by AsPointer(), not the original
		result, ok := val.Get()
		require.True(t, ok)
		require.Equal(t, "hello", result) // Original unchanged

		// The pointer itself reflects the modification
		require.Equal(t, "modified", *ptr)
	})
}
