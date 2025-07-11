//
//  (C) Copyright 2024  Pavel Tisnovsky
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

// CalcMagnetJulia computes the Magnet Julia-like fractal set and writes the resulting complex values and iteration counts into the provided image.
// 
// For each pixel in the image, this function iterates a specific complex transformation up to a maximum number of iterations or until escape conditions are met. The final complex value and the number of iterations are stored in the image's pixel arrays. The fractal is mapped over the region [-2, 2] in both real and imaginary axes.
func CalcMagnetJulia(
	params params.Cplx,
	image deepimage.Image) {
	const MIN_VALUE = 1.0 - 100

	cx := params.Cx0
	cy := params.Cy0
	var zy0 float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var zx0 float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = zx0
			var zy float64 = zy0
			var i uint
			for i < params.Maxiter {
				var zxn float64
				var zyn float64
				zx2 := zx * zx
				zy2 := zy * zy
				if zx2+zy2 > 100.0 {
					break
				}
				if ((zx-1.0)*(zx-1.0) + zy*zy) < 0.001 {
					break
				}
				tzx := zx2 - zy2 + cx - 1
				tzy := 2.0*zx*zy + cy
				bzx := 2.0*zx + cx - 2
				bzy := 2.0*zy + cy
				div := bzx*bzx + bzy*bzy
				if div < MIN_VALUE {
					break
				}
				zxn = (tzx*bzx + tzy*bzy) / div
				zyn = (tzy*bzx - tzx*bzy) / div
				zx = (zxn + zyn) * (zxn - zyn)
				zy = 2.0 * zxn * zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(i)
			zx0 += 4.0 / float64(image.Resolution.Width)
		}
		zy0 += 4.0 / float64(image.Resolution.Height)
	}
}
