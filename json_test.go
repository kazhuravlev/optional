package optional

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
