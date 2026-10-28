package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Repo is a handle to a git working tree. Every git invocation goes through run,
// which applies the guard checks first.
type Repo struct {
	Dir string
	bin string
}

// Open returns a Repo for dir. gitBin overrides the "git" binary if non-empty.
func Open(dir, gitBin string) *Repo {
	bin := gitBin
	if strings.TrimSpace(bin) == "" {
		bin = "git"
	}
	return &Repo{Dir: dir, bin: bin}
}

// run executes a guarded git command and returns trimmed stdout.
func (r *Repo) run(ctx context.Context, args ...string) (string, error) {
	if err := CheckArgs(args); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, r.bin, args...)
	cmd.Dir = r.Dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimRight(out.String(), "\n"), nil
}

// IsRepo reports whether Dir is inside a git working tree.
func (r *Repo) IsRepo(ctx context.Context) bool {
	out, err := r.run(ctx, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// CurrentBranch returns the checked-out branch name (empty on detached HEAD).
func (r *Repo) CurrentBranch(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	if out == "HEAD" {
		return "", nil // detached
	}
	return out, nil
}

// HeadSHA returns the current commit SHA.
func (r *Repo) HeadSHA(ctx context.Context) (string, error) {
	return r.run(ctx, "rev-parse", "HEAD")
}

// ShortSHA returns an abbreviated commit SHA for ref.
func (r *Repo) ShortSHA(ctx context.Context, ref string) (string, error) {
	return r.run(ctx, "rev-parse", "--short", ref)
}

// StatusPorcelain returns the porcelain status output.
func (r *Repo) StatusPorcelain(ctx context.Context) (string, error) {
	return r.run(ctx, "status", "--porcelain")
}

// IsClean reports whether the working tree has no uncommitted changes.
func (r *Repo) IsClean(ctx context.Context) (bool, error) {
	out, err := r.StatusPorcelain(ctx)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// DiffStat returns `git diff <range> --stat`.
func (r *Repo) DiffStat(ctx context.Context, rangeSpec string) (string, error) {
	return r.run(ctx, "diff", rangeSpec, "--stat")
}

// Diff returns the full `git diff <range>`.
func (r *Repo) Diff(ctx context.Context, rangeSpec string) (string, error) {
	return r.run(ctx, "diff", rangeSpec)
}

// DiffWorktree returns the full diff of uncommitted changes (staged+unstaged).
func (r *Repo) DiffWorktree(ctx context.Context) (string, error) {
	return r.run(ctx, "diff", "HEAD")
}

// ListBranches returns local branch names matching prefix (prefix may be empty).
func (r *Repo) ListBranches(ctx context.Context, prefix string) ([]string, error) {
	out, err := r.run(ctx, "branch", "--list", prefix+"*", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			names = append(names, s)
		}
	}
	return names, nil
}
