package internal

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var coverFunc = &goToolsCoverFunc{}

type goToolsCoverFunc struct{}

func AnalyseGoToolsCover() *JSONData {
	coverageData := make(map[string]interface{})

	totalCoverageNumber, m := coverFunc.collectFilesCoverage()

	coverageData["总覆盖率"] = totalCoverageNumber
	coverageData["覆盖率达到100%的文件数"] = 0
	coverageData["新达到100%覆盖率的文件"] = make([]string, 0)
	coverageData["缺少测试的文件数"] = 0
	coverageData["未达到100%覆盖率的文件数"] = 0

	for file, coverage := range m {
		if coverage == 100 {
			coverageData["覆盖率达到100%的文件数"] = coverageData["覆盖率达到100%的文件数"].(int) + 1
			files := coverageData["新达到100%覆盖率的文件"].([]string)
			coverageData["新达到100%覆盖率的文件"] = append(files, file)
		}
		if coverage == 0 {
			coverageData["缺少测试的文件数"] = coverageData["缺少测试的文件数"].(int) + 1
		}
		if coverage > 0 && coverage < 100 {
			coverageData["未达到100%覆盖率的文件数"] = coverageData["未达到100%覆盖率的文件数"].(int) + 1
		}
	}

	if *moreContent != "" {
		coverageData["其他信息"] = *moreContent
	}

	return &JSONData{
		AtAll:   false,
		Title:   *title,
		Content: coverageData,
		Secret:  *secret,
		Token:   *token,
	}
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

	if *showVerbose {
		for file, coverage := range m {
			fmt.Printf("File: %s, Coverage: %.2f%%\n", file, coverage)
		}
	}

	if len(m) > 0 {
		fmt.Printf("Total Coverage: %.2f%%\n", totalCoverage)
	} else {
		fmt.Println("No coverage data found.")
	}
	return totalCoverage, m
}

func funcCoverage(matches []string) (fileName string, funcName string, funcCoverageValue float64) {
	fileAndLine := matches[1]
	fileName = strings.Split(fileAndLine, ":")[0]
	funcName = matches[2]
	funcCoverageValue, _ = strconv.ParseFloat(matches[3], 64)
	return
}
