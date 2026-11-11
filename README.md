# task-runner

A terminal UI for tracking a flat weekly task list. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Requirements

- Go 1.22+

## Run

```bash
go run .
```

Tasks persist to `~/.task-runner/tasks.json`. When a new week starts (Monday), completed tasks are cleared and incomplete ones are archived so you can pull them forward from **history**. Review always reflects the current week.

Tasks have a priority of **P1**, **P2**, or **P3**. The list is always ordered parked before active, then P1 → P3. Only **one active P1** is allowed at a time; park a task (`p`) to free the slot — parked P1s do not count toward the limit.

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `space` | Toggle done |
| `1` / `2` / `3` | Set priority of selected task |
| `p` | Park / unpark selected task |
| `a` | Add a task |
| `d` | Delete selected task |
| `r` | Weekly review |
| `h` | Previous weeks (history) |
| `q` | Quit |

In **add** mode: `ctrl+1` / `ctrl+2` / `ctrl+3` sets priority (default P2), `enter` saves, `esc` cancels. Digits type normally in the title.  
In **review** mode: `esc` or `r` returns to the list.  
In **history** mode: `h`/`l` older/newer week, `j`/`k` move, `m` move task to this week, `d` delete leftover, `esc` back to the list.

## Releases

Push a version tag to publish macOS (amd64, arm64) and Windows (amd64) binaries via GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Download the assets from the [Releases](https://github.com/allenyhu/task-runner/releases) page.
