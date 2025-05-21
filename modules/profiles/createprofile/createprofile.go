package createprofile

import (
	"fmt"
	"fwtui/domain/entity"
	"fwtui/domain/notification"
	"fwtui/utils/focusablelist"
	stringsext "fwtui/utils/strings"
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
				cmdOutput := entity.CreateProfile(res.Unwrap())
				return f, notification.CreateCmd(cmdOutput), CreateProfileCreated
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
	if strings.TrimSpace(f.ports) == "" {
		return result.Err[entity.UFWProfile](fmt.Errorf("Ports cannot be empty"))
	}
	return result.Ok(entity.UFWProfile{
		Name:  strings.TrimSpace(f.name),
		Title: strings.TrimSpace(f.title),
		Ports: strings.Split(strings.TrimSpace(f.ports), "|"),
	})
}
