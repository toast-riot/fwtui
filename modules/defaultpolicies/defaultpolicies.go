package defaultpolicies

import (
	"fmt"
	"fwtui/utils/result"
	"strings"
)

type DefaultPolicies struct {
	Incoming string
	Outgoing string
	Routed   string
}

// ParseUfwDefaults runs `ufw status verbose` and extracts default policies.
func ParseUfwDefaults(status string) result.Result[DefaultPolicies] {
	lines := strings.Split(status, "\n")
	var policies DefaultPolicies
	found := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Default:") {
			found = true
			// Example: "Default: deny (incoming), allow (outgoing), disabled (routed)"
			line = strings.TrimPrefix(line, "Default:")
			parts := strings.Split(line, ",")

			for _, part := range parts {
				part = strings.TrimSpace(part)
				switch {
				case strings.Contains(part, "incoming"):
					policies.Incoming = extractPolicy(part)
				case strings.Contains(part, "outgoing"):
					policies.Outgoing = extractPolicy(part)
				case strings.Contains(part, "routed"):
					policies.Routed = extractPolicy(part)
				}
			}
			break
		}
	}

	if !found {
		return result.Err[DefaultPolicies](fmt.Errorf("default policy line not found in ufw output"))
	}

	return result.Ok(policies)
}

func extractPolicy(text string) string {
	return strings.Split(text, " ")[0]
}
