package reisen

// #cgo LDFLAGS: -lswscale
// #include <libswscale/swscale.h>
// #include <inttypes.h>
import "C"

type InterpolationAlgorithm int

const (
	InterpolationFastBilinear    InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_FAST_BILINEAR)
	InterpolationBilinear        InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_BILINEAR)
	InterpolationBicubic         InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_BICUBIC)
	InterpolationX               InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_X)
	InterpolationPoint           InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_POINT)
	InterpolationArea            InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_AREA)
	InterpolationBicubicBilinear InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_BICUBLIN)
	InterpolationGauss           InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_GAUSS)
	InterpolationSinc            InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_SINC)
	InterpolationLanczos         InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_LANCZOS)
	InterpolationSpline          InterpolationAlgorithm = InterpolationAlgorithm(C.SWS_SPLINE)
)
