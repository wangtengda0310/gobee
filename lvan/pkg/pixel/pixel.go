// Package pixel provides pixel inspection functionality for images.
package pixel

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// RegionType defines the type of region to inspect.
type RegionType int

const (
	RegionSinglePoint RegionType = iota
	RegionRectangle
	RegionMultiplePoints
	RegionAll
)

// Region defines an area or points to inspect.
type Region struct {
	Type   RegionType
	X      int
	Y      int
	Width  int
	Height int
	Points []Point
}

// Point represents a single pixel coordinate.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// PixelInfo contains color information for a pixel.
type PixelInfo struct {
	X   int    `json:"x"`
	Y   int    `json:"y"`
	R   uint8  `json:"r"`
	G   uint8  `json:"g"`
	B   uint8  `json:"b"`
	A   uint8  `json:"a"`
	Hex string `json:"hex"`
	HSL HSL    `json:"hsl,omitempty"`
}

// HSL represents a color in HSL color space.
type HSL struct {
	H float64 `json:"h"` // Hue 0-360
	S float64 `json:"s"` // Saturation 0-100
	L float64 `json:"l"` // Lightness 0-100
}

// RegionStats contains statistics for a region.
type RegionStats struct {
	Count       int     `json:"count"`
	AvgR        float64 `json:"avg_r"`
	AvgG        float64 `json:"avg_g"`
	AvgB        float64 `json:"avg_b"`
	AvgA        float64 `json:"avg_a"`
	MinR        uint8   `json:"min_r"`
	MaxR        uint8   `json:"max_r"`
	MinG        uint8   `json:"min_g"`
	MaxG        uint8   `json:"max_g"`
	MinB        uint8   `json:"min_b"`
	MaxB        uint8   `json:"max_b"`
	MinA        uint8   `json:"min_a"`
	MaxA        uint8   `json:"max_a"`
	AvgHex      string  `json:"avg_hex"`
	DominantHex string  `json:"dominant_hex"`
	StdDevR     float64 `json:"std_dev_r"`
	StdDevG     float64 `json:"std_dev_g"`
	StdDevB     float64 `json:"std_dev_b"`
}

// OutputFormat defines how to format the output.
type OutputFormat int

const (
	OutputList OutputFormat = iota
	OutputStats
	OutputJSON
)

// ParseRegion parses a region definition from JSON string.
func ParseRegion(regionStr string) (*Region, error) {
	if regionStr == "" || regionStr == "all" {
		return &Region{Type: RegionAll}, nil
	}

	// Try to parse as rectangle first (most specific)
	var rect struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.Unmarshal([]byte(regionStr), &rect); err == nil {
		if rect.Width > 0 && rect.Height > 0 {
			return &Region{
				Type:   RegionRectangle,
				X:      rect.X,
				Y:      rect.Y,
				Width:  rect.Width,
				Height: rect.Height,
			}, nil
		}
	}

	// Try to parse as multiple points
	var points []Point
	if err := json.Unmarshal([]byte(regionStr), &points); err == nil && len(points) > 0 {
		return &Region{
			Type:   RegionMultiplePoints,
			Points: points,
		}, nil
	}

	// Try to parse as single point (least specific)
	var singlePoint Point
	if err := json.Unmarshal([]byte(regionStr), &singlePoint); err == nil {
		return &Region{
			Type: RegionSinglePoint,
			X:    singlePoint.X,
			Y:    singlePoint.Y,
		}, nil
	}

	return nil, fmt.Errorf("invalid region format: %s", regionStr)
}

// InspectImage inspects pixels in an image according to the region definition.
func Inspect(img image.Image, region *Region) (interface{}, error) {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	switch region.Type {
	case RegionSinglePoint:
		return inspectSinglePoint(img, region.X, region.Y, imgWidth, imgHeight)

	case RegionRectangle:
		return inspectRectangle(img, region.X, region.Y, region.Width, region.Height, imgWidth, imgHeight)

	case RegionMultiplePoints:
		return inspectMultiplePoints(img, region.Points, imgWidth, imgHeight)

	case RegionAll:
		return inspectAll(img, imgWidth, imgHeight)

	default:
		return nil, fmt.Errorf("invalid region type")
	}
}

// inspectSinglePoint inspects a single pixel.
func inspectSinglePoint(img image.Image, x, y, imgWidth, imgHeight int) (*PixelInfo, error) {
	if x < 0 || x >= imgWidth || y < 0 || y >= imgHeight {
		return nil, fmt.Errorf("coordinate (%d, %d) out of bounds (0-%d, 0-%d)", x, y, imgWidth, imgHeight)
	}

	r32, g32, b32, a32 := img.At(x, y).RGBA()
	r, g, b, a := uint8(r32>>8), uint8(g32>>8), uint8(b32>>8), uint8(a32>>8)

	return &PixelInfo{
		X:   x,
		Y:   y,
		R:   r,
		G:   g,
		B:   b,
		A:   a,
		Hex: rgbToHex(r, g, b, a),
		HSL: rgbToHSL(r, g, b),
	}, nil
}

// inspectRectangle inspects a rectangular region.
func inspectRectangle(img image.Image, x, y, width, height, imgWidth, imgHeight int) (*RegionStats, error) {
	// Validate bounds
	if x < 0 || y < 0 || x >= imgWidth || y >= imgHeight {
		return nil, fmt.Errorf("region start (%d, %d) out of bounds", x, y)
	}

	if x+width > imgWidth {
		width = imgWidth - x
	}
	if y+height > imgHeight {
		height = imgHeight - y
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid region size: %dx%d", width, height)
	}

	stats := &RegionStats{
		Count: width * height,
		MinR:  255,
		MinG:  255,
		MinB: 255,
		MinA: 255,
		MaxR: 0,
		MaxG: 0,
		MaxB: 0,
		MaxA: 0,
	}

	var sumR, sumG, sumB, sumA uint64
	var sumSqR, sumSqG, sumSqB uint64

	// Collect pixels from the region
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			r32, g32, b32, a32 := img.At(px, py).RGBA()
			r, g, b, a := uint8(r32), uint8(g32), uint8(b32), uint8(a32)

			sumR += uint64(r)
			sumG += uint64(g)
			sumB += uint64(b)
			sumA += uint64(a)

			sumSqR += uint64(r) * uint64(r)
			sumSqG += uint64(g) * uint64(g)
			sumSqB += uint64(b) * uint64(b)

			if r < stats.MinR {
				stats.MinR = r
			}
			if g < stats.MinG {
				stats.MinG = g
			}
			if b < stats.MinB {
				stats.MinB = b
			}
			if a < stats.MinA {
				stats.MinA = a
			}

			if r > stats.MaxR {
				stats.MaxR = r
			}
			if g > stats.MaxG {
				stats.MaxG = g
			}
			if b > stats.MaxB {
				stats.MaxB = b
			}
			if a > stats.MaxA {
				stats.MaxA = a
			}
		}
	}

	// Calculate averages
	count := float64(stats.Count)
	stats.AvgR = float64(sumR) / count
	stats.AvgG = float64(sumG) / count
	stats.AvgB = float64(sumB) / count
	stats.AvgA = float64(sumA) / count
	stats.AvgHex = rgbToHex(uint8(stats.AvgR), uint8(stats.AvgG), uint8(stats.AvgB), uint8(stats.AvgA))

	// Calculate standard deviation
	meanR := stats.AvgR
	meanG := stats.AvgG
	meanB := stats.AvgB

	var varR, varG, varB float64
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			r32, g32, b32, _ := img.At(px, py).RGBA()
			r := float64(r32 >> 8)
			g := float64(g32 >> 8)
			b := float64(b32 >> 8)
			varR += (r - meanR) * (r - meanR)
			varG += (g - meanG) * (g - meanG)
			varB += (b - meanB) * (b - meanB)
		}
	}

	stats.StdDevR = math.Sqrt(varR / count)
	stats.StdDevG = math.Sqrt(varG / count)
	stats.StdDevB = math.Sqrt(varB / count)

	// Find dominant color
	stats.DominantHex = findDominantColor(img, x, y, width, height)

	return stats, nil
}

// inspectMultiplePoints inspects multiple specific points.
func inspectMultiplePoints(img image.Image, points []Point, imgWidth, imgHeight int) ([]PixelInfo, error) {
	results := make([]PixelInfo, 0, len(points))

	for _, p := range points {
		if p.X < 0 || p.X >= imgWidth || p.Y < 0 || p.Y >= imgHeight {
			results = append(results, PixelInfo{
				X:   p.X,
				Y:   p.Y,
				Hex: "#ERROR:OUT_OF_BOUNDS",
			})
			continue
		}

		pixel, err := inspectSinglePoint(img, p.X, p.Y, imgWidth, imgHeight)
		if err != nil {
			results = append(results, PixelInfo{
				X:   p.X,
				Y:   p.Y,
				Hex: "#ERROR:" + err.Error(),
			})
		} else {
			results = append(results, *pixel)
		}
	}

	return results, nil
}

// inspectAll inspects the entire image.
func inspectAll(img image.Image, imgWidth, imgHeight int) (*RegionStats, error) {
	return inspectRectangle(img, 0, 0, imgWidth, imgHeight, imgWidth, imgHeight)
}

// LoadImage loads an image from file.
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// FormatOutput formats the inspection result according to the output format.
func FormatOutput(result interface{}, format OutputFormat) string {
	switch format {
	case OutputStats:
		if stats, ok := result.(*RegionStats); ok {
			return formatStats(stats)
		}
		return fmt.Sprintf("%v", result)

	case OutputJSON:
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data)

	case OutputList:
		fallthrough
	default:
		return formatListResult(result)
	}
}

// formatListResult formats any result as a list.
func formatListResult(result interface{}) string {
	switch v := result.(type) {
	case *PixelInfo:
		return formatSinglePixel(v)
	case []PixelInfo:
		return formatPixelList(v)
	case *RegionStats:
		return formatStats(v)
	default:
		return fmt.Sprintf("%v", result)
	}
}

// formatSinglePixel formats a single pixel info.
func formatSinglePixel(p *PixelInfo) string {
	var hsl string
	if p.HSL.L > 0 {
		hsl = fmt.Sprintf(" HSL(%.1f, %.1f%%, %.1f%%)", p.HSL.H, p.HSL.S, p.HSL.L)
	}
	return fmt.Sprintf("(%d, %d) RGBA(%d, %d, %d, %d) %s%s",
		p.X, p.Y, p.R, p.G, p.B, p.A, p.Hex, hsl)
}

// formatPixelList formats a list of pixels.
func formatPixelList(pixels []PixelInfo) string {
	var lines []string
	for _, p := range pixels {
		var hsl string
		if p.HSL.L > 0 {
			hsl = fmt.Sprintf(" HSL(%.1f, %.1f%%, %.1f%%)", p.HSL.H, p.HSL.S, p.HSL.L)
		}
		lines = append(lines, fmt.Sprintf("(%d, %d) RGBA(%d, %d, %d, %d) %s%s",
			p.X, p.Y, p.R, p.G, p.B, p.A, p.Hex, hsl))
	}
	return strings.Join(lines, "\n")
}

// formatStats formats region statistics.
func formatStats(s *RegionStats) string {
	lines := []string{
		fmt.Sprintf("Region Statistics (%d pixels):", s.Count),
		fmt.Sprintf("  Average:  RGBA(%.1f, %.1f, %.1f, %.1f) %s", s.AvgR, s.AvgG, s.AvgB, s.AvgA, s.AvgHex),
		fmt.Sprintf("  Min:     RGBA(%d, %d, %d, %d)", s.MinR, s.MinG, s.MinB, s.MinA),
		fmt.Sprintf("  Max:     RGBA(%d, %d, %d, %d)", s.MaxR, s.MaxG, s.MaxB, s.MaxA),
		fmt.Sprintf("  StdDev:  RGB(%.2f, %.2f, %.2f)", s.StdDevR, s.StdDevG, s.StdDevB),
		fmt.Sprintf("  Dominant: %s", s.DominantHex),
	}
	return strings.Join(lines, "\n")
}

// rgbToHex converts RGBA to hex color string.
func rgbToHex(r, g, b, a uint8) string {
	if a == 0 {
		return "#00000000"
	}
	if a == 255 {
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
}

// rgbToHSL converts RGB to HSL color space.
func rgbToHSL(r, g, b uint8) HSL {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	var h, s, l float64

	l = (max + min) / 2.0

	if delta == 0 {
		h = 0
		s = 0
	} else {
		s = delta / (1 - math.Abs(2*l-1))
		if max == rf {
			h = (gf - bf) / delta
		} else if max == gf {
			h = 2 + (bf-rf)/delta
		} else {
			h = 4 + (rf-gf)/delta
		}
	}

	h = h * 60
	if h < 0 {
		h += 360
	}
	s *= 100
	l *= 100

	return HSL{H: h, S: s, L: l}
}

// findDominantColor finds the most common color in a region.
func findDominantColor(img image.Image, x, y, width, height int) string {
	colorCount := make(map[string]int)

	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			r32, g32, b32, a32 := img.At(px, py).RGBA()
			r, g, b, a := uint8(r32>>8), uint8(g32>>8), uint8(b32>>8), uint8(a32>>8)
			hex := rgbToHex(r, g, b, a)
			colorCount[hex]++
		}
	}

	maxCount := 0
	dominant := "#000000"
	for color, count := range colorCount {
		if count > maxCount {
			maxCount = count
			dominant = color
		}
	}
	return dominant
}

// SaveRegionAsImage saves a region as a new image file.
func SaveRegionAsImage(img image.Image, region *Region, outputPath string) error {
	bounds := img.Bounds()

	// Determine the region bounds
	var rect image.Rectangle
	switch region.Type {
	case RegionSinglePoint:
		// Save a 10x10 area around the point
		size := 10
		x1 := region.X - size/2
		if x1 < 0 {
			x1 = 0
		}
		y1 := region.Y - size/2
		if y1 < 0 {
			y1 = 0
		}
		x2 := region.X + size/2
		if x2 > bounds.Dx() {
			x2 = bounds.Dx()
		}
		y2 := region.Y + size/2
		if y2 > bounds.Dy() {
			y2 = bounds.Dy()
		}
		rect = image.Rect(x1, y1, x2, y2)
	case RegionRectangle:
		rect = image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	case RegionMultiplePoints:
		return fmt.Errorf("multiple points region cannot be saved as single image")
	case RegionAll:
		rect = bounds
	}

	// Extract region
	regionImg := image.NewRGBA(rect)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			c := img.At(x, y)
			regionImg.Set(x-rect.Min.X, y-rect.Min.Y, c)
		}
	}

	// Save to file
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == "" {
		ext = ".png"
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if ext == ".png" {
		err = png.Encode(file, regionImg)
	} else {
		return fmt.Errorf("unsupported format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}
