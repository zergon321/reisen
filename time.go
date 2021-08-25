package reisen

// #cgo LDFLAGS: -lavutil
// #include <libavutil/avutil.h>
import "C"

const (
	TimeBase int = C.AV_TIME_BASE
)
