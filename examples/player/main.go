package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/hajimehoshi/ebiten"
	_ "github.com/silbinarywolf/preferdiscretegpu"
	"github.com/zergon321/reisen"
)

const (
	startWidth                        = 1280
	startHeight                       = 720
	frameBufferLength                 = 10
	sampleRate                        = 44100
	sampleBufferLength                = 4096
	SpeakerSampleRate beep.SampleRate = 44100
)

// readVideoAndAudio reads video and audio frames
// from the opened media and sends the decoded
// data to che channels to be played.
func readVideoAndAudio(media *reisen.Media) (<-chan *image.RGBA, <-chan [2]float64, chan error, error) {
	frameBuffer := make(chan *image.RGBA, frameBufferLength)
	sampleBuffer := make(chan [2]float64, sampleBufferLength)
	errs := make(chan error)

	err := media.OpenDecode()

	if err != nil {
		return nil, nil, nil, err
	}

	videoStream := media.VideoStreams()[0]
	err = videoStream.Open()

	if err != nil {
		return nil, nil, nil, err
	}

	audioStream := media.AudioStreams()[0]
	err = audioStream.Open()

	if err != nil {
		return nil, nil, nil, err
	}

	/*err = media.Streams()[0].Rewind(60 * time.Second)

	if err != nil {
		return nil, nil, nil, err
	}*/

	/*err = media.Streams()[0].ApplyFilter("h264_mp4toannexb")

	if err != nil {
		return nil, nil, nil, err
	}*/

	go func() {
		for {
			packet, gotPacket, err := media.ReadPacket()

			if err != nil {
				go func(err error) {
					errs <- err
				}(err)
			}

			if !gotPacket {
				break
			}

			/*hash := sha256.Sum256(packet.Data())
			fmt.Println(base58.Encode(hash[:]))*/

			switch packet.Type() {
			case reisen.StreamVideo:
				s := media.Streams()[packet.StreamIndex()].(*reisen.VideoStream)
				videoFrame, gotFrame, err := s.ReadVideoFrame()

				if err != nil {
					go func(err error) {
						errs <- err
					}(err)
				}

				if !gotFrame {
					break
				}

				if videoFrame == nil {
					continue
				}

				frameBuffer <- videoFrame.Image()

			case reisen.StreamAudio:
				s := media.Streams()[packet.StreamIndex()].(*reisen.AudioStream)
				audioFrame, gotFrame, err := s.ReadAudioFrame()

				if err != nil {
					go func(err error) {
						errs <- err
					}(err)
				}

				if !gotFrame {
					break
				}

				if audioFrame == nil {
					continue
				}

				// Turn the raw byte data into
				// audio samples of type [2]float64.
				reader := bytes.NewReader(audioFrame.Data())

				// See the README.md file for
				// detailed scheme of the sample structure.
				for reader.Len() >= 16 {
					sample := [2]float64{0, 0}
					err = binary.Read(reader, binary.LittleEndian, sample[:])

					if err != nil {
						go func(err error) {
							errs <- err
						}(err)
					}

					sampleBuffer <- sample
				}
			}
		}

		videoStream.Close()
		audioStream.Close()
		media.CloseDecode()
		close(frameBuffer)
		close(sampleBuffer)
		close(errs)
	}()

	return frameBuffer, sampleBuffer, errs, nil
}

// streamSamples creates a new custom streamer for
// playing audio samples provided by the source channel.
//
// See https://github.com/faiface/beep/wiki/Making-own-streamers
// for reference.
func streamSamples(sampleSource <-chan [2]float64) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i := 0; i < len(samples); i++ {
			sample, ok := <-sampleSource

			if !ok {
				return i, false
			}

			samples[i] = sample
		}

		return len(samples), true
	})
}

// Game holds all the data
// necessary for playing video.
type Game struct {
	videoSprite            *ebiten.Image
	ticker                 <-chan time.Time
	errs                   <-chan error
	frameBuffer            <-chan *image.RGBA
	width                  int
	height                 int
	fps                    int
	videoTotalFramesPlayed int
	videoPlaybackFPS       int
	perSecond              <-chan time.Time
	last                   time.Time
	deltaTime              float64
}

// Strarts reading samples and frames
// of the media file.
func (game *Game) Start(fname string) error {
	// Initialize the audio speaker.
	err := speaker.Init(sampleRate,
		SpeakerSampleRate.N(time.Second/10))

	if err != nil {
		return err
	}

	// Open the media file.
	media, err := reisen.NewMedia(fname)

	if err != nil {
		return err
	}

	// Get the FPS for playing
	// video frames.
	videoFPS, _ := media.Streams()[0].FrameRate()

	if err != nil {
		return err
	}

	// SPF for frame ticker.
	spf := 1.0 / float64(videoFPS)
	frameDuration, err := time.
		ParseDuration(fmt.Sprintf("%fs", spf))

	if err != nil {
		return err
	}

	// Start decoding streams.
	var sampleSource <-chan [2]float64
	game.frameBuffer, sampleSource,
		game.errs, err = readVideoAndAudio(media)

	if err != nil {
		return err
	}

	// Start playing audio samples.
	speaker.Play(streamSamples(sampleSource))

	game.ticker = time.Tick(frameDuration)

	// Setup metrics.
	game.last = time.Now()
	game.fps = 0
	game.perSecond = time.Tick(time.Second)
	game.videoTotalFramesPlayed = 0
	game.videoPlaybackFPS = 0

	return nil
}

func (game *Game) Update(screen *ebiten.Image) error {
	// Compute dt.
	game.deltaTime = time.Since(game.last).Seconds()
	game.last = time.Now()

	// Check for incoming errors.
	select {
	case err, ok := <-game.errs:
		if ok {
			return err
		}

	default:
	}

	// Read video frames and draw them.
	select {
	case <-game.ticker:
		frame, ok := <-game.frameBuffer

		if ok {
			rect := frame.Bounds()
			width := int(rect.Max.X - rect.Min.X)
			height := int(rect.Max.Y - rect.Min.Y)

			if game.width != width || game.height != height {
				// Sprite for drawing video frames.
				sprite, err := ebiten.NewImage(width, height, ebiten.FilterDefault)
				if err != nil {
					return err
				}
				ebiten.SetWindowSize(width, height)
				game.videoSprite, game.width, game.height = sprite, width, height
			}
			game.videoSprite.ReplacePixels(frame.Pix)

			game.videoTotalFramesPlayed++
			game.videoPlaybackFPS++
		}

	default:
	}

	// Draw the video sprite.
	op := &ebiten.DrawImageOptions{}
	err := screen.DrawImage(game.videoSprite, op)

	if err != nil {
		return err
	}

	game.fps++

	// Update metrics in the window title.
	select {
	case <-game.perSecond:
		ebiten.SetWindowTitle(fmt.Sprintf("%s | FPS: %d | dt: %f | Frames: %d | Video FPS: %d",
			"Video", game.fps, game.deltaTime, game.videoTotalFramesPlayed, game.videoPlaybackFPS))

		game.fps = 0
		game.videoPlaybackFPS = 0

	default:
	}

	return nil
}

func (game *Game) Layout(a, b int) (int, int) {
	return a, b
}

func main() {
	game := &Game{}
	// Play the video provided in the commandline, or default to demo.mp4
	err := game.Start(append(os.Args[1:], "demo.mp4")[0])
	handleError(err)

	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowTitle("Video")
	err = ebiten.RunGame(game)
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
