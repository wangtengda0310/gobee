package metrics

import "github.com/prometheus/client_golang/prometheus"


type VecInitiator interface {
	initVec() prometheus.Collector
}

type Gauge struct {
	opt
	vec *prometheus.GaugeVec
}

func (g *Gauge) WithLabelValues(lvs ...string) prometheus.Gauge {
	return g.vec.WithLabelValues(lvs...)
}

func (g *Gauge) initVec() prometheus.Collector {
	g.vec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: g.name,
			Help: g.help,
		},
		g.labels, // 标签维度
	)
	return g.vec
}

type Counter struct {
	opt
	vec *prometheus.CounterVec
}

func (c *Counter) WithLabelValues(lvs ...string) prometheus.Counter {
	return c.vec.WithLabelValues(lvs...)
}

func (c *Counter) initVec() prometheus.Collector {
	c.vec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: c.name,
		Help: c.help,
	},
	c.labels, // 标签维度
)
return c.vec
}

type Histogram struct {
	opt
	buckets []float64
	vec     *prometheus.HistogramVec
}

func (h *Histogram) WithLabelValues(lvs ...string) prometheus.Observer {
	return h.vec.WithLabelValues(lvs...)
}

func (h *Histogram) initVec() prometheus.Collector {
	h.vec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    h.name,
		Help:    h.help,
		Buckets: h.buckets,
	},
		h.labels, // 标签维度
	)
	return h.vec
}
