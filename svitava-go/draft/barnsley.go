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

// CalcBarnsleyM1 calculates one line of Barnsley M1 Mandelbrot-like set
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

// CalcBarnsleyM2 calculates one line of Barnsley M2 Mandelbrot-like set
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

// CalcBarnsleyM3 calculates one line of Barnsley M3 Mandelbrot-like set
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
