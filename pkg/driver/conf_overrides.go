package driver

import (
	"fmt"
	"strings"
)

// ConfOverride represents a single section:key=value config override.
type ConfOverride struct {
	Section string
	Key     string
	Value   string
}

// parseConfOverrides parses a comma-separated "section:key=value" string into ConfOverride structs.
func parseConfOverrides(raw string) ([]ConfOverride, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var overrides []ConfOverride
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid override entry %q: missing colon separator", entry)
		}
		rest := parts[1]
		kv := strings.SplitN(rest, "=", 2)
		if len(kv) != 2 || kv[0] == "" {
			return nil, fmt.Errorf("invalid override entry %q: missing equals separator", entry)
		}
		overrides = append(overrides, ConfOverride{
			Section: strings.TrimSpace(parts[0]),
			Key:     strings.TrimSpace(kv[0]),
			Value:   strings.TrimSpace(kv[1]),
		})
	}
	return overrides, nil
}

// applyConfOverrides applies parsed overrides to an INI-style config string.
func applyConfOverrides(config string, overrides []ConfOverride) string {
	if len(overrides) == 0 {
		return config
	}
	lines := strings.Split(config, "\n")
	for _, o := range overrides {
		sectionHeader := "[" + o.Section + "]"
		sectionIdx := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == sectionHeader {
				sectionIdx = i
				break
			}
		}
		if sectionIdx < 0 {
			continue
		}

		nextSectionIdx := len(lines)
		for i := sectionIdx + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				nextSectionIdx = i
				break
			}
		}

		replaced := false
		for i := sectionIdx + 1; i < nextSectionIdx; i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			eqPos := strings.Index(trimmed, "=")
			if eqPos < 0 {
				continue
			}
			existingKey := strings.TrimSpace(trimmed[:eqPos])
			if existingKey == o.Key {
				lines[i] = o.Key + " = " + o.Value
				replaced = true
				break
			}
		}

		if !replaced {
			newLine := o.Key + " = " + o.Value
			lines = append(lines[:nextSectionIdx], append([]string{newLine}, lines[nextSectionIdx:]...)...)
		}
	}
	return strings.Join(lines, "\n")
}
