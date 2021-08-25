package reisen

type AudioFrame struct {
	baseFrame
	data []byte
}

func (frame *AudioFrame) Data() []byte {
	return frame.data
}

func newAudioFrame(stream Stream, pts int64, data []byte) *AudioFrame {
	frame := new(AudioFrame)

	frame.stream = stream
	frame.pts = pts
	frame.data = data

	return frame
}
