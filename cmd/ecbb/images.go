package main

import (
	"fmt"
	"image"
	"io"

	_ "image/jpeg"
	_ "image/png"
)

// parseReaderToImage reads from a io.Reader into a decoded image.Image
func parseReaderToImage(reader io.Reader) (*image.Image, error) {
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	// Defense in depth - we never expect to have parse anything other than a PNG
	// or a JPEG so error accordingly if expectations differ from reality.
	if format != "png" && format != "jpeg" {
		return nil, fmt.Errorf(
			"decoded with unsupported format: %q", err.Error())
	}

	return &img, nil
}

// toRGBA converts an image.Image to an image.RGBA
func toRGBA(input image.Image) *image.RGBA {
	width := input.Bounds().Max.X
	height := input.Bounds().Max.Y
	rgba := image.NewRGBA(input.Bounds())
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			in := input.At(x, y)
			rgba.Set(x, y, in)
		}
	}
	return rgba
}
