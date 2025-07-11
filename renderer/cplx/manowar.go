package cplx

import (
	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcManowarM calculates Manowar Mandelbrot-like set
func CalcManowarM(
	params params.Cplx,
	image deepimage.Image) {

	var cy float64 = -1.0
	for y := uint(0); y < image.Resolution.Height; y++ {
		var cx float64 = -1.5
		for x := uint(0); x < image.Resolution.Width; x++ {
			var c complex128 = complex(cx, cy)
			var z = c
			var z1 = c
			var i uint
			for i < params.Maxiter {
				zx := real(z)
				zy := imag(z)
				if zx*zx+zy*zy > 4.0 {
					break
				}
				z2 := z*z + z1 + c
				z1 = z
				z = z2
				i++
			}
			image.Z[y][x] = deepimage.ZPixel(z)
			image.I[y][x] = deepimage.IPixel(i)
			cx += 2.0 / float64(image.Resolution.Width)
		}
		cy += 2.0 / float64(image.Resolution.Height)
	}
}
