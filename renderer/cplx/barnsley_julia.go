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

// CalcBarnsleyJ1 computes the Barnsley J1 Mandelbrot-like fractal set for a given complex parameter and stores the results in the provided image.
// 
// For each pixel, the function maps coordinates to the complex plane, iterates a conditional transformation up to a maximum number of iterations or until escape, and records the final complex value and iteration count in the image.
func CalcBarnsleyJ1(
	params params.Cplx,
	image deepimage.Image) {

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
				if zx2+zy2 > 4.0 {
					break
				}
				if zx >= 0 {
					zxn = zx*cx - zy*cy - cx
					zyn = zx*cy + zy*cx - cy
				} else {
					zxn = zx*cx - zy*cy + cx
					zyn = zx*cy + zy*cx + cy
				}
				zx = zxn
				zy = zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(i)
			zx0 += 4.0 / float64(image.Resolution.Width)
		}
		zy0 += 4.0 / float64(image.Resolution.Height)
	}
}

// CalcBarnsleyJ2 computes the Barnsley J2 fractal set for the given parameters and stores the results in the provided image.
// 
// For each pixel, it maps coordinates to the complex plane, iterates a conditional transformation based on a linear combination of the current complex values and parameters, and records the final complex value and iteration count in the image.
func CalcBarnsleyJ2(
	params params.Cplx,
	image deepimage.Image) {

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
				if zx2+zy2 > 4.0 {
					break
				}
				if zx*cy+zy*cx >= 0 {
					zxn = zx*cx - zy*cy - cx
					zyn = zx*cy + zy*cx - cy
				} else {
					zxn = zx*cx - zy*cy + cx
					zyn = zx*cy + zy*cx + cy
				}
				zx = zxn
				zy = zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(i)
			zx0 += 4.0 / float64(image.Resolution.Width)
		}
		zy0 += 4.0 / float64(image.Resolution.Height)
	}
}

// CalcBarnsleyJ3 computes the Barnsley J3 fractal set for each pixel in the image grid.
// 
// For each pixel, it maps coordinates to the complex plane, iterates a conditional quadratic transformation up to a maximum iteration count or until escape, and stores the resulting complex value and iteration count in the image arrays. The transformation varies based on the sign of the real part of the complex value.
func CalcBarnsleyJ3(
	params params.Cplx,
	image deepimage.Image) {

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
				if zx2+zy2 > 4.0 {
					break
				}
				if zx > 0 {
					zxn = zx2 - zy2 - 1
					zyn = 2.0 * zx * zy
				} else {
					zxn = zx2 - zy2 - 1 + cx*zx
					zyn = 2.0*zx*zy + cy*zx
				}
				zx = zxn
				zy = zyn
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(complex(zx, zy))
			image.I[y][x] = deepimage.IPixel(i)
			zx0 += 4.0 / float64(image.Resolution.Width)
		}
		zy0 += 4.0 / float64(image.Resolution.Height)
	}
}
