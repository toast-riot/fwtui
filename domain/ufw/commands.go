package ufw

import (
	"fmt"
	"fwtui/utils/oscmd"
	"os"
)

func StatusVerbose() string {
	return oscmd.RunCommand("sudo ufw status verbose")
}

func StatusNumbered() string {
	return oscmd.RunCommand("sudo ufw status numbered")
}

func Reset() string {
	return oscmd.RunCommand("yes | sudo ufw reset")
}

func Enable() string {
	return oscmd.RunCommand("sudo ufw enable")
}
func Disable() string {
	return oscmd.RunCommand("sudo ufw disable")
}

func EnableLogging() string {
	return oscmd.RunCommand("sudo ufw logging on")
}
func DisableLogging() string {
	return oscmd.RunCommand("sudo ufw logging off")
}

func DeleteRuleByNumber(num int) string {
	return oscmd.RunCommand(fmt.Sprintf("yes | sudo ufw delete %d", num))
}

func LoadProfile(name string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw app update \"%s\"", name))
}

func GetProfileInfo(name string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw app info \"%s\"", name))
}

func GetProfileList() string {
	return oscmd.RunCommand("sudo ufw app list")
}

func AllowProfile(name string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw allow \"%s\"", name))
}

func SetDefaultPolicy(direction, action string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw default %s %s", action, direction))
}

func GetStateFromFiles() (string, error) {
	rulesV4, err := os.ReadFile("/etc/ufw/user.rules")
	if err != nil {
		return "", fmt.Errorf("reading user.rules: %w", err)
	}

	rulesV6, err := os.ReadFile("/etc/ufw/user6.rules")
	if err != nil {
		return "", fmt.Errorf("reading user6.rules: %w", err)
	}

	script := `#!/bin/bash
set -e

echo "Restoring UFW rules..."

cat <<'EOF' > /etc/ufw/user.rules
` + string(rulesV4) + `
EOF

cat <<'EOF' > /etc/ufw/user6.rules
` + string(rulesV6) + `
EOF

ufw --force reload
echo "UFW rules restored."
`
	return script, nil
}
