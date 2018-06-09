package imagequant

/*
#include "libimagequant.h"
*/
import "C"

import (
  "image/color"
  "runtime"
  "unsafe"
)


// A HistogramEntry holds usage information of a single color value.
type HistogramEntry struct {
  Color     color.Color // The color value definition
  Count     uint        // Number of occurrence, influences the weight or importance of the color.
}

// Histogram struct is required by several functions. Don't accss the content directly.
type Histogram struct {
  histogram *C.struct_liq_histogram
}


// Creates a histogram object that will be used to collect color statistics from multiple images.
func (att *Attributes) CreateHistogram() *Histogram {
  hist := new(Histogram)
  hist.histogram = C.liq_histogram_create(att.attr)
  runtime.SetFinalizer(hist, freeHistogram)
  return hist
}

// "Learns" colors from the image, which will be later used to generate the palette.
//
// After the image is added to the histogram it may be freed to save memory (but it's more efficient to keep the image object around if it's going to be used for remapping).
// Fixed colors added to the image are also added to the histogram. If total number of fixed colors exceeds 256, this function will fail with ErrBufferTooSmall.
func (att *Attributes) AddImageToHistogram(hist *Histogram, img *Image) error {
  code := C.liq_histogram_add_image(hist.histogram, att.attr, img.image)
  return getError(code)
}

// Alternative to AddImageToHistogram. Instead of counting colors in an image, it directly takes an array of colors and their counts.
//
// This function is only useful if you already have a histogram of the image from another source.
func (att *Attributes) AddColorsToHistogram(hist *Histogram, entries []HistogramEntry, gamma float64) error {
  if entries == nil { return ErrInvalidPointer }
  c_entries := make([]C.struct_liq_histogram_entry, len(entries))
  for k, v := range entries {
    r, g, b, a := v.Color.RGBA()
    c_entries[k].color.r = C.uchar(r)
    c_entries[k].color.g = C.uchar(g)
    c_entries[k].color.b = C.uchar(b)
    c_entries[k].color.a = C.uchar(a)
    c_entries[k].count = C.uint(v.Count)
  }
  code := C.liq_histogram_add_colors(hist.histogram, att.attr,
                                     (*C.struct_liq_histogram_entry)(unsafe.Pointer(&c_entries[0])),
                                     C.int(len(c_entries)), C.double(gamma))
  return getError(code)
}


// Used internally. Frees a Histogram object.
func freeHistogram(h *Histogram) {
  if h.histogram != nil {
    // fmt.Println("Releasing Histogram object.")
    C.liq_histogram_destroy(h.histogram)
    h.histogram = nil
  }
}
