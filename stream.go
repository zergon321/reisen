package reisen

// #cgo LDFLAGS: -lavutil -lavformat -lavcodec -lswscale
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avconfig.h>
// #include <libswscale/swscale.h>
import "C"
import (
	"fmt"
	"time"
)

type StreamType int

const (
	StreamVideo StreamType = C.AVMEDIA_TYPE_VIDEO
	StreamAudio StreamType = C.AVMEDIA_TYPE_AUDIO
)

func (streamType StreamType) String() string {
	switch streamType {
	case StreamVideo:
		return "video"

	case StreamAudio:
		return "audio"

	default:
		return ""
	}
}

// TODO: add an opportunity to
// receive duration in time base units.

type Stream interface {
	Index() int
	Type() StreamType
	CodecName() string
	CodecLongName() string
	BitRate() int64
	Duration() (time.Duration, error)
	TimeBase() (int, int)
	FrameRate() (int, int)
	FrameCount() int64
}

// TODO: opening and reading from the stream.

type baseStream struct {
	inner       *C.AVStream
	codecParams *C.AVCodecParameters
	codec       *C.AVCodec
	codecCtx    *C.AVCodecContext
	packet      *C.AVPacket
	frame       *C.AVFrame
	rgbaFrame   *C.AVFrame
	swsCtx      *C.struct_SwsContext
}

func (stream *baseStream) Index() int {
	return int(stream.inner.index)
}

func (stream *baseStream) Type() StreamType {
	return StreamType(stream.codecParams.codec_type)
}

func (stream *baseStream) CodecName() string {
	if stream.codec.name == nil {
		return ""
	}

	return C.GoString(stream.codec.name)
}

func (stream *baseStream) CodecLongName() string {
	if stream.codec.long_name == nil {
		return ""
	}

	return C.GoString(stream.codec.long_name)
}

func (stream *baseStream) BitRate() int64 {
	return int64(stream.codecParams.bit_rate)
}

func (stream *baseStream) Duration() (time.Duration, error) {
	dur := stream.inner.duration
	tmNum, tmDen := stream.TimeBase()
	factor := float64(tmNum) / float64(tmDen)
	tm := float64(dur) * factor

	return time.ParseDuration(fmt.Sprintf("%fs", tm))
}

func (stream *baseStream) TimeBase() (int, int) {
	return int(stream.inner.time_base.num),
		int(stream.inner.time_base.den)
}

func (stream *baseStream) FrameRate() (int, int) {
	return int(stream.inner.r_frame_rate.num),
		int(stream.inner.r_frame_rate.den)
}

func (stream *baseStream) FrameCount() int64 {
	return int64(stream.inner.nb_frames)
}
