package main

import "testing"

func TestFilterRowsSubstring(t *testing.T) {
	rows := []pickRow{
		{branch: "chore/annotation-ui-polish", text: "◯    chore/annotation-ui-polish"},
		{branch: "feat/set-up-hatchet-worker", text: "│ ◉  feat/set-up-hatchet-worker"},
		{branch: "main", text: "◯─┘  main"},
		githubStacksRow(),
	}
	got := filterRows(rows, "hatchet")
	if len(got) != 1 || got[0].branch != "feat/set-up-hatchet-worker" {
		t.Fatalf("filter hatchet = %#v", got)
	}
	if got := filterRows(rows, ""); len(got) != 4 {
		t.Fatalf("empty query should keep all rows, got %d", len(got))
	}
	if got := filterRows(rows, "GITHUB"); len(got) != 1 || !got[0].openGithub {
		t.Fatalf("filter github = %#v", got)
	}
}

func TestFilterRowsMatchesDetail(t *testing.T) {
	rows := []pickRow{
		{branch: "feat/a", detail: "PR #12 merged", text: "feat/a  PR #12 merged"},
		{branch: "feat/b", detail: untrackedDetail, text: "feat/b  " + untrackedDetail},
	}
	got := filterRows(rows, "merged")
	if len(got) != 1 || got[0].branch != "feat/a" {
		t.Fatalf("filter merged = %#v", got)
	}
	got = filterRows(rows, "track")
	if len(got) != 1 || got[0].branch != "feat/b" {
		t.Fatalf("filter track = %#v", got)
	}
}
