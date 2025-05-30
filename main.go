package main

import (
	"fmt"
	"fwtui/domain/notification"
	"fwtui/domain/ufw"
	"fwtui/modules/createrule"
	"fwtui/modules/defaultpolicies"
	"fwtui/modules/profiles"
	"fwtui/modules/shared/confirmation"
	"fwtui/utils/focusablelist"
	"fwtui/utils/multiselect"
	"fwtui/utils/teacmd"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This action requires root. Please run with sudo.")
		os.Exit(1)
	}
	cmd := exec.Command("sudo", "ufw", "status")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("ufw is not available or sudo failed: %v", err)
	}

	profilesModule, _ := profiles.Init()
	m := model{menuList: focusablelist.FromList(buildMenu()), view: viewStateHome, profilesModule: profilesModule}
	m = m.reloadRules()
	m = m.reloadStatus()
	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// MODEL

type menuItem struct {
	title  string
	action string
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

func (v viewHomeState) isSetDefault() bool {
	return v == viewSetDefault
}

const viewStateHome = "viewStateHome"
const viewStateProfiles = "profiles"
const viewStateCreateRule = "create_rule"
const viewStateDeleteRule = "delete_rule"
const viewSetDefault = "set_default"

// HOME MENU
const menuResetUFW = "RESET_UFW"
const menuQuit = "QUIT"
const menuDisableUFW = "DISABLE"
const menuEnableUFW = "ENABLE"
const menuCreateRule = "CREATE_RULE"
const menuDeleteRule = "DELETE_RULE"
const menuExportRules = "EXPORT_RULES"
const menuDisableLogging = "DISABLE_LOGGING"
const menuEnableLogging = "ENABLE_LOGGING"
const menuSetDefault = "SET_DEFAULT"
const menuProfiles = "PROFILES"

type rule struct {
	number int
	line   string
}

type model struct {
	menuList             *focusablelist.SelectableList[menuItem]
	resetDialog          *confirmation.ConfirmDialog
	view                 viewHomeState
	status               string
	notification         string
	runningNotifications int
	cmdIsRunning         bool

	rules        multiselect.MultiSelectableList[rule]
	deleteDialog *confirmation.ConfirmDialog

	ruleForm          createrule.RuleForm
	profilesModule    profiles.ProfilesModule
	setDefaultsModule defaultpolicies.DefaultModule
}

func (m model) Init() tea.Cmd {
	return nil
}

// UPDATE

type lastActionTimeUpMsg struct{}
type rulesDeletedMsg struct{ Output string }

func (mod model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m := mod

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c", "ctrl+d", "ctrl+q":
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case lastActionTimeUpMsg:
		m.runningNotifications--
		if m.runningNotifications == 0 {
			m.notification = ""
		}

		return m, nil

	case teacmd.CommandExecutionStartedMsg:
		m.cmdIsRunning = true
		return m, nil

	case teacmd.CommandExecutionFinishedMsg:
		m.cmdIsRunning = false
		return m.setNotification(msg.Output)

	case notification.NotificationReceivedMsg:
		return m.setNotification(msg.Text)

	default:
		switch true {
		case m.view.isHome():
			if m.resetDialog != nil {
				newDeleteDialog, _, outMsg := m.resetDialog.UpdateDialog(msg)
				m.resetDialog = newDeleteDialog
				switch outMsg {
				case confirmation.ConfirmationDialogYes:
					output := ufw.Reset()
					m = m.resetMenu()
					m = m.reloadStatus()
					m = m.reloadRules()
					return m.setNotification(output)
				case confirmation.ConfirmationDialogNo:
					m.resetDialog = nil
				case confirmation.ConfirmationDialogEsc:
					m.resetDialog = nil
				}

				return m, nil
			}

			switch msg := msg.(type) {
			case tea.KeyMsg:
				key := msg.String()
				switch key {
				case "up", "k":
					m.menuList.Prev()
				case "down", "j":
					m.menuList.Next()
				case "enter":
					selected := m.menuList.Focused().action
					switch selected {
					case menuResetUFW:
						m.resetDialog = confirmation.NewConfirmDialog("Are you sure you want to reset UFW?")
					case menuDisableUFW:
						ufw.Disable()
						m = m.resetMenu()
						m.menuList.FocusFirst()
					case menuEnableUFW:
						ufw.Enable()
						m = m.resetMenu()
					case menuEnableLogging:
						ufw.EnableLogging()
						m = m.resetMenu()
					case menuDisableLogging:
						ufw.DisableLogging()
						m = m.resetMenu()
					case menuCreateRule:
						m.ruleForm = createrule.NewRuleForm()
						m.view = viewStateCreateRule
					case menuDeleteRule:
						m.view = viewStateDeleteRule
					case menuExportRules:
						return m, func() tea.Msg {
							getwd, err := os.Getwd()
							if err != nil {
								return nil
							}
							err = ufw.ExportCurrentUFWState(getwd + "/ufw_import.sh")
							if err != nil {
								return notification.NotificationReceivedMsg{Text: fmt.Sprintf("Error exporting rules: %s", err)}
							}
							return notification.NotificationReceivedMsg{Text: "Rules exported to ufw_import.sh.\nCareful!!! This file is executable and will reset your current UFW state when run.\nPlease review it before executing."}

						}
					case menuSetDefault:
						m.view = viewSetDefault
						result := defaultpolicies.ParseUfwDefaults(m.status)

						if result.IsErr() {
							return m.setNotification(result.Err().Error())
						}

						m.setDefaultsModule = defaultpolicies.Init(result.Value())
						return m, nil
					case menuProfiles:
						m.view = viewStateProfiles
					case menuQuit:
						return m, tea.Quit
					}
				}
			}
		case m.view.isCreateRule():
			switch msg.(type) {
			case createrule.CreateRuleCreatedMsg:
				m = m.reloadStatus()
				m = m.reloadRules()
				m.view = viewStateHome
				return m, nil
			case createrule.CreateRuleEscMsg:
				m.view = viewStateHome
				return m, nil
			}

			newForm, cmd := m.ruleForm.UpdateRuleForm(msg)
			m.ruleForm = newForm
			return m, cmd

		case m.view.isDeleteRule():
			if m.deleteDialog != nil {
				newDeleteDialog, _, outMsg := m.deleteDialog.UpdateDialog(msg)
				m.deleteDialog = newDeleteDialog
				switch outMsg {
				case confirmation.ConfirmationDialogYes:
					m.deleteDialog = nil

					return m, teacmd.RunOsCmdAndAfter(func() string {
						if m.rules.NoneSelected() {
							return ufw.DeleteRuleByNumber(m.rules.FocusedIndex() + 1)
						} else {
							var output string
							// we have to reverse otherwise the position of the next element for deletion changes
							selectedSlice := m.rules.GetSelectedIndexes()
							sort.Slice(selectedSlice, func(i, j int) bool {
								return selectedSlice[i] > selectedSlice[j]
							})

							lo.ForEach(selectedSlice, func(i int, _ int) {
								ufw.DeleteRuleByNumber(i + 1)
							})
							return output
						}
					}, func(s string) tea.Msg {
						return rulesDeletedMsg{Output: s}
					},
					)

				case confirmation.ConfirmationDialogNo:
					m.deleteDialog = nil
				case confirmation.ConfirmationDialogEsc:
					m.deleteDialog = nil
				}

				return m, nil
			}

			switch msg := msg.(type) {
			case rulesDeletedMsg:
				m.rules.FocusFirst()
				m = m.reloadRules()
				return m, teacmd.OsCmdExecutionFinishedCmd(msg.Output)

			case tea.KeyMsg:
				key := msg.String()
				switch key {
				case "up", "k":
					m.rules.Prev()
				case "down", "j":
					m.rules.Next()
				case "d":
					if m.rules.NoneSelected() {
						m.deleteDialog = confirmation.NewConfirmDialog("Are you sure you want to delete this rule?")
					} else {
						m.deleteDialog = confirmation.NewConfirmDialog("Are you sure you want to delete selected rules?")
					}

				case "esc":
					m.view = viewStateHome
					m = m.reloadStatus()
				case " ":
					m.rules.Toggle()
				}
			}
		case m.view.isProfiles():
			switch msg.(type) {
			case profiles.ProfilesEscMsg:
				m.view = viewStateHome
				m = m.reloadStatus()
				m = m.reloadRules()
				return m, nil
			}
			newModule, cmd := m.profilesModule.UpdateProfilesModule(msg)
			m.profilesModule = newModule
			return m, cmd

		case m.view.isSetDefault():
			switch msg := msg.(type) {
			case defaultpolicies.DefaultPolicyEscMsg:
				m.view = viewStateHome
				return m, nil
			case defaultpolicies.DefaultPoliciesUpdatedMsg:
				m = m.reloadStatus()
				return m, teacmd.OsCmdExecutionFinishedCmd(msg.Output)
			}

			newModule, cmd := m.setDefaultsModule.UpdateDefaultsModule(msg)
			m.setDefaultsModule = newModule
			return m, cmd
		}
	}
	return m, nil
}

func (m model) setNotification(msg string) (model, tea.Cmd) {
	m.notification = msg
	m.runningNotifications++
	return m, tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return lastActionTimeUpMsg{}
	})
}

func (m model) resetMenu() model {
	m.menuList.SetItems(buildMenu())
	m = m.reloadStatus()
	return m
}

func (m model) reloadStatus() model {
	m.status = ufw.StatusVerbose()
	return m
}

func (m model) reloadRules() model {
	lines := strings.Split(ufw.StatusNumbered(), "\n")
	if len(lines) < 6 {
		m.rules.SetItems([]rule{})
		return m
	}

	lines = lines[4:(len(lines) - 2)]
	rules := lo.Map(lines, func(line string, index int) rule {
		return rule{
			number: index,
			line:   line,
		}
	})
	m.rules.SetItems(rules)
	return m
}

func buildMenu() []menuItem {
	enabled, loggingOn := getStatus()

	items := []menuItem{}

	if enabled {
		items = append(items, menuItem{"Disable", menuDisableUFW})
		items = append(items, menuItem{"Set defaults", menuSetDefault})
		items = append(items,
			menuItem{"Profiles", menuProfiles},
			menuItem{"Create rule", menuCreateRule},
			menuItem{"Delete rule", menuDeleteRule},
			menuItem{"Export rules", menuExportRules},
		)
		if loggingOn {
			items = append(items, menuItem{"Disable logging", menuDisableLogging})
		} else {
			items = append(items, menuItem{"Enable logging", menuEnableLogging})
		}
	} else {
		items = append(items, menuItem{"Enable", menuEnableUFW})

	}

	items = append(items,
		menuItem{"Reset UFW", menuResetUFW},
		menuItem{"Quit", menuQuit},
	)

	return items
}

func getStatus() (enabled bool, loggingOn bool) {
	lines := strings.Split(ufw.StatusVerbose(), "\n")
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

// VIEW

func (m model) View() string {
	if m.cmdIsRunning {
		return "Running command, please wait..."
	}

	var output string

	switch true {
	case m.view.isHome():
		if m.resetDialog != nil {
			return m.resetDialog.ViewDialog()
		}
		left := renderMenu(m.menuList)
		right := strings.Split(m.status, "\n")
		output = renderTwoColumns(left, right)
	case m.view.isCreateRule():
		output = m.ruleForm.ViewCreateRule()
	case m.view.isDeleteRule():
		if m.deleteDialog != nil {
			return m.deleteDialog.ViewDialog()
		}
		lines := []string{"Focus rule to delete:"}
		m.rules.ForEach(func(rule rule, index int, isFocused, isSelected bool) {
			focusedPrefix := lo.Ternary(isFocused, ">", " ")
			selectedPrefix := lo.Ternary(isSelected, "*", " ")
			prefix := focusedPrefix + selectedPrefix
			lines = append(lines, fmt.Sprintf("%s %s", prefix, rule.line))
		})
		output = strings.Join(lines, "\n")
		output += "\n\n↑↓ to navigate, d to delete, Space to select, Esc to cancel"
	case m.view.isProfiles():
		output = m.profilesModule.ViewProfiles()
	case m.view.isSetDefault():
		output = m.setDefaultsModule.ViewSetDefaults()
	}

	output += "\n\n" + m.notification
	return output
}

func renderMenu(menu *focusablelist.SelectableList[menuItem]) []string {
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
