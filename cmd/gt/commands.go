package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

func newFlags(usage string) *pflag.FlagSet {
	fs := pflag.NewFlagSet("gt "+usage, pflag.ContinueOnError)
	fs.SortFlags = false
	return fs
}

func parse(fs *pflag.FlagSet, args []string) error {
	err := fs.Parse(args)
	if errors.Is(err, pflag.ErrHelp) {
		os.Exit(0)
	}
	return err
}

// cmdCreate branches off the current position. On the top branch of a stack
// that is `gh stack add`; off a branch in no stack (the trunk, usually) it is
// `gh stack init`, which starts a second stack rather than appending to the
// first. From the middle of a stack it refuses: that would be a fork.
func cmdCreate(args []string) error {
	fs := newFlags("create [name]")
	all := fs.BoolP("all", "a", false, "stage all changes, including untracked files")
	update := fs.BoolP("update", "u", false, "stage updates to tracked files only")
	patch := fs.BoolP("patch", "p", false, "pick hunks to stage before committing")
	insert := fs.BoolP("insert", "i", false, "insert between the current branch and its child")
	msg := fs.StringArrayP("message", "m", nil, "commit message")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *insert {
		return fmt.Errorf("gt create --insert has no gh stack equivalent; insert a branch with `gh stack modify`")
	}

	branch, err := currentBranch()
	if err != nil {
		return err
	}
	pos, _, err := requireStackPosition(branch)
	if err != nil {
		return err
	}
	if err := requireLocalStackMetadata(branch, pos); err != nil {
		return err
	}

	if pos.inStack && !pos.atTop {
		return fmt.Errorf(
			"branch %q is not the top of its stack; branching here would fork the stack, which gh stack cannot represent.\n"+
				"    Run `gt top` first, or restructure with `gh stack modify`.", branch)
	}

	message := joinMessage(*msg)
	name := ""
	if fs.NArg() > 0 {
		name = fs.Arg(0)
	} else if message != "" {
		name = branchNameFrom(message, time.Now().Format("01-02"))
	} else {
		return fmt.Errorf("gt create needs a branch name or -m <message>")
	}

	// Stage first: an aborted `git add -p` should not leave a new branch behind.
	if err := stage(*all, *update, *patch); err != nil {
		return err
	}

	// gh stack only creates and checks out the branch here. Its own -A/-u/-m
	// staging is deliberately unused: on a parent that carries no commits yet
	// it puts the commit on the parent instead of creating the new branch.
	if pos.inStack {
		err = run("gh", "stack", "add", name)
	} else {
		err = run("gh", "stack", "init", "--base", branch, name)
	}
	if err != nil {
		return err
	}

	if !hasStagedChanges() {
		return nil
	}
	commit := []string{"commit"}
	if message != "" {
		commit = append(commit, "-m", message)
	}
	return run("git", commit...)
}

// cmdModify amends (or adds to) the current branch and restacks everything
// above it. gh stack has no equivalent command, so gt drives git directly and
// then asks gh stack to cascade the rebase.
func cmdModify(args []string) error {
	fs := newFlags("modify")
	commitNew := fs.BoolP("commit", "c", false, "create a new commit instead of amending")
	all := fs.BoolP("all", "a", false, "stage all changes before committing")
	update := fs.BoolP("update", "u", false, "stage updates to tracked files only")
	patch := fs.BoolP("patch", "p", false, "pick hunks to stage before committing")
	edit := fs.BoolP("edit", "e", false, "open an editor for the commit message")
	msg := fs.StringArrayP("message", "m", nil, "commit message")
	if err := parse(fs, args); err != nil {
		return err
	}

	branch, err := currentBranch()
	if err != nil {
		return err
	}
	pos, _, err := requireStackPosition(branch)
	if err != nil {
		return err
	}
	if err := requireLocalStackMetadata(branch, pos); err != nil {
		return err
	}
	if !pos.inStack {
		return errNotInStack(branch)
	}

	if err := stage(*all, *update, *patch); err != nil {
		return err
	}

	// Amending a branch that carries no commits of its own would rewrite the
	// parent's commit, so fall back to a new commit exactly as gt does.
	amend := !*commitNew
	if amend {
		n, err := commitsOn(pos.parent)
		if err != nil {
			return err
		}
		if n == 0 {
			amend = false
		}
	}

	commit := []string{"commit"}
	if amend {
		commit = append(commit, "--amend")
	}
	switch message := joinMessage(*msg); {
	case message != "":
		commit = append(commit, "-m", message)
	case amend && !*edit:
		commit = append(commit, "--no-edit")
	}
	if err := run("git", commit...); err != nil {
		return err
	}

	if pos.atTop {
		return nil
	}
	return run("gh", "stack", "rebase", "--upstack", "--no-trunk")
}

func errNotInStack(branch string) error {
	if trunkNames()[branch] {
		return fmt.Errorf("branch %q is not part of a stack; start one with `gt create`", branch)
	}
	return fmt.Errorf(
		"branch %q is not part of a stack.\n"+
			"    Adopt this existing branch with `gt track`.\n"+
			"    `gt create` would start a new branch on top of this one.",
		branch)
}

// cmdSubmit maps onto `gh stack submit`, which always covers the whole stack.
//
// gh stack opens its editor whenever it has a terminal, but Graphite's submit
// does not, so gt passes --auto by default and keeps -e for the times the
// editor is wanted. --auto creates new PRs as drafts, matching what -n did
// before; -p adds --open, which also publishes PRs that were already drafts, so
// it is never implied.
func cmdSubmit(args []string) error {
	fs := newFlags("submit")
	// -d and -n now describe the default. They stay so that the flags people
	// already type keep working.
	draft := fs.BoolP("draft", "d", false, "create new PRs as drafts (the default)")
	publish := fs.BoolP("publish", "p", false, "mark PRs ready for review")
	noEdit := fs.BoolP("no-edit", "n", false, "skip the PR metadata editor (the default)")
	edit := fs.BoolP("edit", "e", false, "open the gh stack submit editor")
	stack := fs.Bool("stack", false, "submit the whole stack (always on with gh stack)")
	updateOnly := fs.BoolP("update-only", "u", false, "only update existing PRs (gh stack still creates missing ones)")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *draft && *publish {
		return fmt.Errorf("--draft and --publish conflict")
	}
	if *edit && *noEdit {
		return fmt.Errorf("--edit and --no-edit conflict")
	}
	if !*stack {
		fmt.Fprintln(os.Stderr,
			"gt: note — gh stack submit covers the whole stack, not just the current branch and below.\n"+
				"    Pass -e to pick branches in the editor.")
	}
	if *updateOnly {
		fmt.Fprintln(os.Stderr,
			"gt: note — gh stack submit creates PRs for branches that do not have them.\n"+
				"    Pass -e and deselect those branches to update existing PRs only.")
	}
	if *draft && *edit {
		fmt.Fprintln(os.Stderr, "gt: note — set draft with the CREATE AS toggle in the submit editor.")
	}
	captured, err := runTee("gh", submitArgs(*edit, *publish)...)
	if err != nil {
		explainSwallowedPush(captured)
	}
	return err
}

// explainSwallowedPush re-runs a dry-run push when gh stack reports only
// git's generic "failed to push some refs". That line hides pre-push hooks
// and --force-with-lease details.
func explainSwallowedPush(stderr string) {
	if !swallowedPushError(stderr) {
		return
	}
	branch := failedPushBranch(stderr)
	if branch == "" {
		if cur, err := currentBranch(); err == nil {
			branch = cur
		}
	}
	if branch == "" {
		return
	}
	fmt.Fprintln(os.Stderr, "trace:")
	_ = run("git", "push", "--dry-run", "--force-with-lease", "origin", branch)
}

func swallowedPushError(stderr string) bool {
	return strings.Contains(stderr, "failed to push some refs") ||
		(strings.Contains(stderr, "failed to run git") && strings.Contains(stderr, "failed to push"))
}

func failedPushBranch(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "✗"))
		if !strings.HasPrefix(line, "failed to push ") {
			continue
		}
		rest := strings.TrimPrefix(line, "failed to push ")
		if strings.HasPrefix(rest, "some refs") {
			continue
		}
		name, _, ok := strings.Cut(rest, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if name != "" && !strings.ContainsAny(name, " \t") {
			return name
		}
	}
	return ""
}

// submitArgs builds the `gh stack submit` invocation. It is split out so the
// default -- skipping the editor -- is covered by a test.
func submitArgs(edit, publish bool) []string {
	gh := []string{"stack", "submit"}
	if !edit {
		gh = append(gh, "--auto")
	}
	if publish {
		gh = append(gh, "--open")
	}
	return gh
}

func cmdSync(args []string) error {
	fs := newFlags("sync")
	deleteAll := fs.BoolP("delete-all", "d", false, "delete stale stack branches without prompting")
	if err := parse(fs, args); err != nil {
		return err
	}
	if err := fastForwardTrunk(); err != nil {
		return err
	}
	branch, err := currentBranch()
	if err != nil {
		return err
	}
	pos, _, err := requireStackPosition(branch)
	if err != nil {
		return err
	}
	if pos.inStack {
		stderr, err := runRecording("gh", "stack", "rebase")
		if err != nil && !ignorableStackSyncError(stderr) {
			return err
		}
	}
	return pruneStaleBranches(*deleteAll)
}

func ignorableStackSyncError(stderr string) bool {
	return strings.Contains(stderr, "is not part of a stack") ||
		strings.Contains(stderr, "cannot force update the branch")
}

func cmdRestack(args []string) error {
	fs := newFlags("restack")
	downstack := fs.BoolP("downstack", "d", false, "restack this branch and its ancestors")
	upstack := fs.BoolP("upstack", "u", false, "restack this branch and its descendants")
	only := fs.BoolP("only", "o", false, "restack this branch only")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *only {
		return fmt.Errorf("gt restack --only has no gh stack equivalent; rebase the single branch with git")
	}
	if *downstack && *upstack {
		return fmt.Errorf("--downstack and --upstack conflict")
	}
	gh := []string{"stack", "rebase"}
	if *downstack {
		gh = append(gh, "--downstack")
	}
	if *upstack {
		gh = append(gh, "--upstack")
	}
	return run("gh", gh...)
}

func cmdContinue(args []string) error { return resume("--continue", args) }
func cmdAbort(args []string) error    { return resume("--abort", args) }

func resume(flag string, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected argument %q", args[0])
	}
	op, err := pausedOperation()
	if err != nil {
		return err
	}
	if op == "" {
		return fmt.Errorf("no gh stack rebase or modify is in progress")
	}
	return run("gh", "stack", op, flag)
}

func cmdCheckout(args []string) error {
	fs := newFlags("checkout [branch]")
	trunk := fs.BoolP("trunk", "t", false, "check out the trunk")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *trunk {
		return run("gh", "stack", "trunk")
	}
	if fs.NArg() == 0 {
		return checkoutInteractive()
	}
	return checkoutTarget(fs.Arg(0))
}

// checkoutInteractive is Graphite's bare `gt checkout`: a bottom-up tree of
// every locally tracked stack (this worktree and the others). `gh stack switch`
// only lists the current stack; the last row opens `gh stack checkout` for
// stacks that exist only on GitHub.
func checkoutInteractive() error {
	if isTerminal(os.Stdin) && isTerminal(os.Stderr) {
		st, err := loadForestState()
		if err != nil {
			return err
		}
		current, err := currentBranch()
		if err != nil {
			return err
		}
		rows := checkoutRows(st, current)
		rows = append(rows, githubStacksRow())
		chosen, err := pickBranch(rows)
		if err != nil {
			return err
		}
		if chosen.openGithub {
			return run("gh", "stack", "checkout")
		}
		return checkoutTarget(chosen.branch)
	}
	return run("gh", "stack", "checkout")
}

// checkoutTarget sends a named branch to `gh stack checkout`, except for the
// trunk and untracked local branches: a bare name is tried as a stack number
// then a PR, so those would land you on the wrong branch.
func checkoutTarget(target string) error {
	if isLocalBranch(target) {
		pos, _, err := requireStackPosition(target)
		if err != nil {
			return err
		}
		if !pos.inStack {
			return run("git", "checkout", target)
		}
		cur, err := loadCurrentWorktreeState()
		if err != nil {
			return err
		}
		if !locate(cur, target).inStack {
			return run("git", "checkout", target)
		}
	}
	return run("gh", "stack", "checkout", target)
}

func cmdGet(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("gt get needs a branch, PR number, or PR URL; to refresh the current stack run `gt sync`")
	}
	return run("gh", "stack", "checkout", args[0])
}

func cmdLog(args []string) error {
	form := ""
	if len(args) > 0 && (args[0] == "short" || args[0] == "long") {
		form, args = args[0], args[1:]
	}
	if form == "long" {
		return run("git", append([]string{"log", "--graph", "--oneline", "--decorate", "--all"}, args...)...)
	}
	gh := []string{"stack", "view"}
	if form == "short" {
		gh = append(gh, "--short")
	}
	return run("gh", append(gh, args...)...)
}

func cmdPR(args []string) error {
	return run("gh", append([]string{"pr", "view", "--web"}, args...)...)
}

// cmdTrack adopts existing Git branches into a new stack. Graphite's
// per-branch track becomes `gh stack init` of the listed branches (or the
// current one). It does not create a new branch; that is `gt create`.
func cmdTrack(args []string) error {
	fs := newFlags("track [branches...]")
	base := fs.StringP("base", "b", "", "trunk for the new stack")
	parent := fs.StringP("parent", "p", "", "Graphite parent branch")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *parent != "" {
		return fmt.Errorf("gt track --parent has no gh stack equivalent; pass --base <trunk> and list branches bottom to top")
	}
	trunk := *base
	if trunk == "" {
		trunk = fallbackTrunk(trunkNames())
	}
	branches := fs.Args()
	if len(branches) == 0 {
		cur, err := currentBranch()
		if err != nil {
			return err
		}
		if trunkNames()[cur] {
			return fmt.Errorf("on trunk %q; start a stacked branch with `gt create`, or pass names: `gt track a b`", cur)
		}
		pos, _, err := requireStackPosition(cur)
		if err != nil {
			return err
		}
		if pos.inStack {
			return fmt.Errorf("branch %q is already in a stack", cur)
		}
		branches = []string{cur}
	} else {
		for _, name := range branches {
			if trunkNames()[name] {
				return fmt.Errorf("%q is a trunk; omit it and pass --base %s", name, name)
			}
			pos, _, err := requireStackPosition(name)
			if err != nil {
				return err
			}
			if pos.inStack {
				return fmt.Errorf("branch %q is already in a stack", name)
			}
		}
	}
	gh := append([]string{"stack", "init", "--base", trunk}, branches...)
	return run("gh", gh...)
}

func cmdInit(args []string) error {
	return fmt.Errorf(
		"gh stack has no repository-level init; a stack is created when you branch.\n" +
			"    Run `gt create <name>` on your trunk to start one, or `gt track` to adopt existing branches.")
}
