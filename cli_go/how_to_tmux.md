# How to launch the tmux app

## Prerequisites

Make sure the binary is compiled and tmux is in your PATH:

```bash
# Compile
go build -o cli .

# Add tmux to PATH (miniconda install)
export PATH="$PATH:/home/bouchehdahed/miniconda3/bin"
```

---

## Global flags

These flags are available on every subcommand:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--session` | `-s` | `dev` | Name of the tmux session to create |
| `--dir` | `-d` | current dir | Working directory applied to all panes |

---

## Modes

### 1. `three` — interactive setup

Launches a 5-step terminal wizard where you configure the session interactively:

1. Number of panes (1–9)
2. Layout (even-horizontal / even-vertical / tiled)
3. Working directory per pane
4. Session name
5. Confirmation

```bash
./cli tmux three
```

With a pre-filled session name:

```bash
./cli tmux three -s my-session
```

**Wizard controls:**

| Key | Action |
|---|---|
| `Tab` / `Enter` | Next step |
| `Shift+Tab` | Previous step |
| `← →` or `h l` | Change pane count or layout |
| `↑ ↓` | Switch between pane path inputs |
| `Esc` / `Ctrl+C` | Cancel |

---

### 2. `three --config` — headless setup from JSON file

Skips the wizard entirely. Reads the session configuration from a JSON file and launches tmux directly.

```bash
./cli tmux three --config path/to/config.json
```

**JSON schema:**

```json
{
  "session": "dev",
  "layout": "even-horizontal",
  "panes": [
    { "path": "/home/user/project", "command": "nvim ." },
    { "path": "/home/user/project", "command": "git log --oneline" },
    { "path": "/home/user/project" }
  ]
}
```

| Field | Required | Description |
|---|---|---|
| `session` | yes | tmux session name |
| `layout` | yes | one of the 5 layout values below |
| `panes` | yes | array of pane definitions (at least 1) |
| `panes[].path` | yes | working directory for the pane |
| `panes[].command` | no | shell command sent to the pane on launch |

**Available layouts:**

| Value | Description |
|---|---|
| `even-horizontal` | panes split side by side |
| `even-vertical` | panes stacked top to bottom |
| `tiled` | panes arranged in a grid |
| `main-horizontal` | one large pane on top, rest below |
| `main-vertical` | one large pane on the left, rest on the right |

The `command` field is optional. When set, the command is automatically sent to that pane as soon as the session is created (e.g. start an editor, run a server, tail logs).

---

### 3. `dev` — fixed development layout

Creates a pre-configured two-window session:

- Window `editor` — one full pane (for your editor)
- Window `shells` — 3 horizontal panes (for build / run / git)

All panes open in the same working directory.

```bash
./cli tmux dev
```

With a custom session name and directory:

```bash
./cli tmux dev -s my-project -d /home/user/project
```

---

### 4. `custom` — N blank panes

Creates one window with N panes all using `even-horizontal` layout. All panes open in the same directory.

```bash
./cli tmux custom --panes 4
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--panes` | `-p` | `3` | Number of panes to create |

With all flags:

```bash
./cli tmux custom -s my-session -d /home/user/project -p 5
```

---

### 5. `attach` — attach to an existing session

Re-attaches your terminal to a session that is already running.

```bash
./cli tmux attach -s my-session
```

---

### 6. `kill` — kill a session

Destroys a running session and all its panes.

```bash
./cli tmux kill -s my-session
```

---

## Quick reference

```bash
./cli tmux three                              # interactive wizard
./cli tmux three --config session.json        # headless from JSON
./cli tmux dev -s dev -d ~/project            # fixed dev layout
./cli tmux custom -p 4 -s work -d ~/project  # 4 blank panes
./cli tmux attach -s dev                      # reattach
./cli tmux kill -s dev                        # kill session
```
