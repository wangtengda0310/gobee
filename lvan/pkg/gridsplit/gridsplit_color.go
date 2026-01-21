package gridsplit

import (
	"image"
	"image/color"
	"math"
)

// DetectGridByColor detects grid lines by finding horizontal and vertical lines
// of a specific color.
func DetectGridByColor(img image.Image, lineColor color.Color, tolerance int) (GridInfo, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var gridInfo GridInfo

	// Detect horizontal lines
	gridInfo.HLines = detectHorizontalColorLines(img, lineColor, tolerance, height, width)
	// Detect vertical lines
	gridInfo.VLines = detectVerticalColorLines(img, lineColor, tolerance, height, width)

	gridInfo.Rows = len(gridInfo.HLines) - 1
	gridInfo.Cols = len(gridInfo.VLines) - 1

	// Validate results
	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return GridInfo{}, ErrGridNotDetected
	}

	return gridInfo, nil
}

// detectHorizontalColorLines scans for horizontal lines of the target color.
func detectHorizontalColorLines(img image.Image, targetColor color.Color, tolerance, height, width int) []int {
	lines := []int{0} // Start with top edge

	for y := 1; y < height-1; y++ {
		// Check if this row is mostly the target color
		matchCount := 0
		required := width / 2 // At least half the row should match

		for x := 0; x < width; x++ {
			if colorsMatch(img.At(x, y), targetColor, tolerance) {
				matchCount++
			}
		}

		if matchCount >= required {
			// Check if we haven't already added a nearby line
			if len(lines) == 0 || y-lines[len(lines)-1] > 10 {
				lines = append(lines, y)
			}
		}
	}

	lines = append(lines, height) // End with bottom edge
	return lines
}

// detectVerticalColorLines scans for vertical lines of the target color.
func detectVerticalColorLines(img image.Image, targetColor color.Color, tolerance, height, width int) []int {
	lines := []int{0} // Start with left edge

	for x := 1; x < width-1; x++ {
		// Check if this column is mostly the target color
		matchCount := 0
		required := height / 2 // At least half the column should match

		for y := 0; y < height; y++ {
			if colorsMatch(img.At(x, y), targetColor, tolerance) {
				matchCount++
			}
		}

		if matchCount >= required {
			// Check if we haven't already added a nearby line
			if len(lines) == 0 || x-lines[len(lines)-1] > 10 {
				lines = append(lines, x)
			}
		}
	}

	lines = append(lines, width) // End with right edge
	return lines
}

// colorsMatch checks if two colors are within tolerance.
func colorsMatch(c1, c2 color.Color, tolerance int) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	// Convert from 16-bit to 8-bit
	r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
	r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

	// If either is transparent, consider them different
	if a1 < 128 || a2 < 128 {
		return false
	}

	dr := abs(int(r1) - int(r2))
	dg := abs(int(g1) - int(g2))
	db := abs(int(b1) - int(b2))

	return dr <= tolerance && dg <= tolerance && db <= tolerance
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DetectGridByColorAdvanced uses a more sophisticated algorithm to detect grid lines.
// It looks for consistent line patterns rather than individual matching pixels.
func DetectGridByColorAdvanced(img image.Image, lineColor color.Color, tolerance int) (GridInfo, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var gridInfo GridInfo

	// Create intensity maps for horizontal and vertical directions
	hIntensity := make([]int, height)
	vIntensity := make([]int, width)

	for y := 0; y < height; y++ {
		count := 0
		for x := 0; x < width; x++ {
			if colorsMatch(img.At(x, y), lineColor, tolerance) {
				count++
			}
		}
		hIntensity[y] = count
	}

	for x := 0; x < width; x++ {
		count := 0
		for y := 0; y < height; y++ {
			if colorsMatch(img.At(x, y), lineColor, tolerance) {
				count++
			}
		}
		vIntensity[x] = count
	}

	// Find peaks in intensity (grid lines)
	hThreshold := width / 4 // Need at least 25% of width to be a line
	vThreshold := height / 4 // Need at least 25% of height to be a line

	gridInfo.HLines = findPeaks(hIntensity, hThreshold, height)
	gridInfo.VLines = findPeaks(vIntensity, vThreshold, width)

	gridInfo.Rows = len(gridInfo.HLines) - 1
	gridInfo.Cols = len(gridInfo.VLines) - 1

	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return GridInfo{}, ErrGridNotDetected
	}

	return gridInfo, nil
}

// findPeaks finds local maxima in the intensity array above threshold.
func findPeaks(intensity []int, threshold, maxPos int) []int {
	peaks := []int{0}

	inPeak := false
	peakStart := 0

	for i := 1; i < maxPos-1; i++ {
		if intensity[i] >= threshold {
			if !inPeak {
				inPeak = true
				peakStart = i
			}
		} else {
			if inPeak {
				inPeak = false
				// Add the middle of the peak region
				peakPos := (peakStart + i - 1) / 2
				if len(peaks) == 0 || peakPos-peaks[len(peaks)-1] > 10 {
					peaks = append(peaks, peakPos)
				}
			}
		}
	}

	peaks = append(peaks, maxPos)
	return peaks
}

// EstimateLineColor estimates the dominant line color by analyzing edges.
func EstimateLineColor(img image.Image) color.Color {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Sample some horizontal and vertical lines
	var rSum, gSum, bSum uint64
	count := 0

	// Sample horizontal lines at 1/4 and 3/4 positions
	for _, y := range []int{height / 4, 3 * height / 4} {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			rSum += uint64(r >> 8)
			gSum += uint64(g >> 8)
			bSum += uint64(b >> 8)
			count++
		}
	}

	// Sample vertical lines at 1/4 and 3/4 positions
	for _, x := range []int{width / 4, 3 * width / 4} {
		for y := 0; y < height; y++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			rSum += uint64(r >> 8)
			gSum += uint64(g >> 8)
			bSum += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return color.White
	}

	return color.NRGBA{
		R: uint8(rSum / uint64(count)),
		G: uint8(gSum / uint64(count)),
		B: uint8(bSum / uint64(count)),
		A: 255,
	}
}

// DetectGridByColorWithEdge combines color detection with edge verification.
// This helps avoid false positives by confirming detected lines have actual edges.
func DetectGridByColorWithEdge(img image.Image, lineColor color.Color, colorTolerance, edgeThreshold int) (GridInfo, error) {
	gridInfo, err := DetectGridByColorAdvanced(img, lineColor, colorTolerance)
	if err != nil {
		return GridInfo{}, err
	}

	// Verify detected lines using edge detection
	verifiedHLines := make([]int, 0, len(gridInfo.HLines))
	verifiedVLines := make([]int, 0, len(gridInfo.VLines))

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Verify horizontal lines
	for _, y := range gridInfo.HLines {
		if y == 0 || y == height {
			verifiedHLines = append(verifiedHLines, y)
			continue
		}
		if hasHorizontalEdgeAt(img, y, edgeThreshold, width, height) {
			verifiedHLines = append(verifiedHLines, y)
		}
	}

	// Verify vertical lines
	for _, x := range gridInfo.VLines {
		if x == 0 || x == width {
			verifiedVLines = append(verifiedVLines, x)
			continue
		}
		if hasVerticalEdgeAt(img, x, edgeThreshold, width, height) {
			verifiedVLines = append(verifiedVLines, x)
		}
	}

	gridInfo.HLines = verifiedHLines
	gridInfo.VLines = verifiedVLines
	gridInfo.Rows = len(verifiedHLines) - 1
	gridInfo.Cols = len(verifiedVLines) - 1

	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return GridInfo{}, ErrGridNotDetected
	}

	return gridInfo, nil
}

// hasHorizontalEdgeAt checks if there's a horizontal edge at the given y position.
func hasHorizontalEdgeAt(img image.Image, y, threshold, width, height int) bool {
	if y <= 0 || y >= height-1 {
		return false
	}

	edgeCount := 0
	sampleStep := max(1, width/100) // Sample at most 100 points

	for x := 0; x < width; x += sampleStep {
		above := img.At(x, y-1)
		below := img.At(x, y+1)

		if colorDistance(above, below) > threshold {
			edgeCount++
		}
	}

	return edgeCount > width/sampleStep/5 // At least 20% of samples should be edges
}

// hasVerticalEdgeAt checks if there's a vertical edge at the given x position.
func hasVerticalEdgeAt(img image.Image, x, threshold, width, height int) bool {
	if x <= 0 || x >= width-1 {
		return false
	}

	edgeCount := 0
	sampleStep := max(1, height/100) // Sample at most 100 points

	for y := 0; y < height; y += sampleStep {
		left := img.At(x-1, y)
		right := img.At(x+1, y)

		if colorDistance(left, right) > threshold {
			edgeCount++
		}
	}

	return edgeCount > height/sampleStep/5 // At least 20% of samples should be edges
}

// colorDistance calculates the Euclidean distance between two colors.
func colorDistance(c1, c2 color.Color) int {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	dr := int(r1) - int(r2)
	dg := int(g1) - int(g2)
	db := int(b1) - int(b2)

	return int(math.Sqrt(float64(dr*dr + dg*dg + db*db)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
