---
description: Terminal multiplexer — manage multiple terminal sessions
tags: [terminal, multiplexer, productivity]
---

# tmux

Prefix key: Ctrl-b (default). All shortcuts below assume prefix first.

## Sessions
    tmux                                  # new session
    tmux new -s work                      # named session
    tmux ls                               # list sessions
    tmux attach -t work                   # attach to session
    tmux kill-session -t work             # kill session

    Prefix + d                            # detach
    Prefix + $                            # rename session
    Prefix + s                            # switch session (interactive)

## Windows (Tabs)
    Prefix + c                            # new window
    Prefix + ,                            # rename window
    Prefix + &                            # kill window
    Prefix + n / p                        # next / previous window
    Prefix + 0-9                          # switch to window by number
    Prefix + l                            # last (previous) window

## Panes (Splits)
    Prefix + %                            # vertical split
    Prefix + "                            # horizontal split
    Prefix + arrow                        # move between panes
    Prefix + z                            # zoom / unzoom pane
    Prefix + x                            # kill pane
    Prefix + q                            # show pane numbers
    Prefix + {                            # swap pane left
    Prefix + }                            # swap pane right
    Prefix + Ctrl-arrow                   # resize pane

## Copy Mode
    Prefix + [                            # enter copy mode (vi/emacs keys)
    q                                     # exit copy mode
    Space                                 # start selection (vi mode)
    Enter                                 # copy selection
    Prefix + ]                            # paste

## Configuration (~/.tmux.conf)
    set -g prefix C-a                     # change prefix to Ctrl-a
    set -g mouse on                       # enable mouse support
    set -g base-index 1                   # start window numbering at 1
    set -g history-limit 50000            # scrollback buffer size
    set-window-option -g mode-keys vi     # vi keys in copy mode

## Useful One-liners
    tmux new -s main \; split-window -h \; split-window -v   # 3-pane layout
    tmux send-keys -t work "ls -la" Enter   # send command to session
