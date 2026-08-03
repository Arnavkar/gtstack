Pull from remote and do a cascading rebase across the stack.

Ensures that each branch in the stack has the tip of the previous
layer in its commit history, rebasing if necessary.

Use --no-trunk to skip fetching and rebasing with the trunk branch.
Only the inter-branch rebases are performed (branch 2 onto branch 1,
branch 3 onto branch 2, etc.).

Usage:
  gh stack rebase [branch] [flags]

Examples:
  # Rebase the entire stack
  $ gh stack rebase

  # Only rebase from trunk to the current branch
  $ gh stack rebase --downstack

  # Only rebase from current branch to the top
  $ gh stack rebase --upstack

  # Rebase stack branches without pulling from or rebasing with trunk
  $ gh stack rebase --no-trunk

  # Continue after resolving conflicts
  $ gh stack rebase --continue

  # Abort and restore all branches
  $ gh stack rebase --abort

Flags:
      --abort                           Abort rebase and restore all branches
      --committer-date-is-author-date   Set the committer date to the author date during rebase
      --continue                        Continue rebase after resolving conflicts
      --downstack                       Only rebase branches from trunk to current branch
  -h, --help                            help for rebase
      --no-trunk                        Skip trunk — only rebase stack branches onto each other
      --preserve-dates                  Alias for --committer-date-is-author-date
      --remote string                   Remote to fetch from (defaults to auto-detected remote)
      --upstack                         Only rebase branches from current branch to top
