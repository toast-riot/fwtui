package defaultpolicies

import (
	"fmt"
	"fwtui/domain/ufw"
	"fwtui/utils/focusablelist"
	"fwtui/utils/teacmd"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

type Direction string

const (
	DirectionIn     Direction = "incoming"
	DirectionOut    Direction = "outgoing"
	DirectionRouted Direction = "routed"
)

var directions = []Direction{DirectionIn, DirectionOut, DirectionRouted}

type Action string

const (
	ActionAllow  Action = "allow"
	ActionDeny   Action = "deny"
	ActionReject Action = "reject"
)

var actions = []Action{ActionAllow, ActionDeny, ActionReject}

type DefaultModule struct {
	fields *focusablelist.SelectableList[Direction]

	actionIncoming *focusablelist.SelectableList[Action]
	actionOutgoing *focusablelist.SelectableList[Action]
	actionRouted   *focusablelist.SelectableList[Action]
}

func Init(policies DefaultPolicies) DefaultModule {
	return DefaultModule{
		fields:         focusablelist.FromList(directions),
		actionIncoming: focusablelist.FromList(actions).Focus(Action(policies.Incoming)),
		actionOutgoing: focusablelist.FromList(actions).Focus(Action(policies.Outgoing)),
		actionRouted:   focusablelist.FromList(actions).Focus(Action(policies.Routed)),
	}
}

// UPDATE

type DefaultPoliciesUpdatedMsg struct{ Output string }
type DefaultPolicyEscMsg struct{}

func (module DefaultModule) UpdateDefaultsModule(msg tea.Msg) (DefaultModule, tea.Cmd) {
	mod := module
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "up":
			mod.fields.Prev()
		case "down":
			mod.fields.Next()
		case "left":
			switch mod.fields.Focused() {
			case DirectionIn:
				mod.actionIncoming.Prev()
			case DirectionOut:
				mod.actionOutgoing.Prev()
			case DirectionRouted:
				mod.actionRouted.Prev()
			}
		case "right":
			switch mod.fields.Focused() {
			case DirectionIn:
				mod.actionIncoming.Next()
			case DirectionOut:
				mod.actionOutgoing.Next()
			case DirectionRouted:
				mod.actionRouted.Next()
			}

		case "enter":
			return mod, teacmd.RunOsCmdAndAfter(func() string {
				return strings.Join([]string{
					ufw.SetDefaultPolicy("incoming", string(mod.actionIncoming.Focused())),
					ufw.SetDefaultPolicy("outgoing", string(mod.actionOutgoing.Focused())),
					ufw.SetDefaultPolicy("routed", string(mod.actionRouted.Focused())),
				}, "\n")
			}, func(s string) tea.Msg {
				return DefaultPoliciesUpdatedMsg{Output: s}
			})

		case "esc":
			return mod, func() tea.Msg {
				return DefaultPolicyEscMsg{}
			}
		}
	}
	return mod, nil
}

func (module DefaultModule) ViewSetDefaults() string {
	var lines []string
	lines = append(lines, "Default Rules:")

	for _, field := range module.fields.GetItems() {
		var value string
		var fieldString string

		switch field {
		case DirectionIn:
			value = string(module.actionIncoming.Focused())
			fieldString = "Incoming"
		case DirectionOut:
			value = string(module.actionOutgoing.Focused())
			fieldString = "Outgoing"
		case DirectionRouted:
			value = string(module.actionRouted.Focused())
			fieldString = "Routed"
		}

		prefix := lo.Ternary(module.fields.Focused() == field, "> ", "  ")
		line := fmt.Sprintf("%s%s: %s", prefix, fieldString, value)
		lines = append(lines, line)
	}

	output := strings.Join(lines, "\n")
	output += "\n\n↑↓ to navigate, ←→ to change selection, Enter to submit, Esc to cancel"
	return output
}
