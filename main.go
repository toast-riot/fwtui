package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

type menuItem struct {
	title  string
	action func() string
}

// VIEW state
type viewState string

func (v viewState) isPortEdit() bool {
	return strings.HasPrefix(string(v), "port_edit_")
}

func (v viewState) isHome() bool {
	return v == home
}

func (v viewState) isDeleteRule() bool {
	return v == deleteRule
}

const home = "home"
const allowPort = "port_edit_allow"
const denyPort = "port_edit_deny"
const deleteRule = "delete_rule"

// ACTIONS
const menuResetUFW = "RESET_UFW"
const menuDisableUFW = "DISABLE"
const menuEnableUFW = "ENABLE"
const menuAllowPort = "ALLOW_PORT"
const menuDenyPort = "DENY_PORT"
const menuDeleteRule = "DELETE_RULE"
const menuDisableLogging = "DISABLE_LOGGING"
const menuEnableLogging = "ENABLE_LOGGING"

type rule struct {
	number int
	line   string
}

type model struct {
	cursor int
	menu   []menuItem
	status string
	rules  []rule

	view      viewState
	inputPort string
}

func main() {
	m := model{menu: buildMenu(), view: home}
	reloadRules(&m)
	reloadStatus(&m)

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
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		switch true {
		case m.view.isHome():
			switch key {
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
				case menuResetUFW:
					runCommand("sudo ufw reset")()
					resetMenu(&m)
				case menuDisableUFW:
					runCommand("sudo ufw disable")()
					resetMenu(&m)
					m.cursor = 0
				case menuEnableUFW:
					runCommand("sudo ufw enable")()
					resetMenu(&m)
				case menuEnableLogging:
					runCommand("sudo ufw logging on")()
					resetMenu(&m)
				case menuDisableLogging:
					runCommand("sudo ufw logging off")()
					resetMenu(&m)
				case menuAllowPort:
					m.view = allowPort
				case menuDenyPort:
					m.view = denyPort
				case menuDeleteRule:
					m.view = deleteRule
					m.cursor = 0
				}
			}
		case m.view.isPortEdit():
			switch key {
			case "enter":
				port, err := strconv.Atoi(m.inputPort)
				if err == nil && port > 0 && port < 65536 {
					var cmd string
					switch m.view {
					case allowPort:
						cmd = fmt.Sprintf("sudo ufw allow %d", port)
					case denyPort:
						cmd = fmt.Sprintf("sudo ufw deny %d", port)
					}
					runCommand(cmd)()
					reloadStatus(&m)
					m.view = home
					m.inputPort = ""
					reloadRules(&m)
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

		case m.view.isDeleteRule():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.rules)-1 {
					m.cursor++
				}
			case "enter":
				// TODO select multiple by space and then delete
				runCommand(fmt.Sprintf("yes | sudo ufw delete %d", m.cursor+1))()
				reloadRules(&m)
			case "esc":
				m.view = home
				m.cursor = 0
				reloadStatus(&m)
			}
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
		}
		return fmt.Sprintf("Enter port to %s: %s\n(Press Enter to confirm)", prefix, m.inputPort)
	case m.view.isDeleteRule():
		lines := []string{"Select rule to delete:"}
		for i, rule := range m.rules {
			prefix := " "
			if i == m.cursor {
				prefix = ">"
			}
			lines = append(lines, fmt.Sprintf("%s %s", prefix, rule.line))
		}
		return strings.Join(lines, "\n")
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
		items = append(items, menuItem{"Disable", func() string { return menuDisableUFW }})
		items = append(items,
			menuItem{"Allow port...", func() string { return menuAllowPort }},
			menuItem{"Deny port...", func() string { return menuDenyPort }},
			menuItem{"Delete rule", func() string { return menuDeleteRule }},
		)
		if loggingOn {
			items = append(items, menuItem{"Disable logging", func() string { return menuDisableLogging }})
		} else {
			items = append(items, menuItem{"Enable logging", func() string { return menuEnableLogging }})
		}
	} else {
		items = append(items, menuItem{"Enable", func() string { return menuEnableUFW }})

	}

	items = append(items,
		menuItem{"Reset UFW", func() string { return menuResetUFW }},
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
	reloadStatus(m)
}

func reloadStatus(m *model) {
	m.status = runCommand("sudo ufw status verbose")()
}

func runCommand(cmdStr string) func() string {
	return func() string {
		cmd := exec.Command("bash", "-c", cmdStr) // Use a shell to interpret the pipe
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %s\n%s", err, out)
		}
		return string(out)
	}
}

func reloadRules(m *model) {
	output := runCommand("sudo ufw status numbered")()
	lines := strings.Split(output, "\n")
	lines = lines[4:(len(lines) - 2)]
	rules := lo.Map(lines, func(line string, index int) rule {
		return rule{
			number: index,
			line:   line,
		}
	})
	m.rules = rules
}
