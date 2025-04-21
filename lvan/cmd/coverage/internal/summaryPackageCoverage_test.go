package internal

import (
	"github.com/stretchr/testify/assert"
	"slices"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	assert.Equal(t, "", parseLine("").packagePath)
	assert.Equal(t, "", parseLine(" ").packagePath)
	assert.Equal(t, "", parseLine("	").packagePath)
	assert.Equal(t, "gforge/common/actor", parseLine("       gforge/common/actor    ").packagePath)
	assert.Equal(t, "gforge/common/actor", parseLine("?       gforge/common/actor     [no test files]").packagePath)
	assert.Equal(t, "gforge/common/app", parseLine("        gforge/common/app               coverage: 0.0% of statements").packagePath)
	assert.Equal(t, "gforge/common/proto", parseLine("ok      gforge/common/proto     0.154s  coverage: 0.0% of statements").packagePath)

	zero := float64(0)
	assert.Equal(t, zero, parseLine("ok      gforge/common   (cached)        coverage: [no statements]").coverage)
	assert.Equal(t, zero, parseLine("?       gforge/common/actor     [no test files]").coverage)
	assert.Equal(t, zero, parseLine("        gforge/common/signal            coverage: 0.0% of statements").coverage)
	assert.Equal(t, 41.1, parseLine("ok      gforge/common/util      (cached)        coverage: 41.1% of statements").coverage)
}

func TestSumSubPackageCoverage(t *testing.T) {
	coverage := []parsedLine{
		{"gforge/common/actor", 0},
		{"gforge/common/app", 0},
		{"gforge/common/proto", 0},
	}
	packageCoverage := sumSubPackageCoverage(coverage, "gforge/common/")
	assert.Equal(t, "gforge/common/", packageCoverage[0].packagePath)
	assert.Equal(t, 0.0, packageCoverage[0].coverage)
}
func TestParse(t *testing.T) {
	var input = `
ok      gforge/common   (cached)        coverage: [no statements]
?       gforge/common/actor     [no test files]
        gforge/common/config            coverage: 0.0% of statements
        gforge/common/app               coverage: 0.0% of statements
        gforge/common/module            coverage: 0.0% of statements
        gforge/common/logger            coverage: 0.0% of statements
        gforge/common/dispatcher                coverage: 0.0% of statements
        gforge/common/event             coverage: 0.0% of statements
        gforge/common/system            coverage: 0.0% of statements
        gforge/common/transport         coverage: 0.0% of statements
        gforge/common/redis             coverage: 0.0% of statements
        gforge/common/signal            coverage: 0.0% of statements
        gforge/common/table_loader              coverage: 0.0% of statements
ok      gforge/common/proto     0.154s  coverage: 0.0% of statements
ok      gforge/common/util      (cached)        coverage: 41.1% of statements
ok      gforge/another/util      (cached)        coverage: 42.0% of statements
    `
	// scan the input
	reader := strings.NewReader(input)
	coverage, packageCoverage := ScanStdin(reader)
	assert.True(t, slices.Contains(coverage, parsedLine{"gforge/common/util", 41.1}))
	assert.True(t, slices.Contains(coverage, parsedLine{"gforge/another/util", 42}))

	assert.True(t, slices.Contains(packageCoverage, parsedLine{"gforge/common", 41.1}))
	assert.True(t, slices.Contains(packageCoverage, parsedLine{"gforge/another", 42}))

}
