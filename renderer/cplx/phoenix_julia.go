package cplx

import (
	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// CalcPhoenixJ calculates Phoenix Julia-like set
func CalcPhoenixJ(
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
			zx0 += 4.0 / float64(image.Resolution.Width)
		}
		zy0 += 4.0 / float64(image.Resolution.Height)
	}
}
