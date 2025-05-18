package oscmd

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func RunCommand(cmdStr string) func() string {
	return func() string {
		cmd := exec.Command("bash", "-c", cmdStr) // Use a shell to interpret the pipe
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %s\n%s", err, out)
		}
		return string(out)
	}
}

type CommandExecutedMsg struct {
	Cmds   []string
	Output string
}

func OsCmdExecutedMsg(cmds []string, output string) tea.Cmd {
	return func() tea.Msg {
		return CommandExecutedMsg{
			Cmds:   cmds,
			Output: output,
		}
	}
}
