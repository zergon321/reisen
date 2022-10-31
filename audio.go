package reisen

// #cgo pkg-config: libavformat libavcodec libavutil libswresample
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avutil.h>
// #include <libswresample/swresample.h>
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	// StandardChannelCount is used for
	// audio conversion while decoding
	// audio frames.
	StandardChannelCount = 2
)

// AudioStream is a stream containing
// audio frames consisting of audio samples.
type AudioStream struct {
	baseStream
	swrCtx     *C.SwrContext
	buffer     *C.uint8_t
	bufferSize C.int
}

// ChannelCount returns the number of channels
// (1 for mono, 2 for stereo, etc.).
func (audio *AudioStream) ChannelCount() int {
	return int(audio.codecParams.channels)
}

// SampleRate returns the sample rate of the
// audio stream.
func (audio *AudioStream) SampleRate() int {
	return int(audio.codecParams.sample_rate)
}

// FrameSize returns the number of samples
// contained in one frame of the audio.
func (audio *AudioStream) FrameSize() int {
	return int(audio.codecParams.frame_size)
}

// Open opens the audio stream to decode
// audio frames and samples from it.
func (audio *AudioStream) Open() error {
	err := audio.open()

	if err != nil {
		return err
	}

	audio.swrCtx = C.swr_alloc_set_opts(nil,
		C.AV_CH_FRONT_LEFT|C.AV_CH_FRONT_RIGHT,
		C.AV_SAMPLE_FMT_DBL, audio.codecCtx.sample_rate,
		channelLayout(audio), audio.
			codecCtx.sample_fmt, audio.codecCtx.
			sample_rate, 0, nil)

	if audio.swrCtx == nil {
		return fmt.Errorf(
			"couldn't allocate an SWR context")
	}

	status := C.swr_init(audio.swrCtx)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't initialize the SWR context", status)
	}

	audio.buffer = nil

	return nil
}

// ReadFrame reads a new frame from the stream.
func (audio *AudioStream) ReadFrame() (Frame, bool, error) {
	return audio.ReadAudioFrame()
}

// ReadAudioFrame reads a new audio frame from the stream.
func (audio *AudioStream) ReadAudioFrame() (*AudioFrame, bool, error) {
	ok, err := audio.read()

	if err != nil {
		return nil, false, err
	}

	if ok && audio.skip {
		return nil, true, nil
	}

	// No more data.
	if !ok {
		return nil, false, nil
	}

	maxBufferSize := C.av_samples_get_buffer_size(
		nil, StandardChannelCount,
		audio.frame.nb_samples,
		C.AV_SAMPLE_FMT_DBL, 1)

	if maxBufferSize < 0 {
		return nil, false, fmt.Errorf(
			"%d: couldn't get the max buffer size", maxBufferSize)
	}

	if maxBufferSize > audio.bufferSize {
		C.av_free(unsafe.Pointer(audio.buffer))
		audio.buffer = nil
	}

	if audio.buffer == nil {
		audio.buffer = (*C.uint8_t)(unsafe.Pointer(
			C.av_malloc(bufferSize(maxBufferSize))))
		audio.bufferSize = maxBufferSize

		if audio.buffer == nil {
			return nil, false, fmt.Errorf(
				"couldn't allocate an AV buffer")
		}
	}

	gotSamples := C.swr_convert(audio.swrCtx,
		&audio.buffer, audio.frame.nb_samples,
		&audio.frame.data[0], audio.frame.nb_samples)

	if gotSamples < 0 {
		return nil, false, fmt.Errorf(
			"%d: couldn't convert the audio frame", gotSamples)
	}

	data := C.GoBytes(unsafe.Pointer(
		audio.buffer), maxBufferSize)
	frame := newAudioFrame(audio,
		int64(audio.frame.pts),
		int(audio.frame.coded_picture_number),
		int(audio.frame.display_picture_number), data)

	return frame, true, nil
}

// Close closes the audio stream and
// stops decoding audio frames.
func (audio *AudioStream) Close() error {
	err := audio.close()

	if err != nil {
		return err
	}

	C.av_free(unsafe.Pointer(audio.buffer))
	audio.buffer = nil
	C.swr_free(&audio.swrCtx)
	audio.swrCtx = nil

	return nil
}
