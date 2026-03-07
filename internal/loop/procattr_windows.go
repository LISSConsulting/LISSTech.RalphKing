//go:build windows

package loop

import (
	"os/exec"
	"syscall"
)

// isolateProcess puts the child process in a new process group so that
// console Ctrl+C is not forwarded to it. Ralph handles graceful stop itself.
func isolateProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
