package reisen

type ErrorType int

const (
	// ErrorAgain is returned when
	// the decoder needs more data
	// to serve the frame.
	ErrorAgain ErrorType = -11
	// ErrorInvalidValue is returned
	// when the function call argument
	// is invalid.
	ErrorInvalidValue ErrorType = -22
	// ErrorEndOfFile is returned upon
	// reaching the end of the media file.
	ErrorEndOfFile ErrorType = -541478725
)
