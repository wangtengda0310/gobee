package gridsplit

import (
	"image"
	"image/color"
	"math"
)

// DetectGridByEdge detects grid lines using Sobel edge detection.
func DetectGridByEdge(img image.Image, threshold int) (GridInfo, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var gridInfo GridInfo

	// Calculate horizontal edge projection
	hProj := horizontalEdgeProjection(img, width, height, threshold)
	vProj := verticalEdgeProjection(img, width, height, threshold)

	// Find grid lines from projection profiles
	gridInfo.HLines = findLinesFromProjection(hProj, height, threshold)
	gridInfo.VLines = findLinesFromProjection(vProj, width, threshold)

	gridInfo.Rows = len(gridInfo.HLines) - 1
	gridInfo.Cols = len(gridInfo.VLines) - 1

	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return GridInfo{}, ErrGridNotDetected
	}

	return gridInfo, nil
}

// horizontalEdgeProjection computes the horizontal edge strength projection.
func horizontalEdgeProjection(img image.Image, width, height, threshold int) []int {
	proj := make([]int, height)

	// Use Sobel operator for horizontal edge detection
	for y := 1; y < height-1; y++ {
		edgeCount := 0

		for x := 1; x < width-1; x++ {
			// Sobel kernels for horizontal edges (Gy)
			// Gy = [[-1, -2, -1],
			//       [ 0,  0,  0],
			//       [ 1,  2,  1]]

			topLeft := luminance(img.At(x-1, y-1))
			top := luminance(img.At(x, y-1))
			topRight := luminance(img.At(x+1, y-1))

			bottomLeft := luminance(img.At(x-1, y+1))
			bottom := luminance(img.At(x, y+1))
			bottomRight := luminance(img.At(x+1, y+1))

			gy := (bottomLeft + 2*bottom + bottomRight) - (topLeft + 2*top + topRight)

			if abs(gy) > threshold {
				edgeCount++
			}
		}

		proj[y] = edgeCount
	}

	return proj
}

// verticalEdgeProjection computes the vertical edge strength projection.
func verticalEdgeProjection(img image.Image, width, height, threshold int) []int {
	proj := make([]int, width)

	// Use Sobel operator for vertical edge detection
	for x := 1; x < width-1; x++ {
		edgeCount := 0

		for y := 1; y < height-1; y++ {
			// Sobel kernels for vertical edges (Gx)
			// Gx = [[-1, 0, 1],
			//       [-2, 0, 2],
			//       [-1, 0, 1]]

			topLeft := luminance(img.At(x-1, y-1))
			topRight := luminance(img.At(x+1, y-1))

			midLeft := luminance(img.At(x-1, y))
			midRight := luminance(img.At(x+1, y))

			bottomLeft := luminance(img.At(x-1, y+1))
			bottomRight := luminance(img.At(x+1, y+1))

			gx := (topRight + 2*midRight + bottomRight) - (topLeft + 2*midLeft + bottomLeft)

			if abs(gx) > threshold {
				edgeCount++
			}
		}

		proj[x] = edgeCount
	}

	return proj
}

// findLinesFromProjection finds grid line positions from edge projection.
func findLinesFromProjection(projection []int, size, threshold int) []int {
	lines := []int{0}

	// Calculate adaptive threshold based on projection statistics
	mean, stdDev := projectionStats(projection)
	adaptiveThreshold := mean + stdDev

	if adaptiveThreshold < threshold {
		adaptiveThreshold = threshold
	}

	// Smooth the projection to reduce noise
	smoothed := smoothProjection(projection, 5)

	// Find peaks (potential grid lines)
	for i := 10; i < size-10; i++ {
		if smoothed[i] > adaptiveThreshold {
			// Check if this is a local maximum
			isPeak := true
			for j := -5; j <= 5; j++ {
				if j != 0 && i+j >= 0 && i+j < size {
					if smoothed[i+j] >= smoothed[i] {
						isPeak = false
						break
					}
				}
			}

			if isPeak {
				// Ensure minimum spacing between lines
				if len(lines) == 0 || i-lines[len(lines)-1] > 20 {
					lines = append(lines, i)
				}
			}
		}
	}

	lines = append(lines, size)
	return lines
}

// projectionStats calculates mean and standard deviation of projection.
func projectionStats(proj []int) (mean, stdDev int) {
	n := len(proj)
	if n == 0 {
		return 0, 0
	}

	sum := 0
	for _, v := range proj {
		sum += v
	}
	mean = sum / n

 variance := 0
	for _, v := range proj {
		diff := v - mean
		variance += diff * diff
	}
	variance /= n

	stdDev = int(math.Sqrt(float64(variance)))
	return mean, stdDev
}

// smoothProjection applies a moving average filter to smooth the projection.
func smoothProjection(proj []int, windowSize int) []int {
	if windowSize < 1 {
		windowSize = 1
	}

	n := len(proj)
	smoothed := make([]int, n)

	for i := 0; i < n; i++ {
		sum := 0
		count := 0

		for j := -windowSize / 2; j <= windowSize/2; j++ {
			idx := i + j
			if idx >= 0 && idx < n {
				sum += proj[idx]
				count++
			}
		}

		if count > 0 {
			smoothed[i] = sum / count
		} else {
			smoothed[i] = proj[i]
		}
	}

	return smoothed
}

// luminance calculates the perceived brightness of a color.
func luminance(c color.Color) int {
	r, g, b, _ := c.RGBA()
	// Convert to 8-bit
	r = r >> 8
	g = g >> 8
	b = b >> 8

	// Use standard luminance formula: 0.299*R + 0.587*G + 0.114*B
	return int(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

// DetectGridByEdgeAdvanced uses more sophisticated edge detection with
// morphological operations to clean up the detected edges.
func DetectGridByEdgeAdvanced(img image.Image, threshold int) (GridInfo, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Compute gradient magnitude using Sobel
	gradMag := computeGradientMagnitude(img, width, height)

	// Create horizontal and vertical projections
	hProj := make([]int, height)
	vProj := make([]int, width)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if gradMag[idx] > threshold {
				hProj[y]++
				vProj[x]++
			}
		}
	}

	// Apply morphological closing to connect nearby edges
	hProj = morphClose(hProj, 10)
	vProj = morphClose(vProj, 10)

	var gridInfo GridInfo
	gridInfo.HLines = findLinesFromProjection(hProj, height, threshold)
	gridInfo.VLines = findLinesFromProjection(vProj, width, threshold)

	gridInfo.Rows = len(gridInfo.HLines) - 1
	gridInfo.Cols = len(gridInfo.VLines) - 1

	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return GridInfo{}, ErrGridNotDetected
	}

	return gridInfo, nil
}

// computeGradientMagnitude computes the Sobel gradient magnitude for each pixel.
func computeGradientMagnitude(img image.Image, width, height int) []int {
	gradMag := make([]int, width*height)

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// Get 3x3 neighborhood
			var pixels [9]int
			idx := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					pixels[idx] = luminance(img.At(x+dx, y+dy))
					idx++
				}
			}

			// Sobel kernels
			// Gx = [[-1, 0, 1], [-2, 0, 2], [-1, 0, 1]]
			// Gy = [[-1, -2, -1], [0, 0, 0], [1, 2, 1]]

			gx := -pixels[0] + pixels[2] - 2*pixels[3] + 2*pixels[5] - pixels[6] + pixels[8]
			gy := -pixels[0] - 2*pixels[1] - pixels[2] + pixels[6] + 2*pixels[7] + pixels[8]

			magnitude := int(math.Sqrt(float64(gx*gx + gy*gy)))
			gradMag[y*width+x] = magnitude
		}
	}

	return gradMag
}

// morphClose applies morphological closing (dilation followed by erosion).
func morphClose(signal []int, size int) []int {
	// Dilate
	dilated := dilate(signal, size)
	// Erode
	eroded := erode(dilated, size)
	return eroded
}

// dilate applies morphological dilation.
func dilate(signal []int, size int) []int {
	n := len(signal)
	result := make([]int, n)

	for i := 0; i < n; i++ {
		maxVal := signal[i]
		for j := -size / 2; j <= size/2; j++ {
			idx := i + j
			if idx >= 0 && idx < n && signal[idx] > maxVal {
				maxVal = signal[idx]
			}
		}
		result[i] = maxVal
	}

	return result
}

// erode applies morphological erosion.
func erode(signal []int, size int) []int {
	n := len(signal)
	result := make([]int, n)

	for i := 0; i < n; i++ {
		minVal := signal[i]
		for j := -size / 2; j <= size/2; j++ {
			idx := i + j
			if idx >= 0 && idx < n && signal[idx] < minVal {
				minVal = signal[idx]
			}
		}
		result[i] = minVal
	}

	return result
}

// DetectGridByHough uses Hough transform to detect straight lines in the image.
// This is more robust for images with weak or broken grid lines.
func DetectGridByHough(img image.Image, threshold int) (GridInfo, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Compute edges
	edges := computeGradientMagnitude(img, width, height)

	// Detect horizontal lines (using simple projection for H-space)
	hSpace := make([]int, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if edges[idx] > threshold {
				hSpace[y]++
			}
		}
	}

	// Detect vertical lines
	vSpace := make([]int, width)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			idx := y*width + x
			if edges[idx] > threshold {
				vSpace[x]++
			}
		}
	}

	var gridInfo GridInfo
	gridInfo.HLines = findLinesFromProjection(hSpace, height, threshold)
	gridInfo.VLines = findLinesFromProjection(vSpace, width, threshold)

	gridInfo.Rows = len(gridInfo.HLines) - 1
	gridInfo.Cols = len(gridInfo.VLines) - 1

	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return GridInfo{}, ErrGridNotDetected
	}

	return gridInfo, nil
}

// EstimateEdgeThreshold estimates an appropriate edge detection threshold
// by analyzing the gradient magnitude distribution.
func EstimateEdgeThreshold(img image.Image) int {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	gradMag := computeGradientMagnitude(img, width, height)

	// Calculate percentiles
	sorted := make([]int, len(gradMag))
	copy(sorted, gradMag)

	// Simple insertion sort (could be optimized for large images)
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	// Use 75th percentile as threshold
	p75Idx := len(sorted) * 3 / 4
	if p75Idx >= len(sorted) {
		p75Idx = len(sorted) - 1
	}

	threshold := sorted[p75Idx]

	// Ensure minimum threshold
	if threshold < 30 {
		threshold = 30
	}

	return threshold
}

// DetectGridAuto automatically detects the grid without requiring threshold.
// It estimates the threshold and tries both edge and color detection.
func DetectGridAuto(img image.Image) (GridInfo, error) {
	// Try edge detection first with estimated threshold
	threshold := EstimateEdgeThreshold(img)
	gridInfo, err := DetectGridByEdgeAdvanced(img, threshold)
	if err == nil && gridInfo.Rows > 0 && gridInfo.Cols > 0 {
		return gridInfo, nil
	}

	// Fall back to color detection
	lineColor := EstimateLineColor(img)
	gridInfo, err = DetectGridByColorAdvanced(img, lineColor, 30)
	if err == nil && gridInfo.Rows > 0 && gridInfo.Cols > 0 {
		return gridInfo, nil
	}

	// Try basic edge detection with lower threshold
	gridInfo, err = DetectGridByEdge(img, threshold/2)
	if err == nil {
		return gridInfo, nil
	}

	return GridInfo{}, ErrGridNotDetected
}
