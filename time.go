package reisen

// #cgo LDFLAGS: -lavutil
// #include <libavutil/avutil.h>
import "C"

const (
	// TimeBase is a global time base
	// used for describing media containers.
	TimeBase int = C.AV_TIME_BASE
)
