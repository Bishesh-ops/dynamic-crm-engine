package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	QueueLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crm_event_bus_queue_length",
		Help: "The current number of events waiting in the bus channel",
	})

	BatchFlushDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "crm_batch_flush_duration_seconds",
		Help:    "Histogram of database batch flush latencies",
		Buckets: prometheus.DefBuckets,
	})

	EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "crm_events_processed_total",
		Help: "Total number of events processed, partitioned by status",
	}, []string{"status"})
)
