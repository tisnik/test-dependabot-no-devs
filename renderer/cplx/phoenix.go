//
//  (C) Copyright 2025  Pavel Tisnovsky
//
//  All rights reserved. This program and the accompanying materials
//  are made available under the terms of the Eclipse Public License v1.0
//  which accompanies this distribution, and is available at
//  http://www.eclipse.org/legal/epl-v10.html
//
//  Contributors:
//      Pavel Tisnovsky
//

package cplx

import (
	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcPhoenixM computes the Phoenix Mandelbrot-like fractal set and stores the results in the provided image.
// 
// For each pixel in the image, this function maps its coordinates to the complex plane, iteratively applies the Phoenix Mandelbrot formula, and records the resulting complex value and iteration count in the image's data arrays. The calculation covers the region from -2.0 to 2.0 on both axes of the complex plane.
func CalcPhoenixM(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = cx
			var zy float64 = cy
			var ynx = 0.0
			var yny = 0.0
			var i uint
			for i < params.Maxiter {
				zx2 := zx * zx
				zy2 := zy * zy
				zxn := zx2 - zy2 + cx + cy*ynx
				zyn := 2.0*zx*zy + cy*yny
				if zx2+zy2 > 4.0 {
					break
				}
				ynx = zx
				yny = zy
				zx = zxn
				zy = zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(i)
			cx += 4.0 / float64(image.Resolution.Width)
		}
		cy += 4.0 / float64(image.Resolution.Height)
	}
}
