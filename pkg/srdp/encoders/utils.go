package encoders

// bgraToNV12 converts RGBA 32-bit data to NV12 semi-planar format.
// NV12 format: Y plane (width*height) + interleaved UV plane (width*height/2)
// TODO: Optimize this function for performance.
// TODO: Verify color accuracy.
func bgraToNV12(width, height int, data []byte) (y, uv []byte) {
	y = make([]byte, width*height)
	uv = make([]byte, width*height/2) // Interleaved U and V

	var pixelIdx int
	var r, g, b float64
	var yy, u, v float64

	for j := range height {
		for i := range width {
			pixelIdx = (j*width + i) * 4

			b = float64(data[pixelIdx+0])
			g = float64(data[pixelIdx+1])
			r = float64(data[pixelIdx+2])

			// Calculate Y
			yy = 0.299*r + 0.587*g + 0.114*b
			y[j*width+i] = byte(yy)

			// Calculate U and V for every 2x2 block (subsample)
			if j%2 == 0 && i%2 == 0 {
				u = -0.169*r - 0.331*g + 0.5*b + 128
				v = 0.5*r - 0.419*g - 0.081*b + 128

				uvIdx := ((j/2)*(width/2) + (i / 2)) * 2
				uv[uvIdx] = byte(u)
				uv[uvIdx+1] = byte(v)
			}
		}
	}

	data = nil

	return
}
