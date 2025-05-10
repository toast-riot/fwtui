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

const allowPort = "allow"
const denyPort = "deny"
const deleteAllowPort = "delete_allow"
const deleteDenyPort = "delete_deny"

type model struct {
	cursor int
	menu   []menuItem
	status string

	editing   string
	inputPort string
}

func main() {
	m := model{menu: buildMenu(), status: runCommand("sudo ufw status verbose")()}
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if m.editing != "" {
			switch key {
			case "enter":
				port, err := strconv.Atoi(m.inputPort)
				if err == nil && port > 0 && port < 65536 {
					var cmd string
					switch m.editing {
					case deleteAllowPort:
						cmd = fmt.Sprintf("sudo ufw delete allow %d", port)
					case deleteDenyPort:
						cmd = fmt.Sprintf("sudo ufw delete deny %d", port)
					case allowPort:
						cmd = fmt.Sprintf("sudo ufw allow %d", port)
					case denyPort:
						cmd = fmt.Sprintf("sudo ufw deny %d", port)
					}
					runCommand(cmd)()
					m.status = runCommand("sudo ufw status verbose")()
					m.editing = ""
					m.inputPort = ""
				} else {
					m.status = "Invalid port number"
				}

			case "backspace":
				if len(m.inputPort) > 0 {
					m.inputPort = m.inputPort[:len(m.inputPort)-1]
				}
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				m.inputPort += key
			case "esc":
				m.editing = ""
				m.inputPort = ""
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
			case "RESET_UFW":
				runCommand("sudo ufw reset")()
				resetMenu(&m)
			case "DISABLE":
				runCommand("sudo ufw disable")()
				resetMenu(&m)
				m.cursor = 0
			case "ENABLE":
				runCommand("sudo ufw enable")()
				resetMenu(&m)
			case "ENABLE_LOGGING":
				runCommand("sudo ufw logging on")()
				resetMenu(&m)
			case "DISABLE_LOGGING":
				runCommand("sudo ufw logging off")()
				resetMenu(&m)
			case "ALLOW_PORT":
				m.editing = allowPort
				resetStatus(&m)
			case "DENY_PORT":
				m.editing = denyPort
				resetStatus(&m)
			case "DELETE_ALLOW_PORT":
				m.editing = deleteAllowPort
				resetStatus(&m)
			case "DELETE_DENY_PORT":
				m.editing = deleteDenyPort
				resetStatus(&m)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.editing != "" {
		var prefix string
		switch m.editing {
		case allowPort:
			prefix = "allow"
		case denyPort:
			prefix = "deny"
		case deleteAllowPort:
			prefix = "delete allow"
		case deleteDenyPort:
			prefix = "delete deny"
		}
		return fmt.Sprintf("Enter port to %s: %s\n(Press Enter to confirm)", prefix, m.inputPort)
	}

	// Create two columns: left for menu, right for status
	s := "\nUFW Firewall Menu:\n\n"
	for i, item := range m.menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, item.title)
	}

	// Split into two columns
	linesLeft := strings.Split(s, "\n")
	linesRight := strings.Split(m.status, "\n")
	maxLines := len(linesLeft)
	if len(linesRight) > maxLines {
		maxLines = len(linesRight)
	}

	view := ""
	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(linesLeft) {
			left = linesLeft[i]
		}
		if i < len(linesRight) {
			right = linesRight[i]
		}
		view += fmt.Sprintf("%-30s | %s\n", left, right)
	}
	return view
}

func buildMenu() []menuItem {
	enabled, loggingOn := getStatus()

	items := []menuItem{}

	if enabled {
		items = append(items, menuItem{"Disable", func() string { return "DISABLE" }})
		items = append(items,
			menuItem{"Allow port...", func() string { return "ALLOW_PORT" }},
			menuItem{"Deny port...", func() string { return "DENY_PORT" }},
			menuItem{"Delete allow...", func() string { return "DELETE_ALLOW_PORT" }},
			menuItem{"Delete deny...", func() string { return "DELETE_DENY_PORT" }},
		)
		if loggingOn {
			items = append(items, menuItem{"Disable logging", func() string { return "DISABLE_LOGGING" }})
		} else {
			items = append(items, menuItem{"Enable logging", func() string { return "ENABLE_LOGGING" }})
		}
	} else {
		items = append(items, menuItem{"Enable", func() string { return "ENABLE" }})

	}

	items = append(items,
		menuItem{"Reset UFW", func() string { return "RESET_UFW" }},
		menuItem{"Quit", func() string { os.Exit(0); return "" }},
	)

	return items
}

func getStatus() (enabled bool, loggingOn bool) {
	output := runCommand("sudo ufw status verbose")()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Status: active") {
			enabled = true
		}
		if strings.HasPrefix(line, "Logging:") && strings.Contains(line, "on") {
			loggingOn = true
		}
	}
	return
}

func resetMenu(m *model) {
	m.menu = buildMenu()
	resetStatus(m)
}

func resetStatus(m *model) {
	m.status = runCommand("sudo ufw status verbose")()
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
