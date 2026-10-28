//go:build go1.24

package optional

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSONOmitZero(t *testing.T) {
	t.Parallel()

	type payload struct {
		A Val[int]      `json:"a,omitzero"`
		B Val[string]   `json:"b,omitzero"`
		C Val[[]string] `json:"c,omitzero"`
	}

	t.Run("empty_containers_are_omitted", func(t *testing.T) {
		t.Parallel()

		res, err := json.Marshal(payload{})
		require.NoError(t, err)
		assert.Equal(t, `{}`, string(res))

		res, err = json.Marshal(payload{A: Empty[int](), B: Empty[string](), C: Empty[[]string]()})
		require.NoError(t, err)
		assert.Equal(t, `{}`, string(res))
	})

	t.Run("present_zero_values_are_kept", func(t *testing.T) {
		t.Parallel()

		res, err := json.Marshal(payload{A: New(0), B: New(""), C: New([]string(nil))})
		require.NoError(t, err)
		assert.JSONEq(t, `{"a":0,"b":"","c":null}`, string(res))
	})

	t.Run("after_null_unmarshal_over_existing_value", func(t *testing.T) {
		t.Parallel()

		val := payload{A: New(8080), B: New("x"), C: New([]string{"a"})}
		require.NoError(t, json.Unmarshal([]byte(`{"a":null,"b":null,"c":null}`), &val))

		res, err := json.Marshal(val)
		require.NoError(t, err)
		assert.Equal(t, `{}`, string(res))
	})
}
