package create_rule

import (
	"fmt"
	oscmd "fwtui/utils/cmd"
	"fwtui/utils/listext"
	"fwtui/utils/result"
	"fwtui/utils/selectable_list"
	stringsext "fwtui/utils/strings"
	"net"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

// MODEL

type Field string

const (
	RuleFormPort      = "Port"
	RuleFormProtocol  = "Protocol"
	RuleFormAction    = "Action"
	RuleFormDir       = "Direction"
	RuleSourceIP      = "SourceIP"
	RuleDestinationIP = "DestinationIP"
	RuleInterface     = "Interface"
	RuleFormComment   = "Comment"
)

type RuleForm struct {
	port          string
	protocol      *selectable_list.SelectableList[Protocol]
	action        *selectable_list.SelectableList[Action]
	dir           *selectable_list.SelectableList[Direction]
	comment       string
	sourceIP      string
	destinationIP string
	interface_    string
	selectedField *selectable_list.SelectableList[Field]
}

func NewRuleForm() RuleForm {
	return RuleForm{
		protocol:      selectable_list.NewSelectableList(protocols),
		action:        selectable_list.NewSelectableList(actions),
		dir:           selectable_list.NewSelectableList(directions),
		selectedField: selectable_list.NewSelectableList(fieldsForDirection(DirectionIn)),
	}
}

// UPDATE

type CreateRuleOutMsg = string

const CreateRuleCreated CreateRuleOutMsg = "create_rule_created"
const CreateRuleEsc = "create_rule_esc"

func (f RuleForm) UpdateRuleForm(msg tea.Msg) (RuleForm, tea.Cmd, CreateRuleOutMsg) {
	form := f
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "up":
			form.selectedField.Prev()
		case "down":
			form.selectedField.Next()
		case "left":
			switch form.selectedField.Selected() {
			case RuleFormProtocol:
				form.protocol.Prev()
			case RuleFormAction:
				form.action.Prev()
			case RuleFormDir:
				form.dir.Prev()
				form.selectedField.SetItems(fieldsForDirection(form.dir.Selected()))
			}
			return form, nil, ""
		case "right":
			switch form.selectedField.Selected() {
			case RuleFormProtocol:
				form.protocol.Next()
			case RuleFormAction:
				form.action.Next()
			case RuleFormDir:
				form.dir.Next()
				form.selectedField.SetItems(fieldsForDirection(form.dir.Selected()))
			}
			return form, nil, ""

		case "backspace":
			switch form.selectedField.Selected() {
			case RuleFormPort:
				form.port = stringsext.TrimLastChar(form.port)
			case RuleFormComment:
				form.comment = stringsext.TrimLastChar(form.comment)
			case RuleSourceIP:
				form.sourceIP = stringsext.TrimLastChar(form.sourceIP)
			case RuleDestinationIP:
				form.destinationIP = stringsext.TrimLastChar(form.sourceIP)
			case RuleInterface:
				form.interface_ = stringsext.TrimLastChar(form.interface_)
			}
		case "enter":
			res := f.BuildUfwCommand()
			if res.IsOk() {
				cmd := fmt.Sprintf(res.Unwrap())
				output := oscmd.RunCommand(cmd)()
				return f, oscmd.OsCmdExecutedMsg(listext.Singleton(cmd), output), CreateRuleCreated
			} else {
				return f, oscmd.OsCmdExecutedMsg(nil, res.Err().Error()), ""

			}
		case "esc":
			return form, nil, CreateRuleEsc
		default:
			switch form.selectedField.Selected() {
			case RuleFormPort:
				form.port += key
			case RuleFormComment:
				form.comment += key
			case RuleSourceIP:
				form.sourceIP += key
			case RuleDestinationIP:
				form.destinationIP += key
			case RuleInterface:
				form.interface_ += key
			}
		}
	}
	return form, nil, ""
}

func fieldsForDirection(dir Direction) []Field {
	baseFields := []Field{
		RuleFormPort,
		RuleFormProtocol,
		RuleFormAction,
		RuleFormDir,
		RuleFormComment,
	}

	switch dir {
	case DirectionIn:
		return append(baseFields, RuleSourceIP, RuleInterface)
	case DirectionOut:
		return append(baseFields, RuleDestinationIP)
	default:
		return baseFields // fallback in case of invalid input
	}
}

// VIEW

func (f RuleForm) ViewCreateRule() string {
	var lines []string

	for _, field := range f.selectedField.GetItems() {
		var value string
		var fieldString string

		switch field {
		case RuleFormPort:
			value = f.port
			fieldString = "Port"
		case RuleFormProtocol:
			value = string(f.protocol.Selected())
			fieldString = "Protocol"
		case RuleFormAction:
			value = string(f.action.Selected())
			fieldString = "Action"
		case RuleFormDir:
			value = string(f.dir.Selected())
			fieldString = "Direction"
		case RuleFormComment:
			value = f.comment
			fieldString = "Comment (Optional)"
		case RuleSourceIP:
			value = f.sourceIP
			fieldString = "Source IP (Optional)"
		case RuleDestinationIP:
			value = f.destinationIP
			fieldString = "Destination IP (Optional)"
		case RuleInterface:
			value = f.interface_
			fieldString = "Interface (Optional)"
		}

		prefix := lo.Ternary(f.selectedField.Selected() == field, "> ", "  ")
		line := fmt.Sprintf("%s%s: %s", prefix, fieldString, value)
		lines = append(lines, line)
	}

	output := strings.Join(lines, "\n")
	output += "\n\n↑↓ to navigate, type to edit, Enter to submit, Esc to cancel"
	return output
}

func (f RuleForm) BuildUfwCommand() result.Result[string] {
	// Validate port
	if strings.Contains(f.port, ":") {
		split := strings.Split(f.port, ":")

		portNum1, err := strconv.Atoi(split[0])
		if err != nil || portNum1 < 1 || portNum1 > 65535 {
			return result.Err[string](fmt.Errorf("invalid port: %s", split[0]))
		}

		portNum2, err := strconv.Atoi(split[1])
		if err != nil || portNum2 < 1 || portNum2 > 65535 {
			return result.Err[string](fmt.Errorf("invalid port: %s", split[1]))
		}

		if portNum1 > portNum2 {
			return result.Err[string](fmt.Errorf("invalid port range: %s", f.port))
		}

	} else {
		portNum, err := strconv.Atoi(f.port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return result.Err[string](fmt.Errorf("invalid port: %s", f.port))
		}
	}

	// Start building the command
	parts := []string{"sudo", "ufw", string(f.action.Selected())}

	// Direction-specific parts
	switch f.dir.Selected() {
	case DirectionIn:
		if f.interface_ != "" {
			parts = append(parts, "in", "on", f.interface_)
		}

		if f.sourceIP != "" {
			if _, _, err := net.ParseCIDR(f.sourceIP); err != nil {
				if net.ParseIP(f.sourceIP) == nil {
					return result.Err[string](fmt.Errorf("invalid source IP: %s", f.sourceIP))
				}
			}
			parts = append(parts, "from", f.sourceIP)
		} else {
			parts = append(parts, "from", "any")
		}
		parts = append(parts, "to", "any")
	case DirectionOut:
		parts = append(parts, "from", "any")
		if f.destinationIP != "" {
			if _, _, err := net.ParseCIDR(f.destinationIP); err != nil {
				if net.ParseIP(f.destinationIP) == nil {
					return result.Err[string](fmt.Errorf("invalid destination IP: %s", f.destinationIP))
				}
			}
			parts = append(parts, "to", f.destinationIP)
		} else {
			parts = append(parts, "to", "any")
		}
	default:
		return result.Err[string](fmt.Errorf("invalid direction"))
	}

	// Port and protocol
	if f.protocol.Selected() == ProtocolBoth {
		parts = append(parts, "port", f.port)
	} else {
		parts = append(parts, "port", f.port, "proto", string(f.protocol.Selected()))
	}

	// Comment (optional)
	if f.comment != "" {
		parts = append(parts, "comment", fmt.Sprintf("'%s'", f.comment))
	}

	return result.Ok(strings.Join(parts, " "))
}
