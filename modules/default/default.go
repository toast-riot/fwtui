package _default

import (
	"fmt"
	oscmd "fwtui/utils/cmd"
	"fwtui/utils/listext"
	"fwtui/utils/selectable_list"
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
	fields *selectable_list.SelectableList[Direction]

	actionIncoming *selectable_list.SelectableList[Action]
	actionOutgoing *selectable_list.SelectableList[Action]
	actionRouted   *selectable_list.SelectableList[Action]
}

func Init(policies DefaultPolicies) DefaultModule {
	return DefaultModule{
		fields:         selectable_list.NewSelectableList(directions),
		actionIncoming: selectable_list.NewSelectableList(actions).Select(Action(policies.Incoming)),
		actionOutgoing: selectable_list.NewSelectableList(actions).Select(Action(policies.Outgoing)),
		actionRouted:   selectable_list.NewSelectableList(actions).Select(Action(policies.Routed)),
	}
}

type DefaultRuleOutMsg = string

const DefaultRuleEsc = "default_rule_esc"

func (module DefaultModule) UpdateDefaultsModule(msg tea.Msg) (DefaultModule, tea.Cmd, DefaultRuleOutMsg) {
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
			switch mod.fields.Selected() {
			case DirectionIn:
				mod.actionIncoming.Prev()
			case DirectionOut:
				mod.actionOutgoing.Prev()
			case DirectionRouted:
				mod.actionRouted.Prev()
			}
		case "right":
			switch mod.fields.Selected() {
			case DirectionIn:
				mod.actionIncoming.Next()
			case DirectionOut:
				mod.actionOutgoing.Next()
			case DirectionRouted:
				mod.actionRouted.Next()
			}

		case "enter":
			cmd1 := fmt.Sprintf("sudo ufw default %s incoming", string(mod.actionIncoming.Selected()))
			output1 := oscmd.RunCommand(cmd1)()
			cmd2 := fmt.Sprintf("sudo ufw default %s outgoing", string(mod.actionOutgoing.Selected()))
			output2 := oscmd.RunCommand(cmd2)()
			cmd3 := fmt.Sprintf("sudo ufw default %s routed", string(mod.actionRouted.Selected()))
			output3 := oscmd.RunCommand(cmd3)()
			return mod, oscmd.OsCmdExecutedMsg(
				listext.Slice(cmd1, cmd2, cmd3),
				strings.Join([]string{output1, output2, output3}, "\n"),
			), ""

		case "esc":
			return mod, nil, DefaultRuleEsc
		}
	}
	return mod, nil, ""
}

func (module DefaultModule) ViewSetDefaults() string {
	var lines []string
	lines = append(lines, "Default Rules:")

	for _, field := range module.fields.GetItems() {
		var value string
		var fieldString string

		switch field {
		case DirectionIn:
			value = string(module.actionIncoming.Selected())
			fieldString = "Incoming"
		case DirectionOut:
			value = string(module.actionOutgoing.Selected())
			fieldString = "Outgoing"
		case DirectionRouted:
			value = string(module.actionRouted.Selected())
			fieldString = "Routed"
		}

		prefix := lo.Ternary(module.fields.Selected() == field, "> ", "  ")
		line := fmt.Sprintf("%s%s: %s", prefix, fieldString, value)
		lines = append(lines, line)
	}

	output := strings.Join(lines, "\n")
	output += "\n\n↑↓ to navigate, type to edit, Enter to submit, Esc to cancel"
	return output
}
