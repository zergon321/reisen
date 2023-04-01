package reisen

/*
#cgo pkg-config: libavformat libavcodec libavutil libswscale
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavformat/avio.h>
#include <libavutil/avconfig.h>
#include <libswscale/swscale.h>
#include <libavcodec/bsf.h>
#include <string.h>
#include <stdio.h>

static int setFormat(AVFormatContext *ctx, void *format) {
	ctx->iformat = (AVInputFormat*)format;
	return 0;
}

extern int readCallBack(void*, uint8_t*, int);
extern int64_t seekCallBack(void*, int64_t, int);
extern int writeCallBack(void*, uint8_t*, int);

*/
import "C"
import (
	"bytes"
	"fmt"
	"io"
	"time"
	"unsafe"
)

const (
	probeIOBufferSize    = 50000000
	avio_flag_read       = 1
	avio_flag_write      = 2
	avio_flag_read_write = (avio_flag_read | avio_flag_write)
)

var (
	handlersMap map[uintptr]*AVIOHandlers
)

type StreamMediaOptions struct {
	Alias      string
	ReadSeeker io.ReadSeeker
	IOHandlers *AVIOHandlers
}

type AVIOHandlers struct {
	ReadSeeker io.ReadSeeker
	bufferSize int

	ReadPacket  func(int) ([]byte, int)
	WritePacket func([]byte) int
	Seek        func(int64, int) int64
	ReadFormat  func(*Media) (*unsafe.Pointer, error)
}

type defaultIOHandler struct {
	AVIOHandlers
}

func (options *defaultIOHandler) ReadPacket(bufferSize int) ([]byte, int) {
	contents := make([]byte, bufferSize)
	len, err := options.ReadSeeker.Read(contents)
	if err != nil {
		empty := make([]byte, 0)
		return empty, C.AVERROR_EOF
	}
	return contents, len
}

func (options *defaultIOHandler) Seek(offset int64, whence int) int64 {
	if whence == 65536 { // avseek
		return -1
	}
	rs, err := options.ReadSeeker.Seek(offset, whence)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return rs
}

func (options *defaultIOHandler) ReadFormat(media *Media) (*unsafe.Pointer, error) {
	data := make([]byte, probeIOBufferSize)
	options.ReadSeeker.Read(data)
	options.ReadSeeker.Seek(0, io.SeekStart)

	res := C.CBytes(data)
	emptyString := C.CString("")

	probeData := C.AVProbeData{}
	probeData.buf = (*C.uint8_t)(res)
	probeData.mime_type = emptyString
	probeData.buf_size = C.int(probeIOBufferSize)
	probeData.filename = emptyString

	format := C.av_probe_input_format(&probeData, 1)
	result := unsafe.Pointer(format)

	C.free(res)
	defer C.free(unsafe.Pointer(emptyString))

	options.ReadSeeker.Seek(0, io.SeekStart)

	return &result, nil
}

type AVIOContext struct {
	avAVIOContext *C.struct_AVIOContext
	handlerKey    uintptr
}

// Media is a media file containing
// audio, video and other types of streams.
type Media struct {
	ctx     *C.AVFormatContext
	packet  *C.AVPacket
	mediaIO *AVIOContext
	name    string
	streams []Stream
}

// StreamCount returns the number of streams.
func (media *Media) StreamCount() int {
	return int(media.ctx.nb_streams)
}

// Streams returns a slice of all the available
// media data streams.
func (media *Media) Streams() []Stream {
	streams := make([]Stream, len(media.streams))
	copy(streams, media.streams)

	return streams
}

// VideoStreams returns all the
// video streams of the media file.
func (media *Media) VideoStreams() []*VideoStream {
	videoStreams := []*VideoStream{}

	for _, stream := range media.streams {
		if videoStream, ok := stream.(*VideoStream); ok {
			videoStreams = append(videoStreams, videoStream)
		}
	}

	return videoStreams
}

// AudioStreams returns all the
// audio streams of the media file.
func (media *Media) AudioStreams() []*AudioStream {
	audioStreams := []*AudioStream{}

	for _, stream := range media.streams {
		if audioStream, ok := stream.(*AudioStream); ok {
			audioStreams = append(audioStreams, audioStream)
		}
	}

	return audioStreams
}

// Duration returns the overall duration
// of the media file.
func (media *Media) Duration() (time.Duration, error) {
	dur := media.ctx.duration
	tm := float64(dur) / float64(TimeBase)

	return time.ParseDuration(fmt.Sprintf("%fs", tm))
}

// FormatName returns the name of the media format.
func (media *Media) FormatName() string {
	if media.ctx.iformat.name == nil {
		return ""
	}

	return C.GoString(media.ctx.iformat.name)
}

// FormatLongName returns the long name
// of the media container.
func (media *Media) FormatLongName() string {
	if media.ctx.iformat.long_name == nil {
		return ""
	}

	return C.GoString(media.ctx.iformat.long_name)
}

// FormatMIMEType returns the MIME type name
// of the media container.
func (media *Media) FormatMIMEType() string {
	if media.ctx.iformat.mime_type == nil {
		return ""
	}

	return C.GoString(media.ctx.iformat.mime_type)
}

// findStreams retrieves the stream information
// from the media container.
func (media *Media) findStreams() error {
	streams := []Stream{}
	status := C.avformat_find_stream_info(media.ctx, nil)

	if status < 0 {
		return fmt.Errorf(
			"couldn't find stream information")
	}

	innerStreams := unsafe.Slice(
		media.ctx.streams, media.ctx.nb_streams)

	for _, innerStream := range innerStreams {
		codecParams := innerStream.codecpar
		codec := C.avcodec_find_decoder(codecParams.codec_id)

		if codec == nil {
			return fmt.Errorf(
				"couldn't find codec by ID = %d",
				codecParams.codec_id)
		}

		switch codecParams.codec_type {
		case C.AVMEDIA_TYPE_VIDEO:
			videoStream := new(VideoStream)
			videoStream.inner = innerStream
			videoStream.codecParams = codecParams
			videoStream.codec = codec
			videoStream.media = media

			streams = append(streams, videoStream)

		case C.AVMEDIA_TYPE_AUDIO:
			audioStream := new(AudioStream)
			audioStream.inner = innerStream
			audioStream.codecParams = codecParams
			audioStream.codec = codec
			audioStream.media = media

			streams = append(streams, audioStream)

		default:
			return fmt.Errorf("unknown stream type")
		}
	}

	media.streams = streams

	return nil
}

// OpenDecode opens the media container for decoding.
//
// CloseDecode() should be called afterwards.
func (media *Media) OpenDecode() error {
	media.packet = C.av_packet_alloc()

	if media.packet == nil {
		return fmt.Errorf(
			"couldn't allocate a new packet")
	}

	return nil
}

// ReadPacket reads the next packet from the media stream.
func (media *Media) ReadPacket() (*Packet, bool, error) {
	status := C.av_read_frame(media.ctx, media.packet)

	if status < 0 {
		if status == C.int(ErrorAgain) {
			return nil, true, nil
		}

		// No packets anymore.
		return nil, false, nil
	}

	// Filter the packet if needed.
	packetStream := media.streams[media.packet.stream_index]
	outPacket := media.packet

	if packetStream.filter() != nil {
		filter := packetStream.filter()
		packetIn := packetStream.filterIn()
		packetOut := packetStream.filterOut()

		status = C.av_packet_ref(packetIn, media.packet)

		if status < 0 {
			return nil, false,
				fmt.Errorf("%d: couldn't reference the packet",
					status)
		}

		status = C.av_bsf_send_packet(filter, packetIn)

		if status < 0 {
			return nil, false,
				fmt.Errorf("%d: couldn't send the packet to the filter",
					status)
		}

		status = C.av_bsf_receive_packet(filter, packetOut)

		if status < 0 {
			return nil, false,
				fmt.Errorf("%d: couldn't receive the packet from the filter",
					status)
		}

		outPacket = packetOut
	}

	return newPacket(media, outPacket), true, nil
}

// CloseDecode closes the media container for decoding.
func (media *Media) CloseDecode() error {
	C.av_free(unsafe.Pointer(media.packet))
	media.packet = nil

	return nil
}

// Close closes the media container.
func (media *Media) Close() {
	C.avformat_free_context(media.ctx)
	media.ctx = nil
	if media.mediaIO != nil {
		delete(handlersMap, media.mediaIO.handlerKey)
		C.av_free(unsafe.Pointer(media.mediaIO.avAVIOContext.buffer))
		C.av_free(unsafe.Pointer(media.mediaIO.avAVIOContext))
	}
}

// NewMedia returns a new media container analyzer
// for the specified media file.
func NewMedia(filename string) (*Media, error) {
	media := &Media{
		ctx:  C.avformat_alloc_context(),
		name: filename,
	}

	if media.ctx == nil {
		return nil, fmt.Errorf(
			"couldn't create a new media context")
	}

	fname := C.CString(filename)
	status := C.avformat_open_input(&media.ctx, fname, nil, nil)

	if status < 0 {
		return nil, fmt.Errorf(
			"couldn't open file %s", filename)
	}

	C.free(unsafe.Pointer(fname))
	err := media.findStreams()

	if err != nil {
		return nil, err
	}

	return media, nil
}

func getSize(stream io.Reader) int64 {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return int64(buf.Len())
}

func NewMediaFromReadSeeker(name string, bufferSize int, reader io.ReadSeeker) (*Media, error) {
	defaultHandler := &defaultIOHandler{}
	defaultHandler.AVIOHandlers = AVIOHandlers{
		bufferSize: bufferSize,
		ReadSeeker: reader,
		ReadPacket: defaultHandler.ReadPacket,
		ReadFormat: defaultHandler.ReadFormat,
		Seek:       defaultHandler.Seek,
	}
	return NewMediaFromOptions(StreamMediaOptions{
		Alias:      name,
		ReadSeeker: reader,
		IOHandlers: &defaultHandler.AVIOHandlers,
	})
}

func NewMediaFromOptions(options StreamMediaOptions) (*Media, error) {
	if options.IOHandlers == nil {
		return nil, fmt.Errorf("invalid handlers")
	}
	media := &Media{
		ctx:  C.avformat_alloc_context(),
		name: options.Alias,
	}

	if media.ctx == nil {
		return nil, fmt.Errorf(
			"couldn't create a new media context")
	}

	ioCtx, err := NewAVIOContext(media, options)
	if err != nil {
		fmt.Println(err)
		panic("Error creating avio context")
	}
	media.mediaIO = ioCtx

	media.ctx.pb = ioCtx.avAVIOContext
	media.ctx.flags |= C.AVFMT_FLAG_CUSTOM_IO

	tFormat, err := options.IOHandlers.ReadFormat(media)
	if err != nil {
		return nil, fmt.Errorf(
			"couldn't determine media format")
	}

	C.setFormat(media.ctx, *tFormat)

	if media.ctx.iformat == nil {
		return nil, fmt.Errorf(
			"couldn't determine media format")
	}

	status := C.avformat_open_input(&media.ctx, nil, nil, nil)

	if status < 0 {
		return nil, fmt.Errorf(
			"couldn't open file %s", options.Alias)
	}

	err = media.findStreams()

	if err != nil {
		return nil, err
	}
	return media, nil
}

func NewAVIOContext(media *Media, options StreamMediaOptions) (*AVIOContext, error) {
	result := &AVIOContext{}

	buffer := (*C.uchar)(C.av_malloc(C.size_t(options.IOHandlers.bufferSize)))

	if buffer == nil {
		return nil, fmt.Errorf("unable to allocate buffer")
	}

	// we have to explicitly set it to nil, to force library using default handlers
	var ptrRead, ptrWrite, ptrSeek *[0]byte = nil, nil, nil

	if handlersMap == nil {
		handlersMap = make(map[uintptr]*AVIOHandlers)
	}
	ptr := uintptr(unsafe.Pointer(media.ctx))
	handlersMap[ptr] = options.IOHandlers
	result.handlerKey = ptr

	flag := 0

	if options.IOHandlers.ReadPacket != nil {
		ptrRead = (*[0]byte)(C.readCallBack)
	}

	if options.IOHandlers.WritePacket != nil {
		ptrWrite = (*[0]byte)(C.writeCallBack)
		flag = avio_flag_write
	}

	if options.IOHandlers.Seek != nil {
		ptrSeek = (*[0]byte)(C.seekCallBack)
	}

	if options.IOHandlers.ReadPacket != nil && options.IOHandlers.WritePacket != nil {
		flag = avio_flag_read_write
	}

	if result.avAVIOContext = C.avio_alloc_context(buffer, C.int(options.IOHandlers.bufferSize), C.int(flag), unsafe.Pointer(media.ctx), ptrRead, ptrWrite, ptrSeek); result.avAVIOContext == nil {
		return nil, fmt.Errorf("unable to initialize avio context")
	}

	return result, nil
}

//export readCallBack
func readCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.ReadPacket == nil {
		panic("No reader handler initialized")
	}

	b, n := handlers.ReadPacket(int(buf_size))
	if n >= 0 {
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))
	}

	return C.int(n)
}

//export writeCallBack
func writeCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.WritePacket == nil {
		panic("No writer handler initialized.")
	}

	return C.int(handlers.WritePacket(C.GoBytes(unsafe.Pointer(buf), buf_size)))
}

//export seekCallBack
func seekCallBack(opaque unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.Seek == nil {
		panic("No seek handler initialized.")
	}

	return C.int64_t(handlers.Seek(int64(offset), int(whence)))
}
