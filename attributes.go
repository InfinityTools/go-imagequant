package imagequant

/*
#include "libimagequant.h"
*/
import "C"

import (
  "runtime"
)

const (
  SPEED_SLOWEST = 1
  SPEED_DEFAULT = 3
  SPEED_FASTEST = 10
)

const (
  QUALITY_BEST  = 100
  QUALITY_GOOD  = 80
  QUALITY_WORST = 0
)

// The Attributes struct is used to call the majority of quantization functions.
// This is the only structure that can be released manually.
type Attributes struct {
  attr      *C.struct_liq_attr
}


// Returns an object that will hold initial settings (attributes) for the library.
//
// IMPORTANT: The object must be freed by Release after it is no longer needed.
func CreateAttributes() *Attributes {
  att := new(Attributes)
  att.attr = C.liq_attr_create()
  runtime.SetFinalizer(att, freeAttribute)
  return att
}

// Creates an independent copy of the calling object.
//
// IMPORTANT: The copy must also be freed by the Release function.
func (att *Attributes) CopyAttribute() *Attributes {
  att2 := new(Attributes)
  att2.attr = C.liq_attr_copy(att.attr)
  runtime.SetFinalizer(att2, freeAttribute)
  return att2
}

// Call this function to manually release the attributes.
// Otherwise, Golang's own garbage collector will take care of it eventually.
func (att *Attributes) Release() {
  freeAttribute(att)
}


// Specifies the maximum number of colors to use. The default is 256.
//
// Instead of setting a fixed limit it's better to use SetQuality.
// Returns ErrValueOutOfRange if number of colors is outside the range 2-256.
func (att *Attributes) SetMaxColors(colors int) error {
  code := C.liq_set_max_colors(att.attr, C.int(colors))
  return getError(code)
}

// Returns the value set by SetMaxColors.
func (att *Attributes) GetMaxColors() int {
  retVal := C.liq_get_max_colors(att.attr)
  return int(retVal)
}

// Higher speed levels disable expensive algorithms and reduce quantization precision. The default speed is 3.
//
// Speed 1 gives marginally better quality at significant CPU cost. Speed 10 has usually 5% lower quality, but is 8 times faster than the default.
// High speeds combined with SetQuality will use more colors than necessary and will be less likely to meet minimum required quality.
// The range of SPEED_xxx constants covers the common uses.
// Features dependent on speed:
//    Noise-sensitive dithering speed     1 to 5
//    Forced posterization                8-10 or if image has more than million colors
//    Quantization error known            1-7 or if minimum quality is set
//    Additional quantization techniques  1-6
// Returns ErrValueOutOfRange if the speed is outside the 1-10 range.
func (att *Attributes) SetSpeed(speed int) error {
  code := C.liq_set_speed(att.attr, C.int(speed))
  return getError(code)
}

// Returns the value set by SetSpeed.
func (att *Attributes) GetSpeed() int {
  retVal := C.liq_get_speed(att.attr)
  return int(retVal)
}

// Ignores the given number of least significant bits in all channels, posterizing image to 2^bits levels.
//
// 0 gives full quality. Use 2 for VGA or 16-bit RGB565 displays, 4 if image is going to be output
// on a RGB444/RGBA4444 display (e.g. low-quality textures on Android).
//
// Returns LIQ_VALUE_OUT_OF_RANGE if the value is outside the 0-4 range.
func (att *Attributes) SetMinPosterization(bits int) error {
  code := C.liq_set_min_posterization(att.attr, C.int(bits))
  return getError(code)
}

// Returns the value set by SetMinPosterization.
func (att *Attributes) GetMinPosterization() int {
  retVal := C.liq_get_min_posterization(att.attr)
  return int(retVal)
}

// Quality is in range 0 (worst) to 100 (best) and values are analoguous to JPEG quality (i.e. 80 is usually good enough).
//
// Quantization will attempt to use the lowest number of colors needed to achieve maximum quality. max value of 100 is the default
// and means conversion as good as possible. If it's not possible to convert the image with at least minimum quality (i.e. 256 colors
// is not enough to meet the minimum quality), then liq_image_quantize will fail. The default minimum is 0 (proceeds regardless of quality).
//
// Quality measures how well the generated palette fits image given to liq_image_quantize. If a different image is remapped with
// WriteRemappedImage, then actual quality may be different.
// Regardless of the quality settings the number of colors won't exceed the maximum (see SetMaxColors).
// The range of QUALITY_xxx constants covers the common uses.
//
// Returns ErrValueOutOfRange if target is lower than minimum or any of them is outside the 0-100 range.
// Returns ErrInvalidPointer if attr appears to be invalid.
func (att *Attributes) SetQuality(min, max int) error {
  code := C.liq_set_quality(att.attr, C.int(min), C.int(max))
  return getError(code)
}

// Returns the minimum/maximum range of quality set by SetQuality.
func (att *Attributes) GetQuality() (min, max int) {
  min = int(C.liq_get_min_quality(att.attr))
  max = int(C.liq_get_max_quality(att.attr))
  return
}

// Setting to false makes alpha colors sorted before opaque colors. "true" mixes colors together except completely transparent color,
// which is moved to the end of the palette. This is a workaround for programs that blindly assume the last palette entry is transparent.
func (att *Attributes) SetLastIndexTransparent(set bool) {
  v := 0
  if set { v = 1 }
  C.liq_set_last_index_transparent(att.attr, C.int(v))
}


// Used internally. Frees a Attributes object.
func freeAttribute(att *Attributes) {
  if att.attr != nil {
    // fmt.Println("Releasing Attributes object.")
    C.liq_attr_destroy(att.attr)
    att.attr = nil
    runtime.GC()
  }
}
