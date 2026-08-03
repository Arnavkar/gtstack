Add a new branch on top of the current stack.

When -m is omitted but -A or -u is used, your editor opens for the
commit message. When -m is provided without an explicit branch name,
the branch name is auto-generated from the commit message.

Usage:
  gh stack add [branch] [flags]

Examples:
  # Add a new named branch to the stack
  $ gh stack add my-feature

  # Add a branch and commit staged changes
  $ gh stack add -Am "Add user authentication" my-feature

  # Auto-generate branch name from the commit message
  $ gh stack add -m "Fix login bug"

  # Add a branch and open editor to write commit message
  $ gh stack add -A my-feature

Flags:
  -A, --all              Stage all changes including untracked files
  -h, --help             help for add
  -m, --message string   Create a commit with this message
  -u, --update           Stage changes to tracked files only
