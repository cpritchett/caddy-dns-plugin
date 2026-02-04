package dns

import (
"context"
"fmt"
"net"
"net/netip"
"strings"
"sync"
"time"

"github.com/cpritchett/caddy-dns-plugin/internal/providers"
"github.com/libdns/libdns"
)

// RecordType represents the type of DNS record
type RecordType string

const (
RecordTypeA     RecordType = "A"
RecordTypeAAAA  RecordType = "AAAA"
RecordTypeCNAME RecordType = "CNAME"
)

// RecordState represents the state of a DNS record
type RecordState string

const (
RecordStatePresent RecordState = "present"
RecordStatePending RecordState = "pending"
RecordStateError   RecordState = "error"
)

// DNSRecord represents a DNS record managed by the system
type DNSRecord struct {
ID           string
Hostname     string
RecordType   RecordType
Value        string
TTL          int
Proxied      bool
ProviderName string
LastSyncAt   time.Time
State        RecordState
SourceID     string // Container ID
}

// SyncRequest represents a request to sync DNS records for a container
type SyncRequest struct {
Hostname      string
ProviderName  string
RecordType    RecordType
Target        string
SourceID      string
Labels        map[string]string
RequestedAt   time.Time
TTL           *int
Proxied       *bool
}

// ContainerInfo represents container information needed for DNS sync
type ContainerInfo struct {
ID         string
Name       string
Labels     map[string]string
IPV4       []string
IPV6       []string
State      string
IsRunning  bool
}

// Manager orchestrates DNS record creation and deletion
type Manager struct {
providers map[string]providers.Provider
records   map[string]*DNSRecord // key: hostname:provider
mu        sync.RWMutex
}

// NewManager creates a new DNS manager
func NewManager(providerList []providers.Provider) *Manager {
providerMap := make(map[string]providers.Provider)
for _, p := range providerList {
providerMap[p.Name()] = p
}

return &Manager{
providers: providerMap,
records:   make(map[string]*DNSRecord),
}
}

// ComputeDesiredState computes the desired DNS records from container information
func (m *Manager) ComputeDesiredState(containers []ContainerInfo, labelPrefix string) ([]SyncRequest, error) {
var requests []SyncRequest

for _, container := range containers {
if !container.IsRunning {
continue
}

// Parse container labels to determine if DNS sync is requested
hostname, ok := container.Labels[labelPrefix+".hostname"]
if !ok || strings.TrimSpace(hostname) == "" {
// Try to infer from caddy label
if caddyLabel, hasCaddy := container.Labels["caddy"]; hasCaddy {
hostname = inferHostnameFromCaddy(caddyLabel)
}
}

if strings.TrimSpace(hostname) == "" {
continue
}

providerName, ok := container.Labels[labelPrefix+".provider"]
if !ok || strings.TrimSpace(providerName) == "" {
continue
}

// Check if enabled
if enabledStr, ok := container.Labels[labelPrefix+".enable"]; ok {
if strings.ToLower(strings.TrimSpace(enabledStr)) == "false" {
continue
}
}

// Determine target IPs
var targets []string
var recordType RecordType

// Prefer IPv4
if len(container.IPV4) > 0 {
targets = container.IPV4
recordType = RecordTypeA
} else if len(container.IPV6) > 0 {
targets = container.IPV6
recordType = RecordTypeAAAA
}

if len(targets) == 0 {
continue
}

// Parse optional TTL
var ttl *int
if ttlStr, ok := container.Labels[labelPrefix+".ttl"]; ok {
if parsed, err := parseInt(ttlStr); err == nil {
ttl = &parsed
}
}

// Parse optional proxied flag
var proxied *bool
if proxiedStr, ok := container.Labels[labelPrefix+".proxied"]; ok {
if parsed, err := parseBool(proxiedStr); err == nil {
proxied = &parsed
}
}

// Create sync request for first target IP
// Multiple IPs would need multiple records, but we'll use first for simplicity
requests = append(requests, SyncRequest{
Hostname:     hostname,
ProviderName: providerName,
RecordType:   recordType,
Target:       targets[0],
SourceID:     container.ID,
Labels:       container.Labels,
RequestedAt:  time.Now(),
TTL:          ttl,
Proxied:      proxied,
})
}

return requests, nil
}

// Sync processes sync requests and creates/updates DNS records
func (m *Manager) Sync(ctx context.Context, requests []SyncRequest) error {
m.mu.Lock()
defer m.mu.Unlock()

// Build map of desired records
desired := make(map[string]SyncRequest)
for _, req := range requests {
key := recordKey(req.Hostname, req.ProviderName)
// If duplicate, keep the one with most recent timestamp
if existing, ok := desired[key]; ok {
if req.RequestedAt.After(existing.RequestedAt) {
desired[key] = req
}
} else {
desired[key] = req
}
}

// Create or update records
for key, req := range desired {
if err := m.createOrUpdateRecord(ctx, req); err != nil {
// Log error but continue with other records
return fmt.Errorf("create/update record %s: %w", key, err)
}
}

return nil
}

// CreateRecord creates a new DNS record
func (m *Manager) CreateRecord(ctx context.Context, req SyncRequest) error {
m.mu.Lock()
defer m.mu.Unlock()

return m.createOrUpdateRecord(ctx, req)
}

// createOrUpdateRecord creates or updates a DNS record (caller must hold lock)
func (m *Manager) createOrUpdateRecord(ctx context.Context, req SyncRequest) error {
provider, ok := m.providers[req.ProviderName]
if !ok {
return fmt.Errorf("provider %q not found", req.ProviderName)
}

// Validate hostname against zone filters
if !m.validateHostname(req.Hostname, provider.ZoneFilters()) {
return fmt.Errorf("hostname %q does not match zone filters for provider %q", req.Hostname, req.ProviderName)
}

// Extract zone from hostname
zone := extractZone(req.Hostname, provider.ZoneFilters())
if zone == "" {
return fmt.Errorf("could not determine zone for hostname %q", req.Hostname)
}

// Build libdns record
ttl := 300 // default TTL
if req.TTL != nil {
ttl = *req.TTL
}

// Parse IP address
ipAddr, err := netip.ParseAddr(req.Target)
if err != nil {
return fmt.Errorf("invalid IP address %q: %w", req.Target, err)
}

record := libdns.Address{
Name: strings.TrimSuffix(req.Hostname, "."+zone),
IP:   ipAddr,
TTL:  time.Duration(ttl) * time.Second,
}

// Use the provider adapter to create/append the record
adapter := provider.Adapter()
records := []libdns.Record{record}

created, err := adapter.AppendRecords(ctx, zone, records)
if err != nil {
// Mark as error state
dnsRecord := &DNSRecord{
Hostname:     req.Hostname,
RecordType:   req.RecordType,
Value:        req.Target,
TTL:          ttl,
ProviderName: req.ProviderName,
LastSyncAt:   time.Now(),
State:        RecordStateError,
SourceID:     req.SourceID,
}
if req.Proxied != nil {
dnsRecord.Proxied = *req.Proxied
}
key := recordKey(req.Hostname, req.ProviderName)
m.records[key] = dnsRecord
return fmt.Errorf("append record: %w", err)
}

// Store the created record
if len(created) > 0 {
// Extract record info
recordID := ""
if addr, ok := created[0].(libdns.Address); ok && addr.ProviderData != nil {
// Try to extract ID from provider data if available
if idMap, ok := addr.ProviderData.(map[string]interface{}); ok {
if id, ok := idMap["id"].(string); ok {
recordID = id
}
}
}

dnsRecord := &DNSRecord{
ID:           recordID,
Hostname:     req.Hostname,
RecordType:   req.RecordType,
Value:        req.Target,
TTL:          ttl,
ProviderName: req.ProviderName,
LastSyncAt:   time.Now(),
State:        RecordStatePresent,
SourceID:     req.SourceID,
}
if req.Proxied != nil {
dnsRecord.Proxied = *req.Proxied
}
key := recordKey(req.Hostname, req.ProviderName)
m.records[key] = dnsRecord
}

return nil
}

// DeleteRecord deletes a DNS record for a container
func (m *Manager) DeleteRecord(ctx context.Context, hostname, providerName, containerID string) error {
m.mu.Lock()
defer m.mu.Unlock()

key := recordKey(hostname, providerName)
record, ok := m.records[key]
if !ok {
// Record not found, nothing to delete
return nil
}

// Verify the record belongs to this container
if record.SourceID != containerID {
return fmt.Errorf("record does not belong to container %q", containerID)
}

provider, ok := m.providers[providerName]
if !ok {
return fmt.Errorf("provider %q not found", providerName)
}

// Extract zone from hostname
zone := extractZone(hostname, provider.ZoneFilters())
if zone == "" {
return fmt.Errorf("could not determine zone for hostname %q", hostname)
}

// Parse IP address for deletion
ipAddr, err := netip.ParseAddr(record.Value)
if err != nil {
return fmt.Errorf("invalid IP address %q: %w", record.Value, err)
}

// Build libdns record for deletion
libdnsRecord := libdns.Address{
Name: strings.TrimSuffix(hostname, "."+zone),
IP:   ipAddr,
}

// Use the provider adapter to delete the record
adapter := provider.Adapter()
records := []libdns.Record{libdnsRecord}

_, err = adapter.DeleteRecords(ctx, zone, records)
if err != nil {
return fmt.Errorf("delete record: %w", err)
}

// Remove from tracking
delete(m.records, key)

return nil
}

// DeleteRecordsForContainer deletes all DNS records associated with a container
func (m *Manager) DeleteRecordsForContainer(ctx context.Context, containerID string) error {
m.mu.Lock()
defer m.mu.Unlock()

var errs []error

// Find all records for this container
for key, record := range m.records {
if record.SourceID != containerID {
continue
}

provider, ok := m.providers[record.ProviderName]
if !ok {
errs = append(errs, fmt.Errorf("provider %q not found", record.ProviderName))
continue
}

// Extract zone from hostname
zone := extractZone(record.Hostname, provider.ZoneFilters())
if zone == "" {
errs = append(errs, fmt.Errorf("could not determine zone for hostname %q", record.Hostname))
continue
}

// Parse IP address for deletion
ipAddr, err := netip.ParseAddr(record.Value)
if err != nil {
errs = append(errs, fmt.Errorf("invalid IP address %q: %w", record.Value, err))
continue
}

// Build libdns record for deletion
libdnsRecord := libdns.Address{
Name: strings.TrimSuffix(record.Hostname, "."+zone),
IP:   ipAddr,
}

// Use the provider adapter to delete the record
adapter := provider.Adapter()
records := []libdns.Record{libdnsRecord}

_, err = adapter.DeleteRecords(ctx, zone, records)
if err != nil {
errs = append(errs, fmt.Errorf("delete record %s: %w", key, err))
continue
}

// Remove from tracking
delete(m.records, key)
}

if len(errs) > 0 {
return fmt.Errorf("delete records: %v", errs)
}

return nil
}

// GetRecords returns all tracked DNS records
func (m *Manager) GetRecords() []DNSRecord {
m.mu.RLock()
defer m.mu.RUnlock()

records := make([]DNSRecord, 0, len(m.records))
for _, record := range m.records {
records = append(records, *record)
}

return records
}

// validateHostname checks if hostname matches any zone filter
func (m *Manager) validateHostname(hostname string, zoneFilters []string) bool {
hostname = strings.ToLower(strings.TrimSpace(hostname))
if hostname == "" {
return false
}

for _, filter := range zoneFilters {
filter = strings.ToLower(strings.TrimSpace(filter))
if filter == "" {
continue
}

// Simple suffix matching (can be enhanced with glob/regex later)
if strings.HasSuffix(hostname, "."+filter) || hostname == filter {
return true
}
}

return false
}

// Helper functions

func recordKey(hostname, provider string) string {
return hostname + ":" + provider
}

func extractZone(hostname string, zoneFilters []string) string {
hostname = strings.ToLower(strings.TrimSpace(hostname))

// Find the longest matching zone filter
var longestMatch string
for _, filter := range zoneFilters {
filter = strings.ToLower(strings.TrimSpace(filter))
if filter == "" {
continue
}

if strings.HasSuffix(hostname, "."+filter) || hostname == filter {
if len(filter) > len(longestMatch) {
longestMatch = filter
}
}
}

return longestMatch
}

func inferHostnameFromCaddy(value string) string {
trimmed := strings.TrimSpace(value)
if trimmed == "" {
return ""
}

replaced := strings.ReplaceAll(trimmed, ",", " ")
fields := strings.Fields(replaced)
if len(fields) == 0 {
return ""
}

host := strings.TrimPrefix(fields[0], "https://")
host = strings.TrimPrefix(host, "http://")
return host
}

func parseInt(s string) (int, error) {
var i int
_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &i)
return i, err
}

func parseBool(s string) (bool, error) {
s = strings.ToLower(strings.TrimSpace(s))
switch s {
case "true", "1", "yes", "on":
return true, nil
case "false", "0", "no", "off":
return false, nil
default:
return false, fmt.Errorf("invalid boolean value: %q", s)
}
}

// DetermineRecordType determines the DNS record type based on IP address
func DetermineRecordType(ipAddr string) RecordType {
ip := net.ParseIP(ipAddr)
if ip == nil {
return RecordTypeA // default
}

if ip.To4() != nil {
return RecordTypeA
}

return RecordTypeAAAA
}
