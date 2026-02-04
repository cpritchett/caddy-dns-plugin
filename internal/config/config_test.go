package config

import (
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestLoadUsesEnvWhenNoCaddyfile(t *testing.T) {
	t.Setenv("CADDY_DNS_LABEL_PREFIX", "env_prefix")
	t.Setenv("CADDY_DNS_RECONCILE_INTERVAL", "12m")
	t.Setenv("CADDY_DNS_DOCKER_SOCKET", "/tmp/docker.sock")

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.LabelPrefix != "env_prefix" {
		t.Fatalf("label prefix = %q, want %q", cfg.LabelPrefix, "env_prefix")
	}
	if time.Duration(cfg.ReconcileInterval) != 12*time.Minute {
		t.Fatalf("reconcile interval = %s, want %s", time.Duration(cfg.ReconcileInterval), 12*time.Minute)
	}
	if cfg.DockerSocket != "/tmp/docker.sock" {
		t.Fatalf("docker socket = %q, want %q", cfg.DockerSocket, "/tmp/docker.sock")
	}
}

func TestCaddyfileOverridesEnv(t *testing.T) {
	t.Setenv("CADDY_DNS_LABEL_PREFIX", "env_prefix")
	t.Setenv("CADDY_DNS_RECONCILE_INTERVAL", "12m")
	t.Setenv("CADDY_DNS_DOCKER_SOCKET", "/tmp/docker.sock")

	input := `dns_sync {
	label_prefix caddy_dns_custom
	reconcile_interval 30s
	docker_socket /var/run/custom.sock
	provider cloudflare-primary cloudflare *.example.com
}`

	d := caddyfile.NewTestDispenser(input)
	cfg, err := Load(d)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.LabelPrefix != "caddy_dns_custom" {
		t.Fatalf("label prefix = %q, want %q", cfg.LabelPrefix, "caddy_dns_custom")
	}
	if time.Duration(cfg.ReconcileInterval) != 30*time.Second {
		t.Fatalf("reconcile interval = %s, want %s", time.Duration(cfg.ReconcileInterval), 30*time.Second)
	}
	if cfg.DockerSocket != "/var/run/custom.sock" {
		t.Fatalf("docker socket = %q, want %q", cfg.DockerSocket, "/var/run/custom.sock")
	}

	if len(cfg.Providers) != 1 {
		t.Fatalf("providers = %d, want %d", len(cfg.Providers), 1)
	}
	provider := cfg.Providers[0]
	if provider.Name != "cloudflare-primary" {
		t.Fatalf("provider name = %q, want %q", provider.Name, "cloudflare-primary")
	}
	if provider.Type != "cloudflare" {
		t.Fatalf("provider type = %q, want %q", provider.Type, "cloudflare")
	}
	if len(provider.ZoneFilters) != 1 || provider.ZoneFilters[0] != "*.example.com" {
		t.Fatalf("zone filters = %v, want [*.example.com]", provider.ZoneFilters)
	}
}

func TestLoadRejectsProviderWithoutZoneFilters(t *testing.T) {
	input := `dns_sync {
	provider cloudflare-primary cloudflare
}`

	d := caddyfile.NewTestDispenser(input)
	_, err := Load(d)
	if err == nil {
		t.Fatal("expected error for provider without zone filters")
	}
}

func TestLoadRejectsDuplicateProviderNames(t *testing.T) {
	input := `dns_sync {
	provider cloudflare-primary cloudflare *.example.com
	provider cloudflare-primary cloudflare *.example.org
}`

	d := caddyfile.NewTestDispenser(input)
	_, err := Load(d)
	if err == nil {
		t.Fatal("expected error for duplicate provider names")
	}
}

func TestLoadRejectsNonPositiveTTL(t *testing.T) {
	input := `dns_sync {
	provider cloudflare-primary cloudflare *.example.com {
		ttl 0
	}
}`

	d := caddyfile.NewTestDispenser(input)
	_, err := Load(d)
	if err == nil {
		t.Fatal("expected error for non-positive ttl")
	}
}
