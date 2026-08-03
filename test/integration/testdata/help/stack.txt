Stacked PRs let you break a large change into a chain of pull requests
that build on each other. Use `gh stack` to create and manage your stack
locally, then push to GitHub to create your stack of PRs.

Usage:
  gh stack [command]

Examples:
  # Start a new stack targeting your default branch
  $ gh stack init

  # Or turn an existing set of branches into a stack
  $ gh stack init branch1 branch2 branch3

  # Make changes and commit, then add a branch to the stack
  $ gh stack add branch4

  # Push all branches and create/update PRs on GitHub
  $ gh stack submit

  # Keep your local in sync with remote
  $ gh stack sync

Stack management:
  add         Add a new branch on top of the current stack
  checkout    Checkout a stack by stack number, PR number, PR URL, or branch name
  init        Initialize a new stack
  modify      Interactively restructure a stack
  unstack     Remove a stack locally and on GitHub
  view        View the current stack

Remote operations:
  link        Link PRs into a stack on GitHub without local tracking
  merge       Merge a stack of pull requests
  push        Push active branches in the current stack to the remote
  rebase      Rebase a stack of branches
  submit      Create a stack of PRs on GitHub
  sync        Sync the current stack with the remote

Navigation:
  bottom      Check out the bottom branch of the stack (closest to the trunk)
  down        Check out a branch further down in the stack (closer to the trunk)
  switch      Interactively switch to another branch in the stack
  top         Check out the top branch of the stack (furthest from the trunk)
  trunk       Check out the trunk branch of the stack
  up          Check out a branch further up in the stack (further from the trunk)

Utilities:
  alias       Create a shell alias for gh stack
  feedback    Submit feedback for gh-stack

Flags:
  -h, --help      help for stack
  -v, --version   version for stack

Use "gh stack [command] --help" for more information about a command.

Learn more:
  Documentation: https://gh.io/stacks
  Feedback: https://gh.io/stacks-feedback
