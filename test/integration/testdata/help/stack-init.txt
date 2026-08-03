Initialize a new stack of branches in the current repository.

You can pass multiple branch names to create a multi-layer stack in one
command. Existing branches are adopted automatically; missing branches are
created. By default, the first branch is based on the default branch, and
each subsequent branch is based on the previous one.

Use --base to specify a different trunk branch.

Usage:
  gh stack init [branches...] [flags]

Examples:
  # Create a stack with a new branch
  $ gh stack init my-feature

  # Create a multi-layer stack at once
  $ gh stack init auth-layer api-routes ui-components

  # Adopt existing branches into a stack (bottom to top)
  $ gh stack init feat/auth feat/api feat/ui

  # Specify a different trunk branch
  $ gh stack init --base develop my-feature

Flags:
  -b, --base string   Trunk branch for stack (defaults to default branch)
  -h, --help          help for init
