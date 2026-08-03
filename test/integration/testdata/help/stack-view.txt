View the current stack as a list showing branches and PR status.

Status icons:
  ✓  PR merged
  ◎  PR queued
  ○  PR open
  ⚠  Needs rebase

The current branch is highlighted. Use --short for a compact one-line-per-branch
view, or --json for machine-readable output.

Usage:
  gh stack view [flags]

Examples:
  # Show the stack (default interactive view)
  $ gh stack view

  # Show compact output
  $ gh stack view --short

  # Output as JSON
  $ gh stack view --json

Flags:
  -h, --help    help for view
      --json    Output stack data as JSON
  -s, --short   Show compact output
