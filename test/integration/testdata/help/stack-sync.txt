Fetch, rebase, push, and sync PR state for the current stack.

This command performs a safe synchronization:

  1. Fetches the latest changes from the remote
  2. Reconciles the stack on GitHub with your local stack: pulls down
     branches for any PRs added to the stack on GitHub, or prompts you to
     resolve a divergence in an interactive terminal
  3. Fast-forwards the trunk branch to match the remote
  4. Cascade-rebases stack branches onto their updated parents
  5. Pushes all branches atomically (using --force-with-lease --atomic)
  6. Syncs PR state from GitHub
  7. Links the stack's open PRs into a stack on GitHub (creating or updating
     the remote stack object) when two or more PRs exist

If PRs have been added to the stack on GitHub, their branches are pulled
down and appended to your local stack so it mirrors the remote. A clean
"remote is ahead" update happens automatically without prompting. If the
local and remote stacks have diverged, sync prompts (in an interactive
terminal) to use the remote as the source of truth, delete the stack on
GitHub and recreate it later with sync/submit, or cancel. Cancelling — or a
divergence in a non-interactive terminal — aborts the sync without pushing
branches or updating PRs.

If a rebase conflict is detected, all branches are restored to their
original state and you are advised to run "gh stack rebase" to resolve
conflicts interactively.

Sync never opens pull requests — use "gh stack submit" for that. It only
links PRs that already exist. The final message reflects what happened:
"Stack synced" means the stack object on GitHub now matches your local
stack, while "Branches synced" means the branches were rebased and pushed
but no remote stack object was created or updated (for example, when fewer
than two PRs exist yet).

Use --prune to delete local branches for merged PRs. Stack metadata is
preserved so that rebase and display logic continue to work correctly.
If you are on a branch that would be pruned, your checkout is moved to
the first active branch in the stack, or the trunk if all are merged.

Usage:
  gh stack sync [flags]

Flags:
  -h, --help            help for sync
      --prune           Delete local branches for merged PRs
      --remote string   Remote to fetch from and push to (defaults to auto-detected remote)
