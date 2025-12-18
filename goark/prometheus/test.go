package metrics

import (
	"io"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
)

// 在指标环境下运行函数
func MetricRun(t *testing.T, logfile string, f func(prometheus.Registerer)) {
	r := prometheus.NewRegistry()

	file, e := os.Create(logfile)
	assert.Nil(t, e)
	defer func() { _ = file.Close() }()

	f(r)

	mfs, e := r.Gather()
	assert.Nil(t, e)
	encoder := expfmt.NewEncoder(file, expfmt.FmtText)
	for _, mf := range mfs {
		e := encoder.Encode(mf)
		assert.NoError(t, e)
	}
	assert.Nil(t, file.Sync())
}

// 在指标环境下运行基准测试
func BenchFile(b *testing.B, logfile string, f func(prometheus.Registerer)) {
	file, e := os.Create(logfile)
	assert.Nil(b, e)
	defer func() { _ = file.Close() }()

	Bench(b, file, f)

	assert.Nil(b, file.Sync())
}

// 左指标环境下运行基准测试
func Bench(b *testing.B, file io.Writer, f func(prometheus.Registerer)) {
	r := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = r

	f(r)

	mfs, e := r.Gather()
	assert.Nil(b, e)
	encoder := expfmt.NewEncoder(file, expfmt.FmtText)
	for _, mf := range mfs {
		e := encoder.Encode(mf)
		assert.NoError(b, e)
	}
}
