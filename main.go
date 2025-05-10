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

type viewState string

func (v viewState) isPortEdit() bool {
	return strings.HasPrefix(string(v), "port_edit_")
}

func (v viewState) isHome() bool {
	return v == home
}

const home = "home"
const allowPort = "port_edit_allow"
const denyPort = "port_edit_deny"
const deleteAllowPort = "port_edit_delete_allow"
const deleteDenyPort = "port_edit_delete_deny"

type model struct {
	cursor int
	menu   []menuItem
	status string

	view      viewState
	inputPort string
}

func main() {
	m := model{menu: buildMenu(), view: home, status: runCommand("sudo ufw status verbose")()}
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
		switch true {
		case m.view.isHome():
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
					m.view = allowPort
				case "DENY_PORT":
					m.view = denyPort
				case "DELETE_ALLOW_PORT":
					m.view = deleteAllowPort
				case "DELETE_DENY_PORT":
					m.view = deleteDenyPort
				}
			}
		case m.view.isPortEdit():
			switch key {
			case "enter":
				port, err := strconv.Atoi(m.inputPort)
				if err == nil && port > 0 && port < 65536 {
					var cmd string
					switch m.view {
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
					m.view = home
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
				m.view = home
				m.inputPort = ""
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	switch true {
	case m.view.isHome():
		left := renderMenu(m.menu, m.cursor)
		right := strings.Split(m.status, "\n")
		return renderTwoColumns(left, right)
	case m.view.isPortEdit():
		var prefix string
		switch m.view {
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

	return "Unknown state"
}

func renderMenu(menu []menuItem, cursor int) []string {
	var lines []string
	lines = append(lines, "", "UFW Firewall Menu:", "")
	for i, item := range menu {
		prefix := " "
		if i == cursor {
			prefix = ">"
		}
		lines = append(lines, fmt.Sprintf("%s %s", prefix, item.title))
	}
	return lines
}

func renderTwoColumns(left []string, right []string) string {
	var b strings.Builder
	maxLines := max(len(left), len(right))
	for i := 0; i < maxLines; i++ {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		fmt.Fprintf(&b, "%-30s | %s\n", l, r)
	}
	return b.String()
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
