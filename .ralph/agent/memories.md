# Memories

## Patterns

### mem-1770046804-3470
> Config loader validates non-empty label_prefix/docker_socket, positive reconcile_interval, unique provider names, requires zone_filters, and TTL > 0.
<!-- tags: config, validation | created: 2026-02-02 -->

## Decisions

## Fixes

### mem-1770047308-0f05
> failure: cmd=go test ./..., exit=1, error=missing docker transitive deps (opencontainers image-spec, docker go-connections/go-units, pkg/errors, gogo/protobuf), next=go get docker dependencies or avoid docker SDK in watcher
<!-- tags: go, dependencies, testing | created: 2026-02-02 -->

### mem-1770047284-1011
> failure: cmd=go test ./..., exit=1, error=missing github.com/docker/docker/api/types packages, next=go get github.com/docker/docker@<version> to add docker SDK
<!-- tags: go, dependencies, testing | created: 2026-02-02 -->

### mem-1770031829-0874
> Go tests with caddy v2.9.0 fail unless github.com/quic-go/quic-go is pinned to v0.48.2; go mod tidy pulls v0.54.0 and breaks http3 types.
<!-- tags: go, dependencies, testing | created: 2026-02-02 -->

## Context
