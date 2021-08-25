package reisen

// #cgo LDFLAGS: -lavutil -lavformat -lavcodec
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avconfig.h>
import "C"
import (
	"fmt"
	"time"
	"unsafe"
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
	Open() error
	ReadFrame() (Frame, bool, error)
	Close() error
}

type baseStream struct {
	media       *Media
	inner       *C.AVStream
	codecParams *C.AVCodecParameters
	codec       *C.AVCodec
	codecCtx    *C.AVCodecContext
	packet      *C.AVPacket
	frame       *C.AVFrame
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

func (stream *baseStream) open() error {
	stream.codecCtx = C.avcodec_alloc_context3(stream.codec)

	if stream.codecCtx == nil {
		return fmt.Errorf("couldn't open a codec context")
	}

	status := C.avcodec_parameters_to_context(
		stream.codecCtx, stream.codecParams)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't send codec parameters to the context", status)
	}

	stream.packet = C.av_packet_alloc()

	if stream.packet == nil {
		return fmt.Errorf(
			"couldn't allocate a new packet")
	}

	stream.frame = C.av_frame_alloc()

	if stream.frame == nil {
		return fmt.Errorf(
			"couldn't allocate a new frame")
	}

	return nil
}

func (stream *baseStream) read() (bool, error) {
	status := C.av_read_frame(stream.media.ctx, stream.packet)

	if status < 0 {
		if stream.packet.data == nil {
			return false, fmt.Errorf(
				"%d: couldn't extract the frame", status)
		}

		// No packets anymore.
		return false, nil
	}

	// If the packet doesn't belong tj the stream.
	if stream.packet.stream_index != stream.inner.index {
		return true, nil
	}

	status = C.avcodec_send_packet(
		stream.codecCtx, stream.packet)

	if status < 0 {
		return false, fmt.Errorf(
			"%d: couldn't send the packet to the codec context", status)
	}

	status = C.avcodec_receive_frame(
		stream.codecCtx, stream.frame)

	if status < 0 {
		return false, fmt.Errorf(
			"%d: couldn't receive the frame from the codec context", status)
	}

	return true, nil
}

func (stream *baseStream) close() error {
	C.av_free(unsafe.Pointer(stream.frame))
	C.av_free(unsafe.Pointer(stream.packet))

	status := C.avcodec_close(stream.codecCtx)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't close the codec", status)
	}

	return nil
}
