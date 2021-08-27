package reisen

// #cgo LDFLAGS: -lavformat -lavcodec -lavutil -lswscale
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avconfig.h>
// #include <libswscale/swscale.h>
import "C"

// Packet is a piece of data
// acquired from the media container.
//
// It can be either a video frame or
// an audio frame.
type Packet struct {
	media *Media
}

// StreamIndex returns the index of the
// stream the packet belongs to.
func (pkt *Packet) StreamIndex() int {
	return int(pkt.media.packet.stream_index)
}

// Type returns the type of the packet
// (video or audio).
func (pkt *Packet) Type() StreamType {
	return pkt.media.Streams()[pkt.StreamIndex()].Type()
}
