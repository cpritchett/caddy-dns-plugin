---

description: "Task list for Caddy DNS Sync Module"
---

# Tasks: Caddy DNS Sync Module

**Input**: Design documents from `/specs/001-caddy-dns-module/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are OPTIONAL unless requested; focus on implementation tasks aligned to acceptance criteria.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Single project: `cmd/`, `internal/`, `pkg/`, `tests/` at repository root

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go module (Go 1.22.x) and dependencies in `go.mod`
- [x] T002 Create project directories per plan in `cmd/`, `internal/{config,docker,labels,providers/{cloudflare,unifi},dns,reconcile}/`, `pkg/contracts/`, `tests/{unit,integration,contract}/`
- [x] T003 Add Caddy module entrypoint skeleton in `cmd/caddy-dns-sync/main.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T004 Implement configuration loader with Caddyfile + env precedence in `internal/config/config.go`
- [x] T005 Implement label parser with configurable prefix and caddy inference in `internal/labels/parser.go`
- [x] T006 Scaffold Docker/Swarm event watcher with filters and debounce in `internal/docker/watcher.go`
- [x] T007 Define provider interface and libdns adapter contract in `internal/providers/provider.go`
- [x] T008 Implement DNS orchestration pipeline (desired state compute, create/delete) in `internal/dns/manager.go`
- [x] T009 Expose health and metrics handlers per OpenAPI in `internal/dns/handlers.go`

---

## Phase 3: User Story 1 - Automatic Public DNS for Docker Services (Priority: P1) 🎯 MVP

**Goal**: Create Cloudflare DNS records for labeled containers automatically.

**Independent Test**: Deploy a container with required labels; DNS record appears in Cloudflare within 30 seconds.

### Implementation for User Story 1

- [x] T010 [US1] Implement Cloudflare provider adapter (A/AAAA/CNAME, proxied/TTL) in `internal/providers/cloudflare/client.go`
- [ ] T011 [P] [US1] Map container label parsing to provider requests in `internal/labels/parser.go`
- [ ] T012 [US1] Integrate event watcher with Cloudflare create/delete in `internal/dns/manager.go`
- [ ] T013 [US1] Update quickstart with Cloudflare flow in `specs/001-caddy-dns-module/quickstart.md`

**Checkpoint**: User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Internal DNS via UniFi Controller (Priority: P2)

**Goal**: Register internal DNS entries in UniFi controller for labeled containers.

**Independent Test**: Deploy container with UniFi provider label; static DNS entry appears in UniFi controller.

### Implementation for User Story 2

- [ ] T014 [US2] Implement UniFi provider adapter for static DNS entries in `internal/providers/unifi/client.go`
- [ ] T015 [P] [US2] Wire UniFi provider into provider registry in `internal/providers/provider.go`
- [ ] T016 [US2] Integrate UniFi provider into DNS manager lifecycle in `internal/dns/manager.go`

**Checkpoint**: User Story 2 should be independently functional.

---

## Phase 5: User Story 3 - Multi-Provider Zone Validation (Priority: P2)

**Goal**: Enforce hostname validation against provider zone filters and warn when DNS sync is unconfigured.

**Independent Test**: Attempt invalid hostname against zone filter; receive rejection with clear error.

### Implementation for User Story 3

- [ ] T017 [US3] Implement zone filter validator per provider in `internal/providers/validator.go`
- [ ] T018 [P] [US3] Add warning logging for `caddy` labels lacking DNS provider in `internal/labels/parser.go`
- [ ] T019 [US3] Enforce conflict rejection and logging for duplicate hostnames per provider in `internal/dns/manager.go`

**Checkpoint**: User Story 3 should be independently functional.

---

## Phase 6: User Story 4 - DNS Reconciliation (Priority: P3)

**Goal**: Periodically reconcile provider DNS with running containers and correct drift.

**Independent Test**: Manually delete a DNS record; reconciliation loop recreates it within interval.

### Implementation for User Story 4

- [ ] T020 [US4] Implement reconciliation scheduler with configurable interval and backoff in `internal/reconcile/scheduler.go`
- [ ] T021 [P] [US4] Compute desired vs actual record diffs and apply in `internal/reconcile/diff.go`
- [ ] T022 [US4] Integrate reconciliation triggers with metrics and logs in `internal/reconcile/metrics.go`

**Checkpoint**: User Story 4 should be independently functional.

---

## Phase 7: User Story 5 - Configuration Flexibility (Priority: P3)

**Goal**: Support configuration via Caddyfile and environment variables with Caddyfile precedence.

**Independent Test**: Run plugin with env-only config; then with both env and Caddyfile ensuring Caddyfile wins.

### Implementation for User Story 5

- [ ] T023 [US5] Implement environment variable parsing overrides in `internal/config/config.go`
- [ ] T024 [P] [US5] Ensure Caddyfile parsing precedence over env in `internal/config/config.go`
- [ ] T025 [US5] Document configuration precedence in `specs/001-caddy-dns-module/quickstart.md`

**Checkpoint**: User Story 5 should be independently functional.

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T026 Harden structured logging and error messages across providers in `internal/dns/manager.go`
- [ ] T027 Optimize startup path to stay under 2s in `cmd/caddy-dns-sync/main.go`
- [ ] T028 Update documentation with metrics endpoints and reconcile usage in `specs/001-caddy-dns-module/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1) → Foundational (Phase 2) → User Stories (Phase 3–7) → Polish (Final)
- User stories proceed in priority order: US1 (MVP) → US2/US3 (parallel after foundation) → US4 → US5.

### User Story Dependencies

- US1: depends on foundational.
- US2: depends on foundational and provider registry (T007/T008).
- US3: depends on foundational and US1 label parsing groundwork.
- US4: depends on foundational and DNS manager from US1/US2.
- US5: depends on foundational config loader.

### Within Each User Story

- Labels/config before provider integration; provider registry before DNS mutations; documentation updates last.

### Parallel Opportunities

- [P] tasks T011, T015, T018, T021, T024 can run in parallel once prerequisites are met (different files, no overlapping edits).

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Setup + Foundational
2. Deliver US1 (Cloudflare public DNS) as MVP
3. Validate against independent test (record appears in 30s)

### Incremental Delivery

1. Foundation ready → US1 (MVP)
2. Add US2 (UniFi) and US3 (validation) in parallel
3. Add US4 (reconciliation)
4. Add US5 (configuration precedence)

### Parallel Team Strategy

- Developer A: US1
- Developer B: US2
- Developer C: US3/US5
- Developer D: US4 after foundation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify acceptance criteria per story after each checkpoint
- Commit after each task or logical group
