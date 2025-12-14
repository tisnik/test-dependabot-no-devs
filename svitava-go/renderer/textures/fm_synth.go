package textures

import (
	"math"

	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// renderFMPattern fills ZImage structure with a sinusoidal (FM-style) pattern
// by setting each pixel's Iter.
//
// The function treats ZImage data structure as a height-by-width grid (rows =
// height, cols = width) and samples the continuous rectangle defined by (xmin,
// ymin) to (xmax, ymax). For each pixel at image coordinate (x,y) it computes
// a value using a nested sine formula, converts that value to an integer,
// masks it to the low 8 bits, and stores it in the pixel's Iter field. Iter
// values are therefore in the range 0–255 and are intended as indices into a
// palette.
//
// Parameters:
// - zimage: destination image buffer; must have dimensions [height][width].
// - width, height: image dimensions used to compute sampling step sizes.
// - xmin, ymin, xmax, ymax: bounds of the sampled coordinate region.
//
// The function does not return a value; it mutates zimage parameter in place
// instead.
func CalcFMSynth(
	params params.FractalParameter,
	image deepimage.Image) {

	width := image.Resolution.Width
	height := image.Resolution.Height

	stepx, stepy := getSteps(params, image)

	y1 := params.Ymin
	for y := range height {
		x1 := params.Xmin
		for x := range width {
			val := 100.0 + 100.0*math.Sin(float64(x1)/4.0+2*math.Sin(float64(x1)/15.0+float64(y1)/40.0))
			i := int(val) & 255
			image.Z[y][x] = deepimage.ZPixel(complex(x1, y1))
			image.I[y][x] = deepimage.IPixel(i)
			x1 += stepx
		}
		y1 += stepy
	}
}
