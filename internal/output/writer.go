// Package output writes JSON and NDJSON to an io.Writer with a --raw bypass.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Format controls how values are written.
type Format int

const (
	FormatJSON Format = iota
	FormatNDJSON
)

// Writer writes values to Out in the chosen Format.
type Writer struct {
	Out    io.Writer
	Format Format
}

// Write encodes v and writes it.
//   - FormatJSON: single line JSON, no trailing newline between calls — callers should
//     only call Write once unless they know what they're doing.
//   - FormatNDJSON: one JSON object per line, each call appends a newline.
func (w *Writer) Write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if _, err := w.Out.Write(data); err != nil {
		return err
	}
	if w.Format == FormatNDJSON {
		_, err = fmt.Fprintln(w.Out)
	}
	return err
}

// WriteRaw writes pre-encoded JSON bytes followed by a newline.
// Used by the --raw flag to pass the API response through untouched (modulo
// a trailing newline).
func (w *Writer) WriteRaw(data []byte) error {
	if _, err := w.Out.Write(data); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w.Out)
	return err
}
