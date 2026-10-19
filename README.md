# task-runner

A terminal UI for tracking a flat weekly task list. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Requirements

- Go 1.22+

## Run

```bash
go run .
```

Tasks persist to `~/.task-runner/tasks.json`. When a new week starts (Monday), the list resets so review always reflects the current week.

Tasks have a priority of **P1**, **P2**, or **P3**. The list is always ordered P1 → P3.

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `space` | Toggle done |
| `1` / `2` / `3` | Set priority of selected task |
| `a` | Add a task |
| `d` | Delete selected task |
| `r` | Weekly review |
| `q` | Quit |

In **add** mode: `1`/`2`/`3` sets priority (default P2), `enter` saves, `esc` cancels.  
In **review** mode: `esc` or `r` returns to the list.
