package main

import (
	"reflect"
	"testing"
)

func TestParseLocalBranches(t *testing.T) {
	got := parseLocalBranches("main\torigin/main\t\nfeat/a\torigin/feat/a\t[gone]\nlocal\t\t\nfeat/b\torigin/feat/b\t[ahead 1]\n")
	want := []localBranch{
		{name: "main", upstream: "origin/main"},
		{name: "feat/a", upstream: "origin/feat/a", gone: true},
		{name: "local"},
		{name: "feat/b", upstream: "origin/feat/b"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseLocalBranches = %#v, want %#v", got, want)
	}
}

func TestStaleLocalsGoneUpstream(t *testing.T) {
	got := staleLocals([]localBranch{
		{name: "main", upstream: "origin/main"},
		{name: "feat/a", upstream: "origin/feat/a", gone: true},
		{name: "local"},
	}, nil, map[string]bool{"main": true})
	if len(got) != 1 || got[0].name != "feat/a" || got[0].reason != "origin/feat/a is gone" {
		t.Errorf("staleLocals = %#v", got)
	}
}

func TestStaleLocalsMergedAndClosedPR(t *testing.T) {
	branches := []localBranch{
		{name: "merged-one", upstream: "origin/merged-one"},
		{name: "closed-one", upstream: "origin/closed-one"},
		{name: "open-one", upstream: "origin/open-one"},
	}
	prs := []pullRequest{
		{Number: 1, State: "MERGED", HeadRefName: "merged-one"},
		{Number: 2, State: "CLOSED", HeadRefName: "closed-one"},
		{Number: 3, State: "OPEN", HeadRefName: "open-one"},
	}
	got := staleLocals(branches, prs, map[string]bool{"main": true})
	if len(got) != 2 {
		t.Fatalf("got %d stale, want 2: %#v", len(got), got)
	}
	if got[0].reason != "PR #1 merged" || got[1].reason != "PR #2 closed" {
		t.Errorf("reasons = %q, %q", got[0].reason, got[1].reason)
	}
}

func TestStaleLocalsMatchesUpstreamBranchName(t *testing.T) {
	got := staleLocals([]localBranch{
		{name: "local-name", upstream: "origin/feat/api"},
	}, []pullRequest{
		{Number: 9, State: "MERGED", HeadRefName: "feat/api"},
	}, map[string]bool{"main": true})
	if len(got) != 1 || got[0].name != "local-name" || got[0].reason != "PR #9 merged" {
		t.Errorf("staleLocals = %#v", got)
	}
}

func TestStaleLocalsPrefersPRReasonWhenGone(t *testing.T) {
	got := staleLocals([]localBranch{
		{name: "feat/a", upstream: "origin/feat/a", gone: true},
	}, []pullRequest{
		{Number: 4, State: "MERGED", HeadRefName: "feat/a"},
	}, map[string]bool{"main": true})
	if len(got) != 1 || got[0].reason != "PR #4 merged" {
		t.Errorf("staleLocals = %#v", got)
	}
}

func TestStaleLocalsSkipsTrunkEvenIfGone(t *testing.T) {
	got := staleLocals([]localBranch{
		{name: "main", upstream: "origin/main", gone: true},
	}, nil, map[string]bool{"main": true})
	if len(got) != 0 {
		t.Errorf("trunk was stale: %#v", got)
	}
}

func TestUpstreamBranch(t *testing.T) {
	if got := upstreamBranch("origin/feat/a"); got != "feat/a" {
		t.Errorf("upstreamBranch(origin/feat/a) = %q", got)
	}
}

func TestFallbackTrunk(t *testing.T) {
	if got := fallbackTrunk(map[string]bool{"develop": true, "main": true}); got != "main" {
		t.Errorf("fallbackTrunk = %q, want main", got)
	}
}
