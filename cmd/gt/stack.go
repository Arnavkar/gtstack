package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// schemaVersion is the `.git/gh-stack` layout this wrapper understands.
// gh stack writes the version into the file, so a bump means our reading of
// the state is no longer trustworthy and we must stop rather than guess.
const schemaVersion = 1

type trackedStack struct {
	Trunk struct {
		Branch string `json:"branch"`
	} `json:"trunk"`
	Branches []struct {
		Branch string `json:"branch"`
	} `json:"branches"`
}

type stackState struct {
	SchemaVersion int            `json:"schemaVersion"`
	Stacks        []trackedStack `json:"stacks"`
}

func loadState() (*stackState, error) {
	dir, err := gitStackDir()
	if err != nil {
		return nil, fmt.Errorf("not a git repository")
	}
	st, err := readStackFile(filepath.Join(dir, "gh-stack"))
	if os.IsNotExist(err) {
		// No file means either no stack has been created here yet, or the
		// extension is missing. Reading it as "no stacks" would make gt report
		// every branch as untracked, so rule the second case out first. This
		// only costs a process before the repo's first stack exists.
		if _, ierr := ensureExtension(); ierr != nil {
			return nil, ierr
		}
		return &stackState{SchemaVersion: schemaVersion}, nil
	}
	return st, err
}

func readStackFile(path string) (*stackState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st stackState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("reading gh stack state: %w", err)
	}
	if st.SchemaVersion != schemaVersion {
		return nil, fmt.Errorf(
			"gh stack state is schema v%d but this gt understands v%d; upgrade gt or use gh stack directly",
			st.SchemaVersion, schemaVersion)
	}
	return &st, nil
}

// loadForestState unions stacks from this checkout and every linked worktree
// so `gt checkout` can draw the whole local forest, not just the current
// worktree's gh-stack file.
func loadForestState() (*stackState, error) {
	files, err := gitStackFiles()
	if err != nil {
		return nil, fmt.Errorf("not a git repository")
	}
	out := &stackState{SchemaVersion: schemaVersion}
	found := false
	for _, path := range files {
		st, err := readStackFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		found = true
		mergeStacks(out, st)
	}
	if !found {
		if _, ierr := ensureExtension(); ierr != nil {
			return nil, ierr
		}
	}
	return out, nil
}

func mergeStacks(dst, src *stackState) {
	seen := map[string]bool{}
	for _, s := range dst.Stacks {
		seen[stackKey(s)] = true
	}
	for _, s := range src.Stacks {
		k := stackKey(s)
		if seen[k] {
			continue
		}
		seen[k] = true
		dst.Stacks = append(dst.Stacks, s)
	}
}

func stackKey(s trackedStack) string {
	key := s.Trunk.Branch
	for _, b := range s.Branches {
		key += "\x00" + b.Branch
	}
	return key
}

// position describes where a branch sits in the tracked stacks.
type position struct {
	inStack bool
	// forked is true when the branch is a member of one stack and also the
	// trunk of another, or a member of two stacks. gh stack cannot resolve
	// these without an interactive prompt, so gt refuses to act.
	forked bool
	atTop  bool
	// parent is the branch below this one, or the trunk for the bottom branch.
	parent string
}

func locate(st *stackState, branch string) position {
	var p position
	members, trunks := 0, 0
	for _, s := range st.Stacks {
		if s.Trunk.Branch == branch {
			trunks++
		}
		for i, b := range s.Branches {
			if b.Branch != branch {
				continue
			}
			members++
			p.inStack = true
			p.atTop = i == len(s.Branches)-1
			if i == 0 {
				p.parent = s.Trunk.Branch
			} else {
				p.parent = s.Branches[i-1].Branch
			}
		}
	}
	p.forked = members > 1 || (members == 1 && trunks > 0)
	return p
}

func errForked(branch string) error {
	return fmt.Errorf(
		"branch %q belongs to more than one stack; gt only supports linear stacks.\n"+
			"    Run the gh stack command directly and pick a stack interactively.", branch)
}

// pausedOperation reports which gh stack operation, if any, is halted waiting
// for conflict resolution. gh stack drops a marker file per operation.
func pausedOperation() (string, error) {
	dir, err := gitStackDir()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	for _, op := range []string{"rebase", "modify"} {
		if _, err := os.Stat(filepath.Join(dir, "gh-stack-"+op+"-state")); err == nil {
			return op, nil
		}
	}
	return "", nil
}
