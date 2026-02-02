# Quickstart: Caddy DNS Sync Module

## Prerequisites
- Go 1.22.x toolchain installed (matches research decision)
- Docker Engine with socket access at `/var/run/docker.sock`
- Cloudflare API token (DNS edit) and/or UniFi controller credentials

## Build & Load
1) Fetch deps and build the module:
```
go get github.com/caddyserver/caddy/v2
go get github.com/libdns/cloudflare
go get github.com/libdns/unifi
go build ./...
```
2) Register the module in Caddy (example `caddy.json` snippet):
```json
{
  "apps": {
    "http": {},
    "dns_sync": {
      "module": "caddy.dns.sync",
      "providers": [
        {"name": "cloudflare-public", "type": "cloudflare", "token": "${CLOUDFLARE_API_TOKEN}", "zone_filters": ["*.example.com"], "proxied": true},
        {"name": "unifi-internal", "type": "unifi", "controller_url": "https://unifi.local:8443", "username": "${UNIFI_USER}", "password": "${UNIFI_PASS}", "zone_filters": ["*.home.lab"]}
      ],
      "label_prefix": "caddy_dns",
      "reconcile_interval": "5m"
    }
  }
}
```

## Configure via environment
- `CADDY_DNS_LABEL_PREFIX` (default `caddy_dns`)
- `CADDY_DNS_RECONCILE_INTERVAL` (e.g., `5m`)
- Provider credentials (e.g., `CLOUDFLARE_API_TOKEN`, `UNIFI_USER`, `UNIFI_PASS`)

## Deploy a container with labels
```
```
Expected: A/AAAA record created in Cloudflare within 30s; removing the container deletes the record.

## Manual reconciliation
Trigger on-demand reconciliation (if control API exposed):
```
```

## Observability
- Health: `GET /health`
- Metrics: `GET /metrics` (Prometheus format)
- Logs: Caddy logs should show label parsing, validation warnings, and provider actions.
