> **Languages:** [English](../en/branch-workflow.md) · [Русский](../ru/branch-workflow.md) · [Türkçe](../tr/branch-workflow.md) · [Azərbaycanca](../az/branch-workflow.md)

# Branch workflow

Telepath owns all git. The agents' own git and auto-commit features are disabled
so that version control is consistent and safe. The mode is set by `BRANCH_MODE`.

## `per-task` (default)

Each task runs on its own isolated branch:

1. **Preflight.** The working tree must be clean. If it is dirty, Telepath
   refuses with a clear message — unless `BRANCH_STASH_DIRTY=true`, in which case
   it stashes the changes first. It records the base branch and base commit SHA.
2. **Create branch.** It creates and checks out a branch named
   `<BRANCH_PREFIX><slug>-<yymmdd-hhmm>`, e.g.
   `telepath/add-health-endpoint-260115-1430`. The slug is derived from the task
   text (lowercased, non-alphanumerics collapsed to `-`, truncated).
3. **Run the agent.** The agent only edits files; its git is off.
4. **Commit on the branch.** On success, if `BRANCH_COMMIT=true` (default),
   Telepath stages everything and commits **on the task branch** with a message
   that is a short task summary plus the trailer:

   ```
   <summary>

   via Telepath (<agent>)
   ```

   If there is nothing to change, it says so and makes no commit.
5. **Never touch main.** It never pushes, merges, rebases, force-anything, or
   commits onto `main`/`master`.
6. **Report.** It posts the branch name, commit SHA, `git diff <base>..HEAD
   --stat`, and the full diff as a `.diff` document, with **Keep branch** /
   **Delete branch & restore base** buttons.

You review the branch and **merge/push it yourself**.

### The `BRANCH_COMMIT=false` trade-off

If `BRANCH_COMMIT=false`, the agent's changes are left **uncommitted** on the
task branch. Telepath tells you so. Because all chats share one working tree,
uncommitted changes are weaker isolation: a later task's preflight will see a
dirty tree and refuse (or stash). Prefer the default `true` unless you have a
specific reason.

## `off`

Telepath runs on the current branch, creates no branch, and makes no commit. You
still get the live stream and an **end-of-run working-tree diff**. This is the
legacy / minimal mode for when you manage branches yourself.

## The hard guards

Every git command goes through a single guard (`internal/git/guard.go`) that
**refuses**:

- `push`, `merge`, `rebase`, `pull`, `fetch`, `cherry-pick`, `revert`
- any `--force` / `-f` / `--force-with-lease`
- `reset --hard`
- committing onto, creating, or deleting `main` / `master`

**The guarantee: `main` and the remote are never modified. The worst a task can
do is leave a reviewable local branch.**

## Cleanup

- **Keep branch** leaves it for your review.
- **Delete branch & restore base** discards the task branch's changes, checks the
  base branch back out, and deletes the task branch. (Refused while a task is
  running — `/cancel` first.)
- `/branches` lists existing `telepath/*` branches at any time.
