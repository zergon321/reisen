package reisen

import (
	"fmt"
	"time"
)

type Frame interface {
	Data() []byte
	PresentationOffset() (time.Duration, error)
}

type baseFrame struct {
	stream Stream
	pts    int64
}

func (frame *baseFrame) PresentationOffset() (time.Duration, error) {
	tbNum, tbDen := frame.stream.TimeBase()
	tb := float64(tbNum) / float64(tbDen)
	tm := float64(frame.pts) * tb

	return time.ParseDuration(fmt.Sprintf("%fs", tm))
}
