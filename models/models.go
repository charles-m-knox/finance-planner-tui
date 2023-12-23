package models

import (
	"finance-planner-tui/lib"
)

type Profile struct {
	TX              []lib.TX `yaml:"transactions"`
	Name            string   `yaml:"name"`
	Modified        bool     `yaml:"-"`
	SelectedRow     int      `yaml:"selectedRow"`
	SelectedColumn  int      `yaml:"selectedColumn"`
	StartingBalance string   `yaml:"startingBalance"`
	StartDay        string   `yaml:"startDay"`
	StartMonth      string   `yaml:"startMonth"`
	StartYear       string   `yaml:"startYear"`
	EndDay          string   `yaml:"endDay"`
	EndMonth        string   `yaml:"endMonth"`
	EndYear         string   `yaml:"endYear"`
}

type Config struct {
	Keybindings         map[string][]string `yaml:"keybindings"`
	Profiles            []Profile           `yaml:"profiles"`
	UndoBufferMaxLength int                 `yaml:"undoBufferMaxLength"`
	Version             string              `yaml:"version"`
	// if true, results calculations will be faster for large date ranges,
	// as the terminal will not need to periodically re-render the page to
	// show status/progress messages for its work-in-progress calculations
	DisableResultsStatusMessages bool `yaml:"disableResultsStatusMessages"`
}
