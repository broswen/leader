package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SuccessfulUpdateCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "successful_update_count",
	}, []string{"id"})
	FailedUpdateCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "failed_update_count",
	}, []string{"id"})
)
