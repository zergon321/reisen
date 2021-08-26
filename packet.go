package reisen

// #cgo LDFLAGS: -lavformat -lavcodec -lavutil -lswscale
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avconfig.h>
// #include <libswscale/swscale.h>
import "C"

type Packet struct {
	media *Media
}

func (pkt *Packet) StreamIndex() int {
	return int(pkt.media.packet.stream_index)
}

func (pkt *Packet) Type() StreamType {
	return pkt.media.Streams()[pkt.StreamIndex()].Type()
}
