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
brew install libav
```

## Installation

Just casually run this command:

```bash
go get github.com/zergon321/reisen
```