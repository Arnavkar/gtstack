# gt — Graphite command names on top of `gh stack`

A small wrapper that accepts the Graphite CLI commands you already type and runs
`gh stack` (and plain git) underneath.

It supports **linear stacks only**. Where `gh stack` cannot represent what
Graphite would do, `gt` stops with an error naming the native command to run
instead. It never guesses.

## Requirements

- `gh`, and the `github/gh-stack` extension. `gt` detects a missing extension
  and offers to install it — see below.
- Built against `gh-stack` **v0.1.0**, state schema **v1**. `gt` reads
  `.git/gh-stack` to decide between `gh stack init` and `gh stack add`, and
  refuses to run if the schema version changes.
- Go 1.23 to build.

## Missing extension

`gt` is a shim for `gh stack`, so it checks that the extension is there and
asks before installing it:

```
gt: the gh stack extension is not installed. Install it now? [Y/n]
```

Answer yes and it runs `gh extension install github/gh-stack`, then retries the
command you typed. Undo with `gh extension remove github/gh-stack`.

The check costs nothing on the happy path. It runs only when a `gh stack` call
fails, or when a repository has no stack state yet — otherwise `gt` would report
every branch as untracked when the real problem is a missing extension.

Without a terminal on stdin there is nobody to ask, so `gt` declines and prints
the install command instead of installing unattended.

## Build

```sh
go build -o bin/gt .
```

## Install

This binary is called `gt` on purpose, so it takes over the name the Graphite
CLI uses. Put it ahead of Homebrew in `PATH`:

```sh
export PATH="/Users/cat/projects/graphite-gh-stack/bin:$PATH"
```

Check which one you get with `which gt`. `gt --version` reports
`gt 0.1.0 (gh-stack shim)`; the real Graphite CLI reports a bare version number.
The Graphite CLI stays installed and reachable at `/opt/homebrew/bin/gt`.

## Every command is echoed

Before each step runs, `gt` prints it to stderr in cyan:

```
$ git add -A
$ gh stack add feat-b
$ git commit -m 'Add the login form'
```

So you can see the native command, learn it, and copy it. Steps are echoed one
at a time immediately before they run, not as a plan up front, because some of
them are conditional — a plan could name a command that never runs.

Output goes to stderr, so `gt log --json | jq` still works. Color turns off when
stderr is not a terminal, or when `NO_COLOR` is set.

## Command mapping

| Graphite | runs |
| --- | --- |
| `gt create` / `gt c` | `gh stack init` (new stack) or `gh stack add` (on top of one) |
| `gt modify` / `gt m` | `git commit [--amend]` + `gh stack rebase --upstack --no-trunk` |
| `gt submit` / `gt s` / `gt ss` | `gh stack submit` |
| `gt sync` | `gh stack sync` (`-d` → `--prune`) |
| `gt restack` | `gh stack rebase` (`-d`/`-u` → `--downstack`/`--upstack`) |
| `gt continue` / `gt abort` | `gh stack rebase` or `gh stack modify`, whichever is paused |
| `gt checkout` / `gt co` | `gh stack checkout`, `gh stack switch`, or `git checkout` |
| `gt get <pr>` | `gh stack checkout <pr>` |
| `gt log` / `gt ls` / `gt ll` | `gh stack view` / `--short` / `git log --graph` |
| `gt up` `down` `top` `bottom` `trunk` | the same `gh stack` subcommand |
| `gt merge` | `gh stack merge` |
| `gt pr` | `gh pr view --web` |

Flags not listed are rejected rather than silently dropped.

## Behaviour differences worth knowing

**`gt create` cannot fork.** `gh stack add` only appends to the top of a stack.
Running `gt create` from the middle of one would fork it, so `gt` errors and
tells you to run `gt top` first.

**`gt create` drives the commit itself.** It stages, asks `gh stack` to create
just the branch, then commits. `gh stack add -m` is not used: on a parent that
has no commits yet, it puts the commit on the parent instead of creating the new
branch.

**`gt modify` is not a forward.** `gh stack` has no amend command, so `gt`
amends with git and then asks `gh stack` to cascade the rebase. If the branch
carries no commits of its own it creates a new commit instead of amending, so
the parent's commit is never rewritten.

**`gt submit` always submits the whole stack.** Graphite submits the current
branch and everything below it; `gh stack submit` covers the whole stack. `gt`
prints a note. Deselect branches in the submit editor, or pass `-n`.

**`gt submit -n` creates drafts.** It maps to `gh stack submit --auto`, which
drafts new PRs. This is intended. `gt` does not add `--open` to match Graphite's
ready-for-review default, because `--open` would also publish PRs that are
already drafts. Pass `-p` when you want them open.

**`gt co` with no argument lists only the current stack.** Graphite lists every
tracked branch. `gh stack switch` is stack-scoped and there is no wider picker.

**`gt co <branch>` routes around `gh stack checkout`** when the branch is the
trunk or is not tracked in a stack, because `gh stack checkout` reads a bare
name as a stack number first and would land you somewhere else.

## Not supported

These have no `gh stack` equivalent. `gt` names the reason and the nearest
native command:

`absorb`, `fold`, `move`, `reorder`, `rename`, `delete`, `split`, `squash`,
`pop`, `revert`, `undo`, `freeze`, `unfreeze`, `track`, `untrack`, `unlink`,
`info`, `children`, `parent`, `dash`, `auth`, `config`, `aliases`.

`fold`, `move`, `reorder`, `rename` and `delete` exist in `gh stack` only inside
the interactive `gh stack modify` TUI, which a non-interactive command cannot
drive.

## Forked stacks

`gh stack` stores each stack as a flat ordered list with no parent pointers, so
one branch cannot have two children inside a stack. A fork can be expressed as a
second stack whose trunk is a mid-stack branch, but then every `gh stack`
command on the fork point needs an interactive prompt to pick a stack, and
`gh stack rebase` only picks up the new parent after it has been pushed.

`gt` detects that case and stops:

```
gt: branch "feat-a" belongs to more than one stack; gt only supports linear stacks.
    Run the gh stack command directly and pick a stack interactively.
```
