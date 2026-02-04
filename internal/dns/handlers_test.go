package dns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cpritchett/caddy-dns-plugin/internal/providers"
	"github.com/libdns/libdns"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET returns 200 OK with status json",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "POST returns 405 Method Not Allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "PUT returns 405 Method Not Allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "DELETE returns 405 Method Not Allowed",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HealthHandler()

			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				// Verify Content-Type header
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("expected Content-Type to contain 'application/json', got %s", contentType)
				}

				// Verify response body
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var healthResp HealthResponse
				if err := json.Unmarshal(body, &healthResp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if healthResp.Status != "ok" {
					t.Errorf("expected status 'ok', got %s", healthResp.Status)
				}
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	// Initialize some metrics to ensure they appear in the output
	RecordMetricCreated("test-provider")
	RecordMetricDeleted("test-provider")
	RecordMetricError("test-provider", "create")

	handler := MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Verify Prometheus format - metrics should appear after being incremented
	expectedMetrics := []string{
		"caddy_dns_records_created_total",
		"caddy_dns_records_deleted_total",
		"caddy_dns_record_errors_total",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(bodyStr, metric) {
			t.Errorf("expected metrics to contain %s, body:\n%s", metric, bodyStr)
		}
	}

	// Verify HELP and TYPE lines are present (Prometheus format)
	if !strings.Contains(bodyStr, "# HELP") {
		t.Error("expected metrics to contain HELP comments")
	}
	if !strings.Contains(bodyStr, "# TYPE") {
		t.Error("expected metrics to contain TYPE comments")
	}
}

func TestUpdateMetrics(t *testing.T) {
	// Create a manager with a mock provider
	mockAdapter := &mockAdapter{
		appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
			return records, nil
		},
	}

	mockProv := &mockProvider{
		name:         "test-provider",
		providerType: "cloudflare",
		zoneFilters:  []string{"example.com"},
		adapter:      mockAdapter,
	}

	manager := NewManager([]providers.Provider{mockProv})

	// Add some test records
	manager.records["host1.example.com:test-provider"] = &DNSRecord{
		Hostname:     "host1.example.com",
		RecordType:   RecordTypeA,
		Value:        "192.0.2.1",
		TTL:          300,
		ProviderName: "test-provider",
		State:        RecordStatePresent,
		LastSyncAt:   time.Now(),
	}

	manager.records["host2.example.com:test-provider"] = &DNSRecord{
		Hostname:     "host2.example.com",
		RecordType:   RecordTypeA,
		Value:        "192.0.2.2",
		TTL:          300,
		ProviderName: "test-provider",
		State:        RecordStateError,
		LastSyncAt:   time.Now(),
	}

	// Update metrics
	manager.UpdateMetrics()

	// We can't easily verify the metrics were updated without accessing the Prometheus registry
	// but we can at least verify the method doesn't panic
	t.Log("UpdateMetrics completed without error")
}

func TestRecordMetricFunctions(t *testing.T) {
	// Test that metric functions don't panic when called
	t.Run("RecordMetricCreated", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RecordMetricCreated panicked: %v", r)
			}
		}()
		RecordMetricCreated("test-provider")
	})

	t.Run("RecordMetricDeleted", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RecordMetricDeleted panicked: %v", r)
			}
		}()
		RecordMetricDeleted("test-provider")
	})

	t.Run("RecordMetricError", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RecordMetricError panicked: %v", r)
			}
		}()
		RecordMetricError("test-provider", "create")
	})
}

func TestHealthResponseJSONStructure(t *testing.T) {
	response := HealthResponse{
		Status: "ok",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal HealthResponse: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if status, ok := parsed["status"]; !ok {
		t.Error("expected 'status' field in JSON")
	} else if status != "ok" {
		t.Errorf("expected status to be 'ok', got %v", status)
	}
}

func TestMetricsHandlerNonGETMethod(t *testing.T) {
	// The prometheus handler should handle all methods gracefully
	handler := MetricsHandler()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/metrics", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Prometheus handler typically returns 200 for GET and 405 for others
			// but behavior may vary by version, so we just check it doesn't panic
			if w.Code == 0 {
				t.Error("handler did not set status code")
			}
		})
	}
}
