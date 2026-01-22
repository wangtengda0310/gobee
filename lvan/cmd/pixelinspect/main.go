package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/pkg/pixel"
)

const Version = "0.0.1"

func main() {
	// Set up usage
	pflag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Pixel Inspection Tool v%s\n\n", Version)
		_, _ = fmt.Fprintf(os.Stderr, "Usage:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  pixelinspect <image_path> [options]\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "Arguments:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  image_path    Path to the image file to inspect\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "Options:\n")
		pflag.PrintDefaults()
		_, _ = fmt.Fprintf(os.Stderr, "\nRegion format (JSON):\n")
		_, _ = fmt.Fprintf(os.Stderr, "  Single point:       {\"x\": 100, \"y\": 200}\n")
		_, _ = fmt.Fprintf(os.Stderr, "  Rectangle:          {\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}\n")
		_, _ = fmt.Fprintf(os.Stderr, "  Multiple points:    [{\"x\": 0, \"y\": 0}, {\"x\": 100, \"y\": 200}]\n")
		_, _ = fmt.Fprintf(os.Stderr, "  All pixels:         \"all\" or empty\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "Examples:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  pixelinspect image.png --region '{\"x\": 100, \"y\": 200}'\n")
		_, _ = fmt.Fprintf(os.Stderr, "  pixelinspect image.png --region '{\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}' --format stats\n")
		_, _ = fmt.Fprintf(os.Stderr, "  pixelinspect image.png --region 'all' --format json\n")
	}

	// Define flags
	regionStr := pflag.String("region", "all", "Region to inspect (JSON string or 'all')")
	formatStr := pflag.String("format", "list", "Output format: list, stats, json")
	outputFile := pflag.String("output", "", "Output file (optional, default: stdout)")
	showVersion := pflag.BoolP("version", "v", false, "Show version")
	showHelp := pflag.BoolP("help", "h", false, "Show help")

	pflag.Parse()

	// Handle version
	if *showVersion {
		fmt.Printf("Pixel Inspection Tool v%s\n", Version)
		return
	}

	// Handle help
	if *showHelp {
		pflag.Usage()
		return
	}

	// Get image path from positional arguments
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: image_path is required")
		pflag.Usage()
		os.Exit(1)
	}

	imagePath := args[0]

	// Parse region
	region, err := pixel.ParseRegion(*regionStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing region: %v\n", err)
		os.Exit(1)
	}

	// Load image
	img, err := pixel.LoadImage(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
		os.Exit(1)
	}

	// Determine output format
	var format pixel.OutputFormat
	switch *formatStr {
	case "stats":
		format = pixel.OutputStats
	case "json":
		format = pixel.OutputJSON
	case "list", "":
		format = pixel.OutputList
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s', use 'list', 'stats', or 'json'\n", *formatStr)
		os.Exit(1)
	}

	// Inspect pixels
	result, err := pixel.Inspect(img, region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inspecting pixels: %v\n", err)
		os.Exit(1)
	}

	// Format output
	output := pixel.FormatOutput(result, format)

	// Write to file or stdout
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(output)
	}
}
