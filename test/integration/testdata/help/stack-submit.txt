Push all branches and create or update a stack of PRs on GitHub.

In an interactive terminal, a single-screen editor opens. Every branch without a
PR is included by default; deselect any you don't want with the checkbox or ^x,
and draft each PR's title, description, and draft state, then submit them all at
once with Ctrl+S. Pass --auto (or run in a non-interactive terminal) to skip the
editor and use auto-generated titles.

If your branches already have open PRs but no stack on GitHub yet (for example,
after deleting the stack) and you deselect every new PR, press Ctrl+B (or click
the "STACK N PRs" button) to link the existing open PRs into a stack.

This command performs several steps:
  1. Pushes all branches to the remote
  2. Creates new PRs for the included branches
  3. Updates base branches for existing PRs
  4. Creates or updates the stack on GitHub

In the editor, new PRs default to ready for review; switch any to draft with the
"CREATE AS" toggle. With --auto, new PRs are created as drafts unless you pass
--open.

Usage:
  gh stack submit [flags]

Examples:
  # Push and create/update PRs (opens the interactive editor)
  $ gh stack submit

  # Skip the editor and use auto-generated PR titles
  $ gh stack submit --auto

  # Mark new and existing PRs as ready for review
  $ gh stack submit --open

Flags:
      --auto            Use auto-generated PR titles without prompting
  -h, --help            help for submit
      --open            Mark new and existing PRs as ready for review
      --remote string   Remote to push to (defaults to auto-detected remote)
