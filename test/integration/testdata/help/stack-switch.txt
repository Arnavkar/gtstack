Show an interactive picker listing all branches in the current stack
and switch to the selected one.

Branches are displayed from top (furthest from trunk) to bottom (closest to
trunk) with their position number. Use the arrow keys to navigate and Enter
to select.

To move one branch up or down without an interactive picker, use
'gh stack up' or 'gh stack down' instead.

Usage:
  gh stack switch [flags]

Examples:
  # Open the branch picker for the current stack
  $ gh stack switch

Flags:
  -h, --help   help for switch
