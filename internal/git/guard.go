// Package git is the ONLY component that runs git in Telepath. Agents' own git
// and auto-commit features are disabled; Telepath owns version control.
//
// The guards here enforce the project's safety contract: main and the remote
// are NEVER modified. The most a task can do is leave a reviewable local branch.
// Telepath never pushes, never merges, never rebases, never force-anything, and
// never commits onto main/master.
package git

import (
	"errors"
	"fmt"
	"strings"
)

// ErrGuard is the sentinel wrapped by all guard violations.
var ErrGuard = errors.New("git guard")

// protectedBranches are never valid write targets (commit/checkout-create).
var protectedBranches = map[string]bool{
	"main":   true,
	"master": true,
}

// IsProtected reports whether branch is a protected base branch.
func IsProtected(branch string) bool {
	return protectedBranches[strings.TrimSpace(branch)]
}

// forbiddenSubcommands are git operations Telepath must never run, because they
// mutate shared history or the remote.
var forbiddenSubcommands = map[string]bool{
	"push":        true,
	"merge":       true,
	"rebase":      true,
	"pull":        true,
	"fetch":       true,
	"cherry-pick": true,
	"revert":      true,
}

// CheckArgs inspects a git argument vector before it runs and refuses anything
// outside Telepath's allowed operation set. It is the single choke point for
// every git invocation in this package.
func CheckArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: empty git command", ErrGuard)
	}
	sub := args[0]
	if forbiddenSubcommands[sub] {
		return fmt.Errorf("%w: %q is never permitted (Telepath never touches the remote or shared history)", ErrGuard, sub)
	}

	// Never any force operation.
	for _, a := range args {
		la := strings.ToLower(a)
		if la == "--force" || la == "-f" || la == "--force-with-lease" {
			return fmt.Errorf("%w: force operations are never permitted", ErrGuard)
		}
	}

	switch sub {
	case "reset":
		// Refuse hard resets entirely; they can destroy the working tree/base.
		for _, a := range args {
			if a == "--hard" {
				return fmt.Errorf("%w: 'git reset --hard' is never permitted", ErrGuard)
			}
		}
	case "branch":
		// Refuse deleting protected branches.
		for i, a := range args {
			if (a == "-D" || a == "-d" || a == "--delete") && i+1 < len(args) {
				if IsProtected(args[i+1]) {
					return fmt.Errorf("%w: refusing to delete protected branch %q", ErrGuard, args[i+1])
				}
			}
		}
	}
	return nil
}

// CheckCommitTarget refuses a commit when the current branch is protected.
func CheckCommitTarget(current string) error {
	if IsProtected(current) {
		return fmt.Errorf("%w: refusing to commit onto protected branch %q", ErrGuard, current)
	}
	return nil
}

// CheckCheckoutTarget refuses creating/switching a new branch whose name is
// protected (i.e. refuses to make main/master the write target).
func CheckCheckoutTarget(branch string) error {
	if IsProtected(branch) {
		return fmt.Errorf("%w: refusing to create/overwrite protected branch %q as a write target", ErrGuard, branch)
	}
	return nil
}
