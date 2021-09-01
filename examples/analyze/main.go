package main

import (
	"fmt"

	"github.com/zergon321/reisen"
)

func main() {
	// Open the media file by its name.
	media, err := reisen.NewMedia("demo.mp4")
	handleError(err)
	defer media.Close()
	dur, err := media.Duration()
	handleError(err)

	// Print the media properties.
	fmt.Println("Duration:", dur)
	fmt.Println("Format name:", media.FormatName())
	fmt.Println("Format long name:", media.FormatLongName())
	fmt.Println("MIME type:", media.FormatMIMEType())
	fmt.Println("Number of streams:", media.StreamCount())
	fmt.Println()

	// Enumerate the media file streams.
	for _, stream := range media.Streams() {
		dur, err := stream.Duration()
		handleError(err)
		tbNum, tbDen := stream.TimeBase()
		fpsNum, fpsDen := stream.FrameRate()

		// Print the properties common
		// for both stream types.
		fmt.Println("Index:", stream.Index())
		fmt.Println("Stream type:", stream.Type())
		fmt.Println("Codec name:", stream.CodecName())
		fmt.Println("Codec long name:", stream.CodecLongName())
		fmt.Println("Stream duration:", dur)
		fmt.Println("Stream bit rate:", stream.BitRate())
		fmt.Printf("Time base: %d/%d\n", tbNum, tbDen)
		fmt.Printf("Frame rate: %d/%d\n", fpsNum, fpsDen)
		fmt.Println("Frame count:", stream.FrameCount())
		fmt.Println()
	}

	// Do decoding.
	err = media.OpenDecode()
	handleError(err)
	gotPacket := true

	for i := 0; i < 9 && gotPacket; i++ {
		// Read packets one by one. A packet
		// can contain either a video frame
		// or an audio frame.
		var pkt *reisen.Packet
		pkt, gotPacket, err = media.ReadPacket()
		handleError(err)

		// Check if the media file
		// is depleted.
		if !gotPacket {
			break
		}

		// Determine what stream
		// the packet belongs to.
		switch pkt.Type() {
		case reisen.StreamVideo:
			s := media.Streams()[pkt.StreamIndex()].(*reisen.VideoStream)

			if !s.Opened() {
				err = s.Open()
				handleError(err)
			}

			videoFrame, gotFrame, err := s.ReadVideoFrame()
			handleError(err)

			// If the media file is
			// depleted.
			if !gotFrame {
				break
			}

			// If the packet doesn't
			// contain a whole frame,
			// just skip it.
			if videoFrame == nil {
				continue
			}

			pts, err := videoFrame.PresentationOffset()
			handleError(err)

			fmt.Println("Presentation duration offset:", pts)
			fmt.Println("Number of pixels:", len(videoFrame.Image().Pix))
			fmt.Println()

		case reisen.StreamAudio:
			s := media.Streams()[pkt.StreamIndex()].(*reisen.AudioStream)

			if !s.Opened() {
				err = s.Open()
				handleError(err)
			}

			audioFrame, gotFrame, err := s.ReadAudioFrame()
			handleError(err)

			if !gotFrame {
				break
			}

			if audioFrame == nil {
				continue
			}

			pts, err := audioFrame.PresentationOffset()
			handleError(err)

			fmt.Println("Presentation duration offset:", pts)
			fmt.Println("Data length:", len(audioFrame.Data()))
			fmt.Println()
		}
	}

	for _, stream := range media.Streams() {
		err = stream.Close()
		handleError(err)
	}

	err = media.CloseDecode()
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
