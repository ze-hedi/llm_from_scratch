package layout

import "github.com/yourusername/chatbot-tui/pkg/tmux"

// ThreePanes returns a session with one window split into 3 equal vertical panes.
func ThreePanes(sessionName string) tmux.SessionSpec {
	return tmux.SessionSpec{
		Name: sessionName,
		Windows: []tmux.WindowSpec{
			{
				Name:   "main",
				Layout: tmux.LayoutEvenHorizontal,
				Panes: []tmux.PaneSpec{
					{},
					{},
					{},
				},
			},
		},
	}
}

// Dev returns a session tailored for development:
//   - Window 0 "editor"  — one big pane (your editor lives here)
//   - Window 1 "shells"  — 3 horizontal panes for build / run / git
func Dev(sessionName, workDir string) tmux.SessionSpec {
	pane := func(cmd string) tmux.PaneSpec {
		return tmux.PaneSpec{WorkDir: workDir, Command: cmd}
	}

	return tmux.SessionSpec{
		Name: sessionName,
		Windows: []tmux.WindowSpec{
			{
				Name:   "editor",
				Layout: tmux.LayoutMainVertical,
				Panes:  []tmux.PaneSpec{{WorkDir: workDir}},
			},
			{
				Name:   "shells",
				Layout: tmux.LayoutEvenHorizontal,
				Panes:  []tmux.PaneSpec{pane(""), pane(""), pane("")},
			},
		},
	}
}

// Custom lets you define any number of panes in one window on the fly.
func Custom(sessionName string, layout tmux.Layout, panes []tmux.PaneSpec) tmux.SessionSpec {
	return tmux.SessionSpec{
		Name: sessionName,
		Windows: []tmux.WindowSpec{
			{Name: "main", Layout: layout, Panes: panes},
		},
	}
}
