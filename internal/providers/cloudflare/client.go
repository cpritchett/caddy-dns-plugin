package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cpritchett/caddy-dns-plugin/internal/config"
	"github.com/cpritchett/caddy-dns-plugin/internal/providers"
	"github.com/libdns/cloudflare"
	"github.com/libdns/libdns"
)

// CloudflareProvider implements the Provider interface for Cloudflare DNS
type CloudflareProvider struct {
	name        string
	providerType string
	zoneFilters []string
	adapter     *CloudflareAdapter
}

// CloudflareAdapter wraps the libdns Cloudflare provider and implements the Adapter interface
type CloudflareAdapter struct {
	provider *cloudflare.Provider
	ttl      *int
	proxied  *bool
}

// NewCloudflareProvider creates a new Cloudflare provider from configuration
func NewCloudflareProvider(cfg config.ProviderConfig) (*CloudflareProvider, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("cloudflare provider %q requires an API token", cfg.Name)
	}

	libdnsProvider := &cloudflare.Provider{
		APIToken: cfg.Token,
	}

	adapter := &CloudflareAdapter{
		provider: libdnsProvider,
		ttl:      cfg.TTL,
		proxied:  cfg.Proxied,
	}

	return &CloudflareProvider{
		name:         cfg.Name,
		providerType: "cloudflare",
		zoneFilters:  cfg.ZoneFilters,
		adapter:      adapter,
	}, nil
}

// Name returns the provider name
func (p *CloudflareProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *CloudflareProvider) Type() string {
	return p.providerType
}

// ZoneFilters returns the zone filters for this provider
func (p *CloudflareProvider) ZoneFilters() []string {
	return p.zoneFilters
}

// Adapter returns the libdns adapter
func (p *CloudflareProvider) Adapter() providers.Adapter {
	return p.adapter
}

// AppendRecords adds DNS records to Cloudflare
func (a *CloudflareAdapter) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Apply Cloudflare-specific settings (TTL and proxied mode) to records
	enriched := a.enrichRecords(records)
	
	// Call the underlying libdns provider
	created, err := a.provider.AppendRecords(ctx, zone, enriched)
	if err != nil {
		return nil, fmt.Errorf("cloudflare append records: %w", err)
	}
	
	return created, nil
}

// SetRecords updates or creates DNS records in Cloudflare
func (a *CloudflareAdapter) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Apply Cloudflare-specific settings to records
	enriched := a.enrichRecords(records)
	
	// Call the underlying libdns provider
	updated, err := a.provider.SetRecords(ctx, zone, enriched)
	if err != nil {
		return nil, fmt.Errorf("cloudflare set records: %w", err)
	}
	
	return updated, nil
}

// DeleteRecords removes DNS records from Cloudflare
func (a *CloudflareAdapter) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Call the underlying libdns provider
	deleted, err := a.provider.DeleteRecords(ctx, zone, records)
	if err != nil {
		return nil, fmt.Errorf("cloudflare delete records: %w", err)
	}
	
	return deleted, nil
}

// enrichRecords applies Cloudflare-specific settings to records
func (a *CloudflareAdapter) enrichRecords(records []libdns.Record) []libdns.Record {
	enriched := make([]libdns.Record, len(records))
	
	for i, record := range records {
		enriched[i] = a.applySettings(record)
	}
	
	return enriched
}

// applySettings applies TTL and other Cloudflare-specific settings to a record
func (a *CloudflareAdapter) applySettings(record libdns.Record) libdns.Record {
	// Get the underlying RR to check and modify TTL
	rr := record.RR()
	
	// Apply TTL if configured and not already set
	if a.ttl != nil && rr.TTL == 0 {
		rr.TTL = time.Duration(*a.ttl) * time.Second
	}
	
	// Convert back to the original type
	// We need to handle different record types
	switch r := record.(type) {
	case libdns.Address:
		r.TTL = rr.TTL
		return r
	case libdns.CNAME:
		r.TTL = rr.TTL
		return r
	case libdns.TXT:
		r.TTL = rr.TTL
		return r
	case libdns.MX:
		r.TTL = rr.TTL
		return r
	case libdns.NS:
		r.TTL = rr.TTL
		return r
	case libdns.SRV:
		r.TTL = rr.TTL
		return r
	case libdns.CAA:
		r.TTL = rr.TTL
		return r
	default:
		// For unknown types, return as-is
		return record
	}
}

// ValidateRecordType checks if the record type is supported
// Cloudflare supports A, AAAA, CNAME, and many other record types
func ValidateRecordType(recordType string) error {
	supported := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CNAME": true,
		"TXT":   true,
		"MX":    true,
		"NS":    true,
		"SRV":   true,
		"CAA":   true,
	}
	
	if !supported[recordType] {
		return fmt.Errorf("unsupported record type %q for Cloudflare provider", recordType)
	}
	
	return nil
}
