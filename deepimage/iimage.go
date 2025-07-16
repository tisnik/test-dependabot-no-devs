package deepimage

// IImage is representation of raster image consisting of IPixels
type IImage [][]IPixel

// NewZImage constructs new instance of ZImage
func NewIImage(resolution Resolution) IImage {
	iimage := make([][]IPixel, resolution.Height)
	for i := uint(0); i < resolution.Height; i++ {
		iimage[i] = make([]IPixel, resolution.Width)
	}
	return iimage
}
