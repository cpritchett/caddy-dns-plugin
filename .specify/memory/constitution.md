# caddy-dns-plugin Constitution

<!--
Sync Impact Report
Version: none -> 1.0.0
Modified principles: (new) I. Test-First Deterministic Delivery; (new) II. Contract-Driven Interfaces; (new) III. Observability and Resilience; (new) IV. Simplicity and Statelessness; (new) V. Documentation and Template Alignment
Added sections: Technical Standards; Development Workflow and Review
Removed sections: none
Templates requiring updates: .specify/templates/plan-template.md ✅; .specify/templates/spec-template.md ✅; .specify/templates/tasks-template.md ✅; .specify/templates/checklist-template.md ⚠ (no change needed)
Follow-up TODOs: none
-->

## Core Principles

### I. Test-First Deterministic Delivery
Tests must precede new behavior: write failing unit/integration tests before implementation, covering Docker event handling, label parsing, and provider adapters. CI must run these suites; changes lacking tests are blocked unless explicitly documented as test-exempt maintenance. Deterministic fakes for Docker and providers are preferred over broad mocks to preserve event ordering fidelity.

### II. Contract-Driven Interfaces
All external surfaces require explicit contracts: OpenAPI for control endpoints, libdns-compatible interfaces for providers, and documented label schemas. Inputs are validated early with clear errors; ambiguous labels or mismatched zone filters are rejected. Contract changes require concurrent doc/spec updates and backward-compatibility notes.

### III. Observability and Resilience
Structured logs, health, and Prometheus metrics are mandatory. Transient failures use exponential backoff and must not crash Caddy. Reconciliation is periodic and idempotent, with drift surfaced via metrics/logs. Rate limits are respected; retries include jitter.

### IV. Simplicity and Statelessness
No persistent storage is introduced. Desired state is derived from Docker/Swarm state and provider records. Favor minimal dependencies, short startup (<2s target), and avoid adding auxiliary services. Prefer straightforward flows over abstractions unless they reduce risk or duplication.

### V. Documentation and Template Alignment
Every feature/change updates spec, plan, contracts, and quickstart to reflect realities. Constitution Check in plans must cite compliance or justified violations. Manual guidance sections between markers must be preserved when regenerating agent or template files.

## Technical Standards

- Language/toolchain: Go 1.22.x; Caddy module API; Docker SDK for Go; libdns providers (Cloudflare, UniFi) as first-class targets.
- Labeling: Default prefix `caddy_dns`; inference from `caddy` labels must remain compatible with caddy-docker-proxy.
- Platforms: Linux hosts with Docker or Swarm; amd64/arm64 supported; no platform-specific forks without justification.
- Performance: DNS create/delete within 30s of events; reconciliation within 5m; startup overhead <2s.

## Development Workflow and Review

- Every plan.md must include a Constitution Check section enumerating adherence to Principles I–V and list any justified exceptions.
- Specs must include measurable outcomes, observability expectations, and validation/error behaviors aligned with Principles II and III.
- Tasks must include tests for new logic (unit/integration) and explicit observability work when behavior changes.
- Reviews verify contracts updated with code, docs refreshed, and no hidden state added. Violations require entries in Complexity Tracking/justifications.

## Governance

- This constitution guides all feature specs, plans, tasks, and runtime docs. Compliance is reviewed in every PR touching behavior or contracts.
- Amendments require: (1) rationale, (2) version bump per semver, (3) updates to affected templates/guides, (4) recorded date.
- Versioning: MAJOR for principle removal/incompatible governance, MINOR for new or materially expanded principles/sections, PATCH for clarifications.
- If ratification date is unknown during future edits, insert TODO with context; otherwise keep ISO dates.

**Version**: 1.0.0 | **Ratified**: 2026-02-01 | **Last Amended**: 2026-02-01
