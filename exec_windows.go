package godotenv

import (
	"os"
	"os/exec"
)

func execv(cmd string, cmdArgs []string) error {
	command := exec.Command(cmd, cmdArgs...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
