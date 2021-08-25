package reisen

// #cgo LDFLAGS: -lavutil -lavformat -lavcodec
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avutil.h>
import "C"

type VideoStream struct {
	baseStream
}

func (video *VideoStream) AspectRatio() (int, int) {
	return int(video.codecParams.sample_aspect_ratio.num),
		int(video.codecParams.sample_aspect_ratio.den)
}

func (video *VideoStream) Width() int {
	return int(video.codecParams.width)
}

func (video *VideoStream) Height() int {
	return int(video.codecParams.height)
}
