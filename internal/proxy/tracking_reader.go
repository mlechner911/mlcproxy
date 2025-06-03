/*
Copyright (c) 2025 Michael Lechner

This software is released under the MIT License.
See the LICENSE file for further details.
*/

package proxy

import (
	"io"
)

// TrackingReader wraps an io.Reader to track the number of bytes read.
// It is used to monitor the amount of data transferred during proxy operations.
type TrackingReader struct {
	r         io.Reader
	bytesRead uint64
}

// NewTrackingReader creates a new TrackingReader that wraps the given io.Reader.
func NewTrackingReader(r io.Reader) *TrackingReader {
	return &TrackingReader{r: r}
}

// Read implements the io.Reader interface and tracks the number of bytes read.
// It keeps a running total of all bytes that have passed through the reader.
func (t *TrackingReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	t.bytesRead += uint64(n)
	return
}

// BytesRead returns the total number of bytes that have been read through this reader.
func (t *TrackingReader) BytesRead() uint64 {
	return t.bytesRead
}
