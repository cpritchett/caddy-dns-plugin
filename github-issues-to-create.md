# GitHub Issues to Create from tasks.md

Generated from: `specs/001-caddy-dns-module/tasks.md`
Repository: https://github.com/cpritchett/caddy-dns-plugin

## Phase 2: Foundational (Blocking Prerequisites)

### Issue 1: T008 - Implement DNS orchestration pipeline

**Title:** [T008] Implement DNS orchestration pipeline

**Labels:** `enhancement`, `foundational`, `phase-2`, `priority-high`

**Description:**
Implement DNS orchestration pipeline with desired state computation and create/delete operations.

**Task Details:**
- Task ID: T008
- Phase: Foundational (Phase 2)
- File: `internal/dns/manager.go`
- Dependencies: Requires T004-T007 to be complete
- Blocking: Required before any user story can be implemented

**Acceptance Criteria:**
- [ ] Desired state computation implemented
- [ ] DNS record creation logic implemented
- [ ] DNS record deletion logic implemented
- [ ] Pipeline integrates with provider interface from T007
- [ ] Unit tests for state computation
- [ ] Integration tests for create/delete operations

**Implementation Notes:**
This is a blocking prerequisite that must be complete before ANY user story can be implemented.

---

### Issue 2: T009 - Expose health and metrics handlers

**Title:** [T009] Expose health and metrics handlers per OpenAPI

**Labels:** `enhancement`, `foundational`, `phase-2`, `priority-high`, `observability`

**Description:**
Expose health check and metrics handlers according to the OpenAPI specification.

**Task Details:**
- Task ID: T009
- Phase: Foundational (Phase 2)
- File: `internal/dns/handlers.go`
- Dependencies: Requires T004-T007 to be complete
- Blocking: Required before any user story can be implemented

**Acceptance Criteria:**
- [ ] Health check endpoint implemented and exposed
- [ ] Metrics endpoint implemented and exposed
- [ ] Handlers follow OpenAPI specification in `specs/001-caddy-dns-module/contracts/`
- [ ] Proper HTTP status codes returned
- [ ] Documentation updated

**Implementation Notes:**
Part of foundational infrastructure. Ensure OpenAPI contracts are followed.

---

## Phase 3: User Story 1 - Automatic Public DNS for Docker Services (Priority: P1) 🎯 MVP

### Issue 3: T010 - Implement Cloudflare provider adapter

**Title:** [T010][US1] Implement Cloudflare provider adapter

**Labels:** `enhancement`, `user-story-1`, `phase-3`, `priority-critical`, `mvp`, `cloudflare`

**Description:**
Implement Cloudflare provider adapter with support for A/AAAA/CNAME records, proxied mode, and TTL configuration.

**Task Details:**
- Task ID: T010
- User Story: US1 - Automatic Public DNS for Docker Services
- Phase: Phase 3
- File: `internal/providers/cloudflare/client.go`
- Dependencies: Requires foundational phase (T004-T009) to be complete
- Part of: MVP delivery

**Acceptance Criteria:**
- [ ] A record support implemented
- [ ] AAAA record support implemented
- [ ] CNAME record support implemented
- [ ] Cloudflare proxied mode support
- [ ] TTL configuration support
- [ ] Integration with libdns Cloudflare provider
- [ ] Error handling for API failures
- [ ] Unit tests for all record types
- [ ] Integration tests with Cloudflare API (or mocked)

**Independent Test:**
Deploy a container with required labels; DNS record appears in Cloudflare within 30 seconds.

---

### Issue 4: T011 - Map container label parsing to provider requests

**Title:** [T011][US1][P] Map container label parsing to provider requests

**Labels:** `enhancement`, `user-story-1`, `phase-3`, `priority-critical`, `mvp`, `parallel-ok`

**Description:**
Map container label parsing to provider requests in the label parser.

**Task Details:**
- Task ID: T011
- User Story: US1 - Automatic Public DNS for Docker Services
- Phase: Phase 3
- File: `internal/labels/parser.go`
- Can run in parallel: Yes (different files, no dependencies)
- Dependencies: Requires foundational phase (T004-T007) to be complete
- Part of: MVP delivery

**Acceptance Criteria:**
- [ ] Label parsing maps to Cloudflare provider requests
- [ ] DNS record type inference implemented
- [ ] Hostname extraction from labels
- [ ] Provider-specific configuration parsed
- [ ] Validation of required label fields
- [ ] Unit tests for label-to-request mapping

---

### Issue 5: T012 - Integrate event watcher with Cloudflare create/delete

**Title:** [T012][US1] Integrate event watcher with Cloudflare create/delete

**Labels:** `enhancement`, `user-story-1`, `phase-3`, `priority-critical`, `mvp`, `cloudflare`

**Description:**
Integrate Docker/Swarm event watcher with Cloudflare DNS record create/delete operations.

**Task Details:**
- Task ID: T012
- User Story: US1 - Automatic Public DNS for Docker Services
- Phase: Phase 3
- File: `internal/dns/manager.go`
- Dependencies: Requires T006, T008, T010 to be complete
- Part of: MVP delivery

**Acceptance Criteria:**
- [ ] Container start events trigger DNS record creation
- [ ] Container stop/remove events trigger DNS record deletion
- [ ] Event debouncing works correctly (30s default)
- [ ] Integration with Cloudflare provider adapter
- [ ] Error handling and retry logic
- [ ] Logging for all DNS operations
- [ ] Integration tests for full flow

**Independent Test:**
Deploy a container with required labels; DNS record appears in Cloudflare within 30 seconds.

---

### Issue 6: T013 - Update quickstart with Cloudflare flow

**Title:** [T013][US1] Update quickstart with Cloudflare flow

**Labels:** `documentation`, `user-story-1`, `phase-3`, `priority-high`, `mvp`

**Description:**
Update the quickstart documentation with the Cloudflare DNS sync workflow.

**Task Details:**
- Task ID: T013
- User Story: US1 - Automatic Public DNS for Docker Services
- Phase: Phase 3
- File: `specs/001-caddy-dns-module/quickstart.md`
- Dependencies: Requires T010-T012 to be complete
- Part of: MVP delivery

**Acceptance Criteria:**
- [ ] Step-by-step Cloudflare setup guide
- [ ] Example container labels documented
- [ ] Configuration examples provided
- [ ] Troubleshooting section added
- [ ] Expected behavior documented

**Checkpoint:**
User Story 1 should be fully functional and testable independently after this task.

---

## Phase 4: User Story 2 - Internal DNS via UniFi Controller (Priority: P2)

### Issue 7: T014 - Implement UniFi provider adapter

**Title:** [T014][US2] Implement UniFi provider adapter

**Labels:** `enhancement`, `user-story-2`, `phase-4`, `priority-high`, `unifi`

**Description:**
Implement UniFi provider adapter for static DNS entries in UniFi controller.

**Task Details:**
- Task ID: T014
- User Story: US2 - Internal DNS via UniFi Controller
- Phase: Phase 4
- File: `internal/providers/unifi/client.go`
- Dependencies: Requires foundational phase and provider interface (T007, T008)

**Acceptance Criteria:**
- [ ] UniFi controller API integration
- [ ] Static DNS entry creation
- [ ] Static DNS entry deletion
- [ ] Integration with libdns UniFi provider
- [ ] Authentication with UniFi controller
- [ ] Error handling for API failures
- [ ] Unit tests for UniFi operations
- [ ] Integration tests with UniFi controller (or mocked)

**Independent Test:**
Deploy container with UniFi provider label; static DNS entry appears in UniFi controller.

---

### Issue 8: T015 - Wire UniFi provider into provider registry

**Title:** [T015][US2][P] Wire UniFi provider into provider registry

**Labels:** `enhancement`, `user-story-2`, `phase-4`, `priority-high`, `unifi`, `parallel-ok`

**Description:**
Wire UniFi provider into the provider registry.

**Task Details:**
- Task ID: T015
- User Story: US2 - Internal DNS via UniFi Controller
- Phase: Phase 4
- File: `internal/providers/provider.go`
- Can run in parallel: Yes (different files, no dependencies)
- Dependencies: Requires T007 and T014 to be complete

**Acceptance Criteria:**
- [ ] UniFi provider registered in provider registry
- [ ] Provider initialization logic
- [ ] Provider configuration validation
- [ ] Provider selection based on labels
- [ ] Unit tests for provider registration

---

### Issue 9: T016 - Integrate UniFi provider into DNS manager lifecycle

**Title:** [T016][US2] Integrate UniFi provider into DNS manager lifecycle

**Labels:** `enhancement`, `user-story-2`, `phase-4`, `priority-high`, `unifi`

**Description:**
Integrate UniFi provider into the DNS manager lifecycle for managing DNS records.

**Task Details:**
- Task ID: T016
- User Story: US2 - Internal DNS via UniFi Controller
- Phase: Phase 4
- File: `internal/dns/manager.go`
- Dependencies: Requires T014, T015 to be complete

**Acceptance Criteria:**
- [ ] DNS manager supports UniFi provider
- [ ] Container events trigger UniFi DNS operations
- [ ] Multi-provider support in DNS manager
- [ ] Error handling per provider
- [ ] Integration tests for UniFi flow

**Checkpoint:**
User Story 2 should be independently functional after this task.

---

## Phase 5: User Story 3 - Multi-Provider Zone Validation (Priority: P2)

### Issue 10: T017 - Implement zone filter validator per provider

**Title:** [T017][US3] Implement zone filter validator per provider

**Labels:** `enhancement`, `user-story-3`, `phase-5`, `priority-high`, `validation`

**Description:**
Implement zone filter validator per provider to enforce hostname validation.

**Task Details:**
- Task ID: T017
- User Story: US3 - Multi-Provider Zone Validation
- Phase: Phase 5
- File: `internal/providers/validator.go`
- Dependencies: Requires foundational phase and provider implementations (T010, T014)

**Acceptance Criteria:**
- [ ] Zone filter validation logic per provider
- [ ] Hostname validation against zone filters
- [ ] Clear error messages for validation failures
- [ ] Support for multiple zone patterns
- [ ] Unit tests for validation logic
- [ ] Integration tests for rejection scenarios

**Independent Test:**
Attempt invalid hostname against zone filter; receive rejection with clear error.

---

### Issue 11: T018 - Add warning logging for caddy labels lacking DNS provider

**Title:** [T018][US3][P] Add warning logging for caddy labels lacking DNS provider

**Labels:** `enhancement`, `user-story-3`, `phase-5`, `priority-medium`, `validation`, `parallel-ok`

**Description:**
Add warning logging for containers with `caddy` labels that lack a DNS provider configuration.

**Task Details:**
- Task ID: T018
- User Story: US3 - Multi-Provider Zone Validation
- Phase: Phase 5
- File: `internal/labels/parser.go`
- Can run in parallel: Yes (different files, no dependencies)
- Dependencies: Requires T005 and provider implementations

**Acceptance Criteria:**
- [ ] Warning logged when caddy label present but no DNS provider
- [ ] Clear message indicating missing DNS configuration
- [ ] No false positives for valid configurations
- [ ] Structured logging format
- [ ] Unit tests for warning conditions

---

### Issue 12: T019 - Enforce conflict rejection for duplicate hostnames

**Title:** [T019][US3] Enforce conflict rejection and logging for duplicate hostnames

**Labels:** `enhancement`, `user-story-3`, `phase-5`, `priority-high`, `validation`

**Description:**
Enforce conflict rejection and logging for duplicate hostnames per provider.

**Task Details:**
- Task ID: T019
- User Story: US3 - Multi-Provider Zone Validation
- Phase: Phase 5
- File: `internal/dns/manager.go`
- Dependencies: Requires T008 and provider implementations

**Acceptance Criteria:**
- [ ] Duplicate hostname detection per provider
- [ ] Conflict rejection with clear error
- [ ] Logging of conflicts with details
- [ ] Support for multi-provider scenarios
- [ ] Unit tests for conflict detection
- [ ] Integration tests for rejection flow

**Checkpoint:**
User Story 3 should be independently functional after this task.

---

## Phase 6: User Story 4 - DNS Reconciliation (Priority: P3)

### Issue 13: T020 - Implement reconciliation scheduler

**Title:** [T020][US4] Implement reconciliation scheduler with configurable interval

**Labels:** `enhancement`, `user-story-4`, `phase-6`, `priority-medium`, `reconciliation`

**Description:**
Implement reconciliation scheduler with configurable interval and backoff strategy.

**Task Details:**
- Task ID: T020
- User Story: US4 - DNS Reconciliation
- Phase: Phase 6
- File: `internal/reconcile/scheduler.go`
- Dependencies: Requires foundational phase and DNS manager (T008)

**Acceptance Criteria:**
- [ ] Configurable reconciliation interval
- [ ] Backoff strategy for failures
- [ ] Graceful startup and shutdown
- [ ] Integration with DNS manager
- [ ] Unit tests for scheduler logic
- [ ] Integration tests for timing

**Independent Test:**
Manually delete a DNS record; reconciliation loop recreates it within interval.

---

### Issue 14: T021 - Compute desired vs actual record diffs

**Title:** [T021][US4][P] Compute desired vs actual record diffs and apply

**Labels:** `enhancement`, `user-story-4`, `phase-6`, `priority-medium`, `reconciliation`, `parallel-ok`

**Description:**
Compute differences between desired and actual DNS records and apply corrections.

**Task Details:**
- Task ID: T021
- User Story: US4 - DNS Reconciliation
- Phase: Phase 6
- File: `internal/reconcile/diff.go`
- Can run in parallel: Yes (different files, no dependencies)
- Dependencies: Requires foundational phase and provider implementations

**Acceptance Criteria:**
- [ ] Diff computation for missing records
- [ ] Diff computation for extra records
- [ ] Diff computation for modified records
- [ ] Apply logic to correct drift
- [ ] Support for all providers
- [ ] Unit tests for diff logic
- [ ] Integration tests for drift correction

---

### Issue 15: T022 - Integrate reconciliation triggers with metrics

**Title:** [T022][US4] Integrate reconciliation triggers with metrics and logs

**Labels:** `enhancement`, `user-story-4`, `phase-6`, `priority-medium`, `reconciliation`, `observability`

**Description:**
Integrate reconciliation triggers with metrics collection and structured logging.

**Task Details:**
- Task ID: T022
- User Story: US4 - DNS Reconciliation
- Phase: Phase 6
- File: `internal/reconcile/metrics.go`
- Dependencies: Requires T020, T021 to be complete

**Acceptance Criteria:**
- [ ] Reconciliation metrics exposed
- [ ] Metrics for drift detected
- [ ] Metrics for corrections applied
- [ ] Structured logging for reconciliation events
- [ ] Integration with health endpoint
- [ ] Unit tests for metrics

**Checkpoint:**
User Story 4 should be independently functional after this task.

---

## Phase 7: User Story 5 - Configuration Flexibility (Priority: P3)

### Issue 16: T023 - Implement environment variable parsing overrides

**Title:** [T023][US5] Implement environment variable parsing overrides

**Labels:** `enhancement`, `user-story-5`, `phase-7`, `priority-medium`, `configuration`

**Description:**
Implement environment variable parsing with override capability for configuration values.

**Task Details:**
- Task ID: T023
- User Story: US5 - Configuration Flexibility
- Phase: Phase 7
- File: `internal/config/config.go`
- Dependencies: Requires T004 to be complete

**Acceptance Criteria:**
- [ ] Environment variable parsing implemented
- [ ] Support for all configuration parameters
- [ ] Validation of environment variable values
- [ ] Documentation of supported env vars
- [ ] Unit tests for env parsing
- [ ] Integration tests for env-only config

**Independent Test:**
Run plugin with env-only config and verify it works correctly.

---

### Issue 17: T024 - Ensure Caddyfile parsing precedence

**Title:** [T024][US5][P] Ensure Caddyfile parsing precedence over env

**Labels:** `enhancement`, `user-story-5`, `phase-7`, `priority-medium`, `configuration`, `parallel-ok`

**Description:**
Ensure Caddyfile configuration takes precedence over environment variables.

**Task Details:**
- Task ID: T024
- User Story: US5 - Configuration Flexibility
- Phase: Phase 7
- File: `internal/config/config.go`
- Can run in parallel: Yes (different files, no dependencies)
- Dependencies: Requires T004, T023 to be complete

**Acceptance Criteria:**
- [ ] Caddyfile values override environment variables
- [ ] Clear precedence logic documented in code
- [ ] Unit tests for precedence rules
- [ ] Integration tests with both config sources

**Independent Test:**
Run with both env and Caddyfile ensuring Caddyfile wins.

---

### Issue 18: T025 - Document configuration precedence

**Title:** [T025][US5] Document configuration precedence in quickstart

**Labels:** `documentation`, `user-story-5`, `phase-7`, `priority-medium`, `configuration`

**Description:**
Document configuration precedence rules in the quickstart guide.

**Task Details:**
- Task ID: T025
- User Story: US5 - Configuration Flexibility
- Phase: Phase 7
- File: `specs/001-caddy-dns-module/quickstart.md`
- Dependencies: Requires T023, T024 to be complete

**Acceptance Criteria:**
- [ ] Configuration precedence clearly explained
- [ ] Examples of both configuration methods
- [ ] Environment variable reference table
- [ ] Troubleshooting section for config issues

**Checkpoint:**
User Story 5 should be independently functional after this task.

---

## Phase N: Polish & Cross-Cutting Concerns

### Issue 19: T026 - Harden structured logging and error messages

**Title:** [T026] Harden structured logging and error messages across providers

**Labels:** `enhancement`, `phase-polish`, `priority-low`, `observability`

**Description:**
Improve structured logging and error messages across all provider implementations.

**Task Details:**
- Task ID: T026
- Phase: Polish (Phase N)
- File: `internal/dns/manager.go` (and other files as needed)
- Dependencies: Requires all provider implementations to be complete

**Acceptance Criteria:**
- [ ] Consistent log format across all components
- [ ] Structured logging with context
- [ ] Clear error messages with actionable guidance
- [ ] Log levels appropriately set
- [ ] No sensitive data in logs

---

### Issue 20: T027 - Optimize startup path

**Title:** [T027] Optimize startup path to stay under 2s

**Labels:** `enhancement`, `phase-polish`, `priority-low`, `performance`

**Description:**
Optimize the plugin startup path to ensure initialization completes in under 2 seconds.

**Task Details:**
- Task ID: T027
- Phase: Polish (Phase N)
- File: `cmd/caddy-dns-sync/main.go`
- Dependencies: Requires all foundational components to be complete

**Acceptance Criteria:**
- [ ] Startup time measured and tracked
- [ ] Lazy initialization where appropriate
- [ ] Parallel initialization of independent components
- [ ] Startup time consistently under 2s
- [ ] Performance benchmarks added

---

### Issue 21: T028 - Update documentation with metrics and reconciliation

**Title:** [T028] Update documentation with metrics endpoints and reconcile usage

**Labels:** `documentation`, `phase-polish`, `priority-low`, `observability`

**Description:**
Update the quickstart documentation with information about metrics endpoints and reconciliation features.

**Task Details:**
- Task ID: T028
- Phase: Polish (Phase N)
- File: `specs/001-caddy-dns-module/quickstart.md`
- Dependencies: Requires T009, T020-T022 to be complete

**Acceptance Criteria:**
- [ ] Metrics endpoint usage documented
- [ ] Available metrics listed and explained
- [ ] Reconciliation feature explained
- [ ] Configuration examples for reconciliation
- [ ] Monitoring best practices included

---

## Summary

**Total Issues to Create:** 21

**Priority Breakdown:**
- Critical (MVP): 4 issues (T010-T013)
- High: 7 issues (T008, T009, T014-T017, T019)
- Medium: 7 issues (T018, T020-T025)
- Low: 3 issues (T026-T028)

**Phase Breakdown:**
- Phase 2 (Foundational): 2 issues
- Phase 3 (User Story 1 - MVP): 4 issues
- Phase 4 (User Story 2): 3 issues
- Phase 5 (User Story 3): 3 issues
- Phase 6 (User Story 4): 3 issues
- Phase 7 (User Story 5): 3 issues
- Phase N (Polish): 3 issues

**Dependencies:**
- Complete Phase 2 (T008, T009) before starting any user stories
- User Story 1 (Phase 3) is the MVP and should be prioritized
- User Stories 2 and 3 can be done in parallel after Phase 2
- User Story 4 depends on User Stories 1 and 2
- User Story 5 can be done in parallel with other stories after Phase 2
- Polish tasks should be done last

**Next Steps:**
1. Create these issues in the GitHub repository
2. Add appropriate labels and milestones
3. Assign to team members based on parallel strategy
4. Link dependent issues using GitHub issue references
5. Create a project board to track progress across user stories
