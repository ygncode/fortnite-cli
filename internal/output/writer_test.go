package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Format: FormatJSON}
	require.NoError(t, w.Write(map[string]any{"a": 1, "b": "hi"}))
	require.JSONEq(t, `{"a":1,"b":"hi"}`, buf.String())
}

func TestWriteNDJSONIterates(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Format: FormatNDJSON}
	require.NoError(t, w.Write(map[string]any{"n": 1}))
	require.NoError(t, w.Write(map[string]any{"n": 2}))
	require.Equal(t, "{\"n\":1}\n{\"n\":2}\n", buf.String())
}

func TestWriteRawBytes(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Format: FormatJSON}
	require.NoError(t, w.WriteRaw([]byte(`{"already":"json"}`)))
	require.Equal(t, `{"already":"json"}`+"\n", buf.String())
}
