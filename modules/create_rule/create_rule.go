package create_rule

import (
	"fmt"
	oscmd "fwtui/utils/cmd"
	"fwtui/utils/selectable_list"
	stringsext "fwtui/utils/strings"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

type Field string

const (
	RuleFormPort     = "Port"
	RuleFormProtocol = "Protocol"
	RuleFormAction   = "Action"
	RuleFormDir      = "Direction"
	RuleFormComment  = "Comment"
)

var fields = []Field{
	RuleFormPort,
	RuleFormProtocol,
	RuleFormAction,
	RuleFormDir,
	RuleFormComment,
}

type RuleForm struct {
	port          string
	protocol      *selectable_list.SelectableList[Protocol]
	action        *selectable_list.SelectableList[Action]
	dir           *selectable_list.SelectableList[Direction]
	comment       string
	selectedField *selectable_list.SelectableList[Field]
}

func NewRuleForm() RuleForm {
	return RuleForm{
		port:          "",
		protocol:      selectable_list.NewSelectableList(protocols),
		action:        selectable_list.NewSelectableList(actions),
		dir:           selectable_list.NewSelectableList(directions),
		comment:       "",
		selectedField: selectable_list.NewSelectableList(fields),
	}
}

type CreateRuleOutMsg = string

const CreateRuleCreated CreateRuleOutMsg = "create_rule_created"
const CreateRuleEsc = "create_rule_esc"

func (f RuleForm) UpdateRuleForm(msg tea.Msg) (RuleForm, tea.Cmd, CreateRuleOutMsg) {
	form := f
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "up", "k":
			form.selectedField.Prev()
		case "down", "j":
			form.selectedField.Next()
		case "left", "l":
			switch form.selectedField.Selected() {
			case RuleFormProtocol:
				form.protocol.Prev()
			case RuleFormAction:
				form.action.Prev()
			case RuleFormDir:
				form.dir.Prev()
			}
		case "right", "h":
			switch form.selectedField.Selected() {
			case RuleFormProtocol:
				form.protocol.Next()
			case RuleFormAction:
				form.action.Next()
			case RuleFormDir:
				form.dir.Next()
			}

		case "backspace":
			switch form.selectedField.Selected() {
			case RuleFormPort:
				form.port = stringsext.TrimLastChar(form.port)
			case RuleFormComment:
				form.comment = stringsext.TrimLastChar(form.comment)
			}
		case "enter":
			if f.isValid() {
				portProtocol := lo.Ternary(form.protocol.Selected() == ProtocolBoth, form.port, form.port+"/"+string(form.protocol.Selected()))
				cmd := fmt.Sprintf("sudo ufw %s %s comment '%s'", form.action.Selected(), portProtocol, form.comment)
				oscmd.RunCommand(cmd)()
				return form, nil, CreateRuleCreated
			}
		case "esc":
			return form, nil, CreateRuleEsc
		default:
			switch form.selectedField.Selected() {
			case RuleFormPort:
				form.port += key
			case RuleFormComment:
				form.comment += key
			}
		}
	}
	return form, nil, ""
}

func (f RuleForm) ViewCreateRule() string {
	var lines []string

	for _, field := range fields {
		var value string

		switch field {
		case RuleFormPort:
			value = f.port
		case RuleFormProtocol:
			value = string(f.protocol.Selected())
		case RuleFormAction:
			value = string(f.action.Selected())
		case RuleFormDir:
			value = string(f.dir.Selected())
		case RuleFormComment:
			value = f.comment
		}

		prefix := lo.Ternary(f.selectedField.Selected() == field, "> ", "  ")
		line := fmt.Sprintf("%s%s: %s", prefix, field, value)
		lines = append(lines, line)
	}

	output := strings.Join(lines, "\n")
	output += "\n\n↑↓ to navigate, type to edit, Enter to submit, Esc to cancel"
	return output
}

func (f RuleForm) isValid() bool {
	atoi, err := strconv.Atoi(f.port)
	if err != nil {
		return false
	}
	if atoi < 1 || atoi > 65535 {
		return false
	}

	return true
}
