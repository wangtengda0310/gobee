package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func convertToFileAndObtainUnixAbsolutePath(coverageOutputFile string) (*os.File, string, error) {
	// Convert to File type
	file, err := os.Open(coverageOutputFile)
	if err != nil {
		return nil, "", fmt.Errorf("error opening file: %w", err)
	}

	// Obtain absolute path
	absolutePath, err := filepath.Abs(file.Name())
	if err != nil {
		return nil, "", fmt.Errorf("error getting absolute path: %w", err)
	}

	// Convert to Unix-style path
	unixPath := filepath.ToSlash(absolutePath)

	return file, unixPath, nil
}

// Executes the go tool cover command and returns its output.
func getCommandOutput() ([]byte, error) {
	// Define the coverage output file name
	coverageOutputFile := "coverage.out"

	// Check if additional arguments are provided
	fmt.Println(flag.Args())
	if len(flag.Args()) > 0 {
		coverageOutputFile = flag.Args()[0]
	}
	fmt.Println(coverageOutputFile)

	// Execute go test command to generate coverage.out
	//testCmd := exec.Command("go", "test", "./...", "-coverprofile="+coverageOutputFile)
	//output, err := testCmd.Output()
	//if err != nil {
	//	return nil, fmt.Errorf("error executing go test: %w", err)
	//}
	//fmt.Println(output)

	// Obtain absolute path
	_, absolutePath, err := convertToFileAndObtainUnixAbsolutePath(coverageOutputFile)
	if err != nil {
		return nil, fmt.Errorf("error obtaining absolute path: %w", err)
	}

	// Execute go tool cover command with the absolute path to the coverage file
	cmd := exec.Command("go", "tool", "cover", "-func="+absolutePath)
	fmt.Println(cmd.Dir, cmd.Path, cmd.Args)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing go tool cover: %w", err)
	}

	return output, nil
}

// Calculates the coverage from the command output using a LineProcessor.
