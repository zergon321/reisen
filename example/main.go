package main

import (
	"fmt"

	"github.com/zergon321/reisen"
)

func main() {
	media, err := reisen.NewMedia("demo.mp4")
	handleError(err)
	defer media.Close()
	dur, err := media.Duration()
	handleError(err)

	fmt.Println("Duration:", dur)
	fmt.Println("Format name:", media.FormatName())
	fmt.Println("Format long name:", media.FormatLongName())
	fmt.Println("MIME type:", media.FormatMIMEType())
	fmt.Println("Number of streams:", media.StreamCount())
	fmt.Println()

	for _, stream := range media.Streams() {
		dur, err := stream.Duration()
		handleError(err)
		tbNum, tbDen := stream.TimeBase()
		fpsNum, fpsDen := stream.FrameRate()

		fmt.Println("Index:", stream.Index())
		fmt.Println("Stream type:", stream.Type())
		fmt.Println("Codec name:", stream.CodecName())
		fmt.Println("Codec long name:", stream.CodecLongName())
		fmt.Println("Stream duration:", dur)
		fmt.Println("Stream bit rate:", stream.BitRate())
		fmt.Printf("Time base: %d/%d\n", tbNum, tbDen)
		fmt.Printf("Frame rate: %d/%d\n", fpsNum, fpsDen)
		fmt.Println("Frame count:", stream.FrameCount())

		switch s := stream.(type) {
		case *reisen.VideoStream:
			num, den := s.AspectRatio()

			fmt.Println("Width:", s.Width())
			fmt.Println("Height:", s.Height())
			fmt.Printf("Aspect ratio: %d:%d\n", num, den)

		case *reisen.AudioStream:
			fmt.Println("Sample rate:", s.SampleRate())
			fmt.Println("Number of channels:", s.ChannelCount())
			fmt.Println("Frame size:", s.FrameSize())
		}

		fmt.Println()
	}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
