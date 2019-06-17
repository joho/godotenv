//+build !windows

package godotenv

import (
	"os"
	"os/exec"
	"syscall"
)

func execv(cmd string, cmdArgs []string) error {
	prog, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}
	args := append([]string{cmd}, cmdArgs...)

	return syscall.Exec(prog, args, os.Environ())
}
