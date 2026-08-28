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
		{branch: "main", text: "◯  main"},
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
		{branch: "main", text: "◯─┘  main"},
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
		{branch: "main", text: "◯─┘  main"},
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
		{branch: "main", current: true, text: "◉  main"},
	}
	assertRows(t, got, want)
}

func TestStackForestEmpty(t *testing.T) {
	st := stateFrom(t, `{"schemaVersion": 1}`)
	if got := stackForest(st, "main"); len(got) != 0 {
		t.Errorf("empty state produced %d rows, want 0", len(got))
	}
}

func assertRows(t *testing.T, got, want []pickRow) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d\n got %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
