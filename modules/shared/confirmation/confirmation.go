package confirmation

import (
	"fwtui/utils/selectable_list"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmDialog struct {
	options *selectable_list.SelectableList[string]
	prompt  string
}

func NewConfirmDialog(prompt string) *ConfirmDialog {
	return &ConfirmDialog{
		options: selectable_list.NewSelectableList([]string{"Yes", "No"}),
		prompt:  prompt,
	}
}

type ConfirmationDialogOutMsg = string

const (
	ConfirmationDialogYes ConfirmationDialogOutMsg = "yes"
	ConfirmationDialogNo  ConfirmationDialogOutMsg = "no"
	ConfirmationDialogEsc ConfirmationDialogOutMsg = "esc"
)

func (f *ConfirmDialog) UpdateDialog(msg tea.Msg) (*ConfirmDialog, tea.Cmd, ConfirmationDialogOutMsg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "up":
			f.options.Prev()
		case "down":
			f.options.Next()
		case "enter":
			switch f.options.Selected() {
			case "Yes":
				return f, nil, ConfirmationDialogYes
			case "No":
				return f, nil, ConfirmationDialogNo
			}
		case "esc":
			return f, nil, ConfirmationDialogEsc
		}
	}
	return f, nil, ""
}

func (f *ConfirmDialog) ViewDialog() string {
	var output string
	output += f.prompt + "\n\n"
	f.options.ForEach(func(item string, _ int, isSelected bool) {
		if isSelected {
			output += "> " + item + "\n"
		} else {
			output += "  " + item + "\n"
		}

	})
	return output
}
