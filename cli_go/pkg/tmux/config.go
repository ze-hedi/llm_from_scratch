package tmux

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileConfig is the schema for a --config JSON file passed to the tmux command.
//
// Example:
//
//	{
//	  "session": "dev",
//	  "layout": "even-horizontal",
//	  "panes": [
//	    { "path": "/home/user/project", "command": "nvim ." },
//	    { "path": "/home/user/project" },
//	    { "path": "/tmp" }
//	  ]
//	}
type FileConfig struct {
	Session string       `json:"session"`
	Layout  string       `json:"layout"`
	Panes   []PaneConfig `json:"panes"`
}

// PaneConfig describes a single pane entry inside a FileConfig.
type PaneConfig struct {
	Path    string `json:"path"`
	Command string `json:"command,omitempty"`
}

var validLayouts = map[string]Layout{
	"even-horizontal": LayoutEvenHorizontal,
	"even-vertical":   LayoutEvenVertical,
	"tiled":           LayoutTiled,
	"main-horizontal": LayoutMainHorizontal,
	"main-vertical":   LayoutMainVertical,
}

// LoadConfig reads a JSON file at path and returns a validated FileConfig.
func LoadConfig(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *FileConfig) validate() error {
	if c.Session == "" {
		return fmt.Errorf("config: \"session\" must not be empty")
	}
	if _, ok := validLayouts[c.Layout]; !ok {
		return fmt.Errorf("config: unknown layout %q — valid values: even-horizontal, even-vertical, tiled, main-horizontal, main-vertical", c.Layout)
	}
	if len(c.Panes) == 0 {
		return fmt.Errorf("config: \"panes\" must contain at least one entry")
	}
	for i, p := range c.Panes {
		if p.Path == "" {
			return fmt.Errorf("config: pane %d is missing \"path\"", i)
		}
	}
	return nil
}

// ToSessionSpec converts the FileConfig into a SessionSpec ready to be realized.
func (c *FileConfig) ToSessionSpec() SessionSpec {
	panes := make([]PaneSpec, len(c.Panes))
	for i, p := range c.Panes {
		panes[i] = PaneSpec{WorkDir: p.Path, Command: p.Command}
	}

	return SessionSpec{
		Name: c.Session,
		Windows: []WindowSpec{
			{Name: "main", Layout: validLayouts[c.Layout], Panes: panes},
		},
	}
}
