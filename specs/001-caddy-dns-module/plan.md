# Implementation Plan: Caddy DNS Sync Module

**Branch**: `001-caddy-dns-module` | **Date**: 2026-02-01 | **Spec**: /specs/001-caddy-dns-module/spec.md
**Input**: Feature specification from `/specs/001-caddy-dns-module/spec.md`

## Summary

Build a Caddy module that watches Docker container lifecycle events, parses DNS-related labels (compatible with caddy-docker-proxy), and syncs DNS records to configured providers (Cloudflare public, UniFi internal). Include hostname validation, reconciliation, metrics, retries with backoff, and dual configuration via Caddyfile or environment variables.

## Technical Context

**Language/Version**: Go 1.23.0
**Primary Dependencies**: Caddy module API, Docker SDK for Go, libdns (Cloudflare + UniFi implementations), caddy-docker-proxy label conventions
**Storage**: N/A (state derived from Docker + providers)
**Testing**: Go `testing` with unit + integration suites (Docker event stream, provider fakes)
**Target Platform**: Linux hosts running Caddy with Docker or Swarm; support amd64/arm64
**Project Type**: Single Go module/library with optional CLI for debugging
**Performance Goals**: Create/delete DNS records within 30s of lifecycle events; reconcile drift within 5 minutes; add <2s to Caddy startup
**Constraints**: Must not interfere with ACME issuance; respect provider rate limits with backoff; operate without persistent storage
**Scale/Scope**: Homelab to small-cluster scale (dozens of containers, multiple zones/providers)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Principles (I–V) to uphold: Test-First Deterministic Delivery; Contract-Driven Interfaces; Observability and Resilience; Simplicity and Statelessness; Documentation and Template Alignment.
- Compliance: No violations planned; tests and contracts will accompany new behavior; no persistent state introduced.
- If any exception arises, document in Complexity Tracking with rationale before proceeding.

## Project Structure

### Documentation (this feature)

```text
specs/001-caddy-dns-module/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── caddy-dns-sync/           # optional CLI for local debugging/reconciliation triggers

internal/
├── config/                   # Caddyfile/env parsing and validation
├── docker/                   # Docker/Swarm event watcher and container inspection
├── labels/                   # label parsing and inference from caddy-docker-proxy
├── providers/
│   ├── cloudflare/           # libdns-backed Cloudflare adapter (proxied/TTL support)
│   └── unifi/                # UniFi static DNS adapter
├── dns/                      # record mapping and mutation orchestration
└── reconcile/                # reconciliation loop, backoff, drift detection

pkg/
└── contracts/                # shared DTOs for API/metrics exposure

tests/
├── unit/
├── integration/              # Docker daemon + provider fakes
└── contract/                 # API/metrics contract tests
```

**Structure Decision**: Single Go module with internal packages per responsibility, shared DTOs in `pkg/contracts`, optional CLI under `cmd/caddy-dns-sync`, and layered tests (unit, integration, contract) aligned to the feature deliverables.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
