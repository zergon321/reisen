package reisen

import (
	"fmt"
	"time"
)

// Frame is an abstract data frame.
type Frame interface {
	Data() []byte
	PresentationOffset() (time.Duration, error)
}

// baseFrame contains the information
// common for all frames of any type.
type baseFrame struct {
	stream               Stream
	pts                  int64
	codedPictureNumber   int
	displayPictureNumber int
}

// PresentationOffset returns the duration offset
// since the start of the media at which the frame
// should be played.
func (frame *baseFrame) PresentationOffset() (time.Duration, error) {
	tbNum, tbDen := frame.stream.TimeBase()
	tb := float64(tbNum) / float64(tbDen)
	tm := float64(frame.pts) * tb

	return time.ParseDuration(fmt.Sprintf("%fs", tm))
}

func (frame *baseFrame) IndexCoded() int {
	return frame.codedPictureNumber
}

func (frame *baseFrame) IndexDisplay() int {
	return frame.displayPictureNumber
}
