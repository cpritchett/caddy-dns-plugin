package labels

import "testing"

func TestParseLabelEnableProviderHostname(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.enable":   "true",
		"caddy_dns.provider": "cloudflare-public",
		"caddy_dns.hostname": "app.example.com",
	}

	parsed, err := Parse("caddy_dns", labels)
	if err != nil {
		t.Fatalf("parse labels: %v", err)
	}

	if !parsed.Enabled {
		t.Fatalf("expected Enabled true")
	}
	if parsed.Provider != "cloudflare-public" {
		t.Fatalf("provider = %q", parsed.Provider)
	}
	if parsed.Hostname != "app.example.com" {
		t.Fatalf("hostname = %q", parsed.Hostname)
	}
}

func TestParseInfersHostnameFromCaddy(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.enable": "true",
		"caddy":            "example.com, www.example.com",
	}

	parsed, err := Parse("caddy_dns", labels)
	if err != nil {
		t.Fatalf("parse labels: %v", err)
	}

	if parsed.Hostname != "example.com" {
		t.Fatalf("hostname = %q", parsed.Hostname)
	}
	if !parsed.HasCaddy {
		t.Fatalf("expected HasCaddy true")
	}
	if parsed.CaddyHostname != "example.com" {
		t.Fatalf("caddy hostname = %q", parsed.CaddyHostname)
	}
}

func TestParseEnableInferredFromProvider(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.provider": "unifi-internal",
	}

	parsed, err := Parse("caddy_dns", labels)
	if err != nil {
		t.Fatalf("parse labels: %v", err)
	}

	if !parsed.Enabled {
		t.Fatalf("expected Enabled true when provider is set")
	}
}

func TestParseEnableExplicitFalse(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.enable":   "false",
		"caddy_dns.provider": "unifi-internal",
	}

	parsed, err := Parse("caddy_dns", labels)
	if err != nil {
		t.Fatalf("parse labels: %v", err)
	}

	if parsed.Enabled {
		t.Fatalf("expected Enabled false when explicitly disabled")
	}
}

func TestParseInvalidEnable(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.enable": "nope",
	}

	_, err := Parse("caddy_dns", labels)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParsePrefixNormalization(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.enable": "true",
	}

	parsed, err := Parse("caddy_dns.", labels)
	if err != nil {
		t.Fatalf("parse labels: %v", err)
	}
	if !parsed.Enabled {
		t.Fatalf("expected Enabled true")
	}
}

func TestParseEmptyHostnameLabel(t *testing.T) {
	labels := map[string]string{
		"caddy_dns.hostname": " ",
	}

	_, err := Parse("caddy_dns", labels)
	if err == nil {
		t.Fatalf("expected error")
	}
}
