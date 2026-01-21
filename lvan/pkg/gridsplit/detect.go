package gridsplit

import (
	"errors"
	"image"
	"image/color"
)

// Common errors for grid detection.
var (
	// ErrGridNotDetected is returned when no grid can be detected in the image.
	ErrGridNotDetected = errors.New("grid not detected in image")
	// ErrInvalidImage is returned when the image is invalid or nil.
	ErrInvalidImage = errors.New("invalid image")
	// ErrUnsupportedFormat is returned when the image format is not supported.
	ErrUnsupportedFormat = errors.New("unsupported image format")
)

// DetectionMethod specifies which detection method to use.
type DetectionMethod int

const (
	// DetectMethodAuto automatically chooses the best detection method.
	DetectMethodAuto DetectionMethod = iota
	// DetectMethodColor uses color-based grid detection.
	DetectMethodColor
	// DetectMethodEdge uses edge-based grid detection.
	DetectMethodEdge
	// DetectMethodHough uses Hough transform for detection.
	DetectMethodHough
)

// DetectionConfig holds configuration for grid detection.
type DetectionConfig struct {
	Method         DetectionMethod
	LineColor      color.Color
	ColorTolerance int
	EdgeThreshold  int
}

// DefaultDetectionConfig returns a default detection configuration.
func DefaultDetectionConfig() DetectionConfig {
	return DetectionConfig{
		Method:         DetectMethodAuto,
		LineColor:      color.White,
		ColorTolerance: 30,
		EdgeThreshold:  50,
	}
}

// DetectGrid detects a grid in the image using the specified method.
func DetectGrid(img image.Image, config DetectionConfig) (GridInfo, error) {
	if img == nil {
		return GridInfo{}, ErrInvalidImage
	}

	switch config.Method {
	case DetectMethodAuto:
		return DetectGridAuto(img)
	case DetectMethodColor:
		if config.LineColor == nil {
			config.LineColor = color.White
		}
		return DetectGridByColorAdvanced(img, config.LineColor, config.ColorTolerance)
	case DetectMethodEdge:
		return DetectGridByEdgeAdvanced(img, config.EdgeThreshold)
	case DetectMethodHough:
		return DetectGridByHough(img, config.EdgeThreshold)
	default:
		return DetectGridAuto(img)
	}
}

// DetectGridSimple is a simplified version that uses default settings.
func DetectGridSimple(img image.Image) (GridInfo, error) {
	return DetectGrid(img, DefaultDetectionConfig())
}

// AnalyzeGrid analyzes the image to provide information about potential grids.
// This can be used to determine if an image contains a grid and what type.
type GridAnalysis struct {
	HasGrid           bool
	DetectedRows      int
	DetectedCols      int
	Confidence        float64
	BestMethod        DetectionMethod
	RecommendedConfig DetectionConfig
}

// AnalyzeImageForGrid analyzes an image to determine if it contains a grid.
func AnalyzeImageForGrid(img image.Image) GridAnalysis {
	analysis := GridAnalysis{
		HasGrid:    false,
		Confidence: 0.0,
	}

	if img == nil {
		return analysis
	}

	// Try different methods and pick the best
	methods := []struct {
		method DetectionMethod
		name   string
	}{
		{DetectMethodEdge, "edge"},
		{DetectMethodColor, "color"},
		{DetectMethodHough, "hough"},
	}

	bestConfidence := 0.0
	var bestGridInfo GridInfo
	var bestMethod DetectionMethod

	for _, m := range methods {
		var gridInfo GridInfo
		var err error

		switch m.method {
		case DetectMethodEdge:
			threshold := EstimateEdgeThreshold(img)
			gridInfo, err = DetectGridByEdgeAdvanced(img, threshold)
		case DetectMethodColor:
			lineColor := EstimateLineColor(img)
			gridInfo, err = DetectGridByColorAdvanced(img, lineColor, 30)
		case DetectMethodHough:
			threshold := EstimateEdgeThreshold(img)
			gridInfo, err = DetectGridByHough(img, threshold)
		}

		if err == nil {
			confidence := calculateGridConfidence(gridInfo, img.Bounds())
			if confidence > bestConfidence {
				bestConfidence = confidence
				bestGridInfo = gridInfo
				bestMethod = m.method
			}
		}
	}

	if bestConfidence > 0.3 {
		analysis.HasGrid = true
		analysis.DetectedRows = bestGridInfo.Rows
		analysis.DetectedCols = bestGridInfo.Cols
		analysis.Confidence = bestConfidence
		analysis.BestMethod = bestMethod
		analysis.RecommendedConfig = DefaultDetectionConfig()
		analysis.RecommendedConfig.Method = bestMethod
	}

	return analysis
}

// calculateGridConfidence calculates a confidence score for detected grid.
func calculateGridConfidence(gridInfo GridInfo, bounds image.Rectangle) float64 {
	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return 0.0
	}

	width := bounds.Dx()
	height := bounds.Dy()

	// Check if cells are roughly equal size
	if len(gridInfo.HLines) < 2 || len(gridInfo.VLines) < 2 {
		return 0.0
	}

	// Calculate average cell dimensions
	var avgHeight, avgWidth float64
	for i := 1; i < len(gridInfo.HLines); i++ {
		h := gridInfo.HLines[i] - gridInfo.HLines[i-1]
		avgHeight += float64(h)
	}
	avgHeight /= float64(len(gridInfo.HLines) - 1)

	for i := 1; i < len(gridInfo.VLines); i++ {
		w := gridInfo.VLines[i] - gridInfo.VLines[i-1]
		avgWidth += float64(w)
	}
	avgWidth /= float64(len(gridInfo.VLines) - 1)

	// Check for regularity (standard deviation of cell sizes)
	heightVariance := 0.0
	widthVariance := 0.0

	for i := 1; i < len(gridInfo.HLines); i++ {
		h := float64(gridInfo.HLines[i] - gridInfo.HLines[i-1])
		diff := h - avgHeight
		heightVariance += diff * diff
	}
	heightVariance /= float64(len(gridInfo.HLines) - 1)

	for i := 1; i < len(gridInfo.VLines); i++ {
		w := float64(gridInfo.VLines[i] - gridInfo.VLines[i-1])
		diff := w - avgWidth
		widthVariance += diff * diff
	}
	widthVariance /= float64(len(gridInfo.VLines) - 1)

	// Lower variance = higher confidence
	regularityScore := 1.0 / (1.0 + (heightVariance+widthVariance)/(avgHeight*avgWidth+avgWidth*avgHeight)*100)

	// Check for reasonable grid size (not too small, not too large)
	sizeScore := 0.0
	if gridInfo.Rows >= 2 && gridInfo.Rows <= 10 && gridInfo.Cols >= 2 && gridInfo.Cols <= 10 {
		sizeScore = 1.0
	} else if gridInfo.Rows >= 2 && gridInfo.Cols >= 2 {
		sizeScore = 0.5
	}

	// Check coverage (grid should cover most of the image)
	coverageScore := 0.0
	if len(gridInfo.HLines) > 0 && len(gridInfo.VLines) > 0 {
		gridWidth := gridInfo.VLines[len(gridInfo.VLines)-1] - gridInfo.VLines[0]
		gridHeight := gridInfo.HLines[len(gridInfo.HLines)-1] - gridInfo.HLines[0]
		coverage := float64(gridWidth*gridHeight) / float64(width*height)
		coverageScore = coverage
		if coverageScore > 1.0 {
			coverageScore = 1.0
		}
	}

	// Combined confidence score
	confidence := (regularityScore*0.4 + sizeScore*0.3 + coverageScore*0.3)

	return confidence
}

// DetectOptimalRowsCols detects the optimal number of rows and columns.
// This is useful when you know there's a grid but not the exact dimensions.
func DetectOptimalRowsCols(img image.Image, maxRows, maxCols int) (rows, cols int, err error) {
	if maxRows <= 0 {
		maxRows = 10
	}
	if maxCols <= 0 {
		maxCols = 10
	}

	analysis := AnalyzeImageForGrid(img)
	if analysis.HasGrid {
		if analysis.DetectedRows > 0 && analysis.DetectedRows <= maxRows {
			rows = analysis.DetectedRows
		}
		if analysis.DetectedCols > 0 && analysis.DetectedCols <= maxCols {
			cols = analysis.DetectedCols
		}

		if rows > 0 && cols > 0 {
			return rows, cols, nil
		}
	}

	// Fallback to image aspect ratio heuristics
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	aspectRatio := float64(width) / float64(height)

	// Common grid patterns
	if aspectRatio > 1.5 {
		// Wide image: likely more columns than rows
		cols = 3
		rows = 2
	} else if aspectRatio < 0.67 {
		// Tall image: likely more rows than columns
		cols = 2
		rows = 3
	} else {
		// Square-ish image: likely equal rows and columns
		cols = 3
		rows = 3
	}

	return rows, cols, nil
}

// ValidateGridInfo checks if the detected grid info is valid.
func ValidateGridInfo(gridInfo GridInfo, imgBounds image.Rectangle) bool {
	if gridInfo.Rows <= 0 || gridInfo.Cols <= 0 {
		return false
	}

	if len(gridInfo.HLines) != gridInfo.Rows+1 || len(gridInfo.VLines) != gridInfo.Cols+1 {
		return false
	}

	width := imgBounds.Dx()
	height := imgBounds.Dy()

	// Check that lines are within bounds and sorted
	for i, y := range gridInfo.HLines {
		if y < 0 || y > height {
			return false
		}
		if i > 0 && y <= gridInfo.HLines[i-1] {
			return false
		}
	}

	for i, x := range gridInfo.VLines {
		if x < 0 || x > width {
			return false
		}
		if i > 0 && x <= gridInfo.VLines[i-1] {
			return false
		}
	}

	// First and last lines should be at edges
	if gridInfo.HLines[0] != 0 || gridInfo.HLines[len(gridInfo.HLines)-1] != height {
		return false
	}
	if gridInfo.VLines[0] != 0 || gridInfo.VLines[len(gridInfo.VLines)-1] != width {
		return false
	}

	return true
}
