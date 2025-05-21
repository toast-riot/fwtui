package notification

import tea "github.com/charmbracelet/bubbletea"

type NotificationReceivedMsg struct {
	Text string
}

func CreateCmd(text string) tea.Cmd {
	return func() tea.Msg {
		return NotificationReceivedMsg{
			Text: text,
		}
	}
}
