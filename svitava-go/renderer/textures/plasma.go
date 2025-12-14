package textures

import (
	"math"
	"math/rand/v2"

	"github.com/tisnik/svitava-go/deepimage"
	"github.com/tisnik/svitava-go/params"
)

// randomGauss returns an approximately Gaussian-distributed float32 in [0,1)
// produced by averaging 50 independent uniform(0,1) samples (Central Limit
// Theorem). The result has mean ~0.5 and reduced variance compared to a single
// randomGauss returns an approximately Gaussian-distributed float32 in the range [0,1).
// The result is produced by averaging a fixed number of independent uniform samples.
func randomGauss() float32 {
	const N = 50
	sum := float32(0)
	for range N {
		sum += rand.Float32()
	}
	return sum / N
}

// spectralSynthesis fills img (width × height) with a fractal/plasma-like field using
// spectral synthesis.
//
// The function generates two small frequency-domain spectra (A and B) of size
// (n/2) × (n/2) whose amplitudes follow a power-law ~ k^{-beta/2} with beta = 2*h+1
// and randomized Gaussian-modulated phases. It then synthesizes the spatial field by
// summing cosine and sine basis functions over those frequencies and writes the
// resulting float values into image in-place.
//
// Parameters:
// - image: destination Image that will be populated;
// - n: size parameter for the spectral grid (the code uses n/2 × n/2 frequency components).
// - h: controls the spectral slope (higher h produces smoother fields).
//
// Notes:
// - The function mutates image directly and does not return a value.
// spectralSynthesis synthesizes a fractal/plasma-like field into the provided image using spectral synthesis.
// It populates the image's R channel in-place by constructing a frequency-domain spectrum of size n and summing sinusoidal basis functions whose amplitudes follow a power-law slope (beta = 2*h + 1).
// The parameter n controls the spectral grid size; choose n so that n/2 matches the intended frequency resolution.
// The parameter h controls the spectral slope (larger h produces smoother fields).
func spectralSynthesis(image deepimage.Image, n uint, h float32) {
	width := image.Resolution.Width
	height := image.Resolution.Height

	A := deepimage.NewRImage(image.Resolution)
	B := deepimage.NewRImage(image.Resolution)
	beta := float64(2.0*h + 1)

	for j := range n / 2 {
		for i := range n / 2 {
			rad_i := math.Pow(float64(i)+1.0, -beta/2.0) * float64(randomGauss())
			rad_j := math.Pow(float64(j)+1.0, -beta/2.0) * float64(randomGauss())
			phase_i := 2.0 * math.Pi * rand.Float64()
			phase_j := 2.0 * math.Pi * rand.Float64()
			A[j][i] = deepimage.RPixel(rad_i * math.Cos(phase_i) * rad_j * math.Cos(phase_j))
			B[j][i] = deepimage.RPixel(rad_i * math.Sin(phase_i) * rad_j * math.Sin(phase_j))
		}
	}

	for j := range height {
		for i := range width {
			z := 0.0
			for k := range n / 2 {
				for l := range n / 2 {
					u := float64(i) * 2.0 * math.Pi / float64(width)
					v := float64(j) * 2.0 * math.Pi / float64(height)
					w := float64(k)*u + float64(l)*v
					z += float64(A[k][l])*math.Cos(w) +
						float64(B[k][l])*math.Sin(w)
				}
			}
			image.R[j][i] = deepimage.RPixel(z)
		}
	}
}

// CalcPlasmaPattern generates a plasma (fractal-like) texture in the provided image.
// It populates the image using spectral synthesis with fixed parameters (n=6, h=0.6)
// and then converts the internal floating-point R image to an integer image representation.
// The supplied params.FractalParameter is currently ignored; the function mutates image in place.
func CalcPlasmaPattern(
	params params.FractalParameter,
	image deepimage.Image) {

	spectralSynthesis(image, 6, 0.6)
	image.RImage2IImage()
}