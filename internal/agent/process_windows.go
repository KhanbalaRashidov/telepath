//go:build windows

package agent

import (
	"os/exec"
	"strconv"
	"syscall"
)

// configureProcAttr sets Windows-specific spawn attributes. A new process group
// lets us address the child and its descendants.
func configureProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// killTree terminates the process and its whole tree. Agent CLIs (claude,
// gemini, codex) are typically Node processes that spawn further children;
// "taskkill /T" reliably reaps the tree on Windows where os.Process.Kill only
// hits the direct child.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	kill := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	if err := kill.Run(); err != nil {
		// Best effort: fall back to a direct kill.
		_ = cmd.Process.Kill()
	}
}
