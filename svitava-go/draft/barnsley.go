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

// CalcBarnsleyM1 computes a single horizontal row of the Barnsley M1 Mandelbrot-like fractal at imaginary coordinate cy and writes the result into zimageLine.
// Each entry in zimageLine is set to a ZPixel containing the iteration count and final complex value for the corresponding column; when processing finishes the function sends true on done.
// The caller must provide zimageLine with length at least width.
func CalcBarnsleyM1(width uint, height uint, maxiter uint, zimageLine []ZPixel, cy float64, done chan bool) {
	var cx float64 = -2.0
	for x := uint(0); x < width; x++ {
		var zx float64 = cx
		var zy float64 = cy
		var i uint
		for i < maxiter {
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
		zimageLine[x] = ZPixel{Iter: i, Z: complex(zx, zy)}
		cx += 4.0 / float64(width)
	}
	done <- true
}

// CalcBarnsleyM2 computes a single horizontal line of the Barnsley M2 fractal and stores per-pixel results in zimageLine.
// The cy parameter is the imaginary coordinate for the line; zimageLine is populated with ZPixel values (Iter and Z) for each column.
// When processing completes the function signals completion by sending true on the done channel.
func CalcBarnsleyM2(width uint, height uint, maxiter uint, zimageLine []ZPixel, cy float64, done chan bool) {
	var cx float64 = -2.0
	for x := uint(0); x < width; x++ {
		var zx float64 = cx
		var zy float64 = cy
		var i uint
		for i < maxiter {
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
		zimageLine[x] = ZPixel{Iter: i, Z: complex(zx, zy)}
		cx += 4.0 / float64(width)
	}
	done <- true
}

// CalcBarnsleyM3 calculates one horizontal line of a Barnsley M3 Mandelbrot-like fractal.
// It computes values for each column at the given imaginary coordinate cy, writing a ZPixel
// with the final complex value and the iteration count into zimageLine for each x, and signals
// completion by sending true on done. Iteration for each pixel starts with zx=cx (cx begins
// at -2.0 and increments by 4.0/width) and repeats until zx*zx+zy*zy > 4.0 or i reaches maxiter;
// when zx > 0 the update is z' = (zx^2 - zy^2 - 1, 2*zx*zy), otherwise the update includes
// additive cx*zx and cy*zx terms.
func CalcBarnsleyM3(width uint, height uint, maxiter uint, zimageLine []ZPixel, cy float64, done chan bool) {
	var cx float64 = -2.0
	for x := uint(0); x < width; x++ {
		var zx float64 = cx
		var zy float64 = cy
		var i uint
		for i < maxiter {
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
		zimageLine[x] = ZPixel{Iter: i, Z: complex(zx, zy)}
		cx += 4.0 / float64(width)
	}
	done <- true
}