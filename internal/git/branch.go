package git

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Prep is the result of preparing a per-task branch.
type Prep struct {
	Branch  string // the created task branch
	Base    string // the branch the task started from
	BaseSHA string // the base commit SHA (full)
}

// Workspace implements the per-task branch workflow on top of a Repo. It returns
// plain data (branch names, SHAs, diffs); the caller turns those into UI events,
// keeping this package free of any agent/UI dependency.
type Workspace struct {
	repo       *Repo
	prefix     string
	stashDirty bool
}

// NewWorkspace builds a Workspace. prefix defaults to "telepath/". When
// stashDirty is true a dirty working tree is stashed instead of refused.
func NewWorkspace(repo *Repo, prefix string, stashDirty bool) *Workspace {
	if strings.TrimSpace(prefix) == "" {
		prefix = "telepath/"
	}
	return &Workspace{repo: repo, prefix: prefix, stashDirty: stashDirty}
}

// Repo exposes the underlying repo for read-only queries.
func (w *Workspace) Repo() *Repo { return w.repo }

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify turns task text into a short branch-safe slug.
func Slugify(text string) string {
	s := strings.ToLower(strings.TrimSpace(text))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 32 {
		s = strings.Trim(s[:32], "-")
	}
	if s == "" {
		s = "task"
	}
	return s
}

// BranchName composes "<prefix><slug>-<yymmdd-hhmm>".
func (w *Workspace) BranchName(taskText string, now time.Time) string {
	return fmt.Sprintf("%s%s-%s", w.prefix, Slugify(taskText), now.Format("060102-1504"))
}

// Prepare runs preflight, records the base, and creates+checks out the task
// branch. It refuses to proceed on a dirty tree unless stashDirty is set.
func (w *Workspace) Prepare(ctx context.Context, taskText string, now time.Time) (*Prep, error) {
	if !w.repo.IsRepo(ctx) {
		return nil, fmt.Errorf("%s is not a git repository", w.repo.Dir)
	}

	clean, err := w.repo.IsClean(ctx)
	if err != nil {
		return nil, err
	}
	if !clean {
		if !w.stashDirty {
			return nil, fmt.Errorf("working tree is dirty; commit/stash your changes first (or enable BRANCH_STASH_DIRTY)")
		}
		if _, err := w.repo.run(ctx, "stash", "push", "-u", "-m", "telepath-autostash"); err != nil {
			return nil, fmt.Errorf("auto-stash failed: %w", err)
		}
	}

	base, err := w.repo.CurrentBranch(ctx)
	if err != nil {
		return nil, err
	}
	if base == "" {
		return nil, fmt.Errorf("HEAD is detached; checkout a branch before running a task")
	}
	baseSHA, err := w.repo.HeadSHA(ctx)
	if err != nil {
		return nil, err
	}

	branch := w.BranchName(taskText, now)
	if err := CheckCheckoutTarget(branch); err != nil {
		return nil, err
	}
	if _, err := w.repo.run(ctx, "checkout", "-b", branch); err != nil {
		return nil, err
	}

	return &Prep{Branch: branch, Base: base, BaseSHA: baseSHA}, nil
}

// CommitAll stages everything and commits on the current (task) branch. It
// returns ("", false, nil) when there is nothing to commit.
func (w *Workspace) CommitAll(ctx context.Context, message string) (sha string, committed bool, err error) {
	current, err := w.repo.CurrentBranch(ctx)
	if err != nil {
		return "", false, err
	}
	if err := CheckCommitTarget(current); err != nil {
		return "", false, err
	}

	clean, err := w.repo.IsClean(ctx)
	if err != nil {
		return "", false, err
	}
	if clean {
		return "", false, nil
	}

	if _, err := w.repo.run(ctx, "add", "-A"); err != nil {
		return "", false, err
	}

	args := []string{"commit", "-m", message}
	if !w.hasIdentity(ctx) {
		// Provide a fallback identity for headless servers that have none,
		// without overriding a user's real git config when it exists.
		args = []string{
			"-c", "user.name=Telepath",
			"-c", "user.email=telepath@localhost",
			"commit", "-m", message,
		}
	}
	if _, err := w.repo.run(ctx, args...); err != nil {
		return "", false, err
	}

	sha, err = w.repo.HeadSHA(ctx)
	if err != nil {
		return "", false, err
	}
	return sha, true, nil
}

func (w *Workspace) hasIdentity(ctx context.Context) bool {
	email, err := w.repo.run(ctx, "config", "user.email")
	return err == nil && strings.TrimSpace(email) != ""
}

// RangeSummary returns `git diff base..HEAD --stat` and the full diff.
func (w *Workspace) RangeSummary(ctx context.Context, baseSHA string) (stat, full string, err error) {
	rangeSpec := baseSHA + "..HEAD"
	stat, err = w.repo.DiffStat(ctx, rangeSpec)
	if err != nil {
		return "", "", err
	}
	full, err = w.repo.Diff(ctx, rangeSpec)
	if err != nil {
		return "", "", err
	}
	return stat, full, nil
}

// RestoreToBase checks out the base branch and deletes the task branch. It
// discards the task branch, so callers must confirm with the user first.
func (w *Workspace) RestoreToBase(ctx context.Context, base, taskBranch string) error {
	if _, err := w.repo.run(ctx, "checkout", "--", "."); err != nil {
		// Non-fatal: there may be nothing to discard.
		_ = err
	}
	if _, err := w.repo.run(ctx, "checkout", base); err != nil {
		return err
	}
	if _, err := w.repo.run(ctx, "branch", "-D", taskBranch); err != nil {
		return err
	}
	return nil
}

// DeleteBranch deletes a task branch (used by the "delete branch" button when
// already on the base branch). Protected branches are refused by the guard.
func (w *Workspace) DeleteBranch(ctx context.Context, branch string) error {
	_, err := w.repo.run(ctx, "branch", "-D", branch)
	return err
}
