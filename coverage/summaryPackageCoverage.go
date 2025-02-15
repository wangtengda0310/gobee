package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type parsedLine struct {
	packagePath string
	coverage    float64
}

func (l parsedLine) String() string {
	return l.packagePath + " coverage: " + strconv.FormatFloat(l.coverage, 'f', -1, 64)

}

func parseLine(s string) parsedLine {
	// find regex "coverage: (\d+.\d+)% of statements"
	var regex = regexp.MustCompile(`coverage: (\d+.\d+)% of statements`)
	var coverage float64
	if regex.MatchString(s) {
		matches := regex.FindStringSubmatch(s)
		c, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			coverage = c
		}
	}

	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) < 1 {
		return parsedLine{"", 0}
	}

	// Handle different line formats
	var packagePath string
	if strings.HasPrefix(s, "?") || strings.HasPrefix(s, "ok") {
		if len(fields) >= 2 {
			packagePath = fields[1]
		}
	} else {
		packagePath = fields[0]
	}

	return parsedLine{packagePath, coverage}
}

func sumSubPackageCoverage(coverage []parsedLine, ss ...string) []parsedLine {
	var m = make(map[string]*parsedLine)
	for _, s := range ss {
		line, e := m[s]
		if !e {
			line = &parsedLine{s, 0}
			m[s] = line
		}
		var sum float64
		var count float64
		for _, c := range coverage {
			if strings.HasPrefix(c.packagePath, s+"/") && c.coverage > 0 {
				sum += c.coverage
				count++
			}
		}
		if count > 0 {
			line.coverage = sum / count
		}
	}
	var result []parsedLine
	for _, line := range m {
		result = append(result, *line)
	}
	return result
}

func scanStdin(reader io.Reader) ([]parsedLine, []parsedLine) {
	scanner := bufio.NewScanner(reader)
	var coverage []parsedLine
	var parentDir = make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		p := parseLine(line)
		if "" == p.packagePath {
			continue
		}
		coverage = append(coverage, p)
		split := strings.Split(p.packagePath, "/")
		if len(split) > 1 {
			parentDir[split[0]+"/"+split[1]] = true
		}
	}
	var parentDirs []string
	for dir := range parentDir {
		parentDirs = append(parentDirs, dir)
	}
	packageCoverage := sumSubPackageCoverage(coverage, parentDirs...)
	return coverage, packageCoverage
}
