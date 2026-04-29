---
description: Distributed version control system
tags: [vcs, development, workflow]
---

# git

## Setup
    git config --global user.name "Name"
    git config --global user.email "email"
    git init                              # initialize repo
    git clone <url>                       # clone remote repo

## Daily Workflow
    git status                            # show working tree status
    git diff                              # unstaged changes
    git diff --staged                     # staged changes
    git add <file>                        # stage file
    git add -p                            # stage interactively (hunks)
    git commit -m "message"              # commit
    git commit --amend                    # amend last commit

## Branching
    git branch                            # list local branches
    git branch -a                         # list all branches
    git checkout -b feature/name          # create and switch branch
    git switch -c feature/name            # modern alternative
    git switch main                       # switch to branch
    git branch -d feature/name            # delete merged branch
    git branch -D feature/name            # force delete

## Merging & Rebasing
    git merge feature/name                # merge into current branch
    git merge --no-ff feature/name        # always create merge commit
    git rebase main                       # rebase onto main
    git rebase -i HEAD~3                  # interactive rebase last 3 commits
    git cherry-pick <sha>                 # apply specific commit

## Remote
    git remote -v                         # list remotes
    git fetch origin                      # fetch without merging
    git pull origin main                  # fetch + merge
    git pull --rebase origin main         # fetch + rebase
    git push origin feature/name          # push branch
    git push -u origin feature/name       # push and set upstream
    git push --force-with-lease           # safe force push

## History
    git log --oneline --graph --all       # visual branch graph
    git log -p <file>                     # patch history for file
    git blame <file>                      # line-by-line authorship
    git show <sha>                        # show commit details

## Stash
    git stash                             # stash working changes
    git stash pop                         # restore latest stash
    git stash list                        # list stashes

## Undoing
    git restore <file>                    # discard working changes
    git restore --staged <file>           # unstage file
    git reset HEAD~1                      # undo last commit, keep changes
    git reset --hard HEAD~1               # undo last commit, discard changes
    git revert <sha>                      # create revert commit (safe)

## Tags & Releases
    git tag v1.0.0                        # lightweight tag
    git tag -a v1.0.0 -m "Release"       # annotated tag
    git push origin --tags                # push all tags
