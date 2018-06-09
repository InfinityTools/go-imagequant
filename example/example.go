package main

import (
  "flag"
  "fmt"
  "image/png"
  "os"

  "github.com/InfinityTools/go-imagequant"
)

func main() {
  ShouldDisplayVersion := flag.Bool("Version", false, "Print library version and exit")
  InFile := flag.String("In", "", "Input PNG filename")
  OutFile := flag.String("Out", "", "Output PNG filename")
  Speed := flag.Int("Speed", 3, "Speed (1 slowest, 10 fastest)")
  Compression := flag.Int("Compression", -3, "Compression level (DefaultCompression = 0, NoCompression = -1, BestSpeed = -2, BestCompression = -3)")

  flag.Parse()
  if len(os.Args) < 2 {
    fmt.Printf("Usage: %s [options]\nOptions:\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(0)
  }

  if *ShouldDisplayVersion {
    fmt.Printf("libimagequant '%s'\n", imagequant.GetVersionString())
    os.Exit(0)
  }

  var cLevel png.CompressionLevel
  switch *Compression {
    case 0:
      cLevel = png.DefaultCompression
    case -1:
      cLevel = png.NoCompression
    case -2:
      cLevel = png.BestSpeed
    case -3:
      cLevel = png.BestCompression
    default:
      cLevel = png.BestCompression
  }

  err := quantizeFile(*InFile, *OutFile, *Speed, cLevel)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }

  os.Exit(0)
}

func quantizeFile(inFile, outFile string, speed int, compression png.CompressionLevel) error {
  fin, err := os.OpenFile(inFile, os.O_RDONLY, 0444)
  if err != nil {
    return fmt.Errorf("os.OpenFile: %s", err.Error())
  }
  defer fin.Close()

  imgIn, err := png.Decode(fin)
  if err != nil {
    return fmt.Errorf("png.Decode: %s", err.Error())
  }

  // This function call initializes the main quantization structure.
  // Call it once before starting the quantization process.
  quant := imagequant.CreateAttributes()
  // It is recommended but not required to release this object after use. Go's garbage collector
  // will eventually release it automatically when it is no longer used.
  defer quant.Release()

  // Add the source image to the quantizer. Second argument "gamma" is left to the default 0.
  qimg := quant.CreateImage(imgIn, 0.0)
  if qimg == nil {
    return fmt.Errorf("quant.CreateImage: Could not initialize image structure")
  }

  // Generate the palette.
  qresult, err := quant.QuantizeImage(qimg)
  if err != nil {
    return fmt.Errorf("quant.QuantizeImage: %s", err.Error())
  }

  // Remap the input image to the palette and return it as paletted image.
  imgOut, err := quant.WriteRemappedImage(qresult, qimg)
  if err != nil {
    return fmt.Errorf("quant.WriteRemappedImage: %s", err.Error())
  }

  fout, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE, 0644)
  if err != nil {
    return fmt.Errorf("os.OpenFile: %s", err.Error())
  }
  defer fout.Close()

  encoder := png.Encoder{CompressionLevel: compression, BufferPool: nil}
  err = encoder.Encode(fout, imgOut)
  if err != nil {
    return fmt.Errorf("png.Encode: %s", err.Error())
  }

  return nil
}
