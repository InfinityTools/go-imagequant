package imagequant

/*
#include "libimagequant.h"
*/
import "C"

import (
  "image"
  "image/color"
  "runtime"
  "unsafe"
)


const (
  DITHER_MIN  = 0.0
  DITHER_MAX  = 1.0
)

// Result struct is required by several functions. Don't access the content directly.
type Result struct {
  result    *C.struct_liq_result
}


// Generates a palette from the histogram. On success returns the fully initialized Result object.
func (att *Attributes) QuantizeHistogram(hist *Histogram) (res *Result, err error) {
  res = new(Result)
  code := C.liq_histogram_quantize(hist.histogram, att.attr, (**C.struct_liq_result)(unsafe.Pointer(&res.result)))
  runtime.SetFinalizer(res, freeResult)
  err = getError(code)
  return
}

// Performs quantization (palette generation) based on current Quantizer settings and pixels of the image.
//
// Returns the Result object if quantization succeeds.
// Error returns ErrQualityTooLow if quantization fails due to limit set in SetQuality.
func (att *Attributes) QuantizeImage(img *Image) (res *Result, err error) {
  res = new(Result)
  code := C.liq_image_quantize(img.image, att.attr, (**C.struct_liq_result)(unsafe.Pointer(&res.result)))
  runtime.SetFinalizer(res, freeResult)
  err = getError(code)
  return
}

// Enables/disables dithering in WriteRemappedImage.
//
// Dithering level must be between 0 and 1 (inclusive). Dithering level 0 enables fast non-dithered remapping.
// Otherwise a variation of Floyd-Steinberg error diffusion is used.
func (att *Attributes) SetDitheringLevel(res *Result, ditherLevel float32) error {
  code := C.liq_set_dithering_level(res.result, C.float(ditherLevel))
  return getError(code)
}

// Sets gamma correction for generated palette and remapped image.
//
// Must be > 0 and < 1, e.g. 0.45455 for gamma 1/2.2 in PNG images. By default output gamma is same as gamma of the input image.
func (att *Attributes) SetOutputGamma(res *Result, gamma float64) error {
  code := C.liq_set_output_gamma(res.result, C.double(gamma))
  return getError(code)
}

// Returns the gamma value for the output image.
func (att *Attributes) GetOutputGamma(res *Result) float64 {
  return float64(C.liq_get_output_gamma(res.result))
}

// Returns a palette optimized for the image that has been quantized or remapped (final refinements are applied to the palette during remapping).
//
// It's valid to call this method before remapping, if you don't plan to remap any images or want to use same palette for multiple images.
// Returns a Palette object with 0 color entries on error.
func (att *Attributes) GetPalette(res *Result) color.Palette {
  var palette color.Palette = nil
  pal := C.liq_get_palette(res.result)
  if pal != nil {
    palette = make(color.Palette, (*pal).count)
    for i := 0; i < len(palette); i++ {
      r, g, b, a := byte((*pal).entries[i].r), byte((*pal).entries[i].g), byte((*pal).entries[i].b), byte((*pal).entries[i].a)
      // compensating rounding errors
      if r > a { r = a }
      if g > a { g = a }
      if b > a { b = a }
      palette[i] = color.RGBA{ r, g, b, a }
    }
  } else {
    palette = make(color.Palette, 0)
  }
  return palette
}

// Remaps the image to palette and returns the converted image as a byte array, 1 pixel per byte.
//
// For best performance call GetPalette after this function, as palette is improved during remapping
// (except when QuantizeHistogram is used).
//
// The returned byte array is assumed to be contiguous, with rows ordered from top to bottom, and no gaps between rows.
// If you need to return a sequence of rows with padding or upside-down order, then use WriteRemappedImageRows.
func (att *Attributes) WriteRemappedImageBuffer(res *Result, img *Image) (buf []byte, err error) {
  buf = make([]byte, att.GetImageWidth(img) * att.GetImageHeight(img))
  code := C.liq_write_remapped_image(res.result, img.image, unsafe.Pointer(&buf[0]), C.size_t(len(buf)))
  err = getError(code)
  return
}

// Similar to WriteRemappedImageBuffer. Returns a remapped image, at 1 byte per pixel, to each row pointed by rows multi-array.
//
// The array must have at least as many elements as height of the image, and each row must have at least as many bytes as width of the image.
// Rows must not overlap.
func (att *Attributes) WriteRemappedImageBufferRows(res *Result, img *Image, rows [][]byte) (rowsOut [][]byte, err error) {
  if rows == nil { err = ErrInvalidPointer; return }
  if len(rows) < att.GetImageHeight(img) { err = ErrBufferTooSmall; return }

  rowPtr := make([]*C.uchar, len(rows))
  width := att.GetImageWidth(img)
  for i := 0; i < len(rows); i++ {
    if rows[i] == nil || len(rows[i]) < width { err = ErrBufferTooSmall; return }
    rowPtr[i] = (*C.uchar)(unsafe.Pointer(&rows[i][0]))
  }
  code := C.liq_write_remapped_image_rows(res.result, img.image, (**C.uchar)(unsafe.Pointer(&rowPtr[0])))
  rowsOut = rows
  err = getError(code)
  return
}

// A convenience function that returns a paletted Go Image object.
func (att *Attributes) WriteRemappedImage(res *Result, img *Image) (imgOut image.Image, err error) {
  buf, err := att.WriteRemappedImageBuffer(res, img)
  if err != nil { return }

  pal := att.GetPalette(res)
  if len(pal) == 0 { err = ErrUnknown; return }

  width, height := att.GetImageWidth(img), att.GetImageHeight(img)
  imgOut = bytesToPaletted(width, height, pal, buf)
  return
}


// Returns mean square error of quantization (square of difference between pixel values in the source image and its remapped version).
//
// Alpha channel, gamma correction and approximate importance of pixels is taken into account, so the result isn't exactly the mean square error of all channels.
// or most images MSE 1-5 is excellent. 7-10 is OK. 20-30 will have noticeable errors. 100 is awful.
//
// This function may return -1 if the value is not available (this happens when a high speed has been requested, the image hasn't been remapped yet,
// and quality limit hasn't been set, see SetSpeed and SetQuality). The value is not updated when multiple images are remapped, it applies only to the image
// used in QuantizeImage or the first image that has been remapped. See GetRemappingError.
func (att *Attributes) GetQuantizationError(res *Result) float64 {
  return float64(C.liq_get_quantization_error(res.result))
}

// Analoguous to GetQuantizationError, but returns quantization error as quality value in the same 0-100 range that is used by SetQuality.
//
// It may return -1 if the value is not available (see note in GetQuantizationError).
func (att *Attributes) GetQuantizationQuality(res *Result) int {
  return int(C.liq_get_quantization_quality(res.result))
}

// Returns mean square error of last remapping done (square of difference between pixel values in the remapped image and its remapped version).
//
// Alpha channel and gamma correction are taken into account, so the result isn't exactly the mean square error of all channels.
func (att *Attributes) GetRemappingError(res *Result) float64 {
  return float64(C.liq_get_remapping_error(res.result))
}

// Analoguous to GetRemappingError, but returns quantization error as quality value in the same 0-100 range that is used by SetQuality.
func (att *Attributes) GetRemappingQuality(res *Result) int {
  return int(C.liq_get_remapping_quality(res.result))
}


// Used internally. Frees a Result object.
func freeResult(r *Result) {
  if r.result != nil {
    // fmt.Println("Releasing Result object.")
    C.liq_result_destroy(r.result)
    r.result = nil
  }
}
