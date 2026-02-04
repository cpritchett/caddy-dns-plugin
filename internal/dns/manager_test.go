package dns

import (
"context"
"errors"
"net/netip"
"testing"
"time"

"github.com/cpritchett/caddy-dns-plugin/internal/providers"
"github.com/libdns/libdns"
)

// Mock provider for testing
type mockProvider struct {
name        string
providerType string
zoneFilters []string
adapter     *mockAdapter
}

func (m *mockProvider) Name() string {
return m.name
}

func (m *mockProvider) Type() string {
return m.providerType
}

func (m *mockProvider) ZoneFilters() []string {
return m.zoneFilters
}

func (m *mockProvider) Adapter() providers.Adapter {
return m.adapter
}

// Mock adapter for testing
type mockAdapter struct {
appendRecords  func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error)
setRecords     func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error)
deleteRecords  func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error)
}

func (m *mockAdapter) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
if m.appendRecords != nil {
return m.appendRecords(ctx, zone, records)
}
return records, nil
}

func (m *mockAdapter) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
if m.setRecords != nil {
return m.setRecords(ctx, zone, records)
}
return records, nil
}

func (m *mockAdapter) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
if m.deleteRecords != nil {
return m.deleteRecords(ctx, zone, records)
}
return records, nil
}

func TestNewManager(t *testing.T) {
providers := []providers.Provider{
&mockProvider{name: "test-provider", zoneFilters: []string{"example.com"}},
}

manager := NewManager(providers)
if manager == nil {
t.Fatal("NewManager returned nil")
}

if len(manager.providers) != 1 {
t.Fatalf("expected 1 provider, got %d", len(manager.providers))
}
}

func TestComputeDesiredState_SingleContainer(t *testing.T) {
manager := NewManager([]providers.Provider{})

containers := []ContainerInfo{
{
ID:        "container123",
Name:      "web-app",
IsRunning: true,
IPV4:      []string{"192.168.1.10"},
Labels: map[string]string{
"caddy_dns.hostname": "app.example.com",
"caddy_dns.provider": "cloudflare",
},
},
}

requests, err := manager.ComputeDesiredState(containers, "caddy_dns")
if err != nil {
t.Fatalf("ComputeDesiredState failed: %v", err)
}

if len(requests) != 1 {
t.Fatalf("expected 1 request, got %d", len(requests))
}

req := requests[0]
if req.Hostname != "app.example.com" {
t.Errorf("hostname = %q, want %q", req.Hostname, "app.example.com")
}
if req.ProviderName != "cloudflare" {
t.Errorf("provider = %q, want %q", req.ProviderName, "cloudflare")
}
if req.RecordType != RecordTypeA {
t.Errorf("recordType = %q, want %q", req.RecordType, RecordTypeA)
}
if req.Target != "192.168.1.10" {
t.Errorf("target = %q, want %q", req.Target, "192.168.1.10")
}
if req.SourceID != "container123" {
t.Errorf("sourceID = %q, want %q", req.SourceID, "container123")
}
}

func TestComputeDesiredState_InferHostnameFromCaddy(t *testing.T) {
manager := NewManager([]providers.Provider{})

containers := []ContainerInfo{
{
ID:        "container123",
Name:      "web-app",
IsRunning: true,
IPV4:      []string{"192.168.1.10"},
Labels: map[string]string{
"caddy":              "example.com, www.example.com",
"caddy_dns.provider": "cloudflare",
},
},
}

requests, err := manager.ComputeDesiredState(containers, "caddy_dns")
if err != nil {
t.Fatalf("ComputeDesiredState failed: %v", err)
}

if len(requests) != 1 {
t.Fatalf("expected 1 request, got %d", len(requests))
}

if requests[0].Hostname != "example.com" {
t.Errorf("hostname = %q, want %q", requests[0].Hostname, "example.com")
}
}

func TestComputeDesiredState_SkipStoppedContainers(t *testing.T) {
manager := NewManager([]providers.Provider{})

containers := []ContainerInfo{
{
ID:        "container123",
Name:      "web-app",
IsRunning: false,
IPV4:      []string{"192.168.1.10"},
Labels: map[string]string{
"caddy_dns.hostname": "app.example.com",
"caddy_dns.provider": "cloudflare",
},
},
}

requests, err := manager.ComputeDesiredState(containers, "caddy_dns")
if err != nil {
t.Fatalf("ComputeDesiredState failed: %v", err)
}

if len(requests) != 0 {
t.Fatalf("expected 0 requests for stopped container, got %d", len(requests))
}
}

func TestComputeDesiredState_SkipDisabled(t *testing.T) {
manager := NewManager([]providers.Provider{})

containers := []ContainerInfo{
{
ID:        "container123",
Name:      "web-app",
IsRunning: true,
IPV4:      []string{"192.168.1.10"},
Labels: map[string]string{
"caddy_dns.hostname": "app.example.com",
"caddy_dns.provider": "cloudflare",
"caddy_dns.enable":   "false",
},
},
}

requests, err := manager.ComputeDesiredState(containers, "caddy_dns")
if err != nil {
t.Fatalf("ComputeDesiredState failed: %v", err)
}

if len(requests) != 0 {
t.Fatalf("expected 0 requests for disabled container, got %d", len(requests))
}
}

func TestComputeDesiredState_IPv6(t *testing.T) {
manager := NewManager([]providers.Provider{})

containers := []ContainerInfo{
{
ID:        "container123",
Name:      "web-app",
IsRunning: true,
IPV6:      []string{"2001:db8::1"},
Labels: map[string]string{
"caddy_dns.hostname": "app.example.com",
"caddy_dns.provider": "cloudflare",
},
},
}

requests, err := manager.ComputeDesiredState(containers, "caddy_dns")
if err != nil {
t.Fatalf("ComputeDesiredState failed: %v", err)
}

if len(requests) != 1 {
t.Fatalf("expected 1 request, got %d", len(requests))
}

if requests[0].RecordType != RecordTypeAAAA {
t.Errorf("recordType = %q, want %q", requests[0].RecordType, RecordTypeAAAA)
}
if requests[0].Target != "2001:db8::1" {
t.Errorf("target = %q, want %q", requests[0].Target, "2001:db8::1")
}
}

func TestComputeDesiredState_WithTTL(t *testing.T) {
manager := NewManager([]providers.Provider{})

containers := []ContainerInfo{
{
ID:        "container123",
Name:      "web-app",
IsRunning: true,
IPV4:      []string{"192.168.1.10"},
Labels: map[string]string{
"caddy_dns.hostname": "app.example.com",
"caddy_dns.provider": "cloudflare",
"caddy_dns.ttl":      "600",
},
},
}

requests, err := manager.ComputeDesiredState(containers, "caddy_dns")
if err != nil {
t.Fatalf("ComputeDesiredState failed: %v", err)
}

if len(requests) != 1 {
t.Fatalf("expected 1 request, got %d", len(requests))
}

if requests[0].TTL == nil {
t.Fatal("expected TTL to be set")
}
if *requests[0].TTL != 600 {
t.Errorf("TTL = %d, want %d", *requests[0].TTL, 600)
}
}

func TestCreateRecord_Success(t *testing.T) {
appendCalled := false
adapter := &mockAdapter{
appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
appendCalled = true
if zone != "example.com" {
t.Errorf("zone = %q, want %q", zone, "example.com")
}
if len(records) != 1 {
t.Fatalf("expected 1 record, got %d", len(records))
}

rec := records[0]
addr, ok := rec.(libdns.Address)
if !ok {
t.Fatalf("expected Address record, got %T", rec)
}

if addr.Name != "app" {
t.Errorf("name = %q, want %q", addr.Name, "app")
}
expectedIP := netip.MustParseAddr("192.168.1.10")
if addr.IP != expectedIP {
t.Errorf("IP = %v, want %v", addr.IP, expectedIP)
}

return records, nil
},
}

provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     adapter,
}

manager := NewManager([]providers.Provider{provider})

req := SyncRequest{
Hostname:     "app.example.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err != nil {
t.Fatalf("CreateRecord failed: %v", err)
}

if !appendCalled {
t.Error("AppendRecords was not called")
}

// Verify record was stored
records := manager.GetRecords()
if len(records) != 1 {
t.Fatalf("expected 1 stored record, got %d", len(records))
}

stored := records[0]
if stored.State != RecordStatePresent {
t.Errorf("state = %q, want %q", stored.State, RecordStatePresent)
}
}

func TestCreateRecord_ProviderNotFound(t *testing.T) {
manager := NewManager([]providers.Provider{})

req := SyncRequest{
Hostname:     "app.example.com",
ProviderName: "nonexistent",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err == nil {
t.Fatal("expected error for nonexistent provider")
}
}

func TestCreateRecord_HostnameValidationFails(t *testing.T) {
provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     &mockAdapter{},
}

manager := NewManager([]providers.Provider{provider})

req := SyncRequest{
Hostname:     "app.different.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err == nil {
t.Fatal("expected error for hostname not matching zone filters")
}
}

func TestCreateRecord_AdapterError(t *testing.T) {
adapter := &mockAdapter{
appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
return nil, errors.New("provider API error")
},
}

provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     adapter,
}

manager := NewManager([]providers.Provider{provider})

req := SyncRequest{
Hostname:     "app.example.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err == nil {
t.Fatal("expected error from adapter")
}

// Verify record was stored with error state
records := manager.GetRecords()
if len(records) != 1 {
t.Fatalf("expected 1 stored record, got %d", len(records))
}

if records[0].State != RecordStateError {
t.Errorf("state = %q, want %q", records[0].State, RecordStateError)
}
}

func TestDeleteRecord_Success(t *testing.T) {
deleteCalled := false
adapter := &mockAdapter{
appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
return records, nil
},
deleteRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
deleteCalled = true
if zone != "example.com" {
t.Errorf("zone = %q, want %q", zone, "example.com")
}
if len(records) != 1 {
t.Fatalf("expected 1 record, got %d", len(records))
}
return records, nil
},
}

provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     adapter,
}

manager := NewManager([]providers.Provider{provider})

// First create a record
req := SyncRequest{
Hostname:     "app.example.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err != nil {
t.Fatalf("CreateRecord failed: %v", err)
}

// Now delete it
err = manager.DeleteRecord(context.Background(), "app.example.com", "cloudflare", "container123")
if err != nil {
t.Fatalf("DeleteRecord failed: %v", err)
}

if !deleteCalled {
t.Error("DeleteRecords was not called")
}

// Verify record was removed
records := manager.GetRecords()
if len(records) != 0 {
t.Fatalf("expected 0 stored records, got %d", len(records))
}
}

func TestDeleteRecord_NotFound(t *testing.T) {
provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     &mockAdapter{},
}

manager := NewManager([]providers.Provider{provider})

// Try to delete non-existent record (should not error)
err := manager.DeleteRecord(context.Background(), "app.example.com", "cloudflare", "container123")
if err != nil {
t.Fatalf("DeleteRecord failed: %v", err)
}
}

func TestDeleteRecord_WrongContainer(t *testing.T) {
adapter := &mockAdapter{
appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
return records, nil
},
}

provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     adapter,
}

manager := NewManager([]providers.Provider{provider})

// Create a record
req := SyncRequest{
Hostname:     "app.example.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err != nil {
t.Fatalf("CreateRecord failed: %v", err)
}

// Try to delete with different container ID
err = manager.DeleteRecord(context.Background(), "app.example.com", "cloudflare", "different-container")
if err == nil {
t.Fatal("expected error when deleting record from different container")
}

// Verify record still exists
records := manager.GetRecords()
if len(records) != 1 {
t.Fatalf("expected 1 stored record, got %d", len(records))
}
}

func TestDeleteRecordsForContainer(t *testing.T) {
deleteCount := 0
adapter := &mockAdapter{
appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
return records, nil
},
deleteRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
deleteCount++
return records, nil
},
}

provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     adapter,
}

manager := NewManager([]providers.Provider{provider})

// Create multiple records for same container
for _, hostname := range []string{"app1.example.com", "app2.example.com"} {
req := SyncRequest{
Hostname:     hostname,
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container123",
}

err := manager.CreateRecord(context.Background(), req)
if err != nil {
t.Fatalf("CreateRecord failed: %v", err)
}
}

// Verify we have 2 records
records := manager.GetRecords()
if len(records) != 2 {
t.Fatalf("expected 2 records, got %d", len(records))
}

// Delete all records for the container
err := manager.DeleteRecordsForContainer(context.Background(), "container123")
if err != nil {
t.Fatalf("DeleteRecordsForContainer failed: %v", err)
}

if deleteCount != 2 {
t.Errorf("DeleteRecords called %d times, want 2", deleteCount)
}

// Verify all records were removed
records = manager.GetRecords()
if len(records) != 0 {
t.Fatalf("expected 0 records, got %d", len(records))
}
}

func TestSync_DeduplicatesRequests(t *testing.T) {
appendCount := 0
adapter := &mockAdapter{
appendRecords: func(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
appendCount++
return records, nil
},
}

provider := &mockProvider{
name:        "cloudflare",
zoneFilters: []string{"example.com"},
adapter:     adapter,
}

manager := NewManager([]providers.Provider{provider})

// Create duplicate requests (same hostname and provider)
requests := []SyncRequest{
{
Hostname:     "app.example.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.10",
SourceID:     "container1",
RequestedAt:  time.Now().Add(-1 * time.Minute),
},
{
Hostname:     "app.example.com",
ProviderName: "cloudflare",
RecordType:   RecordTypeA,
Target:       "192.168.1.20",
SourceID:     "container2",
RequestedAt:  time.Now(), // More recent
},
}

err := manager.Sync(context.Background(), requests)
if err != nil {
t.Fatalf("Sync failed: %v", err)
}

// Should only create one record (the most recent one)
if appendCount != 1 {
t.Errorf("AppendRecords called %d times, want 1", appendCount)
}

records := manager.GetRecords()
if len(records) != 1 {
t.Fatalf("expected 1 record, got %d", len(records))
}

// Should be the more recent request
if records[0].SourceID != "container2" {
t.Errorf("sourceID = %q, want %q", records[0].SourceID, "container2")
}
if records[0].Value != "192.168.1.20" {
t.Errorf("value = %q, want %q", records[0].Value, "192.168.1.20")
}
}

func TestDetermineRecordType(t *testing.T) {
tests := []struct {
name     string
ip       string
expected RecordType
}{
{"IPv4", "192.168.1.1", RecordTypeA},
{"IPv6", "2001:db8::1", RecordTypeAAAA},
{"Invalid", "not-an-ip", RecordTypeA}, // default to A
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := DetermineRecordType(tt.ip)
if result != tt.expected {
t.Errorf("DetermineRecordType(%q) = %q, want %q", tt.ip, result, tt.expected)
}
})
}
}

func TestExtractZone(t *testing.T) {
tests := []struct {
name        string
hostname    string
filters     []string
expected    string
}{
{
name:     "exact match",
hostname: "example.com",
filters:  []string{"example.com"},
expected: "example.com",
},
{
name:     "subdomain",
hostname: "app.example.com",
filters:  []string{"example.com"},
expected: "example.com",
},
{
name:     "longest match",
hostname: "app.sub.example.com",
filters:  []string{"example.com", "sub.example.com"},
expected: "sub.example.com",
},
{
name:     "no match",
hostname: "app.different.com",
filters:  []string{"example.com"},
expected: "",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := extractZone(tt.hostname, tt.filters)
if result != tt.expected {
t.Errorf("extractZone(%q, %v) = %q, want %q", tt.hostname, tt.filters, result, tt.expected)
}
})
}
}
