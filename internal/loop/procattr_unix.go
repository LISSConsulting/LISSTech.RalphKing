//go:build !windows

package loop

import (
	"os/exec"
	"syscall"
)

// isolateProcess puts the child process in its own process group so that
// SIGINT from Ctrl+C is not forwarded to it. Ralph handles graceful stop itself.
func isolateProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
