package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"

	"github.com/adrg/xdg"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

func tmp() {
	log.Println(lib.IsWeekday("Monday"))
}

const (
	// old
	maxLogLines = 10000
	BOX_LIST    = "list"
	BOX_STDOUT  = "stdout"
	BOX_STDERR  = "stderr"

	// new
	PAGE_PROFILES = "Profiles"
	PAGE_RESULTS  = "Results"
)

var (
	// finance planner things
	app                         *tview.Application
	config                      Config
	selectedProfile             *Profile
	pages                       *tview.Pages
	currentlyFocusedBox         string
	profilesPage                *tview.Flex
	transactionsPage            *tview.Flex
	resultsPage                 *tview.Flex
	billEditorPage              *tview.Flex
	transactionsTableSortColumn string
	// the previously focused primitive
	previous tview.Primitive
	// profilesPage items:
	profileList            *tview.List
	statusText             *tview.TextView
	transactionsTable      *tview.Table
	transactionsInputField *tview.InputField

	// old things that need to be removed
	lastCmd                *Command
	list                   *tview.List
	layout                 *tview.Flex
	info                   *tview.TextView
	errors                 *tview.TextView
	exitCode               *tview.TextView
	bottomLeftText         *tview.TextView
	bottomLeftSearch       *tview.InputField
	bottomLeftBox          *tview.Box
	globalProcessesRunning int
	stdoutLines            []string
	stderrLines            []string
	runIndex               sync.Map
	isSearching            bool
	searchTerm             string
	keybindings            []Keybinding
	filteredResults        []string
)

type Profile struct {
	TX       []lib.TX `yaml:"transactions"`
	Name     string   `yaml:"name"`
	modified bool
}

type Config struct {
	Commands                    []ConfigCommand `yaml:"commands"`
	Keybindings                 []Keybinding    `yaml:"keybindings"`
	IdleRefreshRateMs           int             `yaml:"idleRefreshRateMs"`
	ProcessRunningRefreshRateMs int             `yaml:"processRunningRefreshRateMs"`

	Profiles []Profile `yaml:"profiles"`
}

type Keybinding struct {
	Action     string
	Keybinding string
}

func loadConfig() (c Config, err error) {
	// TODO: later on, try current dir, then xdg_config_dir, then xdg_user_dir
	xdgConfig := path.Join(xdg.ConfigHome, "frequencmd", "config.yml")
	xdgHome := path.Join(xdg.Home, "frequencmd", "config.yml")
	curConf := "config.yml"

	b, err := os.ReadFile(curConf)
	if err == nil {
		err = yaml.Unmarshal(b, &c)
		if err != nil {
			return c, fmt.Errorf(
				"failed to read config from %v: %v",
				curConf,
				err.Error(),
			)
		}
		return c, nil
	}

	b, err = os.ReadFile(xdgConfig)
	if err == nil {
		err = yaml.Unmarshal(b, &c)
		if err != nil {
			return c, fmt.Errorf(
				"failed to read config from %v: %v",
				xdgConfig,
				err.Error(),
			)
		}
		return c, nil
	}

	b, err = os.ReadFile(xdgHome)
	if err == nil {
		err = yaml.Unmarshal(b, &c)
		if err != nil {
			return c, fmt.Errorf(
				"failed to read config from %v: %v",
				xdgHome,
				err.Error(),
			)
		}
		return c, nil
	}

	return c, fmt.Errorf(
		"failed to read config from %v, %v, and %v: %v",
		curConf,
		xdgConfig,
		xdgHome,
		err.Error(),
	)
}

func logOutput(output io.ReadCloser, lines *[]string, prefix string, view *tview.TextView, color tcell.Color) {
	*lines = []string{}
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		*lines = append(*lines, line)
		var sb strings.Builder
		linesLen := len(*lines)
		for i := linesLen - 1; i >= linesLen-maxLogLines && i >= 0; i-- {
			fmt.Fprintf(&sb, "%v%v:[%v] %v\n", prefix, getNowStr(), color, (*lines)[i])
		}
		view.SetText(sb.String())
		app.QueueUpdateDraw(func() {})
		// log.Print(line)
	}
}

func getNowStr() string {
	return time.Now().Format("15:04:05")
}

func setLastCommandText(cmd *Command) {
	lastCmd = cmd
	exitCode.SetTitle(fmt.Sprintf("exit code for: %v", cmd.Label))
}

func setBottomLeftText(t string) {
	bottomLeftText.SetText(fmt.Sprintf("[white][ctrl+c][gray] to quit | %v", t))
}

func pidRunningDrawLoop() {
	for {
		// setStatusText(fmt.Sprintf("%v loops", loops))
		sleepTime := time.Duration(config.IdleRefreshRateMs) * time.Millisecond

		processesRunning := 0
		keysToDelete := []int64{}
		shouldRedrawApp := false
		runIndex.Range(func(key, value any) bool {
			if value == true {
				if app != nil {
					shouldRedrawApp = true
				}

				processesRunning += 1
				errors.ScrollToEnd()
				info.ScrollToEnd()
			} else {
				keysToDelete = append(keysToDelete, key.(int64))
			}

			return true
		})

		// last process finished running; do a one-time update on the status bar
		if processesRunning == 0 && globalProcessesRunning > 0 {
			globalProcessesRunning = 0
			setBottomLeftText(fmt.Sprintf("%v running", globalProcessesRunning))
			app.QueueUpdateDraw(func() {})
		}

		// sync up with the global counter
		globalProcessesRunning = processesRunning
		if globalProcessesRunning > 0 {
			setBottomLeftText(fmt.Sprintf("%v running", globalProcessesRunning))
			app.QueueUpdateDraw(func() {})
		}

		if shouldRedrawApp {
			// draw a little faster if we know something is running
			sleepTime = time.Duration(config.ProcessRunningRefreshRateMs) * time.Millisecond
			setBottomLeftText(fmt.Sprintf("%v running", globalProcessesRunning))
			app.QueueUpdateDraw(func() {})
		}

		for key := range keysToDelete {
			keyToDelete := key
			runIndex.Delete(keyToDelete)
		}

		time.Sleep(sleepTime)
	}
}

func runCommand(command *Command /* command string, args []string, env []string */) {
	jobId := time.Now().UnixNano()
	runIndex.Store(jobId, true)

	setLastCommandText(command)
	exitCode.SetText(fmt.Sprintf("[gray]%v [aqua] running command:[white] %v", getNowStr(), command.Label))

	info.Clear()
	errors.Clear()

	cmd := exec.Command(command.Command, command.Args...)
	cmd.Env = append(cmd.Env, command.Env...)

	cmd.Stdout = info
	cmd.Stderr = errors
	// Run the command
	err := cmd.Run()
	if err != nil {
		runIndex.Store(jobId, false)
		errors.SetText(fmt.Sprintf("error running command: %v", err.Error()))
		exitCode.SetText(fmt.Sprintf("[red] Exit code: %v", cmd.ProcessState.ExitCode()))
		app.QueueUpdateDraw(func() {})
		return
	}
	runIndex.Store(jobId, false)

	exitCode.SetText(fmt.Sprintf("[green] Exit code: %v", cmd.ProcessState.ExitCode()))
	app.QueueUpdateDraw(func() {})
}

func FuzzyFind(input string, commands []Command) []string {
	commandList := []string{}
	for _, c := range commands {
		commandList = append(commandList, c.Label)
	}
	return fuzzy.Find(input, commandList)
}

type Command struct {
	Color   tcell.Color
	Label   string
	Command string
	Args    []string
	Env     []string
}

type ConfigCommand struct {
	Label   string   `yaml:"label"`
	Command string   `yaml:"command"`
	Shell   string   `yaml:"shell"`
	Args    string   `yaml:"args"`
	Env     []string `yaml:"env"`
}

func getFilteredList(l *tview.List, commands []Command, filterString string) {
	if l != nil {
		l.Clear()
	}

	filteredCommands := FuzzyFind(filterString, commands)

	filteredResults = []string{}

	for i := range commands {
		c := &(commands[i])
		matchedMarker := ""
		if !slices.Contains(filteredCommands, c.Label) {
			continue
		}

		l.AddItem(fmt.Sprintf("%v%v", (*c).Label, matchedMarker), "", 0, func() { go runCommand(c) }).ShowSecondaryText(false) // .SetMainTextColor(c.Color)
		filteredResults = append(filteredResults, (*c).Label)
	}

	l.SetBorder(true)
}

// func getLayout(commands []Command) {
// 	getFilteredList(list, commands, searchTerm)

// 	info = tview.NewTextView().SetTextAlign(tview.AlignLeft).SetText("").SetDynamicColors(true)
// 	errors = tview.NewTextView().SetTextAlign(tview.AlignLeft).SetText("").SetDynamicColors(true)
// 	exitCode = tview.NewTextView().SetTextAlign(tview.AlignLeft).SetText("").SetDynamicColors(true)
// 	bottomLeftText = tview.NewTextView().SetTextAlign(tview.AlignLeft).SetDynamicColors(true)
// 	bottomLeftSearch = tview.NewInputField()

// 	setBottomLeftText("0 processes running")

// 	info.SetBorder(true).SetTitle("stdout")
// 	errors.SetBorder(true).SetTitle("stderr")
// 	exitCode.SetBorder(true).SetTitle("exit code")
// 	bottomLeftText.SetBorder(true)
// 	bottomLeftSearch.SetBorder(true)
// 	bottomLeftSearch.SetDisabled(true)

// 	logViews := tview.NewFlex().SetDirection(tview.FlexRow).
// 		AddItem(info, 0, 5, false).
// 		AddItem(errors, 0, 5, false)
// 		// AddItem(exitCode, 0, 1, false)

// 	mainColumns := tview.NewFlex().
// 		SetDirection(tview.FlexColumn).
// 		AddItem(list, 0, 1, true).
// 		AddItem(logViews, 0, 2, false)

// 	bottomRow := tview.NewFlex().SetDirection(tview.FlexColumn)

// 	bottomLeftFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
// 		AddItem(bottomLeftText, 0, 1, false).
// 		AddItem(bottomLeftSearch, 0, 1, false)

// 	bottomRow.AddItem(bottomLeftFlex, 0, 2, false).
// 		AddItem(exitCode, 0, 1, false)

// 	layout = tview.NewFlex().
// 		SetDirection(tview.FlexRow).
// 		AddItem(mainColumns, 0, 1, false).
// 		AddItem(bottomRow, 3, 0, false)
// }

func endSearch(msg string) {
	isSearching = false
	searchTerm = ""
	bottomLeftSearch.SetDisabled(true)
	app.SetFocus(list)
	setBottomLeftText(msg)
}

func populateProfilesPage(doProfile, doBills bool) {
	if doProfile {
		profileList.Clear()
	}
	// if doBills {
	// 	billsSummary.Clear()
	// }

	// var sb strings.Builder
	for i := range config.Profiles {
		profile := &(config.Profiles[i])
		if doProfile {
			profileList.AddItem(fmt.Sprintf("[blue]%v", profile.Name), "", 0, func() {
				selectedProfile = profile
				populateProfilesPage(false, true)
				getTransactionsTable()
			})
		}

		if selectedProfile == nil || selectedProfile.Name != profile.Name {
			continue
		}

		if !doBills {
			continue
		}

		// for i, tx := range profile.TX {
		// 	// only show top 25 transactions for profile
		// 	if i > 25 {
		// 		break
		// 	}
		// 	if tx.Amount < 0 {
		// 		fmt.Fprintf(&sb, "[yellow]%v [gray]| [yellow]%v\n", tx.Name, lib.FormatAsCurrency(tx.Amount))
		// 	} else {
		// 		fmt.Fprintf(&sb, "[green]%v [gray]| [green]%v\n", tx.Name, lib.FormatAsCurrency(tx.Amount))
		// 	}
		// }
	}

	// billsSummary.SetText(sb.String())
}

func getTransactionsTable() {
	transactionsTable.Clear()

	// determine the current sort
	// currentSort := strings.TrimSuffix(strings.TrimSuffix(transactionsTableSortColumn, c.Asc), c.Desc)
	currentSort := ""
	currentSortDir := ""
	if strings.HasSuffix(transactionsTableSortColumn, c.Asc) {
		currentSort = strings.Split(transactionsTableSortColumn, c.Asc)[0]
		// currentSortDir = c.Asc
		currentSortDir = "↑"
	} else if strings.HasSuffix(transactionsTableSortColumn, c.Desc) {
		currentSort = strings.Split(transactionsTableSortColumn, c.Desc)[0]
		currentSortDir = "↓"
	}

	cellColumnOrderText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ORDER, c.ColumnOrder)
	if currentSort == c.ColumnOrder {
		cellColumnOrderText = fmt.Sprintf("%v%v", currentSortDir, cellColumnOrderText)
	}
	cellColumnAmountText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_AMOUNT, c.ColumnAmount)
	if currentSort == c.ColumnAmount {
		cellColumnAmountText = fmt.Sprintf("%v%v", currentSortDir, cellColumnAmountText)
	}
	cellColumnActiveText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ACTIVE, c.ColumnActive)
	if currentSort == c.ColumnActive {
		cellColumnActiveText = fmt.Sprintf("%v%v", currentSortDir, cellColumnActiveText)
	}
	cellColumnNameText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_NAME, c.ColumnName)
	if currentSort == c.ColumnName {
		cellColumnNameText = fmt.Sprintf("%v%v", currentSortDir, cellColumnNameText)
	}
	cellColumnFrequencyText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_FREQUENCY, c.ColumnFrequency)
	if currentSort == c.ColumnFrequency {
		cellColumnFrequencyText = fmt.Sprintf("%v%v", currentSortDir, cellColumnFrequencyText)
	}
	cellColumnIntervalText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_INTERVAL, c.ColumnInterval)
	if currentSort == c.ColumnInterval {
		cellColumnIntervalText = fmt.Sprintf("%v%v", currentSortDir, cellColumnIntervalText)
	}
	cellColumnMondayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_MONDAY, c.ColumnMonday)
	if currentSort == c.ColumnMonday {
		cellColumnMondayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnMondayText)
	}
	cellColumnTuesdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_TUESDAY, c.ColumnTuesday)
	if currentSort == c.ColumnTuesday {
		cellColumnTuesdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnTuesdayText)
	}
	cellColumnWednesdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_WEDNESDAY, c.ColumnWednesday)
	if currentSort == c.ColumnWednesday {
		cellColumnWednesdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnWednesdayText)
	}
	cellColumnThursdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_THURSDAY, c.ColumnThursday)
	if currentSort == c.ColumnThursday {
		cellColumnThursdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnThursdayText)
	}
	cellColumnFridayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_FRIDAY, c.ColumnFriday)
	if currentSort == c.ColumnFriday {
		cellColumnFridayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnFridayText)
	}
	cellColumnSaturdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_SATURDAY, c.ColumnSaturday)
	if currentSort == c.ColumnSaturday {
		cellColumnSaturdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnSaturdayText)
	}
	cellColumnSundayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_SUNDAY, c.ColumnSunday)
	if currentSort == c.ColumnSunday {
		cellColumnSundayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnSundayText)
	}
	cellColumnStartsText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_STARTS, c.ColumnStarts)
	if currentSort == c.ColumnStarts {
		cellColumnStartsText = fmt.Sprintf("%v%v", currentSortDir, cellColumnStartsText)
	}
	cellColumnEndsText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ENDS, c.ColumnEnds)
	if currentSort == c.ColumnEnds {
		cellColumnEndsText = fmt.Sprintf("%v%v", currentSortDir, cellColumnEndsText)
	}
	cellColumnNoteText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_NOTE, c.ColumnNote)
	if currentSort == c.ColumnNote {
		cellColumnNoteText = fmt.Sprintf("%v%v", currentSortDir, cellColumnNoteText)
	}

	cellColumnOrder := tview.NewTableCell(cellColumnOrderText)
	cellColumnAmount := tview.NewTableCell(cellColumnAmountText)
	cellColumnActive := tview.NewTableCell(cellColumnActiveText)
	cellColumnName := tview.NewTableCell(cellColumnNameText)
	cellColumnFrequency := tview.NewTableCell(cellColumnFrequencyText)
	cellColumnInterval := tview.NewTableCell(cellColumnIntervalText)
	cellColumnMonday := tview.NewTableCell(cellColumnMondayText)
	cellColumnTuesday := tview.NewTableCell(cellColumnTuesdayText)
	cellColumnWednesday := tview.NewTableCell(cellColumnWednesdayText)
	cellColumnThursday := tview.NewTableCell(cellColumnThursdayText)
	cellColumnFriday := tview.NewTableCell(cellColumnFridayText)
	cellColumnSaturday := tview.NewTableCell(cellColumnSaturdayText)
	cellColumnSunday := tview.NewTableCell(cellColumnSundayText)
	cellColumnStarts := tview.NewTableCell(cellColumnStartsText)
	cellColumnEnds := tview.NewTableCell(cellColumnEndsText)
	cellColumnNote := tview.NewTableCell(cellColumnNoteText)
	// cellColumnID := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ID, c.ColumnID))
	// cellColumnCreatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",c.ColumnCreatedAt))
	// cellColumnUpdatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",c.ColumnUpdatedAt))

	cellColumnName.SetExpansion(1)
	cellColumnNote.SetExpansion(1)

	transactionsTable.SetCell(0, 0, cellColumnOrder)
	transactionsTable.SetCell(0, 1, cellColumnAmount)
	transactionsTable.SetCell(0, 2, cellColumnActive)
	transactionsTable.SetCell(0, 3, cellColumnName)
	transactionsTable.SetCell(0, 4, cellColumnFrequency)
	transactionsTable.SetCell(0, 5, cellColumnInterval)
	transactionsTable.SetCell(0, 6, cellColumnMonday)
	transactionsTable.SetCell(0, 7, cellColumnTuesday)
	transactionsTable.SetCell(0, 8, cellColumnWednesday)
	transactionsTable.SetCell(0, 9, cellColumnThursday)
	transactionsTable.SetCell(0, 10, cellColumnFriday)
	transactionsTable.SetCell(0, 11, cellColumnSaturday)
	transactionsTable.SetCell(0, 12, cellColumnSunday)
	transactionsTable.SetCell(0, 13, cellColumnStarts)
	transactionsTable.SetCell(0, 14, cellColumnEnds)
	transactionsTable.SetCell(0, 15, cellColumnNote)
	// transactionsTable.SetCell(0, 16, cellColumnID)
	// transactionsTable.SetCell(0, 17, cellColumnCreatedAt)
	// transactionsTable.SetCell(0, 18, cellColumnUpdatedAt)

	if selectedProfile != nil {
		if transactionsTableSortColumn != "" {
			sort.SliceStable(
				selectedProfile.TX,
				func(i, j int) bool {
					tj := (selectedProfile.TX)[j]
					ti := (selectedProfile.TX)[i]

					switch transactionsTableSortColumn {

					// invisible order column (default when no sort is set)
					case c.None:
						return tj.Order > ti.Order

					// Order
					case fmt.Sprintf("%v%v", c.ColumnOrder, c.Asc):
						return ti.Order > tj.Order
					case fmt.Sprintf("%v%v", c.ColumnOrder, c.Desc):
						return ti.Order < tj.Order

					// active
					case fmt.Sprintf("%v%v", c.ColumnActive, c.Asc):
						return ti.Active
					case fmt.Sprintf("%v%v", c.ColumnActive, c.Desc):
						return tj.Active

					// weekdays
					case fmt.Sprintf("%v%v", c.WeekdayMonday, c.Asc):
						return ti.HasWeekday(c.WeekdayMondayInt)
					case fmt.Sprintf("%v%v", c.WeekdayMonday, c.Desc):
						return tj.HasWeekday(c.WeekdayMondayInt)
					case fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Asc):
						return ti.HasWeekday(c.WeekdayTuesdayInt)
					case fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Desc):
						return tj.HasWeekday(c.WeekdayTuesdayInt)
					case fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Asc):
						return ti.HasWeekday(c.WeekdayWednesdayInt)
					case fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Desc):
						return tj.HasWeekday(c.WeekdayWednesdayInt)
					case fmt.Sprintf("%v%v", c.WeekdayThursday, c.Asc):
						return ti.HasWeekday(c.WeekdayThursdayInt)
					case fmt.Sprintf("%v%v", c.WeekdayThursday, c.Desc):
						return tj.HasWeekday(c.WeekdayThursdayInt)
					case fmt.Sprintf("%v%v", c.WeekdayFriday, c.Asc):
						return ti.HasWeekday(c.WeekdayFridayInt)
					case fmt.Sprintf("%v%v", c.WeekdayFriday, c.Desc):
						return tj.HasWeekday(c.WeekdayFridayInt)
					case fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Asc):
						return ti.HasWeekday(c.WeekdaySaturdayInt)
					case fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Desc):
						return tj.HasWeekday(c.WeekdaySaturdayInt)
					case fmt.Sprintf("%v%v", c.WeekdaySunday, c.Asc):
						return ti.HasWeekday(c.WeekdaySundayInt)
					case fmt.Sprintf("%v%v", c.WeekdaySunday, c.Desc):
						return tj.HasWeekday(c.WeekdaySundayInt)

					// other columns
					case fmt.Sprintf("%v%v", c.ColumnAmount, c.Asc):
						return ti.Amount > tj.Amount
					case fmt.Sprintf("%v%v", c.ColumnAmount, c.Desc):
						return ti.Amount < tj.Amount

					case fmt.Sprintf("%v%v", c.ColumnFrequency, c.Asc):
						return ti.Frequency > tj.Frequency
					case fmt.Sprintf("%v%v", c.ColumnFrequency, c.Desc):
						return ti.Frequency < tj.Frequency

					case fmt.Sprintf("%v%v", c.ColumnInterval, c.Asc):
						return ti.Interval > tj.Interval
					case fmt.Sprintf("%v%v", c.ColumnInterval, c.Desc):
						return ti.Interval < tj.Interval
					case fmt.Sprintf("%v%v", c.ColumnNote, c.Asc):
						return strings.ToLower(ti.Note) > strings.ToLower(tj.Note)
					case fmt.Sprintf("%v%v", c.ColumnNote, c.Desc):
						return strings.ToLower(ti.Note) < strings.ToLower(tj.Note)

					case fmt.Sprintf("%v%v", c.ColumnName, c.Asc):
						return strings.ToLower(ti.Name) > strings.ToLower(tj.Name)
					case fmt.Sprintf("%v%v", c.ColumnName, c.Desc):
						return strings.ToLower(ti.Name) < strings.ToLower(tj.Name)

					case fmt.Sprintf("%v%v", c.ColumnID, c.Asc):
						return strings.ToLower(ti.ID) > strings.ToLower(tj.ID)
					case fmt.Sprintf("%v%v", c.ColumnID, c.Desc):
						return strings.ToLower(ti.ID) < strings.ToLower(tj.ID)

					case fmt.Sprintf("%v%v", c.ColumnCreatedAt, c.Asc):
						return tj.CreatedAt.After(tj.CreatedAt)
					case fmt.Sprintf("%v%v", c.ColumnCreatedAt, c.Desc):
						return ti.CreatedAt.Before(tj.CreatedAt)

					case fmt.Sprintf("%v%v", c.ColumnUpdatedAt, c.Asc):
						return ti.UpdatedAt.After(tj.UpdatedAt)
					case fmt.Sprintf("%v%v", c.ColumnUpdatedAt, c.Desc):
						return ti.UpdatedAt.Before(tj.UpdatedAt)

					case fmt.Sprintf("%v%v", c.ColumnStarts, c.Asc):
						ist := fmt.Sprintf("%v-%v-%v", tj.StartsYear, tj.StartsMonth, tj.StartsDay)
						jst := fmt.Sprintf("%v-%v-%v", ti.StartsYear, ti.StartsMonth, ti.StartsDay)
						return ist > jst
					case fmt.Sprintf("%v%v", c.ColumnStarts, c.Desc):
						ist := fmt.Sprintf("%v-%v-%v", tj.StartsYear, tj.StartsMonth, tj.StartsDay)
						jst := fmt.Sprintf("%v-%v-%v", ti.StartsYear, ti.StartsMonth, ti.StartsDay)
						return ist < jst

					case fmt.Sprintf("%v%v", c.ColumnEnds, c.Asc):
						jend := fmt.Sprintf("%v-%v-%v", tj.EndsYear, tj.EndsMonth, tj.EndsDay)
						iend := fmt.Sprintf("%v-%v-%v", ti.EndsYear, ti.EndsMonth, ti.EndsDay)
						return iend > jend
					case fmt.Sprintf("%v%v", c.ColumnEnds, c.Desc):
						jend := fmt.Sprintf("%v-%v-%v", tj.EndsYear, tj.EndsMonth, tj.EndsDay)
						iend := fmt.Sprintf("%v-%v-%v", ti.EndsYear, ti.EndsMonth, ti.EndsDay)
						return iend < jend

					default:
						return false
						// return txs[j].Date.After(txs[i].Date)
					}
				},
			)
		}
		// start by populating the table with the columns first
		for i, tx := range selectedProfile.TX {
			isPositiveAmount := tx.Amount >= 0
			amountColor := c.COLOR_COLUMN_AMOUNT
			if isPositiveAmount {
				amountColor = c.COLOR_COLUMN_AMOUNT_POSITIVE
			}

			cellOrder := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ORDER, tx.Order)).SetAlign(tview.AlignCenter)
			cellAmount := tview.NewTableCell(fmt.Sprintf("%v%v", amountColor, lib.FormatAsCurrency(tx.Amount))).SetAlign(tview.AlignCenter)

			activeText := "X"
			if !tx.Active {
				activeText = " "
			}

			cellActive := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ACTIVE, activeText)).SetAlign(tview.AlignCenter)
			cellName := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_NAME, tx.Name)).SetAlign(tview.AlignLeft)
			cellFrequency := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_FREQUENCY, tx.Frequency)).SetAlign(tview.AlignCenter)
			cellInterval := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_INTERVAL, tx.Interval)).SetAlign(tview.AlignCenter)

			mondayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_MONDAY)
			if !tx.HasWeekday(c.WeekdayMondayInt) {
				mondayText = "[white] "
			}
			tuesdayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_TUESDAY)
			if !tx.HasWeekday(c.WeekdayTuesdayInt) {
				tuesdayText = "[white] "
			}
			wednesdayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_WEDNESDAY)
			if !tx.HasWeekday(c.WeekdayWednesdayInt) {
				wednesdayText = "[white] "
			}
			thursdayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_THURSDAY)
			if !tx.HasWeekday(c.WeekdayThursdayInt) {
				thursdayText = "[white] "
			}
			fridayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_FRIDAY)
			if !tx.HasWeekday(c.WeekdayFridayInt) {
				fridayText = "[white] "
			}
			saturdayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_SATURDAY)
			if !tx.HasWeekday(c.WeekdaySaturdayInt) {
				saturdayText = "[white] "
			}
			sundayText := fmt.Sprintf("%vX", c.COLOR_COLUMN_SUNDAY)
			if !tx.HasWeekday(c.WeekdaySundayInt) {
				sundayText = "[white] "
			}

			cellMonday := tview.NewTableCell(mondayText).SetAlign(tview.AlignCenter)
			cellTuesday := tview.NewTableCell(tuesdayText).SetAlign(tview.AlignCenter)
			cellWednesday := tview.NewTableCell(wednesdayText).SetAlign(tview.AlignCenter)
			cellThursday := tview.NewTableCell(thursdayText).SetAlign(tview.AlignCenter)
			cellFriday := tview.NewTableCell(fridayText).SetAlign(tview.AlignCenter)
			cellSaturday := tview.NewTableCell(saturdayText).SetAlign(tview.AlignCenter)
			cellSunday := tview.NewTableCell(sundayText).SetAlign(tview.AlignCenter)

			cellStarts := tview.NewTableCell(fmt.Sprintf("%v%v-%v-%v", c.COLOR_COLUMN_STARTS, tx.StartsYear, tx.StartsMonth, tx.StartsDay)).SetAlign(tview.AlignCenter)
			cellEnds := tview.NewTableCell(fmt.Sprintf("%v%v-%v-%v", c.COLOR_COLUMN_ENDS, tx.EndsYear, tx.EndsMonth, tx.EndsDay)).SetAlign(tview.AlignCenter)

			cellNote := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_NOTE, tx.Note))

			// cellID := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ID, tx.ID))
			// cellCreatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",fmt.Sprintf("%v", tx.CreatedAt)))
			// cellUpdatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",fmt.Sprintf("%v", tx.UpdatedAt)))

			cellName.SetExpansion(1)
			cellNote.SetExpansion(1)

			if tx.Selected {
				cellOrder.SetBackgroundColor(tcell.ColorDarkBlue)
				cellAmount.SetBackgroundColor(tcell.ColorDarkBlue)
				cellActive.SetBackgroundColor(tcell.ColorDarkBlue)
				cellName.SetBackgroundColor(tcell.ColorDarkBlue)
				cellFrequency.SetBackgroundColor(tcell.ColorDarkBlue)
				cellInterval.SetBackgroundColor(tcell.ColorDarkBlue)
				cellMonday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellTuesday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellWednesday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellThursday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellFriday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellSaturday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellSunday.SetBackgroundColor(tcell.ColorDarkBlue)
				cellStarts.SetBackgroundColor(tcell.ColorDarkBlue)
				cellEnds.SetBackgroundColor(tcell.ColorDarkBlue)
				cellNote.SetBackgroundColor(tcell.ColorDarkBlue)
				// cellID.SetBackgroundColor(tcell.ColorDarkBlue)
				// cellCreatedAt.SetBackgroundColor(tcell.ColorDarkBlue)
				// cellUpdatedAt.SetBackgroundColor(tcell.ColorDarkBlue)
			}

			transactionsTable.SetCell(i+1, 0, cellOrder)
			transactionsTable.SetCell(i+1, 1, cellAmount)
			transactionsTable.SetCell(i+1, 2, cellActive)
			transactionsTable.SetCell(i+1, 3, cellName)
			transactionsTable.SetCell(i+1, 4, cellFrequency)
			transactionsTable.SetCell(i+1, 5, cellInterval)
			transactionsTable.SetCell(i+1, 6, cellMonday)
			transactionsTable.SetCell(i+1, 7, cellTuesday)
			transactionsTable.SetCell(i+1, 8, cellWednesday)
			transactionsTable.SetCell(i+1, 9, cellThursday)
			transactionsTable.SetCell(i+1, 10, cellFriday)
			transactionsTable.SetCell(i+1, 11, cellSaturday)
			transactionsTable.SetCell(i+1, 12, cellSunday)
			transactionsTable.SetCell(i+1, 13, cellStarts)
			transactionsTable.SetCell(i+1, 14, cellEnds)
			transactionsTable.SetCell(i+1, 15, cellNote)
			// transactionsTable.SetCell(i+1, 16, cellID)
			// transactionsTable.SetCell(i+1, 17, cellCreatedAt)
			// transactionsTable.SetCell(i+1, 18, cellUpdatedAt)
		}

		transactionsTable.SetSelectedFunc(func(row, column int) {
			// get the current profile & transaction
			i := 0

			// based on the row, find the actual transaction definition
			// example: row 5 = TX 4 because of table's headers
			for i = range selectedProfile.TX {
				txi := row - 1
				if i == txi {
					i = txi
					break
				}
			}

			switch column {
			case c.COLUMN_ORDER:
				if row == 0 {
					setTransactionsTableSort(c.ColumnOrder)
					return
				}
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 1 {
							activateTransactionsInputFieldNoAutocompleteReset("invalid order given:", fmt.Sprint(selectedProfile.TX[i].Order))
							return
						}

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Order = int(d)
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.COLOR_COLUMN_ORDER,
									selectedProfile.TX[j].Order,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				activateTransactionsInputField("order:", fmt.Sprint(selectedProfile.TX[i].Order))
			case c.COLUMN_AMOUNT:
				if row == 0 {
					setTransactionsTableSort(c.ColumnAmount)
					return
				}

				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						a := lib.ParseDollarAmount(transactionsInputField.GetText(), false)
						isPositiveAmount := a >= 0
						amountColor := c.COLOR_COLUMN_AMOUNT
						if isPositiveAmount {
							amountColor = c.COLOR_COLUMN_AMOUNT_POSITIVE
						}

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Amount = int(a)
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									amountColor,
									lib.FormatAsCurrency(selectedProfile.TX[j].Amount),
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				activateTransactionsInputField("amount (start with + or $+ for positive):", lib.FormatAsCurrency(selectedProfile.TX[i].Amount))
			case c.COLUMN_ACTIVE:
				if row == 0 {
					setTransactionsTableSort(c.ColumnActive)
					return
				}

				newValue := !selectedProfile.TX[i].Active
				selectedProfile.TX[i].Active = !selectedProfile.TX[i].Active

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {

						activeText := "X"
						if !newValue {
							activeText = " "
						}
						selectedProfile.TX[j].Active = newValue

						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ACTIVE, activeText))
					}
				}

				modified()
			case c.COLUMN_NAME:
				if row == 0 {
					setTransactionsTableSort(c.ColumnName)
					return
				}
				activateTransactionsInputField("edit name:", selectedProfile.TX[i].Name)
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						break
					default:
						selectedProfile.TX[i].Name = transactionsInputField.GetText()

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Name = selectedProfile.TX[i].Name
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_NAME, selectedProfile.TX[i].Name))
							}
						}

						modified()
					}
					deactivateTransactionsInputField()
				})
			case c.COLUMN_FREQUENCY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnFrequency)
					return
				}
				activateTransactionsInputField("weekly|monthly|yearly:", selectedProfile.TX[i].Frequency)
				saveFunc := func(newValue string) {
					// save the changes
					validatedFrequency := strings.TrimSpace(strings.ToUpper(newValue))
					switch validatedFrequency {
					case c.WEEKLY:
						fallthrough
					case c.MONTHLY:
						fallthrough
					case c.YEARLY:
						break
					default:
						transactionsInputField.SetLabel("invalid value - can only be weekly, monthly, or yearly:")
						return
					}
					selectedProfile.TX[i].Frequency = validatedFrequency

					// update all selected values as well as the current one
					for j := range selectedProfile.TX {
						if selectedProfile.TX[j].Selected || j == i {
							selectedProfile.TX[j].Frequency = selectedProfile.TX[i].Frequency
							transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_FREQUENCY, selectedProfile.TX[i].Frequency))
						}
					}

					modified()
				}
				transactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
					return fuzzy.Find(strings.TrimSpace(strings.ToUpper(currentText)), []string{
						c.MONTHLY,
						c.YEARLY,
						c.WEEKLY,
					})
				})
				transactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
					saveFunc(text)
					deactivateTransactionsInputField()
					return true
				})
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						break
					default:
						saveFunc(transactionsInputField.GetText())
					}
					deactivateTransactionsInputField()
				})
			case c.COLUMN_INTERVAL:
				if row == 0 {
					setTransactionsTableSort(c.ColumnInterval)
					return
				}
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 1 {
							activateTransactionsInputFieldNoAutocompleteReset("invalid interval given:", fmt.Sprint(selectedProfile.TX[i].Interval))
							return
						}

						selectedProfile.TX[i].Interval = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Interval = selectedProfile.TX[i].Interval
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.COLOR_COLUMN_INTERVAL,
									selectedProfile.TX[i].Interval,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				activateTransactionsInputField("interval:", fmt.Sprint(selectedProfile.TX[i].Interval))
			case c.COLUMN_MONDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnMonday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayMondayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayMondayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayMondayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayMondayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_MONDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayMondayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_MONDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_TUESDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnTuesday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayTuesdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayTuesdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayTuesdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayTuesdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_TUESDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayTuesdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_TUESDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_WEDNESDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnWednesday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayWednesdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayWednesdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayWednesdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayWednesdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_WEDNESDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayWednesdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_WEDNESDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_THURSDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnThursday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayThursdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayThursdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayThursdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayThursdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_THURSDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayThursdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_THURSDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_FRIDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnFriday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayFridayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayFridayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayFridayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayFridayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_FRIDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayFridayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_FRIDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_SATURDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnSaturday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdaySaturdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdaySaturdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdaySaturdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdaySaturdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_SATURDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdaySaturdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_SATURDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_SUNDAY:
				if row == 0 {
					setTransactionsTableSort(c.ColumnSunday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdaySundayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdaySundayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdaySundayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdaySundayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%vX", c.COLOR_COLUMN_SUNDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdaySundayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_SUNDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_STARTS:
				if row == 0 {
					setTransactionsTableSort(c.ColumnStarts)
					return
				}
				// first, prompt for the year
				// then, prompt for month
				// then, prompt for day
				dayFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 1 || d > 31 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid day given:", fmt.Sprint(selectedProfile.TX[i].StartsDay))
							return
						}

						selectedProfile.TX[i].StartsDay = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsDay = selectedProfile.TX[i].StartsDay
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_STARTS,
									selectedProfile.TX[j].StartsYear,
									selectedProfile.TX[j].StartsMonth,
									selectedProfile.TX[j].StartsDay,
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
					}
				}

				monthFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d > 12 || d < 1 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", fmt.Sprint(selectedProfile.TX[i].StartsMonth))
							return
						}

						selectedProfile.TX[i].StartsMonth = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsMonth = selectedProfile.TX[i].StartsMonth
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_STARTS,
									selectedProfile.TX[j].StartsYear,
									selectedProfile.TX[j].StartsMonth,
									selectedProfile.TX[j].StartsDay,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", fmt.Sprint(selectedProfile.TX[i].StartsDay))
						defer transactionsInputField.SetDoneFunc(dayFunc)
					}
				}

				yearFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", fmt.Sprint(selectedProfile.TX[i].StartsYear))
							return
						}

						selectedProfile.TX[i].StartsYear = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsYear = selectedProfile.TX[i].StartsYear
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_STARTS,
									selectedProfile.TX[j].StartsYear,
									selectedProfile.TX[j].StartsMonth,
									selectedProfile.TX[j].StartsDay,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", fmt.Sprint(selectedProfile.TX[i].StartsMonth))
						defer transactionsInputField.SetDoneFunc(monthFunc)
					}
				}

				transactionsInputField.SetDoneFunc(yearFunc)
				activateTransactionsInputField("year:", fmt.Sprint(selectedProfile.TX[i].StartsYear))
			case c.COLUMN_ENDS:
				if row == 0 {
					setTransactionsTableSort(c.ColumnEnds)
					return
				}
				// first, prompt for the year
				// then, prompt for month
				// then, prompt for day
				dayFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 1 || d > 31 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid day given:", fmt.Sprint(selectedProfile.TX[i].EndsDay))
							return
						}

						selectedProfile.TX[i].EndsDay = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsDay = selectedProfile.TX[i].EndsDay
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_ENDS,
									selectedProfile.TX[j].EndsYear,
									selectedProfile.TX[j].EndsMonth,
									selectedProfile.TX[j].EndsDay,
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
					}
				}

				monthFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d > 12 || d < 1 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", fmt.Sprint(selectedProfile.TX[i].EndsMonth))
							return
						}

						selectedProfile.TX[i].EndsMonth = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsMonth = selectedProfile.TX[i].EndsMonth
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_ENDS,
									selectedProfile.TX[j].EndsYear,
									selectedProfile.TX[j].EndsMonth,
									selectedProfile.TX[j].EndsDay,
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", fmt.Sprint(selectedProfile.TX[i].EndsDay))
						defer transactionsInputField.SetDoneFunc(dayFunc)
					}
				}

				yearFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", fmt.Sprint(selectedProfile.TX[i].EndsYear))
							return
						}

						selectedProfile.TX[i].EndsYear = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsYear = selectedProfile.TX[i].EndsYear
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_ENDS,
									selectedProfile.TX[j].EndsYear,
									selectedProfile.TX[j].EndsMonth,
									selectedProfile.TX[j].EndsDay,
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", fmt.Sprint(selectedProfile.TX[i].EndsMonth))
						defer transactionsInputField.SetDoneFunc(monthFunc)
					}
				}

				transactionsInputField.SetDoneFunc(yearFunc)
				activateTransactionsInputField("year:", fmt.Sprint(selectedProfile.TX[i].EndsYear))
			case c.COLUMN_NOTE:
				if row == 0 {
					setTransactionsTableSort(c.ColumnNote)
					return
				}
				activateTransactionsInputField("edit note:", selectedProfile.TX[i].Note)
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						break
					default:
						// save the changes
						selectedProfile.TX[i].Note = transactionsInputField.GetText()
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Note = selectedProfile.TX[i].Note
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_NOTE, selectedProfile.TX[j].Note))
							}
						}

						modified()
					}
					deactivateTransactionsInputField()
				})
			case c.COLUMN_ID:
				// pass for now
			case c.COLUMN_CREATEDAT:
				// pass for now
			case c.COLUMN_UPDATEDAT:
				// pass for now
			default:
				break
			}
		})
	}

	transactionsTable.SetTitle("Transactions")
	transactionsTable.SetBorders(false).
		SetSelectable(true, true). // set row & cells to be selectable
		SetSeparator(' ')
}

func resetTransactionsInputFieldAutocomplete() {
	// transactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
	// 	return true
	// })
	transactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
		return []string{}
	})
}

func deactivateTransactionsInputField() {
	// transactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
	// 	return []string{}
	// })
	// transactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
	// 	return true
	// })
	transactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	transactionsInputField.SetLabel("[gray] editor appears here when editing")
	transactionsInputField.SetText("")

	if previous != nil {
		app.SetFocus(previous)
	}
}

// focuses the transactions input field, updates its label, and sets
// its background color to something noticeable
func activateTransactionsInputField(msg, value string) {
	resetTransactionsInputFieldAutocomplete()

	transactionsInputField.SetFieldBackgroundColor(tcell.ColorDimGray)
	transactionsInputField.SetLabel(fmt.Sprintf("[aqua] %v", msg))
	transactionsInputField.SetText(value)

	// don't mess with the previously stored focus if the text field is already
	// focused
	currentFocus := app.GetFocus()
	if currentFocus == transactionsInputField {
		return
	}

	previous = currentFocus
	app.SetFocus(transactionsInputField)
}

// focuses the transactions input field, updates its label, and sets
// its background color to something noticeable - in some cases, the
// resetTransactionsInputFieldAutocomplete cannot be called without risking
// an infinite loop, so this function does not call it
func activateTransactionsInputFieldNoAutocompleteReset(msg, value string) {
	transactionsInputField.SetFieldBackgroundColor(tcell.ColorDimGray)
	transactionsInputField.SetLabel(fmt.Sprintf("[aqua] %v", msg))
	transactionsInputField.SetText(value)

	// don't mess with the previously stored focus if the text field is already
	// focused
	currentFocus := app.GetFocus()
	if currentFocus == transactionsInputField {
		return
	}

	previous = currentFocus
	app.SetFocus(transactionsInputField)
}

func setStatusNoChanges() {
	statusText.SetText("[gray] no changes")
}

func modified() {
	if selectedProfile != nil {
		selectedProfile.modified = true
		// transactionsTable.SetTitle("Transactions*")
		statusText.SetText("[white] [ctrl+s] to save")
	}
}

// returns a simple flex view with two columns:
// - a list of profiles (left side)
// - a quick summary of bills / stats for the highlighted profile (right side)
func getProfilesFlex() {
	profileList = tview.NewList()
	profileList.SetBorder(true)
	profileList.ShowSecondaryText(false)

	statusText = tview.NewTextView()
	statusText.SetBorder(true)
	statusText.SetDynamicColors(true)
	setStatusNoChanges()

	profilesLeftSide := tview.NewFlex().SetDirection(tview.FlexRow)
	profilesLeftSide.AddItem(profileList, 0, 1, true).
		AddItem(statusText, 3, 0, true)

	transactionsPage = tview.NewFlex().SetDirection(tview.FlexRow)
	transactionsTable = tview.NewTable().SetFixed(1, 1)
	transactionsInputField = tview.NewInputField()

	transactionsTable.SetBorder(true)
	transactionsInputField.SetBorder(true)

	transactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	transactionsInputField.SetLabel("[gray] editor appears here when editing")

	getTransactionsTable()

	transactionsPage.AddItem(transactionsTable, 0, 1, false).
		AddItem(transactionsInputField, 3, 0, false)

	populateProfilesPage(true, true)

	profilesPage = tview.NewFlex().SetDirection(tview.FlexColumn)
	profilesPage.AddItem(profilesLeftSide, 0, 1, true).
		AddItem(transactionsPage, 0, 10, false)
}

func setTransactionsTableSort(column string) {
	transactionsTableSortColumn = lib.GetNextSort(transactionsTableSortColumn, column)
	defer getTransactionsTable()
}

func main() {
	runIndex = sync.Map{}
	var err error

	config, err = loadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err.Error())
	}

	if len(config.Profiles) > 0 {
		selectedProfile = &(config.Profiles[0])
	}

	app = tview.NewApplication()
	pages = tview.NewPages()
	getProfilesFlex()
	resultsPage = tview.NewFlex()
	billEditorPage = tview.NewFlex()

	pages.AddPage(PAGE_PROFILES, profilesPage, true, true).
		AddPage(PAGE_RESULTS, resultsPage, true, true)

	pages.SwitchToPage(PAGE_PROFILES)

	app.SetFocus(profileList)
	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Rune() == '/' {
			// if !isSearching {
			// 	isSearching = true
			// 	searchTerm = ""
			// 	// app.SetFocus(bottomLeftSearch)
			// 	// getLayout(commands)
			// 	setBottomLeftText("[aqua]searching:")
			// 	bottomLeftSearch.SetText("")
			// 	bottomLeftSearch.SetDisabled(false)
			// 	// searchTerm = fmt.Sprintf("%v%v", searchTerm, string(e.Rune()))
			// 	app.SetFocus(bottomLeftSearch)
			// 	bottomLeftSearch.SetChangedFunc(func(text string) {
			// 		if list != nil {
			// 			list.Clear()
			// 		}
			// 		searchTerm = text
			// 		// getFilteredList(list, *commands, searchTerm)
			// 	})
			// 	return nil
			// }
		} else if e.Key() == tcell.KeyEnter {
			// if isSearching {
			// 	if len(filteredResults) == 0 {
			// 		// getFilteredList(list, *commands, "")
			// 	}
			// 	endSearch(fmt.Sprintf("searched: %v", searchTerm))
			// 	return nil
			// }
		} else if e.Key() == tcell.KeyEscape {
			currentFocus := app.GetFocus()
			switch currentFocus {
			case transactionsInputField:
				return e
			case transactionsTable:
				anySelected := false
				for i := range selectedProfile.TX {
					if selectedProfile.TX[i].Selected {
						anySelected = true
						selectedProfile.TX[i].Selected = false
					}
				}
				if !anySelected {
					app.SetFocus(profileList)
					return nil
				}
				getTransactionsTable()
				cr, cc := transactionsTable.GetSelection()
				transactionsTable.Select(cr, cc)
				app.SetFocus(transactionsTable)
			default:
				app.Stop()
			}

			// if isSearching {
			// endSearch(fmt.Sprintf("canceled search: %v", searchTerm))
			// return nil
			// } else if
			// else {
			// app.Stop()
			// }
		} else if e.Key() == tcell.KeyLeft {
			// if isSearching {
			// 	return e
			// }
			// switch currentlyFocusedBox {
			// case BOX_STDOUT:
			// 	app.SetFocus(list)
			// 	currentlyFocusedBox = BOX_LIST
			// case BOX_STDERR:
			// 	app.SetFocus(list)
			// 	currentlyFocusedBox = BOX_LIST
			// case BOX_LIST:
			// 	fallthrough
			// default:
			// 	app.SetFocus(info)
			// 	currentlyFocusedBox = BOX_STDOUT
			// }
			// return nil
		} else if e.Key() == tcell.KeyRight {
			// if isSearching {
			// 	return e
			// }
			// switch currentlyFocusedBox {
			// case BOX_STDOUT:
			// 	app.SetFocus(list)
			// 	currentlyFocusedBox = BOX_LIST
			// case BOX_STDERR:
			// 	app.SetFocus(list)
			// 	currentlyFocusedBox = BOX_LIST
			// case BOX_LIST:
			// 	fallthrough
			// default:
			// 	app.SetFocus(info)
			// 	currentlyFocusedBox = BOX_STDOUT
			// }
			// return nil
		} else if e.Key() == tcell.KeyTab {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsInputField:
					return nil
				case profileList:
					app.SetFocus(transactionsTable)
				case transactionsTable:
					// get the height & width of the transactions table
					r := transactionsTable.GetRowCount() - 1
					c := transactionsTable.GetColumnCount() - 1
					cr, cc := transactionsTable.GetSelection()
					nc := cc + 1
					nr := cr
					if nc > c {
						nc -= 1
						nr += 1
						if nr > r {
							nc = 0
							nr = 0
						}
					}
					transactionsTable.Select(nr, nc)
					app.SetFocus(transactionsTable)
				default:
					app.SetFocus(profileList)
				}
				return nil
			case PAGE_RESULTS:
				return e
			}
		} else if e.Key() == tcell.KeyBacktab {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsInputField:
					return nil
				case profileList:
					app.SetFocus(transactionsTable)
				case transactionsTable:
					// get the height & width of the transactions table
					r := transactionsTable.GetRowCount() - 1
					c := transactionsTable.GetColumnCount() - 1
					cr, cc := transactionsTable.GetSelection()
					nc := cc - 1
					nr := cr
					if nc <= 0 {
						nc += 1
						nr -= 1
						if nr <= 0 {
							nc = c
							nr = r
						}
					}
					transactionsTable.Select(nr, nc)
					app.SetFocus(transactionsTable)
				default:
					app.SetFocus(profileList)
				}
				return nil
			case PAGE_RESULTS:
				return e
			}
		} else if e.Key() == tcell.KeyPgUp {
			// if isSearching {
			// 	return e
			// }
			// switch currentlyFocusedBox {
			// case BOX_STDOUT:
			// 	app.SetFocus(info)
			// 	r, c := info.GetScrollOffset()
			// 	_, _, _, h := info.GetRect()
			// 	newRow := r - h + 2 // the borders add some extra distance
			// 	if newRow < 0 {
			// 		newRow = 0
			// 	}
			// 	info.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_STDERR:
			// 	app.SetFocus(errors)
			// 	r, c := errors.GetScrollOffset()
			// 	_, _, _, h := errors.GetRect()
			// 	newRow := r - h + 2 // the borders add some extra distance
			// 	if newRow < 0 {
			// 		newRow = 0
			// 	}
			// 	errors.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_LIST:
			// 	fallthrough
			// default:
			// 	app.SetFocus(list)
			// 	return e
			// }
		} else if e.Key() == tcell.KeyPgDn {
			// if isSearching {
			// 	return e
			// }
			// switch currentlyFocusedBox {
			// case BOX_STDOUT:
			// 	app.SetFocus(info)
			// 	r, c := info.GetScrollOffset()
			// 	_, _, _, h := info.GetRect()
			// 	newRow := r + h - 2 // the borders add some extra distance
			// 	info.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_STDERR:
			// 	app.SetFocus(errors)
			// 	r, c := info.GetScrollOffset()
			// 	_, _, _, h := errors.GetRect()
			// 	newRow := r + h - 2 // the borders add some extra distance
			// 	errors.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_LIST:
			// 	fallthrough
			// default:
			// 	app.SetFocus(list)
			// 	return e
			// }
		} else if e.Key() == tcell.KeyUp {
			switch app.GetFocus() {
			case transactionsInputField:
				return nil
			default:
				return e
			}
			// if isSearching {
			// 	return e
			// }
			// switch currentlyFocusedBox {
			// case BOX_STDOUT:
			// 	app.SetFocus(info)
			// 	r, c := info.GetScrollOffset()
			// 	newRow := r - 1
			// 	info.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_STDERR:
			// 	app.SetFocus(errors)
			// 	r, c := errors.GetScrollOffset()
			// 	newRow := r - 1
			// 	errors.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_LIST:
			// 	fallthrough
			// default:
			// 	app.SetFocus(list)
			// 	return e
			// }
		} else if e.Key() == tcell.KeyDown {
			switch app.GetFocus() {
			case transactionsInputField:
				return nil
			default:
				return e
			}
			// if isSearching {
			// 	return e
			// }
			// switch currentlyFocusedBox {
			// case BOX_STDOUT:
			// 	app.SetFocus(info)
			// 	r, c := info.GetScrollOffset()
			// 	newRow := r + 1
			// 	info.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_STDERR:
			// 	app.SetFocus(errors)
			// 	r, c := info.GetScrollOffset()
			// 	newRow := r + 1
			// 	errors.ScrollTo(newRow, c)
			// 	return nil
			// case BOX_LIST:
			// 	fallthrough
			// default:
			// 	app.SetFocus(list)
			// 	return e
			// }
		} else if e.Key() == tcell.KeyHome {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsTable:
					cr, _ := transactionsTable.GetSelection()
					transactionsTable.Select(cr, 0)
					app.SetFocus(transactionsTable)
					return nil
				default:
					return e
				}
			case PAGE_RESULTS:
				return e
			}
		} else if e.Key() == tcell.KeyEnd {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsTable:
					c := transactionsTable.GetColumnCount() - 1
					cr, _ := transactionsTable.GetSelection()
					transactionsTable.Select(cr, c)
					app.SetFocus(transactionsTable)
					return nil
				default:
					return e
				}
			case PAGE_RESULTS:
				return e
			}
		} else if e.Key() == tcell.KeyCtrlS {
			b, err := yaml.Marshal(config)
			if err != nil {
				statusText.SetText("failed to marshal")
				return nil
			}

			err = os.WriteFile("config.yml", b, os.FileMode(0o644))
			if err != nil {
				statusText.SetText("failed to save")
				return nil
			}

			selectedProfile.modified = false
			statusText.SetText("[gray] saved changes")
			return nil
		} else if e.Key() == tcell.KeyCtrlD || e.Key() == tcell.KeyCtrlN || e.Rune() == 'a' || e.Rune() == 'n' {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsInputField:
					return e
				case transactionsTable:
					// duplicate the current transaction
					// get the height & width of the transactions table
					cr, cc := transactionsTable.GetSelection()
					actual := cr - 1 // skip header
					nt := []lib.TX{}
					duplicating := e.Key() == tcell.KeyCtrlD
					// iterate through the list once to find how many selected
					// items there are
					numSelected := 0
					for i := range selectedProfile.TX {
						if selectedProfile.TX[i].Selected {
							numSelected += 1
						}
					}
					for i := range selectedProfile.TX {
						if (i == actual && numSelected <= 1) || (selectedProfile.TX[i].Selected && duplicating) {
							now := time.Now()
							oneMonth := now.Add(time.Hour * 24 * 31)

							// keep track of the highest order in a temporary
							// slice
							largestOrderHolder := []lib.TX{}
							largestOrderHolder = append(largestOrderHolder, selectedProfile.TX...)
							largestOrderHolder = append(largestOrderHolder, nt...)

							newTX := lib.TX{
								Order:       lib.GetLargestOrder(largestOrderHolder) + 1,
								Amount:      500,
								Active:      true,
								Name:        "New",
								Frequency:   c.MONTHLY,
								Interval:    1,
								StartsDay:   now.Day(),
								StartsMonth: int(now.Month()),
								StartsYear:  now.Year(),
								EndsDay:     oneMonth.Day(),
								EndsMonth:   int(oneMonth.Month()),
								EndsYear:    oneMonth.Year(),
								ID:          uuid.NewString(),
								CreatedAt:   now,
								UpdatedAt:   now,
							}

							if duplicating {
								newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1
								newTX.Amount = selectedProfile.TX[i].Amount
								newTX.Active = selectedProfile.TX[i].Active
								newTX.Name = selectedProfile.TX[i].Name
								newTX.Note = selectedProfile.TX[i].Note
								newTX.RRule = selectedProfile.TX[i].RRule
								newTX.Frequency = selectedProfile.TX[i].Frequency
								newTX.Interval = selectedProfile.TX[i].Interval
								newTX.Weekdays = selectedProfile.TX[i].Weekdays
								newTX.StartsDay = selectedProfile.TX[i].StartsDay
								newTX.StartsMonth = selectedProfile.TX[i].StartsMonth
								newTX.StartsYear = selectedProfile.TX[i].StartsYear
								newTX.EndsDay = selectedProfile.TX[i].EndsDay
								newTX.EndsMonth = selectedProfile.TX[i].EndsMonth
								newTX.EndsYear = selectedProfile.TX[i].EndsYear
							}

							nt = append(nt, newTX)
						}
					}
					if len(nt) > 0 {
						selectedProfile.TX = slices.Insert(selectedProfile.TX, actual, nt...)
						getTransactionsTable()
						transactionsTable.Select(cr, cc)
						app.SetFocus(transactionsTable)
					}
				default:
					return e
				}
			case PAGE_RESULTS:
				return e
			}
		} else if e.Key() == tcell.KeyDelete {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsInputField:
					return e
				case transactionsTable:
					// duplicate the current transaction
					// get the height & width of the transactions table
					cr, cc := transactionsTable.GetSelection()
					actual := cr - 1 // skip header
					for i := len(selectedProfile.TX) - 1; i > 0; i-- {
						if selectedProfile.TX[i].Selected || i == actual {
							selectedProfile.TX = slices.Delete(selectedProfile.TX, i, i+1)
						}
					}
					// for i := range selectedProfile.TX {
					// 	if i == actual {
					// 		selectedProfile.TX = slices.Delete(selectedProfile.TX, i, i+1)
					// 	}
					// }
					getTransactionsTable()
					transactionsTable.Select(cr, cc)
					app.SetFocus(transactionsTable)
				default:
					app.SetFocus(profileList)
				}
				return nil
			case PAGE_RESULTS:
				return e
			}
		} else if e.Rune() == ' ' {
			pageName, _ := pages.GetFrontPage()
			switch pageName {
			case PAGE_PROFILES:
				switch app.GetFocus() {
				case transactionsInputField:
					return e
				case transactionsTable:
					// duplicate the current transaction
					// get the height & width of the transactions table
					cr, cc := transactionsTable.GetSelection()
					actual := cr - 1 // skip header
					for i := range selectedProfile.TX {
						if i == actual {
							selectedProfile.TX[i].Selected = !selectedProfile.TX[i].Selected
							break
						}
					}
					getTransactionsTable()
					transactionsTable.Select(cr, cc)
					app.SetFocus(transactionsTable)
				default:
					return e
				}
			case PAGE_RESULTS:
				return e
			}
		} else {
			// if isSearching {
			// 	setBottomLeftText("[aqua]searching:")
			// 	bottomLeftSearch.SetDisabled(false)
			// 	bottomLeftSearch.Focus(func(p tview.Primitive) {})
			// 	searchTerm = fmt.Sprintf("%v%v", searchTerm, string(e.Rune()))
			// 	bottomLeftSearch.SetChangedFunc(func(text string) {
			// 		if list != nil {
			// 			list.Clear()
			// 		}
			// 		searchTerm = text
			// 		getFilteredList(list, commands, searchTerm)
			// 		bottomLeftSearch.SetText(searchTerm)
			// 		bottomLeftSearch.Focus(func(p tview.Primitive) {})
			// 	})
			// 	return nil
			// }
			// if !isSearching {
			// 	app.SetFocus(list)
			// }
		}
		return e
	})

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
