package git

import (
	"errors"
	"testing"
)

func TestCheckArgsRefusesDangerousOps(t *testing.T) {
	bad := [][]string{
		{"push", "origin", "main"},
		{"merge", "feature"},
		{"rebase", "-i", "HEAD~3"},
		{"pull"},
		{"fetch", "--all"},
		{"reset", "--hard", "HEAD~1"},
		{"commit", "-m", "x", "--force"},
		{"checkout", "-f", "main"},
		{"branch", "-D", "main"},
		{"branch", "-d", "master"},
	}
	for _, args := range bad {
		if err := CheckArgs(args); err == nil {
			t.Errorf("CheckArgs(%v) = nil, want guard error", args)
		} else if !errors.Is(err, ErrGuard) {
			t.Errorf("CheckArgs(%v) error does not wrap ErrGuard: %v", args, err)
		}
	}
}

func TestCheckArgsAllowsSafeOps(t *testing.T) {
	ok := [][]string{
		{"status", "--porcelain"},
		{"rev-parse", "HEAD"},
		{"checkout", "-b", "telepath/fix-123"},
		{"add", "-A"},
		{"commit", "-m", "message"},
		{"diff", "abc..HEAD", "--stat"},
		{"branch", "--list", "telepath/*"},
		{"branch", "-D", "telepath/old"},
		{"stash", "push", "-u"},
	}
	for _, args := range ok {
		if err := CheckArgs(args); err != nil {
			t.Errorf("CheckArgs(%v) = %v, want nil", args, err)
		}
	}
}

func TestCommitTargetGuard(t *testing.T) {
	for _, b := range []string{"main", "master"} {
		if err := CheckCommitTarget(b); err == nil {
			t.Errorf("CheckCommitTarget(%q) = nil, want error", b)
		}
	}
	if err := CheckCommitTarget("telepath/x"); err != nil {
		t.Errorf("CheckCommitTarget(telepath/x) = %v, want nil", err)
	}
}

func TestCheckoutTargetGuard(t *testing.T) {
	if err := CheckCheckoutTarget("master"); err == nil {
		t.Error("CheckCheckoutTarget(master) = nil, want error")
	}
	if err := CheckCheckoutTarget("telepath/feature"); err != nil {
		t.Errorf("CheckCheckoutTarget(telepath/feature) = %v, want nil", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Fix the null check":  "fix-the-null-check",
		"  trim  ":            "trim",
		"":                    "task",
		"UPPER/slash:colon!!": "upper-slash-colon",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
	// Long input is truncated to <=32 chars.
	if got := Slugify("this is a very long task description that should be truncated"); len(got) > 32 {
		t.Errorf("Slugify long = %q (len %d), want <=32", got, len(got))
	}
}
