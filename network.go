package reisen

// #include <libavformat/avformat.h>
import "C"
import "fmt"

func NetworkInitialize() error {
	code := C.avformat_network_init()

	if code < 0 {
		return fmt.Errorf("error occurred: 0x%X", code)
	}

	return nil
}

func NetworkDeinitialize() error {
	code := C.avformat_network_deinit()

	if code < 0 {
		return fmt.Errorf("error occurred: 0x%X", code)
	}

	return nil
}
