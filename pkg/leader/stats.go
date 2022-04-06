package leader

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HealthyWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "healthy_workers_total",
	})
	UnhealthyWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "unhealthy_workers_total",
	})
	WantedWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wanted_workers",
	})
	SelectedWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "selected_workers",
	})
	UpdateCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_update_count",
	}, []string{"id"})
)
