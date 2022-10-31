# Reisen [![GoDoc](https://godoc.org/github.com/zergon321/reisen?status.svg)](https://pkg.go.dev/github.com/zergon321/reisen)

A simple library to extract video and audio frames from media containers (based on **libav**, i.e. **ffmpeg**).

## Dependencies

The library requires **libav** components to work:

- **libavformat**
- **libavcodec**
- **libavutil**
- **libswresample**
- **libswscale**

For **Arch**-based **Linux** distributions:

```bash
sudo pacman -S ffmpeg
```

For **Debian**-based **Linux** distributions:

```bash
sudo apt install libswscale-dev libavcodec-dev libavformat-dev libswresample-dev libavutil-dev
```

For **macOS**:

```bash
brew install ffmpeg
```

For **Windows** see the [detailed tutorial](https://medium.com/@maximgradan/how-to-easily-bundle-your-cgo-application-for-windows-8515d2b19f1e).

## Installation

Just casually run this command:

```bash
go get github.com/zergon321/reisen
```

## Usage

Any media file is composed of streams containing media data, e.g. audio, video and subtitles. The whole presentation data of the file is divided into packets. Each packet belongs to one of the streams and represents a single frame of its data. The process of decoding implies reading packets and decoding them into either video frames or audio frames.

The library provides read video frames as **RGBA** pictures. The audio samples are provided as raw byte slices in the format of `AV_SAMPLE_FMT_DBL` (i.e. 8 bytes per sample for one channel, the data type is `float64`). The channel layout is stereo (2 channels). The byte order is little-endian. The detailed scheme of the audio samples sequence is given below.

![Audio sample structure](https://github.com/zergon321/reisen/blob/master/pictures/audio_sample_structure.png)

You are welcome to look at the [examples](https://github.com/zergon321/reisen/tree/master/examples) to understand how to work with the library. Also please take a look at the detailed [tutorial](https://medium.com/@maximgradan/playing-videos-with-golang-83e67447b111).
