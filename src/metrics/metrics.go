package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	totalCounter *prometheus.CounterVec
	duration     *prometheus.HistogramVec
	rows         *prometheus.GaugeVec
)

// PushConfig is used to create prometheus pusher.
type PushConfig struct {
	Instance string
	URL      string
	Job      string
	Interval time.Duration
}

func init() {
	totalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "service_total_count",
	}, []string{"name"})
	prometheus.MustRegister(totalCounter)

	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_duration_milliseconds",
			Buckets: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		}, []string{"name"})
	prometheus.MustRegister(duration)

	rows = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "service_rows",
	}, []string{"name"})
	prometheus.MustRegister(rows)
}

// InitPusher creates prometheus pusher and starts goroutine to push metrics.
func InitPusher(ctx context.Context, cfg *PushConfig) {
	pusher := push.New(cfg.URL, cfg.Job).Gatherer(prometheus.DefaultGatherer).Grouping("instance", cfg.Instance)

	// start pusher
	go func() {
		ticker := time.NewTicker(cfg.Interval)

		for {
			select {
			case <-ticker.C:
				if err := pusher.Push(); err != nil {
					fmt.Println("Could not push to Pushgateway:", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ServiceMiddleware executes given service and updates necessary metrics.
func ServiceMiddleware(name string, service func() int) func() int {
	return func() (n int) {
		start := time.Now()
		defer func(start time.Time) {
			v := time.Since(start).Seconds() * 1e3
			duration.WithLabelValues(name).Observe(v)
		}(start)

		n = service()
		rows.WithLabelValues(name).Set(float64(n))
		totalCounter.WithLabelValues(name).Inc()

		return
	}
}
