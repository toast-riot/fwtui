package main

import (
	"fmt"
	"fwtui/utils/set"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

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

func (v viewState) isProfilesHome() bool {
	return v == profilesHome
}

func (v viewState) isInstalledProfilesList() bool {
	return v == installedProfilesList
}

func (v viewState) isInstallProfile() bool {
	return v == installProfile
}

func (v viewState) isDeleteRule() bool {
	return v == deleteRule
}

const home = "home"

const profilesHome = "profiles_home"

const installedProfilesList = "installed_profiles_list"
const installProfile = "install_profile"

const allowPort = "port_edit_allow"
const denyPort = "port_edit_deny"
const deleteRule = "delete_rule"

// HOME MENU
const menuResetUFW = "RESET_UFW"
const menuDisableUFW = "DISABLE"
const menuEnableUFW = "ENABLE"
const menuAllowPort = "ALLOW_PORT"
const menuDenyPort = "DENY_PORT"
const menuDeleteRule = "DELETE_RULE"
const menuDisableLogging = "DISABLE_LOGGING"
const menuEnableLogging = "ENABLE_LOGGING"

// PROFILE MENU
var profileHomeActions = []string{menuInstalledProfiles, menuInstallProfile}

const menuProfiles = "PROFILES"
const menuInstalledProfiles = "INSTALLED_PROFILES"
const menuInstallProfile = "INSTALL_PROFILE"

// EVENT
const lastActionTimeUp = "LAST_ACTION_TIME_UP"

type rule struct {
	number int
	line   string
}

type model struct {
	cursor            int
	menu              []menuItem
	status            string
	rules             []rule
	installedProfiles []UFWProfile
	profilesToInstall []UFWProfile
	lastAction        string
	selectedItems     set.Set[int]

	view      viewState
	inputPort string
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This action requires root. Please run with sudo.")
		os.Exit(1)
	}

	m := model{menu: buildMenu(), selectedItems: set.NewSet[int](), view: home}
	m = m.reloadRules()
	m = m.reloadStatus()
	m = m.reloadInstalledProfiles()
	m = m.reloadProfilesToInstall()

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) resetMenu() model {
	m.menu = buildMenu()
	m = m.reloadStatus()
	return m
}

func (m model) reloadStatus() model {
	m.status = runCommand("sudo ufw status verbose")()
	return m
}

func (m model) reloadInstalledProfiles() model {
	profiles, _ := loadInstalledProfiles()
	m.installedProfiles = profiles
	return m
}

func (m model) reloadProfilesToInstall() model {
	m.profilesToInstall = installableProfiles()
	return m
}

func (m model) reloadRules() model {
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
	return m
}

func (m model) setLastAction(msg string) (model, tea.Cmd) {
	m.lastAction = msg
	return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return lastActionTimeUp
	})
}

func (mod model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m := mod
	switch msg := msg.(type) {
	case string:
		switch msg {
		case lastActionTimeUp:
			m.lastAction = ""
		}

	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c", "ctrl+d", "ctrl+q":
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
					m = m.resetMenu()
				case menuDisableUFW:
					runCommand("sudo ufw disable")()
					m = m.resetMenu()
					m.cursor = 0
				case menuEnableUFW:
					runCommand("sudo ufw enable")()
					m = m.resetMenu()
				case menuEnableLogging:
					runCommand("sudo ufw logging on")()
					m = m.resetMenu()
				case menuDisableLogging:
					runCommand("sudo ufw logging off")()
					m = m.resetMenu()
				case menuAllowPort:
					m.view = allowPort
				case menuDenyPort:
					m.view = denyPort
				case menuDeleteRule:
					m.view = deleteRule
					m.cursor = 0
				case menuProfiles:
					m.view = profilesHome
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
					m = m.reloadStatus()
					m.view = home
					m.inputPort = ""
					m = m.reloadRules()
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
				if m.selectedItems.IsEmpty() {
					runCommand(fmt.Sprintf("yes | sudo ufw delete %d", m.cursor+1))()
					m = m.reloadRules()

					if m.cursor > len(m.rules)-1 {
						m.cursor = len(m.rules) - 1
					}
				} else {
					// we have to reverse otherwise the position of the next element for deletion changes
					selectedSlice := m.selectedItems.ToSlice()
					sort.Slice(selectedSlice, func(i, j int) bool {
						return selectedSlice[i] > selectedSlice[j]
					})

					lo.ForEach(selectedSlice, func(i int, _ int) {
						runCommand(fmt.Sprintf("yes | sudo ufw delete %d", i))()
					})
					m = m.reloadRules()
					m.cursor = 0
					m.selectedItems = set.NewSet[int]()
				}

			case "esc":
				m.view = home
				m.cursor = 0
				m = m.reloadStatus()
				m.selectedItems = set.NewSet[int]()
			case " ":
				m.selectedItems = m.selectedItems.Toggle(m.cursor + 1)
			}

		case m.view.isProfilesHome():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(profileHomeActions)-1 {
					m.cursor++
				}
			case "esc":
				m.view = home
				m.cursor = 0
			case "enter":
				switch profileHomeActions[m.cursor] {
				case menuInstalledProfiles:
					m.view = installedProfilesList
					m.cursor = 0
				case menuInstallProfile:
					m.view = installProfile
					m.cursor = 0
				}
			}
		case m.view.isInstalledProfilesList():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.installedProfiles)-1 {
					m.cursor++
				}
			case "esc":
				m.view = profilesHome
				m.cursor = 0
				m.selectedItems = set.NewSet[int]()
			case " ":
				m.selectedItems = m.selectedItems.Toggle(m.cursor)
			case "enter":
				var output string
				if m.selectedItems.IsEmpty() {
					profile := m.installedProfiles[m.cursor]
					output = runCommand(fmt.Sprintf("sudo ufw allow \"%s\"", profile.Name))()
				} else {
					lo.ForEach(m.selectedItems.ToSlice(), func(i int, _ int) {
						profile := m.installedProfiles[i]
						output += runCommand(fmt.Sprintf("sudo ufw allow \"%s\"", profile.Name))()
					})
					m.cursor = 0
					m.selectedItems = set.NewSet[int]()
				}
				m = m.reloadStatus()
				m = m.reloadRules()
				return m.setLastAction(output)
			}

		case m.view.isInstallProfile():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.profilesToInstall)-1 {
					m.cursor++
				}
			case "esc":
				m.view = profilesHome
				m.cursor = 0
				m.selectedItems = set.NewSet[int]()
			case " ":
				m.selectedItems = m.selectedItems.Toggle(m.cursor)
			case "enter":
				var output string
				if m.selectedItems.IsEmpty() {
					profile := m.profilesToInstall[m.cursor]
					output = createProfile(profile)
				} else {
					lo.ForEach(m.selectedItems.ToSlice(), func(i int, _ int) {
						profile := m.profilesToInstall[i]
						output += "\n" + createProfile(profile)
					})
					m.cursor = 0
					m.selectedItems = set.NewSet[int]()
				}

				m = m.reloadInstalledProfiles()
				m = m.reloadProfilesToInstall()
				return m.setLastAction(output)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var output string

	switch true {
	case m.view.isHome():
		left := renderMenu(m.menu, m.cursor)
		right := strings.Split(m.status, "\n")
		output = renderTwoColumns(left, right)
	case m.view.isPortEdit():
		var prefix string
		switch m.view {
		case allowPort:
			prefix = "allow"
		case denyPort:
			prefix = "deny"
		}
		output = fmt.Sprintf("Enter port to %s: %s\n(Press Enter to confirm)", prefix, m.inputPort)
	case m.view.isDeleteRule():
		lines := []string{"Select rule to delete:"}
		for i, rule := range m.rules {
			prefix := "  "
			if i == m.cursor && m.selectedItems.Has(i+1) {
				prefix = ">*"
			} else if i == m.cursor {
				prefix = "> "
			} else if m.selectedItems.Has(i + 1) {
				prefix = " *"
			}
			lines = append(lines, fmt.Sprintf("%s %s", prefix, rule.line))
		}
		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to delete"
	case m.view.isProfilesHome():
		lines := []string{"Select profile action:"}
		for i, item := range profileHomeActions {
			prefix := " "
			if i == m.cursor {
				prefix = ">"
			}
			var itemName string
			switch item {
			case menuInstalledProfiles:
				itemName = "List installed"
			case menuInstallProfile:
				itemName = "Install"
			}
			lines = append(lines, fmt.Sprintf("%s %s", prefix, itemName))
		}
		output = strings.Join(lines, "\n")
	case m.view.isInstalledProfilesList():
		lines := []string{"Select profile:"}
		for i, profile := range m.installedProfiles {
			prefix := "  "
			if i == m.cursor && m.selectedItems.Has(i) {
				prefix = ">*"
			} else if i == m.cursor {
				prefix = "> "
			} else if m.selectedItems.Has(i) {
				prefix = " *"
			}
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		}
		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to allow"
	case m.view.isInstallProfile():
		lines := []string{"Select profile to install:"}
		for i, profile := range m.profilesToInstall {
			prefix := "  "
			if i == m.cursor && m.selectedItems.Has(i) {
				prefix = ">*"
			} else if i == m.cursor {
				prefix = "> "
			} else if m.selectedItems.Has(i) {
				prefix = " *"
			}
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		}
		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to delete"
	}

	output += "\n" + m.lastAction
	return output
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
			menuItem{"Profiles", func() string { return menuProfiles }},
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
