package main

import "testing"

func TestStackForestSingleStack(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [{"trunk": {"branch": "main"},
                  "branches": [{"branch": "a"}, {"branch": "b"}]}]
    }`)
	got := stackForest(st, "b")
	want := []pickRow{
		{branch: "b", current: true, text: "◉  b"},
		{branch: "a", text: "◯  a"},
		{branch: "main", trunk: true, text: "◯  main"},
	}
	assertRows(t, got, want)
}

func TestStackForestSiblingStacksShareTrunk(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [
        {"trunk": {"branch": "main"}, "branches": [{"branch": "a"}, {"branch": "b"}]},
        {"trunk": {"branch": "main"}, "branches": [{"branch": "c"}]}
      ]
    }`)
	got := stackForest(st, "c")
	want := []pickRow{
		{branch: "b", text: "◯    b"},
		{branch: "a", text: "◯    a"},
		{branch: "c", current: true, text: "│ ◉  c"},
		{branch: "main", trunk: true, text: "◯─┘  main"},
	}
	assertRows(t, got, want)
}

func TestStackForestTwoSingleBranchStacks(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [
        {"trunk": {"branch": "main"}, "branches": [{"branch": "chore/annotation-ui-polish"}]},
        {"trunk": {"branch": "main"}, "branches": [{"branch": "feat/set-up-hatchet-worker"}]}
      ]
    }`)
	got := stackForest(st, "feat/set-up-hatchet-worker")
	want := []pickRow{
		{branch: "chore/annotation-ui-polish", text: "◯    chore/annotation-ui-polish"},
		{branch: "feat/set-up-hatchet-worker", current: true, text: "│ ◉  feat/set-up-hatchet-worker"},
		{branch: "main", trunk: true, text: "◯─┘  main"},
	}
	assertRows(t, got, want)
}

func TestStackForestCurrentOnTrunk(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [{"trunk": {"branch": "main"}, "branches": [{"branch": "a"}]}]
    }`)
	got := stackForest(st, "main")
	want := []pickRow{
		{branch: "a", text: "◯  a"},
		{branch: "main", current: true, trunk: true, text: "◉  main"},
	}
	assertRows(t, got, want)
}

func TestPickRowRenderColor(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [
        {"trunk": {"branch": "main"}, "branches": [{"branch": "a"}]},
        {"trunk": {"branch": "main"}, "branches": [{"branch": "b"}]}
      ]
    }`)
	rows := stackForest(st, "b")
	plain := rows[1].render(false, false)
	if plain != rows[1].text {
		t.Fatalf("uncolored render = %q, want %q", plain, rows[1].text)
	}
	colored := rows[1].render(true, true)
	if colored == plain {
		t.Fatal("colored render produced no ANSI")
	}
	if !containsANSI(colored) || !containsANSI(rows[0].render(true, false)) {
		t.Fatalf("expected ANSI in graph: %q / %q", colored, rows[0].render(true, false))
	}
	if githubStacksRow().render(true, false) == githubStacksRow().text {
		t.Fatal("github row should dim when color is on")
	}
}

func containsANSI(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			return true
		}
	}
	return false
}

func TestStackForestEmpty(t *testing.T) {
	st := stateFrom(t, `{"schemaVersion": 1}`)
	if got := stackForest(st, "main"); len(got) != 0 {
		t.Errorf("empty state produced %d rows, want 0", len(got))
	}
}

func TestEnsureCurrentAddsMissing(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [{"trunk": {"branch": "main"}, "branches": [{"branch": "a"}]}]
    }`)
	got := ensureCurrent(stackForest(st, "scratch"), "scratch")
	if got[0].branch != "scratch" || !got[0].current {
		t.Fatalf("untracked current was not prepended: %+v", got[0])
	}
}

func TestEnsureCurrentLeavesPresent(t *testing.T) {
	st := stateFrom(t, `{
      "schemaVersion": 1,
      "stacks": [{"trunk": {"branch": "main"}, "branches": [{"branch": "a"}]}]
    }`)
	forest := stackForest(st, "a")
	got := ensureCurrent(forest, "a")
	if len(got) != len(forest) {
		t.Fatalf("duplicated current: got %d rows, want %d", len(got), len(forest))
	}
}

func assertRows(t *testing.T, got, want []pickRow) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d\n got %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i].branch != want[i].branch || got[i].text != want[i].text || got[i].current != want[i].current || got[i].openGithub != want[i].openGithub || got[i].trunk != want[i].trunk {
			t.Errorf("row %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
