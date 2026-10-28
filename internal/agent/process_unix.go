//go:build !windows

package agent

import (
	"os/exec"
	"syscall"
)

// configureProcAttr puts the child in its own process group so the whole group
// can be signalled at once.
func configureProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killTree sends SIGKILL to the child's process group, reaping descendants too.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	// Negative pid targets the process group (set up via Setpgid above).
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		_ = cmd.Process.Kill()
	}
}
