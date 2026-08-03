Merge some or all of a stack of pull requests using GitHub's atomic stack
merge. All members of the stack up to and including your chosen pull request are
merged into the base branch in a single, all-or-nothing operation: if any PR
cannot be merged, none are.

With no argument, the stack for the current branch is used. Pass a stack number
to merge a stack you don't have checked out, or a pull request number to merge
directly up to that PR. A bare number is treated first as a stack number, then
as a pull request number.

In an interactive terminal, a short wizard lets you choose how far up the stack
to merge (everything below your selection is always included), pick the merge
method, and confirm, then shows live progress. In a non-interactive terminal, or
with --yes, the whole stack (or everything up to the given PR) is merged without
prompting, using your last-used merge method unless one is specified.

Only basic pull request state is checked before merging (open and not a draft);
GitHub evaluates branch protection and repository rules when the merge runs, so
any such failure is reported back to you. Bypassing merge requirements is not
supported for stacks.

If the base branch uses a merge queue, the stack is added to the queue and merges
once the queue processes it; otherwise it is merged directly.

Usage:
  gh stack merge [<stack-number> | <pr-number>] [flags]

Examples:
  # Merge the current stack (interactive picker)
  $ gh stack merge

  # Merge a stack you don't have checked out, by stack number
  $ gh stack merge 7

  # Merge everything up to and including PR #42
  $ gh stack merge 42

  # Merge the whole current stack without prompting, squashing
  $ gh stack merge --yes --squash

Flags:
  -h, --help                  help for merge
      --merge                 Merge with a merge commit
      --merge-method string   Merge method to use: merge, squash, or rebase
      --rebase                Rebase and merge
      --squash                Squash and merge
  -y, --yes                   Merge without prompting for confirmation
