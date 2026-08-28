package main

import "fmt"

// pickRow is one line in the checkout tree. openGithub means "run
// `gh stack checkout` with no arguments" instead of checking out a branch.
type pickRow struct {
	branch     string
	text       string
	openGithub bool
	current    bool
}

func nodeMark(current bool) string {
	if current {
		return "◉"
	}
	return "◯"
}

// stackForest is a Graphite-style tree: tips at the top, trunk at the bottom,
// each stack a column that stays vertically continuous until it merges.
func stackForest(st *stackState, current string) []pickRow {
	var trunks []string
	byTrunk := map[string][][]string{}
	seenTrunk := map[string]bool{}
	for _, s := range st.Stacks {
		trunk := s.Trunk.Branch
		if trunk == "" {
			continue
		}
		if !seenTrunk[trunk] {
			seenTrunk[trunk] = true
			trunks = append(trunks, trunk)
		}
		var branches []string
		for _, b := range s.Branches {
			if b.Branch != "" {
				branches = append(branches, b.Branch)
			}
		}
		if len(branches) == 0 {
			continue
		}
		byTrunk[trunk] = append(byTrunk[trunk], branches)
	}

	var rows []pickRow
	for _, trunk := range trunks {
		rows = append(rows, renderTrunkGroup(trunk, byTrunk[trunk], current)...)
	}
	return rows
}

func renderTrunkGroup(trunk string, stacks [][]string, current string) []pickRow {
	n := len(stacks)
	active := make([]bool, n)
	var rows []pickRow
	for si, branches := range stacks {
		for i := len(branches) - 1; i >= 0; i-- {
			b := branches[i]
			active[si] = true
			on := b == current
			rows = append(rows, pickRow{
				branch:  b,
				current: on,
				text:    fmt.Sprintf("%s  %s", branchGraph(n, si, active, nodeMark(on)), b),
			})
		}
	}
	on := trunk == current
	rows = append(rows, pickRow{
		branch:  trunk,
		current: on,
		text:    fmt.Sprintf("%s  %s", trunkGraph(n, nodeMark(on)), trunk),
	})
	return rows
}

// branchGraph is one row of the Graphite column graph. Columns are two cells
// wide (`◯ `, `│ `, or `  `) so a track that has no node on this row still
// occupies its slot and reads as a continuous line down to the trunk.
func branchGraph(nCols, nodeCol int, active []bool, mark string) string {
	var b []byte
	for c := 0; c < nCols; c++ {
		if c > 0 {
			b = append(b, ' ')
		}
		switch {
		case c == nodeCol:
			b = append(b, mark...)
		case active[c]:
			b = append(b, []byte("│")...)
		default:
			b = append(b, ' ')
		}
	}
	return string(b)
}

func trunkGraph(nCols int, mark string) string {
	if nCols <= 1 {
		return mark
	}
	var b []byte
	b = append(b, mark...)
	for c := 1; c < nCols; c++ {
		b = append(b, []byte("─")...)
		if c == nCols-1 {
			b = append(b, []byte("┘")...)
		} else {
			b = append(b, []byte("┴")...)
		}
	}
	return string(b)
}

func githubStacksRow() pickRow {
	return pickRow{openGithub: true, text: "…  All stacks on GitHub"}
}
