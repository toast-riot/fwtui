package createprofile

import (
	"fmt"
	"fwtui/domain/entity"
	"fwtui/domain/notification"
	"fwtui/utils/focusablelist"
	stringsext "fwtui/utils/strings"
	"strconv"
	"strings"

	"fwtui/utils/result"

	tea "github.com/charmbracelet/bubbletea"
)

// MODEL

type ProfileField string

const (
	ProfileFormName  ProfileField = "name"
	ProfileFormTitle ProfileField = "Title"
	ProfileFormPorts ProfileField = "Ports"
)

type ProfileForm struct {
	name  string
	title string
	ports string

	selectedField *focusablelist.SelectableList[ProfileField]
}

func NewProfileForm() ProfileForm {
	return ProfileForm{
		selectedField: focusablelist.FromList([]ProfileField{
			ProfileFormName,
			ProfileFormTitle,
			ProfileFormPorts,
		}),
	}
}

// UPDATE

type CreateProfileOutMsg = string

const CreateProfileCreated CreateProfileOutMsg = "create_profile_created"
const CreateProfileEsc CreateProfileOutMsg = "create_profile_esc"

func (f ProfileForm) UpdateProfileForm(msg tea.Msg) (ProfileForm, tea.Cmd, CreateProfileOutMsg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "up":
			f.selectedField.Prev()
		case "down":
			f.selectedField.Next()
		case "backspace":
			switch f.selectedField.Focused() {
			case ProfileFormName:
				f.name = stringsext.TrimLastChar(f.name)
			case ProfileFormTitle:
				f.title = stringsext.TrimLastChar(f.title)
			case ProfileFormPorts:
				f.ports = stringsext.TrimLastChar(f.ports)
			}
		case "enter":
			res := f.BuildUfwProfile()
			if res.IsOk() {
				createProfileRes := entity.CreateProfile(res.Unwrap())
				if createProfileRes.IsErr() {
					return f, notification.CreateCmd(createProfileRes.Err().Error()), ""
				}
				return f, notification.CreateCmd(createProfileRes.Unwrap()), CreateProfileCreated
			}

			return f, notification.CreateCmd(res.Err().Error()), ""
		case "esc":
			return f, nil, CreateProfileEsc
		default:
			switch f.selectedField.Focused() {
			case ProfileFormName:
				f.name += key
			case ProfileFormTitle:
				f.title += key

			case ProfileFormPorts:
				f.ports += key
			}
		}
	}
	return f, nil, ""
}

// VIEW

func (f ProfileForm) ViewCreateProfile() string {
	var lines []string

	for _, field := range f.selectedField.GetItems() {
		var value string
		var label string
		switch field {
		case ProfileFormName:
			value = f.name
			label = "name"
		case ProfileFormTitle:
			value = f.title
			label = "Title"
		case ProfileFormPorts:
			value = f.ports
			label = "Ports"
		}

		prefix := "  "
		if f.selectedField.Focused() == field {
			prefix = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s: %s", prefix, label, value))
	}

	return strings.Join(lines, "\n") + "\n\n↑↓ to navigate, type to edit, Enter to submit, Esc to cancel"
}

// EXPORT

func (f ProfileForm) BuildUfwProfile() result.Result[entity.UFWProfile] {
	if strings.TrimSpace(f.name) == "" {
		return result.Err[entity.UFWProfile](fmt.Errorf("name cannot be empty"))
	}

	err := validatePorts(f.ports)
	if err != nil {
		return result.Err[entity.UFWProfile](fmt.Errorf("invalid ports: %s", err))
	}

	return result.Ok(entity.UFWProfile{
		Name:  strings.TrimSpace(f.name),
		Title: strings.TrimSpace(f.title),
		Ports: strings.Split(strings.TrimSpace(f.ports), "|"),
	})
}

func validatePorts(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("ports cannot be empty")
	}

	// Split groups by '|'
	groups := strings.Split(input, "|")
	for _, group := range groups {
		parts := strings.Split(group, "/")
		portList := parts[0]
		protocol := ""

		// Check optional protocol
		if len(parts) > 2 {
			return fmt.Errorf("too many '/' in group: %s", group)
		}
		if len(parts) == 2 {
			protocol = parts[1]
			if protocol != "tcp" && protocol != "udp" {
				return fmt.Errorf("invalid protocol: %s", protocol)
			}
		}

		// Validate all ports in the list
		portStrs := strings.Split(portList, ",")
		for _, portStr := range portStrs {
			if strings.Contains(portStr, ":") {
				if len(parts) != 2 {
					return fmt.Errorf("port range must specify protocol: %s", portStr)
				}
				// Handle port ranges
				rangeParts := strings.Split(portStr, ":")
				if len(rangeParts) != 2 {
					return fmt.Errorf("invalid port range: %s", portStr)
				}
				startPortStr := strings.TrimSpace(rangeParts[0])
				endPortStr := strings.TrimSpace(rangeParts[1])

				startPort, err := strconv.Atoi(startPortStr)
				if err != nil || startPort < 1 || startPort > 65535 {
					return fmt.Errorf("invalid start port: %s", startPortStr)
				}

				endPort, err := strconv.Atoi(endPortStr)
				if err != nil || endPort < 1 || endPort > 65535 {
					return fmt.Errorf("invalid end port: %s", endPortStr)
				}

				if startPort > endPort {
					return fmt.Errorf("start port cannot be greater than end port in range: %s", portStr)
				}
			} else {
				portStr = strings.TrimSpace(portStr)
				if err := validatePortString(portStr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validatePortString(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %s", portStr)
	}
	return nil
}
