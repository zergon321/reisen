package reisen

import "image"

type VideoFrame struct {
	baseFrame
	img *image.RGBA
}

func (frame *VideoFrame) Data() []byte {
	return frame.img.Pix
}

func (frame *VideoFrame) Image() *image.RGBA {
	return frame.img
}

func newVideoFrame(stream Stream, pts int64, width, height int, pix []byte) *VideoFrame {
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	frame := new(VideoFrame)

	img.Pix = pix
	frame.stream = stream
	frame.pts = pts
	frame.img = img

	return frame
}
