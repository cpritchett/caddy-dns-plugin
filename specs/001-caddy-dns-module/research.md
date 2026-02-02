# Research: Caddy DNS Sync Module

## Go toolchain version

- Decision: Target Go 1.22.x for the module and tooling.
- Rationale: Current stable with security fixes, improved memory management, and mature generics; aligns with Caddy’s recent releases and libdns providers.
- Alternatives considered: Go 1.21.x (older security window); Go tip (too unstable for plugin consumers).

## Testing harness

- Decision: Use Go `testing` with unit fakes plus integration tests driven by Docker Engine API via socket and provider fakes; shell-driven e2e scripts only where container orchestration is required.
- Rationale: Matches Go ecosystem norms, keeps fast unit feedback, and allows exercising label parsing and reconciliation against real Docker events without external dependencies beyond docker/socket; fakes keep provider testing deterministic.
- Alternatives considered: Full end-to-end in-compose for every test (too slow for CI); mocking Docker entirely (misses event ordering/edge cases).

## Governance placeholder

- Decision: Operate with default engineering gates: document assumptions, keep tests for new logic, and expose metrics; revisit once constitution is populated.
- Rationale: Constitution file is empty; adopting minimal, transparent gates prevents paralysis while staying ready to align once guidance exists.
- Alternatives considered: Blocking progress until constitution is defined (stalls feature delivery).

## Docker event ingestion best practices

- Decision: Use Docker SDK event stream with filters for container/service lifecycle, debounce rapid restarts, and reconcile on startup to avoid missed events.
- Rationale: Matches FR-001/FR-003, handles Swarm and standalone uniformly, and reduces flapping noise.
- Alternatives considered: Polling container lists (slower and risks 30s SLA); shelling to `docker` CLI (less reliable, harder to vendor for Caddy).

## Label parsing and inference

- Decision: Normalize a configurable prefix `caddy_dns` with defaults matching caddy-docker-proxy (`caddy` label inference), with strict validation and descriptive errors.
- Rationale: Aligns with FR-002/FR-014, minimizes operator friction, and supports warning behavior for missing DNS config when `caddy` labels exist.
- Alternatives considered: New label schema unrelated to caddy-docker-proxy (increases adoption cost).

## Cloudflare provider via libdns

- Decision: Use libdns/cloudflare with support for A/AAAA/CNAME, proxied flag, TTL override, and zone filter validation before mutation.
- Rationale: Satisfies FR-019–FR-021 + FR-007; leverages libdns contract expected by Caddy.
- Alternatives considered: Direct Cloudflare REST calls (duplicates libdns, more maintenance); ACME DNS provider hooks (violates FR-015).

## UniFi internal DNS

- Decision: Implement a provider adapter that calls UniFi controller’s static DNS endpoints (client credentials), scoped by hostname zone filters.
- Rationale: Meets FR-022 + internal DNS user story; keeps provider interface extensible.
- Alternatives considered: Using DHCP host overrides (less precise for container lifecycle); deferring UniFi to later milestone (misses P2 priority).

## Reconciliation and backoff

- Decision: Run reconciliation on a configurable interval (default 5m) with exponential backoff for transient provider failures and idempotent record diffing.
- Rationale: Aligns SC-003 and FR-009/FR-013 while respecting rate limits.
- Alternatives considered: Constant-interval retries (risks rate-limit penalties); manual-only reconciliation (misses drift correction requirement).
