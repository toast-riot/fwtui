package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type menuItem struct {
	title  string
	action func() string
}

type model struct {
	cursor        int
	menu          []menuItem
	output        string
	inputPort     string
	choosingPort  bool
	actionPrefix  string
	deletingAllow bool
	deletingDeny  bool
	viewingOutput bool
}

func main() {
	items := []menuItem{
		{"Status", runCommand("sudo ufw status")},
		{"Enable", runCommand("sudo ufw enable")},
		{"Disable", runCommand("sudo ufw disable")},
		{"Verbose Status", runCommand("sudo ufw status verbose")},
		{"Allow port...", func() string { return "ALLOW_PORT" }},
		{"Deny port...", func() string { return "DENY_PORT" }},
		{"Delete allow...", func() string { return "DELETE_ALLOW_PORT" }},
		{"Delete deny...", func() string { return "DELETE_DENY_PORT" }},
		{"Enable logging", runCommand("sudo ufw logging on")},
		{"Reset UFW", runCommand("sudo ufw reset")},
		{"Quit", func() string { os.Exit(0); return "" }},
	}

	m := model{menu: items}
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func runCommand(cmdStr string) func() string {
	return func() string {
		parts := strings.Fields(cmdStr)
		cmd := exec.Command(parts[0], parts[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %s\n%s", err, out)
		}
		return string(out)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if m.choosingPort {
			if key == "enter" {
				port, err := strconv.Atoi(m.inputPort)
				if err == nil && port > 0 && port < 65536 {
					var cmd string
					switch {
					case m.deletingAllow:
						cmd = fmt.Sprintf("sudo ufw delete allow %d", port)
					case m.deletingDeny:
						cmd = fmt.Sprintf("sudo ufw delete deny %d", port)
					default:
						cmd = fmt.Sprintf("sudo ufw %s %d", m.actionPrefix, port)
					}
					m.output = runCommand(cmd)()
				} else {
					m.output = "Invalid port number"
				}
				m.choosingPort = false
				m.inputPort = ""
				m.actionPrefix = ""
				m.deletingAllow = false
				m.deletingDeny = false
			} else if key == "backspace" || key == "ctrl+h" {
				if len(m.inputPort) > 0 {
					m.inputPort = m.inputPort[:len(m.inputPort)-1]
				}
			} else if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
				m.inputPort += key
			}
			return m, nil
		} else if m.viewingOutput {
			if key == "esc" {
				m.viewingOutput = false
				m.output = ""
			}
			return m, nil
		}

		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			selected := m.menu[m.cursor].action()
			switch selected {
			case "ALLOW_PORT":
				m.choosingPort = true
				m.actionPrefix = "allow"
			case "DENY_PORT":
				m.choosingPort = true
				m.actionPrefix = "deny"
			case "DELETE_ALLOW_PORT":
				m.choosingPort = true
				m.deletingAllow = true
			case "DELETE_DENY_PORT":
				m.choosingPort = true
				m.deletingDeny = true
			default:
				m.output = selected
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.choosingPort {
		prefix := m.actionPrefix
		if m.deletingAllow {
			prefix = "delete allow"
		} else if m.deletingDeny {
			prefix = "delete deny"
		}
		return fmt.Sprintf("Enter port to %s: %s\n(Press Enter to confirm)", prefix, m.inputPort)
	}

	if m.viewingOutput {
		return fmt.Sprintf("Output:\n%s\n\n(Press Esc to return to menu)", m.output)
	}

	s := "\nUFW Firewall Menu:\n\n"
	for i, item := range m.menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, item.title)
	}

	if m.output != "" {
		s += "\nOutput:\n" + m.output + "\n"
	}

	return s
}
