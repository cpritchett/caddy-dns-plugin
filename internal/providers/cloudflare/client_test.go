package cloudflare

import (
	"context"
	"net/netip"
	"strings"
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
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
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

func TestCloudflareAdapterRecordOperations(t *testing.T) {
	// These tests verify the adapter methods exist and have proper signatures
	// Note: Full integration tests with mocked Cloudflare API would be in integration tests
	
	t.Run("adapter methods are callable", func(t *testing.T) {
		adapter := &CloudflareAdapter{
			provider: &cloudflare.Provider{APIToken: "test-token"},
		}
		
		ctx := context.Background()
		records := []libdns.Record{
			libdns.Address{Name: "test", IP: netip.MustParseAddr("1.2.3.4")},
		}
		
		// Verify methods exist and can be called (will fail due to invalid credentials)
		// This is primarily a compile-time check that interfaces are implemented correctly
		_, _ = adapter.AppendRecords(ctx, "example.com", records)
		_, _ = adapter.SetRecords(ctx, "example.com", records)
		_, _ = adapter.DeleteRecords(ctx, "example.com", records)
	})
}

// Helper functions

func ptrInt(i int) *int {
	return &i
}

func ptrBool(b bool) *bool {
	return &b
}
