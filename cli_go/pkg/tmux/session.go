package tmux

import "fmt"

// Layout controls how panes are arranged inside a window.
type Layout string

const (
	LayoutEvenHorizontal Layout = "even-horizontal"
	LayoutEvenVertical   Layout = "even-vertical"
	LayoutTiled          Layout = "tiled"
	LayoutMainHorizontal Layout = "main-horizontal"
	LayoutMainVertical   Layout = "main-vertical"
)

// PaneSpec describes a single pane: where it starts and what it runs.
type PaneSpec struct {
	WorkDir string // directory to open the pane in (empty = inherit)
	Command string // shell command to send (empty = just open shell)
}

// WindowSpec describes one tmux window and its pane layout.
type WindowSpec struct {
	Name   string
	Layout Layout
	Panes  []PaneSpec
}

// SessionSpec is the full desired state for a tmux session.
type SessionSpec struct {
	Name    string
	Windows []WindowSpec
}

// Realize creates the tmux session described by the spec.
// It returns an error if the session already exists.
// After this returns nil the caller should call Attach(spec.Name).
func (s *SessionSpec) Realize() error {
	if SessionExists(s.Name) {
		return fmt.Errorf("session %q already exists — use attach or choose a different name", s.Name)
	}

	for wi, win := range s.Windows {
		if err := s.realizeWindow(wi, win); err != nil {
			return err
		}
	}
	return nil
}

func (s *SessionSpec) realizeWindow(idx int, win WindowSpec) error {
	if len(win.Panes) == 0 {
		return nil
	}

	firstPane := win.Panes[0]

	if idx == 0 {
		// Create the session itself with the first pane.
		args := []string{"new-session", "-d", "-s", s.Name, "-n", win.Name}
		if firstPane.WorkDir != "" {
			args = append(args, "-c", firstPane.WorkDir)
		}
		if _, err := Run(args...); err != nil {
			return err
		}
	} else {
		// Subsequent windows.
		args := []string{"new-window", "-t", s.Name, "-n", win.Name}
		if firstPane.WorkDir != "" {
			args = append(args, "-c", firstPane.WorkDir)
		}
		if _, err := Run(args...); err != nil {
			return err
		}
	}

	// Send command to first pane if provided.
	if firstPane.Command != "" {
		target := fmt.Sprintf("%s:%d.0", s.Name, idx)
		if _, err := Run("send-keys", "-t", target, firstPane.Command, "Enter"); err != nil {
			return err
		}
	}

	// Split and populate the remaining panes.
	// We apply select-layout after every split so tmux redistributes space
	// before the next split — without this, panes shrink until tmux refuses
	// to split further ("no space for new pane").
	for pi, pane := range win.Panes[1:] {
		windowTarget := fmt.Sprintf("%s:%d", s.Name, idx)
		splitArgs := []string{"split-window", "-t", windowTarget}
		if pane.WorkDir != "" {
			splitArgs = append(splitArgs, "-c", pane.WorkDir)
		}
		if _, err := Run(splitArgs...); err != nil {
			return err
		}

		if _, err := Run("select-layout", "-t", windowTarget, string(win.Layout)); err != nil {
			return err
		}

		if pane.Command != "" {
			paneTarget := fmt.Sprintf("%s:%d.%d", s.Name, idx, pi+1)
			if _, err := Run("send-keys", "-t", paneTarget, pane.Command, "Enter"); err != nil {
				return err
			}
		}
	}

	return nil
}
