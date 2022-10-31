package reisen

// #cgo pkg-config: libswscale
// #include <libswscale/swscale.h>
import "C"

// InterpolationAlgorithm is used when
// we scale a video frame in a different resolution.
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

// String returns the name of the interpolation algorithm.
func (interpolationAlg InterpolationAlgorithm) String() string {
	switch interpolationAlg {
	case InterpolationFastBilinear:
		return "fast bilinear"

	case InterpolationBilinear:
		return "bilinear"

	case InterpolationBicubic:
		return "bicubic"

	case InterpolationX:
		return "x"

	case InterpolationPoint:
		return "point"

	case InterpolationArea:
		return "area"

	case InterpolationBicubicBilinear:
		return "bicubic bilinear"

	case InterpolationSinc:
		return "sinc"

	case InterpolationLanczos:
		return "lanczos"

	case InterpolationSpline:
		return "spline"

	default:
		return ""
	}
}
