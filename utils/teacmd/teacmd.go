package teacmd

import (
	tea "github.com/charmbracelet/bubbletea"
)

type CommandExecutionStartedMsg struct{}
type CommandExecutionFinishedMsg struct {
	Output string
}

func RunOsCmdAndAfter(command func() string, resultMsg func(string) tea.Msg) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return CommandExecutionStartedMsg{}
		},
		func() tea.Msg {
			return resultMsg(command())
		},
	)
}

func OsCmdExecutionFinishedCmd(output string) tea.Cmd {
	return func() tea.Msg {
		return CommandExecutionFinishedMsg{
			Output: output,
		}
	}
}
