# Data Model: Caddy DNS Sync Module

## Entities

### Provider
- Fields: `name` (unique), `type` (`cloudflare`|`unifi`|`custom`), `zoneFilters` (list of wildcard/regex patterns), `credentials` (provider-specific), `ttl` (optional override), `proxied` (Cloudflare only), `reconcileEnabled` (bool), `rateLimit` (optional requests/min hint).
- Relationships: Owns many `DnsRecord`s; referenced by `SyncRequest` derived from containers.
- Validation: `name` required/unique; `type` must map to registered adapter; `zoneFilters` non-empty for mutation; provider-specific credential schema enforced.

### Container
- Fields: `id`, `name`, `labels`, `ipAddresses` (IPv4/IPv6), `state` (created|running|stopped|removed), `swarmService` (optional), `networks`.
- Relationships: Emits `SyncRequest` for zero or more `Provider`s based on labels.
- Validation: IP required for create/update; label prefix must match configured prefix; hostname inferred from `caddy` label when `hostname` absent.
- State transitions: created→running→stopped→removed; removals trigger record deletions.

### SyncRequest (derived event)
- Fields: `hostname`, `providerName`, `recordType` (A/AAAA/CNAME), `target` (IP or cname), `sourceContainerId`, `labels`, `requestedAt`.
- Relationships: Maps container intent to provider; leads to `DnsRecord` mutation.
- Validation: `hostname` must match provider `zoneFilters`; deduplicate by `(hostname, provider)`; reject missing provider with warning when `caddy` label exists.

### DnsRecord
- Fields: `id` (provider-scoped), `hostname`, `recordType`, `value`, `ttl`, `proxied`, `providerName`, `lastSyncAt`, `state` (present|pending|error).
- Relationships: Owned by `Provider`; reconciled against `SyncRequest` and provider reality.
- Validation: TTL within provider limits; `proxied` only for Cloudflare; value must match record type.

### ReconciliationRun
- Fields: `runId`, `startedAt`, `completedAt`, `diffsApplied` (list), `errors` (list), `duration`, `status` (success|partial|failed).
- Relationships: Consumes current `Container` state + provider records; updates `DnsRecord` states.
- Validation: Interval configurable; must backoff when prior run failed with rate-limit errors.

## Derived/Computed State
- `DesiredRecords`: set of `DnsRecord` specs from active containers and labels.
- `Drift`: difference between provider-reported records and `DesiredRecords`.

## Validation Rules
- Hostnames must pass RFC 1123 validation and provider `zoneFilters` before mutation.
- Multiple containers requesting the same hostname → deterministic tie-break (e.g., newest container wins) with warning; reconcile must detect conflicts.
- Missing container IP → defer create and retry with backoff; log warning.
- Provider API failures → retry with exponential backoff; mark `DnsRecord.state=error` for observability.
