package imagequant
// Collection of miscellaneous functions.

import (
  "image"
  "image/color"
)

// NRGBA converts a premultiplied color back to a normalized color with each component in range [0, 255].
func NRGBA(col color.Color) (r, g, b, a byte) {
  if nrgba, ok := col.(color.NRGBA); ok {
    r, g, b, a = nrgba.R, nrgba.G, nrgba.B, nrgba.A
  } else {
    pr, pg, pb, pa := col.RGBA()
    pa >>= 8
    if pa > 0 {
      pr >>= 8
      pr *= 0xff
      pr /= pa
      pg >>= 8
      pg *= 0xff
      pg /= pa
      pb >>= 8
      pb *= 0xff
      pb /= pa
    }
    r = byte(pr)
    g = byte(pg)
    b = byte(pb)
    a = byte(pa)
  }
  return
}

// imageToBytes32 converts the given image into a 32-bit byte array.
func imageToBytes32(img image.Image) []byte {
  w := img.Bounds().Dx()
  h := img.Bounds().Dy()
  retVal := make([]byte, w*h*4)
  dofs := 0
  for y := 0; y < h; y++ {
    for x := 0; x < w; x++ {
      r, g, b, a := img.At(x, y).RGBA()
      retVal[dofs] = byte(r >> 8)
      retVal[dofs+1] = byte(g >> 8)
      retVal[dofs+2] = byte(b >> 8)
      retVal[dofs+3] = byte(a >> 8)
      dofs += 4
    }
  }
  return retVal
}

// bytesToPaletted creates a image.Paletted object out of the given palette and pixel data.
func bytesToPaletted(width, height int, palette color.Palette, pixels []byte) *image.Paletted {
  retVal := image.NewPaletted(image.Rect(0, 0, width, height), palette)
  y0, y1 := retVal.Bounds().Min.Y, retVal.Bounds().Max.Y
  sofs := 0
  dofs := y0 * retVal.Stride
  for y := y0; y < y1; y++ {
    copy(retVal.Pix[dofs:dofs+retVal.Stride], pixels[sofs:sofs+width])
    sofs += width
    dofs += retVal.Stride
  }
  return retVal
}
