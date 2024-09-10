package main

import (
	lib "github.com/charles-m-knox/finance-planner-lib"
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
	Theme               string              `yaml:"theme"`
	// if true, results calculations will be faster for large date ranges,
	// as the terminal will not need to periodically re-render the page to
	// show status/progress messages for its work-in-progress calculations
	DisableResultsStatusMessages bool `yaml:"disableResultsStatusMessages"`
	// to save on memory, each time a change is made, a copy of the config is
	// added to the undo buffer, which can add up over time. If you're on a
	// system that struggles with gzip somehow, you can disable this feature
	// here at the cost of using more memory.
	DisableGzipCompressionInUndoBuffer bool `yaml:"disableGzipCompressionInUndoBuffer"`
}

type TableCell struct {
	Color  string
	Text   string
	Expand int
	Align  int
}
