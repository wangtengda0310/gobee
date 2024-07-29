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
	if len(fields) < 2 {
		return parsedLine{fields[0], 0}

	}
	hasQuestionPrefix := strings.HasPrefix(s, "?")
	if hasQuestionPrefix {
		return parsedLine{fields[1], coverage}
	}
	hasOKPrefix := strings.HasPrefix(s, "ok")
	if hasOKPrefix {
		return parsedLine{fields[1], coverage}
	}
	if !hasQuestionPrefix && !hasOKPrefix {
		return parsedLine{fields[0], coverage}
	}

	return parsedLine{fields[1], 0}
}

func sumSubPackageCoverage(coverage []parsedLine, ss ...string) []parsedLine {
	var m = make(map[string]*parsedLine)
	for _, s := range ss {
		line, e := m[s]
		if !e {
			line = &parsedLine{s, 0}
			m[s] = line
		}
		for _, c := range coverage {

			if strings.HasPrefix(c.packagePath, s) {
				line.coverage += c.coverage
			}
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
	var parentDir = make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		p := parseLine(line)
		if "" == p.packagePath {
			continue

		}
		coverage = append(coverage, p)
		split := strings.Split(p.packagePath, "/")
		var s string
		if len(split) > 1 {
			s = split[0] + "/" + split[1]
		} else {
			s = split[0]
		}
		parentDir[s] = "true"
	}
	var parentDirs []string
	for k := range parentDir {
		parentDirs = append(parentDirs, k)

	}
	packageCoverage := sumSubPackageCoverage(coverage, parentDirs...)
	return coverage, packageCoverage
}
