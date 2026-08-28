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
		{branch: "main", text: "◯  main"},
		{branch: "a", text: "│  ◯  a"},
		{branch: "b", text: "│  │  ◉  b  (current)"},
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
		{branch: "main", text: "◯  main"},
		{branch: "a", text: "│  ◯  a"},
		{branch: "b", text: "│  │  ◯  b"},
		{branch: "c", text: "│  ◉  c  (current)"},
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
		{branch: "main", text: "◉  main  (current)"},
		{branch: "a", text: "│  ◯  a"},
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
		if got[i].branch != want[i].branch || got[i].text != want[i].text || got[i].openGithub != want[i].openGithub {
			t.Errorf("row %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
