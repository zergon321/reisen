package reisen

// #cgo LDFLAGS: -lavutil -lavformat -lavcodec -lswscale
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avutil.h>
// #include <libavutil/imgutils.h>
// #include <libswscale/swscale.h>
// #include <inttypes.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// VideoStream is a streaming holding
// video frames.
type VideoStream struct {
	baseStream
	swsCtx    *C.struct_SwsContext
	rgbaFrame *C.AVFrame
	bufSize   C.int
}

// AspectRatio returns the fraction of the video
// stream frame aspect ratio (1/0 if unknown).
func (video *VideoStream) AspectRatio() (int, int) {
	return int(video.codecParams.sample_aspect_ratio.num),
		int(video.codecParams.sample_aspect_ratio.den)
}

// Width returns the width of the video
// stream frame.
func (video *VideoStream) Width() int {
	return int(video.codecParams.width)
}

// Height returns the height of the video
// stream frame.
func (video *VideoStream) Height() int {
	return int(video.codecParams.height)
}

// OpenDecode opens the video stream for
// decoding with default parameters.
func (video *VideoStream) Open() error {
	return video.OpenDecode(
		int(video.codecParams.width),
		int(video.codecParams.height),
		InterpolationBicubic)
}

// OpenDecode opens the video stream for
// decoding with the specified parameters.
func (video *VideoStream) OpenDecode(width, height int, alg InterpolationAlgorithm) error {
	err := video.open()

	if err != nil {
		return err
	}

	video.rgbaFrame = C.av_frame_alloc()

	if video.rgbaFrame == nil {
		return fmt.Errorf(
			"couldn't allocate a new RGBA frame")
	}

	video.bufSize = C.av_image_get_buffer_size(
		C.AV_PIX_FMT_RGBA, C.int(width), C.int(height), 1)

	if video.bufSize < 0 {
		return fmt.Errorf(
			"%d: couldn't get the buffer size", video.bufSize)
	}

	buf := (*C.uint8_t)(unsafe.Pointer(
		C.av_malloc(bufferSize(video.bufSize))))

	if buf == nil {
		return fmt.Errorf(
			"couldn't allocate an AV buffer")
	}

	status := C.av_image_fill_arrays(&video.rgbaFrame.data[0],
		&video.rgbaFrame.linesize[0], buf, C.AV_PIX_FMT_RGBA,
		C.int(width), C.int(height), 1)

	if status < 0 {
		return fmt.Errorf(
			"%d: couldn't fill the image arrays", status)
	}

	video.swsCtx = C.sws_getContext(video.codecCtx.width,
		video.codecCtx.height, video.codecCtx.pix_fmt,
		C.int(width), C.int(height),
		C.AV_PIX_FMT_RGBA, C.int(alg), nil, nil, nil)

	if video.swsCtx == nil {
		return fmt.Errorf(
			"couldn't create an SWS context")
	}

	return nil
}

// ReadFrame reads the next frame from the stream.
func (video *VideoStream) ReadFrame() (Frame, bool, error) {
	return video.ReadVideoFrame()
}

// ReadVideoFrame reads the next video frame
// from the video stream.
func (video *VideoStream) ReadVideoFrame() (*VideoFrame, bool, error) {
	ok, err := video.read()

	if err != nil {
		return nil, false, err
	}

	if ok && video.skip {
		return nil, true, nil
	}

	// No more data.
	if !ok {
		return nil, false, nil
	}

	C.sws_scale(video.swsCtx, &video.frame.data[0],
		&video.frame.linesize[0], 0,
		video.codecCtx.height,
		&video.rgbaFrame.data[0],
		&video.rgbaFrame.linesize[0])

	data := C.GoBytes(unsafe.
		Pointer(video.rgbaFrame.data[0]),
		video.bufSize)
	frame := newVideoFrame(video, int64(video.frame.pts),
		int(video.frame.coded_picture_number),
		int(video.frame.display_picture_number),
		int(video.codecCtx.width), int(video.codecCtx.height), data)

	return frame, true, nil
}

// Close closes the video stream for decoding.
func (video *VideoStream) Close() error {
	err := video.close()

	if err != nil {
		return err
	}

	C.av_free(unsafe.Pointer(video.rgbaFrame))
	video.rgbaFrame = nil
	C.sws_freeContext(video.swsCtx)
	video.swsCtx = nil

	return nil
}
