package main

import (
	"fmt"
	"fwtui/modules/create_rule"
	"fwtui/modules/profiles"
	oscmd "fwtui/utils/cmd"
	"fwtui/utils/selectable_list"
	"fwtui/utils/set"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

type menuItem struct {
	title  string
	action func() string
}

type viewHomeState string

func (v viewHomeState) isCreateRule() bool {
	return v == viewStateCreateRule
}

func (v viewHomeState) isHome() bool {
	return v == viewStateHome
}

func (v viewHomeState) isProfiles() bool {
	return v == viewStateProfiles
}

func (v viewHomeState) isDeleteRule() bool {
	return v == viewStateDeleteRule
}

const viewStateHome = "viewStateHome"
const viewStateProfiles = "profiles"
const viewStateCreateRule = "create_rule"
const viewStateDeleteRule = "delete_rule"

// HOME MENU
const menuResetUFW = "RESET_UFW"
const menuDisableUFW = "DISABLE"
const menuEnableUFW = "ENABLE"
const menuCreateRule = "CREATE_RULE"
const menuDeleteRule = "DELETE_RULE"
const menuDisableLogging = "DISABLE_LOGGING"
const menuEnableLogging = "ENABLE_LOGGING"
const menuProfiles = "PROFILES"

// EVENT
const lastActionTimeUp = "LAST_ACTION_TIME_UP"

type rule struct {
	number int
	line   string
}

type model struct {
	cursor   int
	menuList *selectable_list.SelectableList[menuItem]
	status   string
	rules    []rule

	lastAction    string
	selectedItems set.Set[int]

	view           viewHomeState
	ruleForm       create_rule.RuleForm
	profilesModule profiles.ProfilesModule
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This action requires root. Please run with sudo.")
		os.Exit(1)
	}

	m := model{menuList: buildMenu(), selectedItems: set.NewSet[int](), view: viewStateHome}
	m = m.reloadRules()
	m = m.reloadStatus()
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
	m.menuList = buildMenu()
	m = m.reloadStatus()
	return m
}

func (m model) reloadStatus() model {
	m.status = oscmd.RunCommand("sudo ufw status verbose")()
	return m
}

func (m model) reloadRules() model {
	output := oscmd.RunCommand("sudo ufw status numbered")()
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

func (mod model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m := mod
	switch msg := msg.(type) {
	case string:
		switch msg {
		case lastActionTimeUp:
			m.lastAction = ""
		}
	case oscmd.CommandExecutedMsg:
		lastAction := []string{"Executed commands:"}
		lo.ForEach(msg.Cmds, func(cmd string, _ int) {
			lastAction = append(lastAction, cmd)
		})
		lastAction = append(lastAction, "With output:")
		lastAction = append(lastAction, msg.Output)
		m.lastAction = strings.Join(lastAction, "\n")
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return lastActionTimeUp
		})

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
				m.menuList.Prev()
			case "down", "j":
				m.menuList.Next()
			case "enter":
				selected := m.menuList.Selected().action()
				switch selected {
				case menuResetUFW:
					oscmd.RunCommand("sudo ufw reset")()
					m = m.resetMenu()
				case menuDisableUFW:
					oscmd.RunCommand("sudo ufw disable")()
					m = m.resetMenu()
					m.cursor = 0
				case menuEnableUFW:
					oscmd.RunCommand("sudo ufw enable")()
					m = m.resetMenu()
				case menuEnableLogging:
					oscmd.RunCommand("sudo ufw logging on")()
					m = m.resetMenu()
				case menuDisableLogging:
					oscmd.RunCommand("sudo ufw logging off")()
					m = m.resetMenu()
				case menuCreateRule:
					m.ruleForm = create_rule.NewRuleForm()
					m.view = viewStateCreateRule
				case menuDeleteRule:
					m.view = viewStateDeleteRule
					m.cursor = 0
				case menuProfiles:
					m.view = viewStateProfiles
					module, cmd := profiles.Init()
					m.profilesModule = module
					return m, cmd
				}
			}
		case m.view.isCreateRule():
			newForm, cmd, outMsg := m.ruleForm.UpdateRuleForm(msg)
			m.ruleForm = newForm
			switch outMsg {
			case create_rule.CreateRuleCreated:
				m = m.reloadStatus()
				m = m.reloadRules()
				m.view = viewStateHome
			case create_rule.CreateRuleEsc:
				m.view = viewStateHome
			}
			return m, cmd

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
					oscmd.RunCommand(fmt.Sprintf("yes | sudo ufw delete %d", m.cursor+1))()
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
						oscmd.RunCommand(fmt.Sprintf("yes | sudo ufw delete %d", i))()
					})
					m = m.reloadRules()
					m.cursor = 0
					m.selectedItems = set.NewSet[int]()
				}

			case "esc":
				m.view = viewStateHome
				m.cursor = 0
				m = m.reloadStatus()
				m.selectedItems = set.NewSet[int]()
			case " ":
				m.selectedItems = m.selectedItems.Toggle(m.cursor + 1)
			}
		case m.view.isProfiles():
			newModule, cmd, outMsg := m.profilesModule.UpdateProfilesModule(msg)
			m.profilesModule = newModule
			switch outMsg {
			case profiles.ProfilesOutMsgEsc:
				m.view = viewStateHome
				m = m.reloadStatus()
				m = m.reloadRules()
			}
			return m, cmd
		}
	}
	return m, nil
}

func (m model) View() string {
	var output string

	switch true {
	case m.view.isHome():
		left := renderMenu(m.menuList)
		right := strings.Split(m.status, "\n")
		output = renderTwoColumns(left, right)
	case m.view.isCreateRule():
		output = m.ruleForm.ViewCreateRule()
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
	case m.view.isProfiles():
		output = m.profilesModule.ViewProfiles()
	}

	output += "\n" + m.lastAction
	return output
}

func renderMenu(menu *selectable_list.SelectableList[menuItem]) []string {
	var lines []string
	lines = append(lines, "", "UFW Firewall Menu:", "")
	menu.ForEach(func(item menuItem, index int, isSelected bool) {
		prefix := lo.Ternary(isSelected, ">", " ")
		lines = append(lines, fmt.Sprintf("%s %s", prefix, item.title))
	})
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

func buildMenu() *selectable_list.SelectableList[menuItem] {
	enabled, loggingOn := getStatus()

	items := []menuItem{}

	if enabled {
		items = append(items, menuItem{"Disable", func() string { return menuDisableUFW }})
		items = append(items,
			menuItem{"Profiles", func() string { return menuProfiles }},
			menuItem{"Create rule", func() string { return menuCreateRule }},
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

	return selectable_list.NewSelectableList(items)
}

func getStatus() (enabled bool, loggingOn bool) {
	output := oscmd.RunCommand("sudo ufw status verbose")()
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
