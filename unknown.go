package reisen

import (
	"fmt"
)

// UnknownStream is a stream containing frames consisting of unknown data.
type UnknownStream struct {
	baseStream
}

// Open is just a stub.
func (unknown *UnknownStream) Open() error {
	return nil
}

// ReadFrame is just a stub.
func (unknown *UnknownStream) ReadFrame() (Frame, bool, error) {
	return nil, false, fmt.Errorf("UnknownStream.ReadFrame() not implemented")
}

// Close is just a stub.
func (unknown *UnknownStream) Close() error {
	return nil
}
