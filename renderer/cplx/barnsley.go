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

// CalcBarnsleyM1 computes the Barnsley M1 Mandelbrot-like fractal set over a 2D grid.
// 
// For each pixel in the image, it maps the pixel to a complex coordinate in the range [-2, 2] and iteratively applies a conditional affine transformation based on the real part of the current value. Iteration stops when the magnitude squared exceeds 4 or the maximum iteration count is reached. The final complex value and iteration count are stored in the image at the corresponding pixel location.
func CalcBarnsleyM1(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = cx
			var zy float64 = cy
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
			cx += 4.0 / float64(image.Resolution.Width)
		}
		cy += 4.0 / float64(image.Resolution.Height)
	}
}

// CalcBarnsleyM2 computes the Barnsley M2 Mandelbrot-like fractal set over a 2D grid and stores the results in the provided image.
// 
// For each pixel, the function maps its coordinates to a complex number in the range [-2, 2], iteratively applies a conditional affine transformation based on the sign of zx*cy + zy*cx, and records the final complex value and iteration count in the image. Iteration stops if the squared magnitude exceeds 4 or the maximum iteration count is reached.
func CalcBarnsleyM2(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = cx
			var zy float64 = cy
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
			cx += 4.0 / float64(image.Resolution.Width)
		}
		cy += 4.0 / float64(image.Resolution.Height)
	}
}

// CalcBarnsleyM3 computes the Barnsley M3 Mandelbrot-like fractal set and stores the results in the provided image.
// 
// For each pixel, it maps the coordinates to the complex plane, iterates a transformation based on the sign of the real part, and records the final complex value and iteration count. The process covers the region [-2, 2] for both real and imaginary axes.
func CalcBarnsleyM3(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -2.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -2.0
		for x := uint(0); x < image.Resolution.Width; x++ {
			var zx float64 = cx
			var zy float64 = cy
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
			cx += 4.0 / float64(image.Resolution.Width)
		}
		cy += 4.0 / float64(image.Resolution.Height)
	}
}
