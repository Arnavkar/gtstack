package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type localBranch struct {
	name     string
	upstream string
	gone     bool
}

type pullRequest struct {
	Number      int    `json:"number"`
	State       string `json:"state"`
	HeadRefName string `json:"headRefName"`
}

type staleBranch struct {
	name   string
	reason string
}

// pruneStaleBranches offers to delete local branches whose upstream is gone
// or whose pull request has been merged or closed. Sync has already fetched;
// this only prunes stale remote-tracking refs so "gone" is visible.
func pruneStaleBranches(deleteAll bool) error {
	if err := pruneRemoteTracking(); err != nil {
		return err
	}
	branches, err := listLocalBranches()
	if err != nil {
		return err
	}
	prs, err := listClosedPullRequests()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gt: could not list pull requests (%v); checking deleted remotes only\n", err)
		prs = nil
	}
	stale := staleLocals(branches, prs, trunkNames())
	if len(stale) == 0 {
		return nil
	}

	fmt.Fprintf(os.Stderr, "gt: %d local branch(es) have a deleted upstream or a merged/closed PR:\n", len(stale))
	for _, s := range stale {
		fmt.Fprintf(os.Stderr, "    %s  %s\n", s.name, s.reason)
	}

	interactive := isTerminal(os.Stdin)
	if !deleteAll && !interactive {
		fmt.Fprintln(os.Stderr, "gt: rerun `gt sync` in a terminal to keep or delete each branch, or pass -d to delete them all.")
		return nil
	}

	current, err := currentBranch()
	if err != nil {
		current = ""
	}
	fallback := fallbackTrunk(trunkNames())
	for _, s := range stale {
		if !deleteAll && !confirm(fmt.Sprintf("gt: delete local branch %q?", s.name), false) {
			continue
		}
		if err := deleteLocalBranch(s.name, current, fallback); err != nil {
			fmt.Fprintf(os.Stderr, "gt: could not delete %s: %v\n", s.name, err)
			continue
		}
		if s.name == current {
			current = fallback
		}
	}
	return nil
}

func pruneRemoteTracking() error {
	out, err := capture("git", "remote")
	if err != nil || out == "" {
		return nil
	}
	for _, remote := range strings.Fields(out) {
		if err := run("git", "remote", "prune", remote); err != nil {
			return err
		}
	}
	return nil
}

func listLocalBranches() ([]localBranch, error) {
	out, err := capture("git", "for-each-ref",
		"--format=%(refname:short)%09%(upstream:short)%09%(upstream:track)",
		"refs/heads")
	if err != nil {
		return nil, err
	}
	return parseLocalBranches(out), nil
}

func parseLocalBranches(out string) []localBranch {
	var branches []localBranch
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		name, rest, _ := strings.Cut(line, "\t")
		upstream, track, _ := strings.Cut(rest, "\t")
		branches = append(branches, localBranch{
			name:     name,
			upstream: upstream,
			gone:     strings.Contains(track, "gone"),
		})
	}
	return branches
}

func listClosedPullRequests() ([]pullRequest, error) {
	out, err := capture("gh", "pr", "list", "--state", "closed", "--limit", "1000",
		"--json", "number,state,headRefName")
	if err != nil {
		return nil, err
	}
	var prs []pullRequest
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

func staleLocals(branches []localBranch, prs []pullRequest, trunks map[string]bool) []staleBranch {
	prByHead := map[string]pullRequest{}
	for _, pr := range prs {
		state := strings.ToUpper(pr.State)
		if state != "MERGED" && state != "CLOSED" {
			continue
		}
		pr.State = state
		prByHead[pr.HeadRefName] = pr
	}
	var out []staleBranch
	for _, b := range branches {
		if trunks[b.name] || b.upstream == "" {
			continue
		}
		reason := ""
		if pr, ok := prByHead[b.name]; ok {
			reason = prReason(pr)
		} else if pr, ok := prByHead[upstreamBranch(b.upstream)]; ok {
			reason = prReason(pr)
		}
		switch {
		case b.gone && reason == "":
			reason = b.upstream + " is gone"
		case reason == "" && !b.gone:
			continue
		}
		out = append(out, staleBranch{name: b.name, reason: reason})
	}
	return out
}

func prReason(pr pullRequest) string {
	return fmt.Sprintf("PR #%d %s", pr.Number, strings.ToLower(pr.State))
}

func upstreamBranch(upstream string) string {
	_, name, ok := strings.Cut(upstream, "/")
	if !ok {
		return upstream
	}
	return name
}

func trunkNames() map[string]bool {
	m := map[string]bool{}
	if st, err := loadForestState(); err == nil {
		for _, s := range st.Stacks {
			if s.Trunk.Branch != "" {
				m[s.Trunk.Branch] = true
			}
		}
	}
	if head, err := capture("git", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil {
		if _, name, ok := strings.Cut(head, "/"); ok {
			m[name] = true
		}
	}
	if len(m) == 0 {
		m["main"] = true
		m["master"] = true
	}
	return m
}

func fallbackTrunk(trunks map[string]bool) string {
	for _, name := range []string{"main", "master"} {
		if trunks[name] {
			return name
		}
	}
	for name := range trunks {
		return name
	}
	return "main"
}

func deleteLocalBranch(name, current, fallback string) error {
	if name == current {
		if fallback == "" || fallback == name {
			return fmt.Errorf("cannot delete the current branch (no trunk to check out)")
		}
		if err := run("git", "checkout", fallback); err != nil {
			return err
		}
	}
	return run("git", "branch", "-D", name)
}
