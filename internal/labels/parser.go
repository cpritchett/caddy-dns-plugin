package labels

import (
	"fmt"
	"strconv"
	"strings"
)

type ParsedLabels struct {
	Enabled       bool
	Provider      string
	Hostname      string
	CaddyHostname string
	HasCaddy      bool
	Raw           map[string]string
}

func Parse(prefix string, labels map[string]string) (ParsedLabels, error) {
	normalized := normalizePrefix(prefix)
	if normalized == "" {
		return ParsedLabels{}, fmt.Errorf("label prefix must not be empty")
	}

	result := ParsedLabels{
		Raw: copyLabels(labels),
	}

	if labels == nil {
		labels = map[string]string{}
	}

	enableKey := normalized + ".enable"
	providerKey := normalized + ".provider"
	hostnameKey := normalized + ".hostname"

	enabledSet := false
	if value, ok := labels[enableKey]; ok {
		enabledSet = true
		parsed, err := parseBool(value)
		if err != nil {
			return ParsedLabels{}, fmt.Errorf("invalid %s value %q: %w", enableKey, value, err)
		}
		result.Enabled = parsed
	}

	if value, ok := labels[providerKey]; ok {
		result.Provider = strings.TrimSpace(value)
	}

	if value, ok := labels[hostnameKey]; ok {
		result.Hostname = strings.TrimSpace(value)
		if result.Hostname == "" {
			return ParsedLabels{}, fmt.Errorf("%s must not be empty", hostnameKey)
		}
	}

	if value, ok := labels["caddy"]; ok {
		result.HasCaddy = true
		result.CaddyHostname = inferHostnameFromCaddy(value)
		if result.Hostname == "" && result.CaddyHostname != "" {
			result.Hostname = result.CaddyHostname
		}
	}

	if !enabledSet && result.Provider != "" {
		result.Enabled = true
	}

	return result, nil
}

func normalizePrefix(prefix string) string {
	normalized := strings.TrimSpace(prefix)
	normalized = strings.TrimSuffix(normalized, ".")
	return normalized
}

func parseBool(value string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(value))
}

func inferHostnameFromCaddy(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	replaced := strings.ReplaceAll(trimmed, ",", " ")
	fields := strings.Fields(replaced)
	if len(fields) == 0 {
		return ""
	}

	host := strings.TrimPrefix(fields[0], "https://")
	host = strings.TrimPrefix(host, "http://")
	return host
}

func copyLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return map[string]string{}
	}

	copied := make(map[string]string, len(labels))
	for key, value := range labels {
		copied[key] = value
	}

	return copied
}
