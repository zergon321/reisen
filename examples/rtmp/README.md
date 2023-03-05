# An RTMP stream reading example

## Dependencies

- **Docker**
- `docker-compose`

## How to launch

First run `docker-compose up` to launch the **RTMP** streaming server. Then run `./stream.sh` to start streaming `video.mp4` over **RTMP**. After it execute `go run main.go` to watch the stream.