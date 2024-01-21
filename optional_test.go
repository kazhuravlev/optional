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
