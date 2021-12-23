package reisen

// #cgo LDFLAGS: -lavformat -lavcodec -lavutil -lswscale
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avconfig.h>
// #include <libswscale/swscale.h>
import "C"
import "unsafe"

// Packet is a piece of encoded data
// acquired from the media container.
//
// It can be either a video frame or
// an audio frame.
type Packet struct {
	media       *Media
	streamIndex int
	data        []byte
	pts         int64
	dts         int64
	pos         int64
	duration    int64
	size        int
	flags       int
}

// StreamIndex returns the index of the
// stream the packet belongs to.
func (pkt *Packet) StreamIndex() int {
	return pkt.streamIndex
}

// Type returns the type of the packet
// (video or audio).
func (pkt *Packet) Type() StreamType {
	return pkt.media.Streams()[pkt.streamIndex].Type()
}

// Data returns the data
// encoded in the packet.
func (pkt *Packet) Data() []byte {
	buf := make([]byte, pkt.size)

	copy(buf, pkt.data)

	return buf
}

// Returns the size of the
// packet data.
func (pkt *Packet) Size() int {
	return pkt.size
}

// newPacket creates a
// new packet info object.
func newPacket(media *Media, cPkt *C.AVPacket) *Packet {
	pkt := &Packet{
		media:       media,
		streamIndex: int(cPkt.stream_index),
		data: C.GoBytes(unsafe.Pointer(
			cPkt.data), cPkt.size),
		pts:      int64(cPkt.pts),
		dts:      int64(cPkt.dts),
		pos:      int64(cPkt.pos),
		duration: int64(cPkt.duration),
		size:     int(cPkt.size),
		flags:    int(cPkt.flags),
	}

	return pkt
}
