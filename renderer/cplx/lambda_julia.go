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

// CalcLambda computes the Julia variant of the Lambda fractal for the specified parameters and stores the results in the provided image.
// 
// For each pixel in the image, the function maps its coordinates to a region of the complex plane, iteratively applies the fractal formula z = c * z * (1 - z), and records both the final complex value and the iteration count at which the escape condition is met. The region spans from -1.0 to 2.0 on the real axis and -1.5 to 1.5 on the imaginary axis.
func CalcLambda(
	params params.Cplx,
	image deepimage.Image) {

	cx := params.Cx0
	cy := params.Cy0
	var c complex128 = complex(cx, cy)

	var zy0 float64 = -1.5
	for y := uint(0); y < image.Resolution.Height; y++ {
		var zx0 float64 = -1.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var z complex128 = complex(zx0, zy0)
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > 4.0 {
					break
				}
				z = c * z * (1 - z)
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			zx0 += 3.0 / float64(image.Resolution.Width)
		}
		zy0 += 3.0 / float64(image.Resolution.Height)
	}
}
