package reisen

// #cgo LDFLAGS: -lavutil -lavformat -lavcodec
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avutil.h>
import "C"

type AudioStream struct {
	baseStream
}

func (audio *AudioStream) ChannelCount() int {
	return int(audio.codecParams.channels)
}

func (audio *AudioStream) SampleRate() int {
	return int(audio.codecParams.sample_rate)
}

func (audio *AudioStream) FrameSize() int {
	return int(audio.codecParams.frame_size)
}
