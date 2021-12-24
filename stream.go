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

// StreamType is a type of
// a media stream.
type StreamType int

const (
	// StreamVideo denotes the stream keeping video frames.
	StreamVideo StreamType = C.AVMEDIA_TYPE_VIDEO
	// StreamAudio denotes the stream keeping audio frames.
	StreamAudio StreamType = C.AVMEDIA_TYPE_AUDIO
)

// String returns the string representation of
// stream type identifier.
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

// Stream is an abstract media data stream.
type Stream interface {
	// Index returns the index
	// number of the stream.
	Index() int
	// Type returns the type
	// identifier of the stream.
	//
	// It's either video or audio.
	Type() StreamType
	// CodecName returns the
	// shortened name of the stream codec.
	CodecName() string
	// CodecLongName returns the
	// long name of the stream codec.
	CodecLongName() string
	// BitRate returns the stream
	// bitrate (in bps).
	BitRate() int64
	// Duration returns the time
	// duration of the stream
	Duration() (time.Duration, error)
	// TimeBase returns the numerator
	// and the denominator of the stream
	// time base fraction to convert
	// time duration in time base units
	// of the stream.
	TimeBase() (int, int)
	// FrameRate returns the approximate
	// frame rate (FPS) of the stream.
	FrameRate() (int, int)
	// FrameCount returns the total number
	// of frames in the stream.
	FrameCount() int64
	// Open opens the stream for decoding.
	Open() error
	// Rewind rewinds the whole media to the
	// specified time location based on the stream.
	Rewind(time.Duration) error
	// ReadFrame decodes the next frame from the stream.
	ReadFrame() (Frame, bool, error)
	// Closes the stream for decoding.
	Close() error
}

// baseStream holds the information
// common for all media data streams.
type baseStream struct {
	media       *Media
	inner       *C.AVStream
	codecParams *C.AVCodecParameters
	codec       *C.AVCodec
	codecCtx    *C.AVCodecContext
	frame       *C.AVFrame
	skip        bool
	opened      bool
}

// Opened returns 'true' if the stream
// is opened for decoding, and 'false' otherwise.
func (stream *baseStream) Opened() bool {
	return stream.opened
}

// Index returns the index of the stream.
func (stream *baseStream) Index() int {
	return int(stream.inner.index)
}

// Type returns the stream media data type.
func (stream *baseStream) Type() StreamType {
	return StreamType(stream.codecParams.codec_type)
}

// CodecName returns the name of the codec
// that was used for encoding the stream.
func (stream *baseStream) CodecName() string {
	if stream.codec.name == nil {
		return ""
	}

	return C.GoString(stream.codec.name)
}

// CodecName returns the long name of the
// codec that was used for encoding the stream.
func (stream *baseStream) CodecLongName() string {
	if stream.codec.long_name == nil {
		return ""
	}

	return C.GoString(stream.codec.long_name)
}

// BitRate returns the bit rate of the stream (in bps).
func (stream *baseStream) BitRate() int64 {
	return int64(stream.codecParams.bit_rate)
}

// Duration returns the duration of the stream.
func (stream *baseStream) Duration() (time.Duration, error) {
	dur := stream.inner.duration

	if dur < 0 {
		dur = 0
	}

	tmNum, tmDen := stream.TimeBase()
	factor := float64(tmNum) / float64(tmDen)
	tm := float64(dur) * factor

	return time.ParseDuration(fmt.Sprintf("%fs", tm))
}

// TimeBase the numerator and the denominator of the
// stream time base factor fraction.
//
// All the duration values of the stream are
// multiplied by this factor to get duration
// in seconds.
func (stream *baseStream) TimeBase() (int, int) {
	return int(stream.inner.time_base.num),
		int(stream.inner.time_base.den)
}

// FrameRate returns the frame rate of the stream
// as a fraction with a numerator and a denominator.
func (stream *baseStream) FrameRate() (int, int) {
	return int(stream.inner.r_frame_rate.num),
		int(stream.inner.r_frame_rate.den)
}

// FrameCount returns the total number of frames
// in the stream.
func (stream *baseStream) FrameCount() int64 {
	return int64(stream.inner.nb_frames)
}

// Rewind rewinds the stream to
// the specified time position.
//
// Can be used on all the types
// of streams. However, it's better
// to use it on the video stream of
// the media file if you don't want
// the streams of the playback to
// desynchronyze.
func (stream *baseStream) Rewind(t time.Duration) error {
	tmNum, tmDen := stream.TimeBase()
	factor := float64(tmDen) / float64(tmNum)
	seconds := t.Seconds()
	dur := int64(seconds * factor)

	status := C.av_seek_frame(stream.media.ctx,
		stream.inner.index, C.long(dur),
		C.AVSEEK_FLAG_FRAME|C.AVSEEK_FLAG_BACKWARD)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't rewind the stream", status)
	}

	return nil
}

// open opens the stream for decoding.
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

	status = C.avcodec_open2(stream.codecCtx, stream.codec, nil)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't open the codec context", status)
	}

	stream.frame = C.av_frame_alloc()

	if stream.frame == nil {
		return fmt.Errorf(
			"couldn't allocate a new frame")
	}

	stream.opened = true

	return nil
}

// read decodes the packet and obtains a
// frame from it.
func (stream *baseStream) read() (bool, error) {
	status := C.avcodec_send_packet(
		stream.codecCtx, stream.media.packet)

	if status < 0 {
		stream.skip = false

		return false, fmt.Errorf(
			"%d: couldn't send the packet to the codec context", status)
	}

	status = C.avcodec_receive_frame(
		stream.codecCtx, stream.frame)

	if status < 0 {
		if status == C.int(ErrorAgain) {
			stream.skip = true
			return true, nil
		}

		stream.skip = false

		return false, fmt.Errorf(
			"%d: couldn't receive the frame from the codec context", status)
	}

	C.av_packet_unref(stream.media.packet)

	stream.skip = false

	return true, nil
}

// close closes the stream for decoding.
func (stream *baseStream) close() error {
	C.av_free(unsafe.Pointer(stream.frame))

	status := C.avcodec_close(stream.codecCtx)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't close the codec", status)
	}

	stream.opened = false

	return nil
}
