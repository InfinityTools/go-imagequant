/*
Package imagequant provides bindings to the external imagequant C library.

Original C library: https://github.com/ImageOptim/libimagequant/
*/
package imagequant
// Alternative Go binding package (by larrabee): https://github.com/ultimate-guitar/go-imagequant
// TODO: Attempt to implement callback support:
// - liq_set_log_callback
// - liq_set_log_flush_callback
// - liq_attr_set_progress_callback
// - liq_result_set_progress_callback
// - liq_image_create_custom

/*
// CGO linker flags are defined for selected platforms windows/linux/freebsd/darwin and architectures 386/amd64.
// Set CGO_LDFLAGS environment variable manually for undefined platforms and architectures or non-standard configurations.
// Note: Use static linking on Windows. Override via CGO_LDFLAGS if needed.
#cgo windows,386 LDFLAGS: -Llibs/windows/386 -limagequant -lm -static
#cgo windows,amd64 LDFLAGS: -Llibs/windows/amd64 -limagequant -lm -static
#cgo linux,386 LDFLAGS: -Llibs/linux/386 -limagequant -lm
#cgo linux,amd64 LDFLAGS: -Llibs/linux/amd64 -limagequant -lm
#cgo freebsd,386 LDFLAGS: -Llibs/freebsd/386 -limagequant -lm
#cgo freebsd,amd64 LDFLAGS: -Llibs/freebsd/amd64 -limagequant -lm
#cgo darwin,386 LDFLAGS: -Llibs/darwin/386 -limagequant -lm
#cgo darwin,amd64 LDFLAGS: -Llibs/darwin/amd64 -limagequant -lm
#include "libimagequant.h"

const char* liqVersionString() {
  return LIQ_VERSION_STRING;
}

*/
import "C"

import (
  "errors"
)


var (
  // Potential error codes
  ErrQualityTooLow      = errors.New("Quality too low")
  ErrValueOutOfRange    = errors.New("Value is out of range")
  ErrOutOfMemory        = errors.New("Out of memory")
  ErrAborted            = errors.New("Aborted")
  ErrBitmapNotAvailable = errors.New("Bitmap is not available")
  ErrBufferTooSmall     = errors.New("Buffer is too small")
  ErrInvalidPointer     = errors.New("Invalid pointer")
  ErrUnsupported        = errors.New("Unsupported")
  ErrUnknown            = errors.New("Unknown error")
)


// GetVersion returns the imagequant library version as major, minor and patch number.
func GetVersion() (major, minor, patch int) {
  value := int(C.liq_version())
  patch = value % 100
  value /= 100
  minor = value % 100
  value /= 100
  major = value
  return
}

// GetVersionString returns the imagequant library version as a preformatted string.
func GetVersionString() string {
  return C.GoString(C.liqVersionString())
}


// Used internally. Translates error codes to Golang errors.
func getError(code C.liq_error) error {
  switch code {
  case C.LIQ_OK:
    return nil
  case C.LIQ_QUALITY_TOO_LOW:
    return ErrQualityTooLow
  case C.LIQ_VALUE_OUT_OF_RANGE:
    return ErrValueOutOfRange
  case C.LIQ_OUT_OF_MEMORY:
    return ErrOutOfMemory
  case C.LIQ_ABORTED:
    return ErrAborted
  case C.LIQ_BITMAP_NOT_AVAILABLE:
    return ErrBitmapNotAvailable
  case C.LIQ_BUFFER_TOO_SMALL:
    return ErrBufferTooSmall
  case C.LIQ_INVALID_POINTER:
    return ErrInvalidPointer
  case C.LIQ_UNSUPPORTED:
    return ErrUnsupported
  default:
    return ErrUnknown
  }
}
