package ufw

import (
	"fmt"
	oscmd "fwtui/utils/cmd"
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
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw app update %s", name))
}

func GetProfileInfo(name string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw app info %s", name))
}

func GetProfileList() string {
	return oscmd.RunCommand("sudo ufw app list")
}

func AllowProfile(name string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw allow %s", name))
}

func SetDefaultPolicy(direction, action string) string {
	return oscmd.RunCommand(fmt.Sprintf("sudo ufw default %s %s", action, direction))
}
