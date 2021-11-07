package reisen

import "image"

// VideoFrame is a single frame
// of a video stream.
type VideoFrame struct {
	baseFrame
	img *image.RGBA
}

// Data returns a byte slice of RGBA
// pixels of the frame image.
func (frame *VideoFrame) Data() []byte {
	return frame.img.Pix
}

// Image returns the RGBA image of the frame.
func (frame *VideoFrame) Image() *image.RGBA {
	return frame.img
}

// newVideoFrame returns a newly created video frame.
func newVideoFrame(stream Stream, pts int64, indCoded, indDisplay, width, height int, pix []byte) *VideoFrame {
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	frame := new(VideoFrame)

	img.Pix = pix
	frame.stream = stream
	frame.pts = pts
	frame.img = img
	frame.indexCoded = indCoded
	frame.indexDisplay = indDisplay

	return frame
}
