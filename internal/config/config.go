package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

const (
	defaultLabelPrefix       = "caddy_dns"
	defaultDockerSocket      = "/var/run/docker.sock"
	defaultReconcileInterval = 5 * time.Minute
)

type Config struct {
	LabelPrefix       string           `json:"label_prefix,omitempty"`
	ReconcileInterval caddy.Duration   `json:"reconcile_interval,omitempty"`
	DockerSocket      string           `json:"docker_socket,omitempty"`
	Providers         []ProviderConfig `json:"providers,omitempty"`
}

type ProviderConfig struct {
	Name          string   `json:"name,omitempty"`
	Type          string   `json:"type,omitempty"`
	ZoneFilters   []string `json:"zone_filters,omitempty"`
	TTL           *int     `json:"ttl,omitempty"`
	Proxied       *bool    `json:"proxied,omitempty"`
	Token         string   `json:"token,omitempty"`
	ControllerURL string   `json:"controller_url,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		LabelPrefix:       defaultLabelPrefix,
		ReconcileInterval: caddy.Duration(defaultReconcileInterval),
		DockerSocket:      defaultDockerSocket,
	}
}

func Load(dispenser *caddyfile.Dispenser) (Config, error) {
	cfg := DefaultConfig()
	if err := cfg.ApplyEnv(); err != nil {
		return Config{}, err
	}
	if dispenser != nil {
		if err := cfg.UnmarshalCaddyfile(dispenser); err != nil {
			return Config{}, err
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) ApplyEnv() error {
	if value, ok := os.LookupEnv("CADDY_DNS_LABEL_PREFIX"); ok && value != "" {
		c.LabelPrefix = value
	}

	if value, ok := os.LookupEnv("CADDY_DNS_RECONCILE_INTERVAL"); ok && value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("parse CADDY_DNS_RECONCILE_INTERVAL: %w", err)
		}
		c.ReconcileInterval = caddy.Duration(duration)
	}

	if value, ok := os.LookupEnv("CADDY_DNS_DOCKER_SOCKET"); ok && value != "" {
		c.DockerSocket = value
	}

	return nil
}

func (c *Config) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "label_prefix":
				value, err := parseSingleArg(d)
				if err != nil {
					return err
				}
				c.LabelPrefix = value
			case "reconcile_interval":
				value, err := parseSingleArg(d)
				if err != nil {
					return err
				}
				duration, err := time.ParseDuration(value)
				if err != nil {
					return d.Errf("invalid reconcile_interval %q: %v", value, err)
				}
				c.ReconcileInterval = caddy.Duration(duration)
			case "docker_socket":
				value, err := parseSingleArg(d)
				if err != nil {
					return err
				}
				c.DockerSocket = value
			case "provider":
				provider, err := parseProviderBlock(d)
				if err != nil {
					return err
				}
				c.Providers = append(c.Providers, provider)
			default:
				return d.Errf("unrecognized subdirective %q", d.Val())
			}
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.LabelPrefix) == "" {
		return fmt.Errorf("label_prefix must not be empty")
	}
	if time.Duration(c.ReconcileInterval) <= 0 {
		return fmt.Errorf("reconcile_interval must be positive")
	}
	if strings.TrimSpace(c.DockerSocket) == "" {
		return fmt.Errorf("docker_socket must not be empty")
	}

	seen := make(map[string]struct{})
	for i, provider := range c.Providers {
		if strings.TrimSpace(provider.Name) == "" {
			return fmt.Errorf("provider[%d] missing name", i)
		}
		if strings.TrimSpace(provider.Type) == "" {
			return fmt.Errorf("provider %q missing type", provider.Name)
		}
		if _, ok := seen[provider.Name]; ok {
			return fmt.Errorf("duplicate provider name %q", provider.Name)
		}
		seen[provider.Name] = struct{}{}
		if len(provider.ZoneFilters) == 0 {
			return fmt.Errorf("provider %q requires at least one zone_filter", provider.Name)
		}
		if provider.TTL != nil && *provider.TTL <= 0 {
			return fmt.Errorf("provider %q ttl must be positive", provider.Name)
		}
	}

	return nil
}

func parseSingleArg(d *caddyfile.Dispenser) (string, error) {
	if !d.NextArg() {
		return "", d.ArgErr()
	}
	value := d.Val()
	if d.NextArg() {
		return "", d.ArgErr()
	}
	return value, nil
}

func parseProviderBlock(d *caddyfile.Dispenser) (ProviderConfig, error) {
	args := d.RemainingArgs()
	if len(args) < 2 {
		return ProviderConfig{}, d.ArgErr()
	}
	provider := ProviderConfig{
		Name: args[0],
		Type: args[1],
	}
	if len(args) > 2 {
		provider.ZoneFilters = append(provider.ZoneFilters, args[2:]...)
	}

	for d.NextBlock(d.Nesting()) {
		switch d.Val() {
		case "zone_filters":
			filters := d.RemainingArgs()
			if len(filters) == 0 {
				return ProviderConfig{}, d.ArgErr()
			}
			provider.ZoneFilters = append([]string{}, filters...)
		case "token":
			value, err := parseSingleArg(d)
			if err != nil {
				return ProviderConfig{}, err
			}
			provider.Token = value
		case "controller_url":
			value, err := parseSingleArg(d)
			if err != nil {
				return ProviderConfig{}, err
			}
			provider.ControllerURL = value
		case "username":
			value, err := parseSingleArg(d)
			if err != nil {
				return ProviderConfig{}, err
			}
			provider.Username = value
		case "password":
			value, err := parseSingleArg(d)
			if err != nil {
				return ProviderConfig{}, err
			}
			provider.Password = value
		case "ttl":
			value, err := parseSingleArg(d)
			if err != nil {
				return ProviderConfig{}, err
			}
			ttl, err := strconv.Atoi(value)
			if err != nil {
				return ProviderConfig{}, d.Errf("invalid ttl %q: %v", value, err)
			}
			provider.TTL = &ttl
		case "proxied":
			value, err := parseSingleArg(d)
			if err != nil {
				return ProviderConfig{}, err
			}
			proxied, err := strconv.ParseBool(value)
			if err != nil {
				return ProviderConfig{}, d.Errf("invalid proxied %q: %v", value, err)
			}
			provider.Proxied = &proxied
		default:
			return ProviderConfig{}, d.Errf("unrecognized provider option %q", d.Val())
		}
	}

	return provider, nil
}
