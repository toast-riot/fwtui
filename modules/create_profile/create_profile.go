package create_profile

import (
	"fmt"
	"fwtui/entity"
	oscmd "fwtui/utils/cmd"
	stringsext "fwtui/utils/strings"
	"strings"

	"fwtui/utils/result"
	"fwtui/utils/selectable_list"

	tea "github.com/charmbracelet/bubbletea"
)

// MODEL

type ProfileField string

const (
	ProfileFormName        ProfileField = "Name"
	ProfileFormTitle       ProfileField = "Title"
	ProfileFormDescription ProfileField = "Description"
	ProfileFormPorts       ProfileField = "Ports"
)

type ProfileForm struct {
	Name        string
	title       string
	description string
	ports       string

	selectedField *selectable_list.SelectableList[ProfileField]
}

func NewProfileForm() ProfileForm {
	return ProfileForm{
		selectedField: selectable_list.NewSelectableList([]ProfileField{
			ProfileFormName,
			ProfileFormTitle,
			ProfileFormDescription,
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
			switch f.selectedField.Selected() {
			case ProfileFormName:
				f.Name = stringsext.TrimLastChar(f.Name)
			case ProfileFormTitle:
				f.title = stringsext.TrimLastChar(f.title)
			case ProfileFormDescription:
				f.description = stringsext.TrimLastChar(f.description)
			case ProfileFormPorts:
				f.ports = stringsext.TrimLastChar(f.ports)
			}
		case "enter":
			res := f.BuildUfwProfile()
			if res.IsOk() {
				cmdOutput := entity.CreateProfile(res.Unwrap())
				return f, oscmd.OsCmdExecutedMsg([]string{}, cmdOutput), CreateProfileCreated
			}

			return f, oscmd.OsCmdExecutedMsg([]string{}, res.Err().Error()), ""
		case "esc":
			return f, nil, CreateProfileEsc
		default:
			switch f.selectedField.Selected() {
			case ProfileFormName:
				f.Name += key
			case ProfileFormTitle:
				f.title += key
			case ProfileFormDescription:
				f.description += key
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
			value = f.Name
			label = "Name"
		case ProfileFormTitle:
			value = f.title
			label = "Title"
		case ProfileFormDescription:
			value = f.description
			label = "Description"
		case ProfileFormPorts:
			value = f.ports
			label = "Ports"
		}

		prefix := "  "
		if f.selectedField.Selected() == field {
			prefix = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s: %s", prefix, label, value))
	}

	return strings.Join(lines, "\n") + "\n\n↑↓ to navigate, type to edit, Enter to submit, Esc to cancel"
}

// EXPORT

func (f ProfileForm) BuildUfwProfile() result.Result[entity.UFWProfile] {
	if strings.TrimSpace(f.Name) == "" {
		return result.Err[entity.UFWProfile](fmt.Errorf("Name cannot be empty"))
	}
	if strings.TrimSpace(f.ports) == "" {
		return result.Err[entity.UFWProfile](fmt.Errorf("Ports cannot be empty"))
	}
	return result.Ok(entity.UFWProfile{
		Name:        strings.TrimSpace(f.Name),
		Title:       strings.TrimSpace(f.title),
		Description: strings.TrimSpace(f.description),
		Ports:       strings.Split(strings.TrimSpace(f.ports), "|"),
	})
}
