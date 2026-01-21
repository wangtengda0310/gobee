// Package gridsplit provides image grid splitting functionality.
// It supports multiple splitting modes: equal spacing, color detection, and edge detection.
package gridsplit

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// SplitMode defines the splitting algorithm.
type SplitMode int

const (
	// ModeEqualSpacing splits image into equal-sized grid cells.
	ModeEqualSpacing SplitMode = iota
	// ModeColorDetection detects splitting lines by color.
	ModeColorDetection
	// ModeEdgeDetection detects grid edges using edge detection algorithm.
	ModeEdgeDetection
)

// String returns the string representation of SplitMode.
func (m SplitMode) String() string {
	switch m {
	case ModeEqualSpacing:
		return "equal"
	case ModeColorDetection:
		return "color"
	case ModeEdgeDetection:
		return "edge"
	default:
		return "unknown"
	}
}

// ParseSplitMode parses a string into SplitMode.
func ParseSplitMode(s string) (SplitMode, error) {
	switch strings.ToLower(s) {
	case "equal", "equal-spacing":
		return ModeEqualSpacing, nil
	case "color", "color-detection":
		return ModeColorDetection, nil
	case "edge", "edge-detection":
		return ModeEdgeDetection, nil
	default:
		return ModeEqualSpacing, fmt.Errorf("unknown split mode: %s", s)
	}
}

// GridInfo contains information about the detected grid structure.
type GridInfo struct {
	HLines []int // Horizontal line positions (Y coordinates)
	VLines []int // Vertical line positions (X coordinates)
	Rows   int   // Number of rows detected
	Cols   int   // Number of columns detected
}

// SplitConfig holds configuration for image splitting.
type SplitConfig struct {
	Mode           SplitMode
	Rows           int         // Number of rows (0 = auto detect)
	Cols           int         // Number of columns (0 = auto detect)
	LineColor      color.Color // Line color for color detection mode
	ColorTolerance int         // Color tolerance 0-255 for color detection
	EdgeThreshold  int         // Edge threshold 0-255 for edge detection
	OutputFormat   string      // Output format: "png", "jpg", "jpeg"
	Padding        int         // Padding around each split image in pixels
}

// DefaultConfig returns a default SplitConfig.
func DefaultConfig() SplitConfig {
	return SplitConfig{
		Mode:           ModeEqualSpacing,
		Rows:           3,
		Cols:           3,
		LineColor:      color.White,
		ColorTolerance: 30,
		EdgeThreshold:  50,
		OutputFormat:   "png",
		Padding:        0,
	}
}

// SplitResult contains the result of splitting an image.
type SplitResult struct {
	Images   []image.Image // Split images
	Paths    []string      // Output file paths (if saved)
	Rows     int           // Number of rows
	Cols     int           // Number of columns
	GridInfo GridInfo      // Grid detection information
}

// SplitImage splits an image according to the given configuration.
func SplitImage(img image.Image, config SplitConfig) (*SplitResult, error) {
	if img == nil {
		return nil, errors.New("image is nil")
	}

	switch config.Mode {
	case ModeEqualSpacing:
		return splitEqualSpacing(img, config.Rows, config.Cols, config.Padding)
	case ModeColorDetection:
		return splitByColor(img, config)
	case ModeEdgeDetection:
		return splitByEdge(img, config)
	default:
		return nil, fmt.Errorf("unsupported split mode: %v", config.Mode)
	}
}

// splitEqualSpacing splits image into equal-sized grid cells.
func splitEqualSpacing(img image.Image, rows, cols, padding int) (*SplitResult, error) {
	if rows <= 0 || cols <= 0 {
		return nil, errors.New("rows and cols must be positive")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	cellWidth := width / cols
	cellHeight := height / rows

	// Build grid info
	var gridInfo GridInfo
	gridInfo.Rows = rows
	gridInfo.Cols = cols
	gridInfo.HLines = make([]int, rows+1)
	gridInfo.VLines = make([]int, cols+1)

	for i := 0; i <= rows; i++ {
		gridInfo.HLines[i] = i * cellHeight
	}
	for i := 0; i <= cols; i++ {
		gridInfo.VLines[i] = i * cellWidth
	}

	images := make([]image.Image, 0, rows*cols)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * cellWidth
			y := row * cellHeight

			// Handle last row/column to include remaining pixels
			w := cellWidth
			h := cellHeight
			if col == cols-1 {
				w = width - x
			}
			if row == rows-1 {
				h = height - y
			}

			// Add padding if specified
			if padding > 0 {
				subImg := image.NewRGBA(image.Rect(
					-padding, -padding,
					w+padding, h+padding,
				))
				// Fill with transparent/background color
				for py := -padding; py < h+padding; py++ {
					for px := -padding; px < w+padding; px++ {
						subImg.Set(px+padding, py+padding, color.RGBA{0, 0, 0, 0})
					}
				}
				// Copy image data
				for py := 0; py < h; py++ {
					for px := 0; px < w; px++ {
						if x+px < width && y+py < height {
							c := img.At(x+px, y+py)
							subImg.Set(px+padding, py+padding, c)
						}
					}
				}
				images = append(images, subImg)
			} else {
				subImg := image.NewRGBA(image.Rect(0, 0, w, h))
				for py := 0; py < h; py++ {
					for px := 0; px < w; px++ {
						if x+px < width && y+py < height {
							c := img.At(x+px, y+py)
							subImg.Set(px, py, c)
						}
					}
				}
				images = append(images, subImg)
			}
		}
	}

	return &SplitResult{
		Images:   images,
		Rows:     rows,
		Cols:     cols,
		GridInfo: gridInfo,
	}, nil
}

// splitByColor splits image by detecting color lines.
func splitByColor(img image.Image, config SplitConfig) (*SplitResult, error) {
	gridInfo, err := DetectGridByColorAdvanced(img, config.LineColor, config.ColorTolerance)
	if err != nil {
		return nil, fmt.Errorf("color detection failed: %w", err)
	}

	// Use detected rows/cols if not specified
	rows := config.Rows
	cols := config.Cols
	if rows <= 0 {
		rows = gridInfo.Rows
	}
	if cols <= 0 {
		cols = gridInfo.Cols
	}

	if rows <= 0 || cols <= 0 {
		return nil, errors.New("could not detect grid dimensions")
	}

	images := make([]image.Image, 0, rows*cols)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x1 := gridInfo.VLines[col]
			y1 := gridInfo.HLines[row]
			x2 := gridInfo.VLines[col+1]
			y2 := gridInfo.HLines[row+1]

			w := x2 - x1
			h := y2 - y1

			if w <= 0 || h <= 0 {
				continue
			}

			subImg := image.NewRGBA(image.Rect(0, 0, w, h))
			for py := 0; py < h; py++ {
				for px := 0; px < w; px++ {
					c := img.At(x1+px, y1+py)
					subImg.Set(px, py, c)
				}
			}
			images = append(images, subImg)
		}
	}

	return &SplitResult{
		Images:   images,
		Rows:     rows,
		Cols:     cols,
		GridInfo: gridInfo,
	}, nil
}

// splitByEdge splits image using edge detection.
func splitByEdge(img image.Image, config SplitConfig) (*SplitResult, error) {
	gridInfo, err := DetectGridByEdgeAdvanced(img, config.EdgeThreshold)
	if err != nil {
		return nil, fmt.Errorf("edge detection failed: %w", err)
	}

	// Use detected rows/cols if not specified
	rows := config.Rows
	cols := config.Cols
	if rows <= 0 {
		rows = gridInfo.Rows
	}
	if cols <= 0 {
		cols = gridInfo.Cols
	}

	if rows <= 0 || cols <= 0 {
		return nil, errors.New("could not detect grid dimensions")
	}

	images := make([]image.Image, 0, rows*cols)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x1 := gridInfo.VLines[col]
			y1 := gridInfo.HLines[row]
			x2 := gridInfo.VLines[col+1]
			y2 := gridInfo.HLines[row+1]

			w := x2 - x1
			h := y2 - y1

			if w <= 0 || h <= 0 {
				continue
			}

			subImg := image.NewRGBA(image.Rect(0, 0, w, h))
			for py := 0; py < h; py++ {
				for px := 0; px < w; px++ {
					c := img.At(x1+px, y1+py)
					subImg.Set(px, py, c)
				}
			}
			images = append(images, subImg)
		}
	}

	return &SplitResult{
		Images:   images,
		Rows:     rows,
		Cols:     cols,
		GridInfo: gridInfo,
	}, nil
}

// LoadImage loads an image from file.
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	_ = format // Keep for reference
	return img, nil
}

// SaveImage saves an image to file.
func SaveImage(img image.Image, path string, format string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	switch strings.ToLower(format) {
	case "png":
		err = png.Encode(file, img)
	case "jpg", "jpeg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	default:
		err = png.Encode(file, img)
	}

	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

// SaveResult saves all images from SplitResult to a directory.
func SaveResult(result *SplitResult, outputDir string, format string, baseName string) error {
	for i, img := range result.Images {
		filename := fmt.Sprintf("%s_%02d.%s", baseName, i, format)
		path := filepath.Join(outputDir, filename)

		if err := SaveImage(img, path, format); err != nil {
			return fmt.Errorf("failed to save image %d: %w", i, err)
		}

		if result.Paths == nil {
			result.Paths = make([]string, len(result.Images))
		}
		result.Paths[i] = path
	}
	return nil
}

// DetectGridSize auto-detects grid dimensions from an image.
func DetectGridSize(img image.Image) (rows, cols int, err error) {
	// Common grid sizes for preference
	commonGrids := []struct {
		rows int
		cols int
	}{
		{3, 3}, {2, 2}, {3, 2}, {2, 3}, {4, 4}, {3, 4}, {4, 3},
	}

	// Max reasonable grid size to prevent over-detection
	const maxRows = 10
	const maxCols = 10

	var bestGrid GridInfo
	bestScore := 0.0

	// Try color detection first (more reliable for typical grid images)
	for _, tolerance := range []int{20, 30, 40, 50} {
		gridInfo, err := DetectGridByColorAdvanced(img, color.White, tolerance)
		if err == nil && gridInfo.Rows > 0 && gridInfo.Cols > 0 {
			// Limit to reasonable grid sizes
			if gridInfo.Rows <= maxRows && gridInfo.Cols <= maxCols {
				score := calculateGridScore(gridInfo, commonGrids)
				if score > bestScore {
					bestScore = score
					bestGrid = gridInfo
				}
			}
		}
	}

	// Try edge detection with different thresholds
	for _, threshold := range []int{50, 70, 100, 150} {
		gridInfo, err := DetectGridByEdgeAdvanced(img, threshold)
		if err == nil && gridInfo.Rows > 0 && gridInfo.Cols > 0 {
			// Limit to reasonable grid sizes
			if gridInfo.Rows <= maxRows && gridInfo.Cols <= maxCols {
				score := calculateGridScore(gridInfo, commonGrids)
				if score > bestScore {
					bestScore = score
					bestGrid = gridInfo
				}
			}
		}
	}

	if bestGrid.Rows > 0 && bestGrid.Cols > 0 {
		return bestGrid.Rows, bestGrid.Cols, nil
	}

	return 0, 0, errors.New("could not auto-detect grid size")
}

// calculateGridScore calculates a score for detected grid, favoring common grid sizes.
func calculateGridScore(gridInfo GridInfo, commonGrids []struct {
	rows int
	cols int
}) float64 {
	// Base score from grid size (prefer smaller, simpler grids)
	sizeScore := 1.0 / float64(gridInfo.Rows*gridInfo.Cols+1)

	// Bonus for common grid sizes
	commonBonus := 0.0
	for _, common := range commonGrids {
		if gridInfo.Rows == common.rows && gridInfo.Cols == common.cols {
			commonBonus = 2.0
			break
		}
	}

	// Bonus for square-ish grids
	aspectRatio := float64(gridInfo.Rows) / float64(gridInfo.Cols)
	aspectBonus := 0.0
	if aspectRatio > 0.5 && aspectRatio < 2.0 {
		aspectBonus = 0.5
	}

	return sizeScore + commonBonus + aspectBonus
}

// ProcessDirectory processes all images in a directory.
func ProcessDirectory(dirPath, outputDir string, config SplitConfig) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's an image file
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		inputPath := filepath.Join(dirPath, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ext)

		img, err := LoadImage(inputPath)
		if err != nil {
			fmt.Printf("Warning: failed to load %s: %v\n", entry.Name(), err)
			continue
		}

		result, err := SplitImage(img, config)
		if err != nil {
			fmt.Printf("Warning: failed to split %s: %v\n", entry.Name(), err)
			continue
		}

		subOutputDir := filepath.Join(outputDir, baseName)
		if err := SaveResult(result, subOutputDir, config.OutputFormat, baseName); err != nil {
			fmt.Printf("Warning: failed to save %s: %v\n", entry.Name(), err)
			continue
		}

		fmt.Printf("Processed %s: %d images saved to %s\n", entry.Name(), len(result.Images), subOutputDir)
	}

	return nil
}
