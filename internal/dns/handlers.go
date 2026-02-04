package dns

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Metrics for DNS operations
	dnsRecordsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "caddy_dns_records_created_total",
			Help: "Total number of DNS records created",
		},
		[]string{"provider"},
	)

	dnsRecordsDeleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "caddy_dns_records_deleted_total",
			Help: "Total number of DNS records deleted",
		},
		[]string{"provider"},
	)

	dnsRecordErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "caddy_dns_record_errors_total",
			Help: "Total number of DNS record errors",
		},
		[]string{"provider", "operation"},
	)

	dnsRecordsTracked = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "caddy_dns_records_tracked",
			Help: "Current number of DNS records being tracked",
		},
		[]string{"provider", "state"},
	)
)

func init() {
	// Register metrics with the default registry
	prometheus.MustRegister(dnsRecordsCreated)
	prometheus.MustRegister(dnsRecordsDeleted)
	prometheus.MustRegister(dnsRecordErrors)
	prometheus.MustRegister(dnsRecordsTracked)
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// HealthHandler returns an HTTP handler for the health check endpoint
// GET /health - Returns 200 OK with {"status": "ok"}
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only accept GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := HealthResponse{
			Status: "ok",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
}

// MetricsHandler returns an HTTP handler for the Prometheus metrics endpoint
// GET /metrics - Returns 200 OK with Prometheus-formatted metrics
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// UpdateMetrics updates Prometheus metrics based on current manager state
func (m *Manager) UpdateMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Count records by provider and state
	providerStateCounts := make(map[string]map[RecordState]int)

	for _, record := range m.records {
		if _, ok := providerStateCounts[record.ProviderName]; !ok {
			providerStateCounts[record.ProviderName] = make(map[RecordState]int)
		}
		providerStateCounts[record.ProviderName][record.State]++
	}

	// Update gauge metrics
	for providerName, stateCounts := range providerStateCounts {
		for state, count := range stateCounts {
			dnsRecordsTracked.WithLabelValues(providerName, string(state)).Set(float64(count))
		}
	}

	// Reset gauges for providers with no records
	for providerName := range m.providers {
		if _, ok := providerStateCounts[providerName]; !ok {
			for _, state := range []RecordState{RecordStatePresent, RecordStatePending, RecordStateError} {
				dnsRecordsTracked.WithLabelValues(providerName, string(state)).Set(0)
			}
		}
	}
}

// RecordMetricCreated increments the created counter for a provider
func RecordMetricCreated(providerName string) {
	dnsRecordsCreated.WithLabelValues(providerName).Inc()
}

// RecordMetricDeleted increments the deleted counter for a provider
func RecordMetricDeleted(providerName string) {
	dnsRecordsDeleted.WithLabelValues(providerName).Inc()
}

// RecordMetricError increments the error counter for a provider and operation
func RecordMetricError(providerName, operation string) {
	dnsRecordErrors.WithLabelValues(providerName, operation).Inc()
}
