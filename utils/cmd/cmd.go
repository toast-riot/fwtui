package oscmd

import (
	"fmt"
	"os/exec"
)

func RunCommand(cmdStr string) string {
	cmd := exec.Command("bash", "-c", cmdStr) // Use a shell to interpret the pipe
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %s\n%s", err, out)
	}
	return string(out)

}
