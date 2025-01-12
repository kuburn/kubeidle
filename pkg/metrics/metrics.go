package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ScaleDownOperationsTotal tracks the total number of scale down operations
	ScaleDownOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubeidle_scale_down_operations_total",
			Help: "The total number of scale down operations",
		},
		[]string{"resource_type", "namespace", "status"},
	)

	// ResourcesCurrentlyScaledDown tracks the number of resources currently scaled down
	ResourcesCurrentlyScaledDown = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubeidle_resources_scaled_down",
			Help: "The number of resources currently scaled down",
		},
		[]string{"resource_type", "namespace"},
	)

	// LastScaleOperationTimestamp tracks when the last scale operation occurred
	LastScaleOperationTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubeidle_last_scale_operation_timestamp",
			Help: "The timestamp of the last scale operation",
		},
		[]string{"resource_type", "namespace", "operation"},
	)

	// ControllerStatus tracks the active state of the controller
	ControllerStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kubeidle_controller_active",
			Help: "The current active state of the controller (1 for active, 0 for inactive)",
		},
	)

	// ReconciliationDuration tracks the duration of reconciliation operations
	ReconciliationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kubeidle_reconciliation_duration_seconds",
			Help:    "The duration of reconciliation operations",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
		[]string{"operation"},
	)
)
