package agent

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ResolveBinary returns the binary to invoke for an agent. If override is set it
// is returned as-is; otherwise name is looked up on PATH (exec.LookPath already
// handles ".exe"/PATHEXT resolution on Windows). The returned string is safe to
// pass to exec.Command even when LookPath fails, so callers that only need a
// best-effort command (e.g. to print an error) still get something usable.
func ResolveBinary(name, override string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return name
}

// IsInstalled reports whether the given binary can be resolved on PATH.
func IsInstalled(name, override string) bool {
	if strings.TrimSpace(override) != "" {
		if _, err := os.Stat(override); err == nil {
			return true
		}
		// Fall through: an override that is a bare name may still be on PATH.
		if _, err := exec.LookPath(override); err == nil {
			return true
		}
		return false
	}
	_, err := exec.LookPath(name)
	return err == nil
}

// NewCommand builds an *exec.Cmd for an agent run: working directory set, the
// process placed in its own group so the whole tree can be killed, and the
// current environment inherited.
func NewCommand(ctx context.Context, dir, bin string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	configureProcAttr(cmd)
	// CommandContext kills only the direct child by default; kill the whole
	// tree so spawned node/git processes do not linger after /cancel.
	cmd.Cancel = func() error {
		killTree(cmd)
		return nil
	}
	return cmd
}

// LineFunc maps a single line of agent output to zero or more normalized events.
// state is adapter-private and persists across calls within one run.
type LineFunc func(line string) []Event

// WrapStderr attaches trailing stderr to an error for context.
func WrapStderr(err error, stderr string) error {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return err
	}
	if len(stderr) > 500 {
		stderr = stderr[len(stderr)-500:]
	}
	return errors.New(err.Error() + ": " + stderr)
}

// StreamLines runs a command for a "simple" adapter (no terminal record of its
// own), maps each stdout line through fn, and ALWAYS finishes with a terminal
// Result (success) or Error (failure). The Result summary is the last
// AssistantText seen, or "completed". File changes for such agents are surfaced
// by the git layer's end-of-run diff, not by per-line events.
func StreamLines(ctx context.Context, dir, bin string, args []string, fn LineFunc) <-chan Event {
	out := make(chan Event, 64)
	go func() {
		defer close(out)
		var lastText string
		wrapped := func(line string) []Event {
			evs := fn(line)
			for _, e := range evs {
				if at, ok := e.(AssistantText); ok && at.Text != "" {
					lastText = at.Text
				}
			}
			return evs
		}
		cmd := NewCommand(ctx, dir, bin, args...)
		stderr, err := Pump(cmd, wrapped, out)
		if err != nil {
			out <- Error{Err: WrapStderr(err, stderr)}
			return
		}
		summary := lastText
		if summary == "" {
			summary = "completed"
		}
		out <- Result{Summary: summary}
	}()
	return out
}

// Pump runs cmd, reads stdout line by line through fn, and forwards the produced
// events to out. stderr is drained and returned as a string for error context.
// Pump blocks until cmd exits; it does not close out.
//
// Lines can be very large (stream-json), so a bufio.Reader with ReadString is
// used instead of bufio.Scanner (which caps token size).
func Pump(cmd *exec.Cmd, fn LineFunc, out chan<- Event) (stderr string, err error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}

	var errBuf strings.Builder
	errDone := make(chan struct{})
	go func() {
		defer close(errDone)
		_, _ = io.Copy(&errBuf, errPipe)
	}()

	r := bufio.NewReader(stdout)
	for {
		line, rerr := r.ReadString('\n')
		if line != "" {
			for _, ev := range fn(strings.TrimRight(line, "\r\n")) {
				out <- ev
			}
		}
		if rerr != nil {
			if !errors.Is(rerr, io.EOF) {
				// Reading error: still wait for the process to report its own.
			}
			break
		}
	}
	<-errDone
	waitErr := cmd.Wait()
	return errBuf.String(), waitErr
}
