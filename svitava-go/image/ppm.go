//
//  (C) Copyright 2019, 2020  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package image

import (
	"fmt"
	"image"
	"log"
	"os"
)

// PPMImageWriter implements image.Writer interface, it writes into selected PPM format
type PPMImageWriter struct{}

// WritePPMImage writes an image represented by standard image.Image structure into file with PPM format.
func (writer PPMImageWriter) WriteImage(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	fmt.Fprintln(file, "P3")
	fmt.Fprintf(file, "%d %d\n", width, height)
	fmt.Fprintln(file, "255")

	for y := range height {
		for x := range width {
			r, g, b, _ := img.At(x, y).RGBA()
			fmt.Fprintf(file, "%d %d %d\n", r>>8, g>>8, b>>8)
		}
	}
	return nil
}

// NewPPMImageWriter is a constructor for PPM image writer
func NewPPMImageWriter() PPMImageWriter {
	return PPMImageWriter{}
}
