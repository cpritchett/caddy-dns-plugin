# Feature Specification: Caddy DNS Sync Module

**Feature Branch**: `001-caddy-dns-module`  
**Created**: 2026-02-01  
**Status**: Draft  
**Input**: User description: "Build Caddy DNS Sync Module for automatic Docker container DNS management"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Public DNS for Docker Services (Priority: P1)

As a homelab operator, I want DNS records automatically created in Cloudflare when I deploy Docker containers with specific labels, so that my services are immediately accessible via their hostnames without manual DNS configuration.

**Why this priority**: This is the core value proposition - eliminating manual DNS management for the most common use case (public-facing services via Cloudflare).

**Independent Test**: Deploy a Docker container with appropriate labels and verify the DNS record appears in Cloudflare within 30 seconds.

**Acceptance Scenarios**:

1. **Given** a container with `caddy_dns.enable=true` and `caddy_dns.provider=cloudflare-public` labels, **When** the container starts, **Then** a DNS A record is created in Cloudflare pointing to the container's IP address.
2. **Given** a running container with DNS sync enabled, **When** the container is removed, **Then** the corresponding DNS record is deleted from Cloudflare.
3. **Given** a container with an existing `caddy` label (hostname), **When** `caddy_dns.enable=true` is added, **Then** the plugin infers the hostname from the caddy label and creates the DNS record.

---

### User Story 2 - Internal DNS via UniFi Controller (Priority: P2)

As a homelab operator with a UniFi network, I want internal services to automatically register DNS entries in my UniFi controller, so that devices on my local network can resolve internal service hostnames.

**Why this priority**: Complements public DNS with internal DNS management - essential for split-horizon DNS setups common in homelabs.

**Independent Test**: Deploy a container with UniFi provider label and verify the static DNS entry appears in the UniFi controller.

**Acceptance Scenarios**:

1. **Given** a container with `caddy_dns.provider=unifi-internal` label, **When** the container starts, **Then** a static DNS entry is created in the UniFi controller.
2. **Given** multiple providers configured for different zones, **When** a container specifies `caddy_dns.provider=unifi-internal`, **Then** the record is created only in the UniFi controller, not in Cloudflare.

---

### User Story 3 - Multi-Provider Zone Validation (Priority: P2)

As an operator managing multiple DNS zones, I want the plugin to validate that hostnames match the configured zone filters for each provider, so that DNS records are created in the correct provider and invalid configurations are rejected.

**Why this priority**: Prevents misconfiguration and ensures DNS records end up in the right place - critical for multi-zone environments.

**Independent Test**: Attempt to create a record with a hostname that doesn't match the provider's zone filter and verify rejection with clear error message.

**Acceptance Scenarios**:

1. **Given** a provider configured with zone filter `*.example.com`, **When** a container requests `app.other.com`, **Then** the request is rejected with a validation error.
2. **Given** a container with `caddy` label but missing `caddy_dns.provider`, **When** the container starts, **Then** a warning is logged indicating DNS sync is not configured.

---

### User Story 4 - DNS Reconciliation (Priority: P3)

As an operator, I want the plugin to periodically verify that DNS records match running containers, so that any manual DNS changes or drift are automatically corrected.

**Why this priority**: Ensures long-term consistency - less critical than initial sync but important for reliability.

**Independent Test**: Manually delete a DNS record and verify it is recreated within the reconciliation interval.

**Acceptance Scenarios**:

1. **Given** a DNS record was manually deleted, **When** the reconciliation loop runs, **Then** the record is recreated to match the running container.
2. **Given** a DNS record has incorrect data (wrong IP), **When** the reconciliation loop runs, **Then** the record is updated to match the current container state.

---

### User Story 5 - Configuration Flexibility (Priority: P3)

As an operator, I want to configure the plugin via Caddyfile or environment variables, so that I can integrate it with my existing configuration management approach.

**Why this priority**: Configuration flexibility improves adoption and integration with existing workflows.

**Independent Test**: Configure the plugin using only environment variables and verify it functions correctly.

**Acceptance Scenarios**:

1. **Given** configuration via environment variables only, **When** Caddy starts, **Then** the plugin initializes with the environment-based configuration.
2. **Given** both Caddyfile and environment variables are present, **When** Caddy starts, **Then** Caddyfile configuration takes precedence.

---

### Edge Cases

- What happens when a container has DNS sync enabled but Docker networking is misconfigured (no IP available)?
- How does the system handle provider API rate limits?
- What happens when multiple containers request the same hostname?
- How does the plugin behave when the DNS provider API is temporarily unavailable?
- What happens when a container label is changed while the container is running?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST monitor Docker API for container lifecycle events (create, start, stop, remove) in real-time
- **FR-002**: System MUST parse DNS-related labels from containers using a configurable label prefix (default: `caddy_dns`)
- **FR-003**: System MUST support both standalone Docker containers and Docker Swarm services
- **FR-004**: System MUST create DNS records when containers with appropriate labels start
- **FR-005**: System MUST delete DNS records when containers are removed
- **FR-006**: System MUST support multiple DNS providers configured with unique names
- **FR-007**: System MUST validate hostnames against provider zone filters before creating records
- **FR-008**: System MUST log warnings for containers with `caddy` labels but missing DNS sync configuration
- **FR-009**: System MUST provide a reconciliation loop to periodically verify DNS state matches container state
- **FR-010**: System MUST support configuration via Caddyfile global options block
- **FR-011**: System MUST support configuration via environment variables
- **FR-012**: System MUST expose operational metrics for monitoring
- **FR-013**: System MUST retry transient failures with exponential backoff
- **FR-014**: System MUST work alongside caddy-docker-proxy without interference
- **FR-015**: System MUST use `libdns` providers for DNS record operations (append/set/delete), not ACME issuance hooks
- **FR-016**: System SHOULD support reusing existing DNS provider credentials/config from Caddy's DNS provider configuration when available, without triggering certificate issuance
- **FR-017**: System MUST be implemented as a Caddy module (Go)
- **FR-018**: System MUST only use a Caddy config adapter if a module implementation is infeasible

### DNS Provider Requirements

- **FR-019**: Cloudflare provider MUST support A/AAAA/CNAME record types
- **FR-020**: Cloudflare provider MUST support proxied and DNS-only modes
- **FR-021**: Cloudflare provider MUST support configurable TTL
- **FR-022**: UniFi provider MUST support static DNS entry creation/deletion
- **FR-023**: Provider interface MUST be extensible for community contributions

### Key Entities

- **Container**: Docker container or Swarm service with DNS-related labels
- **DNS Record**: A mapping of hostname to IP address managed by the plugin
- **Provider**: Configured DNS service (Cloudflare, UniFi, etc.) with credentials and zone filters
- **Zone Filter**: Pattern(s) defining which hostnames a provider can manage

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: DNS records are created within 30 seconds of container start
- **SC-002**: DNS records are deleted within 30 seconds of container removal
- **SC-003**: Reconciliation detects and corrects drift within 5 minutes
- **SC-004**: Plugin startup adds less than 2 seconds to Caddy boot time
- **SC-005**: Operators can deploy services with DNS without any manual DNS configuration (zero-touch DNS provisioning)
- **SC-006**: Invalid hostname/provider configurations are rejected with actionable error messages
- **SC-007**: Plugin handles DNS provider unavailability gracefully without crashing
- **SC-008**: At least two DNS providers (Cloudflare, UniFi) are supported at launch

## Assumptions

- Docker socket is accessible at a configurable path (default: `/var/run/docker.sock`)
- DNS provider credentials are available via environment variables or Caddyfile
- Container IPs are routable from the networks where DNS resolution occurs
- Operators have appropriate permissions on DNS provider accounts
- caddy-docker-proxy is the primary use case for hostname inference via `caddy` labels
- Reusing provider configuration is optional; absence of reusable config does not block DNS sync
