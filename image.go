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


// Image struct is required by several functions. Don't access the content directly.
type Image struct {
  image     *C.struct_liq_image
  buffer      []byte    // set to prevent GC from cleaning up pixel buffer prematurely
  bufferRows  [][]byte  // set to prevent GC from cleaning up pixel buffer prematurely
}


// Creates an object that represents the image pixels to be used for quantization and remapping.
//
// The pixel array must be contiguous run of RGBA pixels (alpha is the last component, 0 = transparent, 255 = opaque).
//
// The rgba array must not be modified or freed until this object is freed with liq_image_destroy. See also liq_image_set_memory_ownership.
//
// width and height are dimensions in pixels. An image 10x10 pixel large will need a 400-byte array.
//
// gamma can be 0 for images with the typical 1/2.2 gamma. Otherwise gamma must be > 0 and < 1, e.g. 0.45455 (1/2.2) or 0.55555 (1/1.8). 
//
// Generated palette will use the same gamma unless SetOutputGamma is used. If SetOutputGamma is not used, then it only affects whether 
// brighter or darker areas of the image will get more palette colors allocated.
//
// Returns nil on failure, e.g. if rgba is nil or too small or width/height is <= 0.
func (att *Attributes) CreateImageBuffer(rgba []byte, width, height int, gamma float64) *Image {
  if width < 0 || height < 0 { return nil }
  if rgba == nil || len(rgba) < width*height*4 { return nil }
  // img := Image{ nil }
  img := new(Image)
  img.image = C.liq_image_create_rgba(att.attr, unsafe.Pointer(&rgba[0]), C.int(width), C.int(height), C.double(gamma))
  if img.image == nil { return nil }
  img.buffer = rgba
  runtime.SetFinalizer(img, freeImage)
  return img
}

// Same as CreateImageBuffer, but takes an array of rows of pixels.
//
// This allows defining images with reversed rows (like in BMP), "stride" different than width or using only fragment of a larger bitmap, etc.
// The rows array must have at least height elements, and each row must be at least width RGBA pixels wide.
func (att *Attributes) CreateImageBufferRows(rgbaRows [][]byte, width, height int, gamma float64) *Image {
  if width < 0 || height < 0 { return nil }
  if rgbaRows == nil || len(rgbaRows) < height { return nil }

  // img := Image{ nil }
  img := new(Image)
  rowPtr := make([]uintptr, len(rgbaRows))
  for i := 0; i < len(rgbaRows); i++ {
    if rgbaRows[i] == nil || len(rgbaRows[i]) < width * 4 { return nil }
    rowPtr[i] = uintptr(unsafe.Pointer(&rgbaRows[i][0]))
  }
  img.image = C.liq_image_create_rgba_rows(att.attr, (*unsafe.Pointer)(unsafe.Pointer(&rowPtr[0])), C.int(width), C.int(height), C.double(gamma))
  if img.image == nil { return nil }
  img.bufferRows = rgbaRows
  runtime.SetFinalizer(img, freeImage)
  return img
}

// Same as CreateImageBuffer, but takes a Go Image interface as source.
func (att *Attributes) CreateImage(img image.Image, gamma float64) *Image {
  buf := imageToBytes32(img)
  width, height := img.Bounds().Dx(), img.Bounds().Dy()
  return att.CreateImageBuffer(buf, width, height, gamma)
}

// Analyze and remap this image with assumption that it will be always presented exactly on top of this background.
//
// When this image is remapped to a palette with a fully transparent color (use AddImageFixedColor to ensure this) 
// pixels that are better represented by the background than the palette will be made transparent. This function can 
// be used to improve quality of animated GIFs by setting previous animation frame as the background.
//
// Returns ErrBufferTooSmall if the background image has a different size than the foreground.
func (att *Attributes) SetImageBackground(img *Image, background *Image) error {
  code := C.liq_image_set_background(img.image, background.image)
  return getError(code)
}

// Importance map controls which areas of the image get more palette colors.
//
// Pixels corresponding to 0 values in the map are completely ignored. 
// The higher the value the more weight is placed on the given pixel, giving it higher chance of influencing the final palette.
// The map is one byte per pixel and must have the same size as the image (width×height bytes).
//
// Returns ErrInvalidPointer if any pointer is nil and ErrBufferTooSmall if the map size does not match the image size.
func (att *Attributes) SetImageImportanceMap(img *Image, importanceMap []byte) error {
  code := C.liq_image_set_importance_map(img.image, (*C.uchar)(unsafe.Pointer(&importanceMap[0])), C.size_t(len(importanceMap)), C.LIQ_COPY_PIXELS)
  return getError(code)
}

// Reserves a color in the output palette created from this image.
//
// It behaves as if the given color was used in the image and was very important.
// RGB values of the Color object are assumed to have the same gamma as the image. It must be called before the image is quantized.
//
// Returns error if more than 256 colors are added. If image is quantized to fewer colors than the number of fixed colors added, then excess fixed colors will be ignored.
func (att *Attributes) AddImageFixedColor(img *Image, col color.Color) error {
  r, g, b, a := col.RGBA()
  c := C.struct_liq_color{}
  c.r, c.g, c.b, c.a = C.uchar(r), C.uchar(g), C.uchar(b), C.uchar(a)
  code := C.liq_image_add_fixed_color(img.image, c)
  return getError(code)
}

// Getter for image width.
func (att *Attributes) GetImageWidth(img *Image) int {
  return int(C.liq_image_get_width(img.image))
}

// Getter for image height.
func (att *Attributes) GetImageHeight(img *Image) int {
  return int(C.liq_image_get_height(img.image))
}


// Used internally. Frees an Image object.
func freeImage(i *Image) {
  if i.image != nil {
    // fmt.Println("Releasing Image object.")
    C.liq_image_destroy(i.image)
    i.image = nil
    i.buffer = nil
    i.bufferRows = nil
  }
}
