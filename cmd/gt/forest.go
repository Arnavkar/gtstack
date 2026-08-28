package main

import (
	"fmt"
	"strings"
)

// pickRow is one line in the checkout tree. openGithub means "run
// `gh stack checkout` with no arguments" instead of checking out a branch.
type pickRow struct {
	branch     string
	text       string
	openGithub bool
}

func currentSuffix(branch, current string) string {
	if branch == current {
		return "  (current)"
	}
	return ""
}

func nodeMark(branch, current string) string {
	if branch == current {
		return "◉"
	}
	return "◯"
}

// stackForest is a Graphite-style trunk-first tree of every locally tracked
// stack. Stacks that share a trunk are siblings under it.
func stackForest(st *stackState, current string) []pickRow {
	type stack []string
	var trunks []string
	byTrunk := map[string][]stack{}
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
		rows = append(rows, pickRow{
			branch: trunk,
			text:   fmt.Sprintf("%s  %s%s", nodeMark(trunk, current), trunk, currentSuffix(trunk, current)),
		})
		for _, branches := range byTrunk[trunk] {
			for i, b := range branches {
				prefix := strings.Repeat("│  ", i+1)
				rows = append(rows, pickRow{
					branch: b,
					text:   fmt.Sprintf("%s%s  %s%s", prefix, nodeMark(b, current), b, currentSuffix(b, current)),
				})
			}
		}
	}
	return rows
}

func githubStacksRow() pickRow {
	return pickRow{openGithub: true, text: "…  All stacks on GitHub"}
}
