package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var coverFunc = &goToolsCoverFunc{}

func count100PercentCoverageFiles(c *CoverageData, file string, coverage float64) {
	if coverage == 100 {
		c.Total100PercentFiles++

	}
}
func countNonTestFiles(c *CoverageData, file string, coverage float64) {
	if coverage == 0 {
		c.NonTestFiles++
	}
}
func countNon100PercentCoverageFiles(c *CoverageData, file string, coverage float64) {
	if coverage > 0 && coverage < 100 {
		c.Non100PercentFiles++
	}
}
func collect100PercentCoverageFiles(c *CoverageData, file string, coverage float64) {
	if coverage == 100 {
		c.New100PercentFiles = append(c.New100PercentFiles, file)

	}
}

func analyseGoToolsCover() *JSONData {
	totalCoverage, m := coverFunc.collectFilesCoverage()
	fmt.Println("Total Coverage: ", totalCoverage)

	c := &CoverageData{}

	for file, coverage := range m {
		count100PercentCoverageFiles(c, file, coverage)
		collect100PercentCoverageFiles(c, file, coverage)
		countNonTestFiles(c, file, coverage)
		countNon100PercentCoverageFiles(c, file, coverage)

	}

	return &JSONData{
		AtAll:   false,
		Title:   "新提交代码测试覆盖率统计",
		Content: c,
		Secret:  *secret,
		Token:   *token,
	}

}

type goToolsCoverFunc struct {
}

func (s *goToolsCoverFunc) collectFilesCoverage() (float64, map[string]float64) {
	m := make(map[string]float64)

	output, err := getCommandOutput()
	if err != nil {
		fmt.Println("Error running go tool cover command", err, output)
		return 0, nil
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	coveragePattern := regexp.MustCompile(`^(.*):\s+(.*)\s+(\d+\.\d+)%$`)
	totalCoveragePattern := regexp.MustCompile(`^total:\s+\(statements\)\s+(\d+\.\d+)%$`)
	fileCoverages := make(map[string]float64)
	fileMethodsCount := make(map[string]float64)
	var totalCoverage float64

	for scanner.Scan() {
		line := scanner.Text()
		if totalMatch := totalCoveragePattern.FindStringSubmatch(line); len(totalMatch) == 2 {
			totalCoverage, _ = strconv.ParseFloat(totalMatch[1], 64)
			continue
		}
		matches := coveragePattern.FindStringSubmatch(line)
		if len(matches) == 4 {
			file, _, coverage := funcCoverage(matches)

			fileCoverages[file] += coverage
			fileMethodsCount[file]++
		}
	}

	for file, totalFileCoverage := range fileCoverages {
		if count, ok := fileMethodsCount[file]; ok && count > 0 {
			m[file] = totalFileCoverage / count // Calculate average coverage per file
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return totalCoverage, m
	}

	for file, coverage := range m {
		fmt.Printf("File: %s, Coverage: %.2f%%\n", file, coverage)
	}

	if len(m) > 0 {
		fmt.Printf("Total Coverage: %.2f%%\n", totalCoverage)
	} else {
		fmt.Println("No coverage data found.")
	}
	return totalCoverage, m
}

// go tool cover -func=coverage.out
func funcCoverage(matches []string) (fileName string, funcName string, funcCoverageValue float64) {
	fileAndLine := matches[1]
	fileName = strings.Split(fileAndLine, ":")[0]
	funcName = matches[2]
	funcCoverageValue, _ = strconv.ParseFloat(matches[3], 64)
	return
}
