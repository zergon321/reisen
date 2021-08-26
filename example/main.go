package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/zergon321/reisen"
)

func main() {
	media, err := reisen.NewMedia("demo.mkv")
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

			err := stream.Open()
			handleError(err)
			gotFrame := true

			for i := 0; i < 7 && gotFrame; i++ {
				var videoFrame *reisen.VideoFrame
				videoFrame, gotFrame, err = s.ReadVideoFrame()
				handleError(err)

				if videoFrame == nil {
					continue
				}

				if !gotFrame {
					break
				}

				pts, err := videoFrame.PresentationOffset()
				handleError(err)

				fmt.Println("Presentation duration offset:", pts)

				file, err := os.Create(fmt.Sprintf("frame_%d.png", i))
				handleError(err)
				err = png.Encode(file, videoFrame.Image())
				handleError(err)
			}

			err = stream.Close()
			handleError(err)

		case *reisen.AudioStream:
			fmt.Println("Sample rate:", s.SampleRate())
			fmt.Println("Number of channels:", s.ChannelCount())
			fmt.Println("Frame size:", s.FrameSize())

			err := stream.Open()
			handleError(err)
			gotFrame := true

			for i := 0; i < 7 && gotFrame; i++ {
				var audioFrame *reisen.AudioFrame
				audioFrame, gotFrame, err = s.ReadAudioFrame()
				handleError(err)

				if audioFrame == nil {
					continue
				}

				if !gotFrame {
					break
				}

				pts, err := audioFrame.PresentationOffset()
				handleError(err)

				fmt.Println("Presentation duration offset:", pts)
				fmt.Println("Data length:", len(audioFrame.Data()))
			}
		}

		fmt.Println()
	}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
