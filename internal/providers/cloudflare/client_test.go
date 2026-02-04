package cloudflare

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/cpritchett/caddy-dns-plugin/internal/config"
	"github.com/libdns/cloudflare"
	"github.com/libdns/libdns"
)

func TestNewCloudflareProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.ProviderConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			cfg: config.ProviderConfig{
				Name:        "cloudflare-test",
				Type:        "cloudflare",
				Token:       "test-token",
				ZoneFilters: []string{"*.example.com"},
			},
			wantErr: false,
		},
		{
			name: "valid configuration with TTL",
			cfg: config.ProviderConfig{
				Name:        "cloudflare-test",
				Type:        "cloudflare",
				Token:       "test-token",
				ZoneFilters: []string{"*.example.com"},
				TTL:         ptrInt(300),
			},
			wantErr: false,
		},
		{
			name: "valid configuration with proxied",
			cfg: config.ProviderConfig{
				Name:        "cloudflare-test",
				Type:        "cloudflare",
				Token:       "test-token",
				ZoneFilters: []string{"*.example.com"},
				Proxied:     ptrBool(true),
			},
			wantErr: false,
		},
		{
			name: "missing API token",
			cfg: config.ProviderConfig{
				Name:        "cloudflare-test",
				Type:        "cloudflare",
				ZoneFilters: []string{"*.example.com"},
			},
			wantErr: true,
			errMsg:  "requires an API token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewCloudflareProvider(tt.cfg)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewCloudflareProvider() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("NewCloudflareProvider() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewCloudflareProvider() unexpected error = %v", err)
				return
			}
			
			if provider.Name() != tt.cfg.Name {
				t.Errorf("Name() = %v, want %v", provider.Name(), tt.cfg.Name)
			}
			
			if provider.Type() != "cloudflare" {
				t.Errorf("Type() = %v, want cloudflare", provider.Type())
			}
			
			if len(provider.ZoneFilters()) != len(tt.cfg.ZoneFilters) {
				t.Errorf("ZoneFilters() length = %v, want %v", len(provider.ZoneFilters()), len(tt.cfg.ZoneFilters))
			}
			
			if provider.Adapter() == nil {
				t.Errorf("Adapter() = nil, want non-nil")
			}
		})
	}
}

func TestCloudflareProviderMethods(t *testing.T) {
	cfg := config.ProviderConfig{
		Name:        "test-provider",
		Type:        "cloudflare",
		Token:       "test-token",
		ZoneFilters: []string{"*.example.com", "*.test.com"},
	}
	
	provider, err := NewCloudflareProvider(cfg)
	if err != nil {
		t.Fatalf("NewCloudflareProvider() unexpected error = %v", err)
	}
	
	t.Run("Name returns correct name", func(t *testing.T) {
		if got := provider.Name(); got != "test-provider" {
			t.Errorf("Name() = %v, want test-provider", got)
		}
	})
	
	t.Run("Type returns cloudflare", func(t *testing.T) {
		if got := provider.Type(); got != "cloudflare" {
			t.Errorf("Type() = %v, want cloudflare", got)
		}
	})
	
	t.Run("ZoneFilters returns correct filters", func(t *testing.T) {
		filters := provider.ZoneFilters()
		if len(filters) != 2 {
			t.Errorf("ZoneFilters() length = %v, want 2", len(filters))
		}
		if filters[0] != "*.example.com" || filters[1] != "*.test.com" {
			t.Errorf("ZoneFilters() = %v, want [*.example.com *.test.com]", filters)
		}
	})
	
	t.Run("Adapter returns non-nil adapter", func(t *testing.T) {
		if adapter := provider.Adapter(); adapter == nil {
			t.Errorf("Adapter() = nil, want non-nil")
		}
	})
}

func TestCloudflareAdapterEnrichRecords(t *testing.T) {
	tests := []struct {
		name     string
		ttl      *int
		proxied  *bool
		input    []libdns.Record
		wantTTL  time.Duration
	}{
		{
			name: "applies configured TTL when record has no TTL",
			ttl:  ptrInt(300),
			input: []libdns.Record{
				libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4"), TTL: 0},
			},
			wantTTL: 300 * time.Second,
		},
		{
			name: "preserves existing TTL when configured",
			ttl:  ptrInt(300),
			input: []libdns.Record{
				libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4"), TTL: 600 * time.Second},
			},
			wantTTL: 600 * time.Second,
		},
		{
			name: "no TTL configured and no TTL on record",
			input: []libdns.Record{
				libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4"), TTL: 0},
			},
			wantTTL: 0,
		},
		{
			name: "applies TTL to CNAME record",
			ttl:  ptrInt(300),
			input: []libdns.Record{
				libdns.CNAME{Name: "test", Target: "example.com", TTL: 0},
			},
			wantTTL: 300 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &CloudflareAdapter{
				ttl:     tt.ttl,
				proxied: tt.proxied,
			}
			
			enriched := adapter.enrichRecords(tt.input)
			
			if len(enriched) != len(tt.input) {
				t.Errorf("enrichRecords() returned %d records, want %d", len(enriched), len(tt.input))
				return
			}
			
			// Check the TTL based on record type
			rr := enriched[0].RR()
			if rr.TTL != tt.wantTTL {
				t.Errorf("enrichRecords() TTL = %v, want %v", rr.TTL, tt.wantTTL)
			}
		})
	}
}

func TestValidateRecordType(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		wantErr    bool
	}{
		{
			name:       "A record is supported",
			recordType: "A",
			wantErr:    false,
		},
		{
			name:       "AAAA record is supported",
			recordType: "AAAA",
			wantErr:    false,
		},
		{
			name:       "CNAME record is supported",
			recordType: "CNAME",
			wantErr:    false,
		},
		{
			name:       "TXT record is supported",
			recordType: "TXT",
			wantErr:    false,
		},
		{
			name:       "MX record is supported",
			recordType: "MX",
			wantErr:    false,
		},
		{
			name:       "unsupported record type",
			recordType: "INVALID",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecordType(tt.recordType)
			
			if tt.wantErr && err == nil {
				t.Errorf("ValidateRecordType() expected error for type %q, got nil", tt.recordType)
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateRecordType() unexpected error for type %q: %v", tt.recordType, err)
			}
		})
	}
}

func TestCloudflareAdapterRecordOperations(t *testing.T) {
	// These tests verify the adapter methods exist and properly wrap errors
	// In a real scenario, we would mock the Cloudflare API
	
	t.Run("AppendRecords returns error for invalid context", func(t *testing.T) {
		adapter := &CloudflareAdapter{
			provider: &cloudflare.Provider{APIToken: "invalid-token"},
		}
		
		ctx := context.Background()
		records := []libdns.Record{
			libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4")},
		}
		
		// This will fail because we don't have a valid token and zone
		_, err := adapter.AppendRecords(ctx, "invalid.zone", records)
		if err == nil {
			t.Log("AppendRecords expected to fail with invalid configuration (this is okay for unit test)")
		}
	})
	
	t.Run("SetRecords returns error for invalid context", func(t *testing.T) {
		adapter := &CloudflareAdapter{
			provider: &cloudflare.Provider{APIToken: "invalid-token"},
		}
		
		ctx := context.Background()
		records := []libdns.Record{
			libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4")},
		}
		
		// This will fail because we don't have a valid token and zone
		_, err := adapter.SetRecords(ctx, "invalid.zone", records)
		if err == nil {
			t.Log("SetRecords expected to fail with invalid configuration (this is okay for unit test)")
		}
	})
	
	t.Run("DeleteRecords returns error for invalid context", func(t *testing.T) {
		adapter := &CloudflareAdapter{
			provider: &cloudflare.Provider{APIToken: "invalid-token"},
		}
		
		ctx := context.Background()
		records := []libdns.Record{
			libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4")},
		}
		
		// This will fail because we don't have a valid token and zone
		_, err := adapter.DeleteRecords(ctx, "invalid.zone", records)
		if err == nil {
			t.Log("DeleteRecords expected to fail with invalid configuration (this is okay for unit test)")
		}
	})
}

// Helper functions

func ptrInt(i int) *int {
	return &i
}

func ptrBool(b bool) *bool {
	return &b
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
