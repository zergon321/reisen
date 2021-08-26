package reisen

// #cgo LDFLAGS: -lavutil -lavformat -lavcodec -lswresample
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
	StandardChannelCount = 2
)

type AudioStream struct {
	baseStream
	swrCtx *C.SwrContext
	buffer *C.uint8_t
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

func (audio *AudioStream) Open() error {
	err := audio.open()

	if err != nil {
		return err
	}

	audio.swrCtx = C.swr_alloc_set_opts(nil,
		C.AV_CH_FRONT_LEFT|C.AV_CH_FRONT_RIGHT,
		C.AV_SAMPLE_FMT_DBL, audio.codecCtx.sample_rate,
		C.long(audio.codecCtx.channel_layout), audio.
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

func (audio *AudioStream) ReadFrame() (Frame, bool, error) {
	return audio.ReadAudioFrame()
}

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

	if audio.buffer == nil {
		var byteSize C.ulong = 8
		audio.buffer = (*C.uint8_t)(unsafe.Pointer(
			C.av_malloc(C.ulong(maxBufferSize) * byteSize)))

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
		int64(audio.frame.pts), data)

	return frame, true, nil
}

func (audio *AudioStream) Close() error {
	err := audio.close()

	if err != nil {
		return err
	}

	C.av_free(unsafe.Pointer(audio.buffer))
	C.swr_free(&audio.swrCtx)

	return nil
}
