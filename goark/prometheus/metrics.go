package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

func init() {
	http.HandleFunc("/metrics", Exporter())
}

const IgnoreMetric = "x-gforge-core2-metrics-ignore"

// Exporter prometheus指标接口
func Exporter() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
		w.Header().Set(IgnoreMetric, "true")
	}
}

func Register(ms ...VecInitiator) {
	for _, m := range ms {
		prometheus.MustRegister(m.initVec())
	}
}

type opt struct {
	name   string
	help   string
	labels []string
}

var defaultMetrics = newMetrics()

func newMetrics() *pushGateWay {
	ctx, cancel := context.WithCancel(context.Background())
	a := &pushGateWay{
		ctx:    ctx,
		cancel: cancel,
	}
	return a
}

type pushGateWay struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (e *pushGateWay) Init() error {
	return nil
}

func (e *pushGateWay) Start() error {
	config := GetConfig()
	addr := config.PushGateWayAddr
	if addr != "" {
		go func() {
			duration := config.PushTickerDuration
			log.Println("prometheus start push gateway", "addr", addr, "ticker", duration)
			pusher := push.New(addr, config.PushJobName)
			pusher = pusher.Gatherer(prometheus.DefaultGatherer) // 关键：使用默认收集器
			tick := time.NewTicker(duration * time.Second)
			for {
				select {
				case <-e.ctx.Done():
					return
				case <-tick.C:
					if err := pusher.Push(); err != nil {
						log.Println("fail push prometheus metrics", "error", err)
					}
				}
			}
		}()
	}
	return nil
}

func (e *pushGateWay) Stop() error {
	e.cancel()
	return nil
}

func (e *pushGateWay) Close() error {
	e.cancel()
	return nil
}

func (e *pushGateWay) Desc() string {
	return ""
}

func (e *pushGateWay) Name() string {
	return "PushGateWay"
}
