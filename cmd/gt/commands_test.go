package main

import (
	"slices"
	"testing"
)

func TestSubmitArgs(t *testing.T) {
	tests := []struct {
		name          string
		edit, publish bool
		want          []string
	}{
		// The bare `gt submit` must not open the editor: Graphite's does not,
		// and matching that is the point of the command.
		{"default", false, false, []string{"stack", "submit", "--auto"}},
		{"publish", false, true, []string{"stack", "submit", "--auto", "--open"}},
		{"edit", true, false, []string{"stack", "submit"}},
		{"edit and publish", true, true, []string{"stack", "submit", "--open"}},
	}
	for _, tt := range tests {
		if got := submitArgs(tt.edit, tt.publish); !slices.Equal(got, tt.want) {
			t.Errorf("%s: submitArgs(%v, %v) = %q, want %q", tt.name, tt.edit, tt.publish, got, tt.want)
		}
	}
}

func TestIgnorableStackSyncError(t *testing.T) {
	if !ignorableStackSyncError(`fatal: cannot force update the branch 'main' used by worktree at '/repo'`) {
		t.Error("worktree force-update should be ignorable")
	}
	if !ignorableStackSyncError(`current branch "main" is not part of a stack`) {
		t.Error("not on a stack should be ignorable")
	}
	if ignorableStackSyncError("rebase conflict") {
		t.Error("a real sync failure must not be ignored")
	}
}
