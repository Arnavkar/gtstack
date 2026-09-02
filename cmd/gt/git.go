package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// run executes a command with the terminal attached, so interactive prompts
// and TUIs from gh stack and git work unchanged. Every command is announced
// first: the point of gt is to be a stepping stone to gh stack, so the native
// command has to be visible.
func run(name string, args ...string) error {
	err := exec1(name, args)
	if err == nil {
		return nil
	}
	// A failing `gh stack` call may only mean the extension is not installed.
	if name == "gh" && len(args) > 0 && args[0] == "stack" {
		installed, ierr := ensureExtension()
		if ierr != nil {
			return ierr
		}
		if installed {
			return exec1(name, args)
		}
	}
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("%s is not installed, or not on PATH", name)
	}
	return err
}

func runRecording(name string, args ...string) (string, error) {
	stderr, err := execRecording(name, args)
	if err == nil {
		return stderr, nil
	}
	if name == "gh" && len(args) > 0 && args[0] == "stack" {
		installed, ierr := ensureExtension()
		if ierr != nil {
			return stderr, ierr
		}
		if installed {
			return execRecording(name, args)
		}
	}
	if errors.Is(err, exec.ErrNotFound) {
		return stderr, fmt.Errorf("%s is not installed, or not on PATH", name)
	}
	return stderr, err
}

func execRecording(name string, args []string) (string, error) {
	announce(name, args)
	var buf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &buf)
	err := cmd.Run()
	return buf.String(), err
}

func exec1(name string, args []string) error {
	announce(name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extensionChecked keeps the check to once per process: gt runs one command,
// and a failed install should not be retried within it.
var extensionChecked bool

// ensureExtension makes `gh stack` resolvable, offering to install
// github/gh-stack when it is not. gt is a shim for that one extension, so there
// is nothing to fall back to. It reports whether an install actually happened,
// so a caller holding a failed command knows to retry it.
func ensureExtension() (installed bool, err error) {
	if extensionChecked {
		return false, nil
	}
	extensionChecked = true
	if stackExtensionAvailable() {
		return false, nil
	}
	if !confirmInstall() {
		return false, fmt.Errorf("the gh stack extension is required.\n" +
			"    Install it with: gh extension install github/gh-stack")
	}
	if err := exec1("gh", []string{"extension", "install", "github/gh-stack"}); err != nil {
		return false, fmt.Errorf("could not install the gh stack extension: %w", err)
	}
	return true, nil
}

// confirmInstall asks before pulling a binary off the network. With no terminal
// there is nobody to ask, so it declines rather than installing unattended.
func confirmInstall() bool {
	return confirm("gt: the gh stack extension is not installed. Install it now?", true)
}

// confirm asks a yes/no question on stderr. With no terminal it returns false
// rather than taking a default into the void.
func confirm(prompt string, defaultYes bool) bool {
	if !isTerminal(os.Stdin) {
		return false
	}
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}
	fmt.Fprintf(os.Stderr, "%s %s ", prompt, hint)
	answer, ok := readLine(os.Stdin)
	if !ok {
		return false
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	case "":
		return defaultYes
	}
	return false
}

// readLine reads one line a byte at a time; a buffered reader would swallow
// input meant for the interactive child processes that run after this. The
// second result is false at end of input, so a closed stdin is not read as the
// empty "just pressed enter" answer.
func readLine(f *os.File) (string, bool) {
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			return strings.TrimSpace(string(line)), len(line) > 0
		}
		if buf[0] == '\n' {
			return strings.TrimSpace(string(line)), true
		}
		line = append(line, buf[0])
	}
}

// stackExtensionAvailable reports whether `gh stack` resolves. It costs a
// process, so callers only reach for it once something has already gone wrong
// or the local stack state looks absent.
func stackExtensionAvailable() bool {
	_, err := capture("gh", "stack", "--version")
	return err == nil
}

// announce writes the command to stderr, copy-pasteable, so stdout stays clean
// for pipes. Commands are echoed one at a time, immediately before they run,
// because some steps are conditional and a plan printed up front could lie.
func announce(name string, args []string) {
	line := name
	for _, a := range args {
		line += " " + shellQuote(a)
	}
	if useColor() {
		fmt.Fprintf(os.Stderr, "\x1b[2m$\x1b[0m \x1b[36m%s\x1b[0m\n", line)
		return
	}
	fmt.Fprintf(os.Stderr, "$ %s\n", line)
}

var colorOnce struct {
	done bool
	on   bool
}

func useColor() bool {
	if !colorOnce.done {
		colorOnce.done = true
		colorOnce.on = os.Getenv("NO_COLOR") == "" && isTerminal(os.Stderr)
	}
	return colorOnce.on
}

// isTerminal needs a real tty check, not a character-device check: /dev/null is
// a character device, and treating it as interactive would make `gt ... </dev/null`
// prompt into the void and take the default.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

const shellSafe = "-_./=:@,+"

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		alnum := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
		if !alnum && !strings.ContainsRune(shellSafe, r) {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}

// capture runs a command and returns its trimmed stdout.
func capture(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

// passthrough forwards a command straight to `gh stack <sub>`.
func passthrough(sub string) func([]string) error {
	return func(args []string) error {
		return run("gh", append([]string{"stack", sub}, args...)...)
	}
}

func currentBranch() (string, error) {
	b, err := capture("git", "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	if b == "" {
		return "", fmt.Errorf("HEAD is detached; check out a branch first")
	}
	return b, nil
}

// fastForwardTrunk fetches origin and moves the stack trunk (usually main) to
// match it. Git refuses `branch -f` on a branch another worktree has checked
// out, so when that happens the fast-forward runs in that worktree instead.
func fastForwardTrunk() error {
	if err := run("git", "fetch", "origin"); err != nil {
		return err
	}
	trunk := fallbackTrunk(trunkNames())
	remote := "origin/" + trunk
	if run2("git", "rev-parse", "--verify", "--quiet", remote) != nil {
		return nil
	}
	local, err := capture("git", "rev-parse", trunk)
	if err != nil {
		return nil
	}
	want, err := capture("git", "rev-parse", remote)
	if err != nil {
		return err
	}
	if local == want {
		return nil
	}
	wt, err := worktreePathForBranch(trunk)
	if err != nil {
		return err
	}
	if wt == "" {
		return run("git", "branch", "--force", "--", trunk, remote)
	}
	return run("git", "-C", wt, "merge", "--ff-only", remote)
}

func worktreePathForBranch(name string) (string, error) {
	out, err := capture("git", "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	return parseWorktreeBranchPaths(out)[name], nil
}

func parseWorktreeBranchPaths(out string) map[string]string {
	m := map[string]string{}
	var path, branch string
	flush := func() {
		if path != "" && branch != "" {
			m[branch] = path
		}
		path, branch = "", ""
	}
	for _, line := range strings.Split(out, "\n") {
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "worktree "):
			path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		}
	}
	flush()
	return m
}

// gitStackDir is the directory that holds this checkout's `gh-stack` state file.
//
// Current gh stack writes that file next to `--git-dir`, which in a linked
// worktree is `.git/worktrees/<name>` rather than the shared repository.
// Reading `--git-common-dir` instead finds the main checkout's file (often
// empty) and reports every worktree branch as untracked.
//
// If this checkout has no file of its own yet, fall back to the common dir so
// a stack created in the main repository is still visible from a new worktree.
func gitStackDir() (string, error) {
	gitDir, err := capture("git", "rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(gitDir, "gh-stack")); err == nil {
		return gitDir, nil
	}
	return capture("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
}

// gitStackFiles is every gh-stack path that might hold local stacks: this
// checkout, the shared repository, and each linked worktree. gh stack writes
// per git-dir, so a worktree only sees its own stacks unless we union them.
func gitStackFiles() ([]string, error) {
	gitDir, err := capture("git", "rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return nil, err
	}
	common, err := capture("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var files []string
	add := func(dir string) {
		if dir == "" || seen[dir] {
			return
		}
		seen[dir] = true
		files = append(files, filepath.Join(dir, "gh-stack"))
	}
	add(gitDir)
	add(common)
	entries, err := os.ReadDir(filepath.Join(common, "worktrees"))
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				add(filepath.Join(common, "worktrees", e.Name()))
			}
		}
	}
	return files, nil
}

func isLocalBranch(name string) bool {
	return run2("git", "show-ref", "--verify", "--quiet", "refs/heads/"+name) == nil
}

// hasStagedChanges reports whether the index differs from HEAD.
func hasStagedChanges() bool {
	err := run2("git", "diff", "--cached", "--quiet")
	return err != nil
}

// run2 is run without inheriting stdout, for commands used as predicates.
func run2(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// commitsOn counts commits on HEAD that are not reachable from parent.
func commitsOn(parent string) (int, error) {
	out, err := capture("git", "rev-list", "--count", parent+"..HEAD")
	if err != nil {
		return 0, err
	}
	var n int
	if _, err := fmt.Sscanf(out, "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}

// stage applies the Graphite staging flags before a commit.
func stage(all, update, patch bool) error {
	switch {
	case patch:
		return run("git", "add", "-p")
	case all:
		return run("git", "add", "-A")
	case update:
		return run("git", "add", "-u")
	}
	return nil
}

// joinMessage joins repeated -m values the way git does.
func joinMessage(parts []string) string {
	return strings.Join(parts, "\n\n")
}

// branchNameFrom builds a branch name from a commit message, matching the
// MM-DD-slug shape that `gh stack add -m` generates.
func branchNameFrom(msg, date string) string {
	var b strings.Builder
	prevUnderscore := false
	for _, r := range strings.ToLower(msg) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevUnderscore = false
		default:
			if !prevUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}
	slug := strings.Trim(b.String(), "_")
	if slug == "" {
		slug = "branch"
	}
	return date + "-" + slug
}
