# go-imagequant
[![GoDoc](https://godoc.org/github.com/InfinityTools/go-imagequant?status.svg)](https://godoc.org/github.com/InfinityTools/go-imagequant)

## About

[libimagequant](https://github.com/ImageOptim/libimagequant/) is a small, portable C library for high-quality conversion of RGBA images to 8-bit indexed-color (palette) images.

This project provides [Go](https://golang.org/) bindings for *libimagequant*. The bindings were adapted to be closer to Go code conventions. As a result, C memory management details were reduced to a single, optional, Release() call, and pixel data can be transferred by using Go's Image interface.

## Building

go-imagequant uses the system version of the static *libimagequant* library. More information about how to build libimagequant can be found on the [libimagequant project page](https://github.com/ImageOptim/libimagequant/).

go-imagequant package path is currently `github.com/InfinityTools/imagequant`. The bindings can be built via `go build`.

This package makes use of CGO, which requires a decent C compiler to be installed. However, using `go install` removes the C compiler requirement for future invocations of `go build`.

## Overview

The basic flow is:
1. Create an Attributes structure.
2. Create an Image structure from a Go `image.Image` object or a Histogram structure.
3. Perform quantization from the Image or Histogram structure (generate palette).
4. Request a paletted Go Image object.

Basic example (full code can be found in `example/example.go`):
```
package main

import "github.com/InfinityTools/go-imagequant"

func main() {
  att := imagequant.CreateAttributes()
  defer att.Release()

  // imgIn is expected to be a valid image.Image object.
  qimg := att.CreateImage(imgIn, 0.0)

  qresult, _ := att.QuantizeImage(qimg)

  // imgOut is a paletted Go image that can be directly exported using Go's image export packages.
  imgOut, _ := att.WriteRemappedImage(qresult, qimg)
}
```

Detailed function descriptions can be found in the respective Go source files.

## Documentation

For docs, see https://godoc.org/github.com/InfinityTools/go-imagequant .

## License

*go-imagequant* is released under the BSD 2-clause license. See LICENSE for more details.

*libimagequant* itself is available under under the *GPL v3 or later* license and a commercial license for non-GPL software. See *libimagequant-license.txt* for more details.
