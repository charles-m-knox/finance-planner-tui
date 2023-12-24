package constants

const (
	ColumnOrder     = "Order"
	ColumnAmount    = "Amount"    // int in cents; 500 = $5.00
	ColumnActive    = "Active"    // bool true/false
	ColumnName      = "Name"      // editable string
	ColumnFrequency = "Frequency" // dropdown, monthly/daily/weekly/yearly
	ColumnInterval  = "Interval"  // integer, occurs every x frequency
	ColumnMonday    = "Monday"    // bool
	ColumnTuesday   = "Tuesday"   // bool
	ColumnWednesday = "Wednesday" // bool
	ColumnThursday  = "Thursday"  // bool
	ColumnFriday    = "Friday"    // bool
	ColumnSaturday  = "Saturday"  // bool
	ColumnSunday    = "Sunday"    // bool
	ColumnStarts    = "Starts"    // string
	ColumnEnds      = "Ends"      // string
	ColumnNote      = "Note"      // editable string
	ColumnID        = "ID"
	ColumnCreatedAt = "CreatedAt"
	ColumnUpdatedAt = "UpdatedAt"

	WeekdayMonday    = "Monday"
	WeekdayTuesday   = "Tuesday"
	WeekdayWednesday = "Wednesday"
	WeekdayThursday  = "Thursday"
	WeekdayFriday    = "Friday"
	WeekdaySaturday  = "Saturday"
	WeekdaySunday    = "Sunday"

	WEEKLY  = "WEEKLY"
	MONTHLY = "MONTHLY"
	YEARLY  = "YEARLY"

	New = "New"

	None = "none"
	Desc = "Desc"
	Asc  = "Asc"

	StartTodayPreset = "today"
	OneYear          = "1year"
	FiveYear         = "5year"

	DefaultConfig          = "config.yml"
	DefaultConfigParentDir = "finance-planner-tui"

	ConfigVersion = "1"
)

const ResetStyle = "[-:-:-:-]"

const (
	WeekdayMondayInt = iota
	WeekdayTuesdayInt
	WeekdayWednesdayInt
	WeekdayThursdayInt
	WeekdayFridayInt
	WeekdaySaturdayInt
	WeekdaySundayInt
)

const (
	// COLUMN_ORDER     = iota // .
	ColumnAmountIndex = iota
	ColumnActiveIndex
	ColumnNameIndex
	ColumnFrequencyIndex
	ColumnIntervalIndex
	ColumnMondayIndex
	ColumnTuesdayIndex
	ColumnWednesdayIndex
	ColumnThursdayIndex
	ColumnFridayIndex
	ColumnSaturdayIndex
	ColumnSundayIndex
	ColumnStartsIndex
	ColumnEndsIndex
	ColumnNoteIndex
	ColumnIDIndex
	ColumnCreatedAtIndex
	ColumnUpdatedAtIndex
)

const (
	ColorColumnOrder     = "[gray]"
	ColorColumnAmount    = "[gold]"
	ColorColumnActive    = "[white]"
	ColorColumnName      = "[#8899dd]"
	ColorColumnFrequency = "[#70dd70]"
	ColorColumnInterval  = "[#de9a9a]"
	ColorColumnMonday    = "[red]"
	ColorColumnTuesday   = "[orange]"
	ColorColumnWednesday = "[yellow]"
	ColorColumnThursday  = "[green]"
	ColorColumnFriday    = "[blue]"
	ColorColumnSaturday  = "[indigo]"
	ColorColumnSunday    = "[violet]"
	ColorColumnStarts    = "[#aaffaa]"
	ColorColumnEnds      = "[#aaffee]"
	ColorColumnNote      = "[white]"
	ColorColumnID        = "[gray]"
	ColorColumnCreatedAt = "[blue]"
	ColorColumnUpdatedAT = "[blue]"

	ColorInactive = "[gray::i]"

	ColorColumnAmountPositive = "[lightgreen]"
)

// results page values

const (
	ColumnDate                = "Date"
	ColumnBalance             = "Balance"
	ColumnCumulativeIncome    = "CumulativeIncome"
	ColumnCumulativeExpenses  = "CumulativeExpenses"
	ColumnDayExpenses         = "DayExpenses"
	ColumnDayIncome           = "DayIncome"
	ColumnDayNet              = "DayNet"
	ColumnDiffFromStart       = "DiffFromStart"
	ColumnDayTransactionNames = "DayTransactionNames"

	ColorColumnResultsDate                = "[#8899dd]"
	ColorColumnResultsBalance             = "[white::b]"
	ColorColumnResultsCumulativeIncome    = "[lightgreen]"
	ColorColumnResultsCumulativeExpenses  = "[gold]"
	ColorColumnResultsDayExpenses         = "[orange]"
	ColorColumnResultsDayIncome           = "[lightgreen]"
	ColorColumnResultsDayNet              = "[#cccccc]"
	ColorColumnResultsDiffFromStart       = "[lightgoldenrodyellow]"
	ColorColumnResultsDayTransactionNames = "[smoke]"
)

// Actions that can be mapped to keybindings.
const (
	ActionRedo       = "redo"
	ActionUndo       = "undo"
	ActionQuit       = "quit"
	ActionSelect     = "select"
	ActionMulti      = "multi"
	ActionMove       = "move"
	ActionDelete     = "delete"
	ActionDuplicate  = "duplicate"
	ActionAdd        = "add"
	ActionEdit       = "edit"
	ActionSave       = "save"
	ActionEnd        = "end"
	ActionHome       = "home"
	ActionLeft       = "left"
	ActionRight      = "right"
	ActionDown       = "down"
	ActionUp         = "up"
	ActionPageDown   = "pagedown"
	ActionPageUp     = "pageup"
	ActionBackTab    = "backtab"
	ActionTab        = "tab"
	ActionEsc        = "escape"
	ActionResults    = "results"
	ActionProfiles   = "profiles"
	ActionGlobalHelp = "globalhelp" // e.g. F1 key instead of ?
	ActionHelp       = "help"       // e.g. ? key that can also be used in input fields
	ActionSearch     = "search"
)

var AllActions = []string{
	ActionRedo,
	ActionUndo,
	ActionQuit,
	ActionSelect,
	ActionMulti,
	ActionMove,
	ActionDelete,
	ActionDuplicate,
	ActionAdd,
	ActionEdit,
	ActionSave,
	ActionEnd,
	ActionHome,
	ActionLeft,
	ActionRight,
	ActionDown,
	ActionUp,
	ActionPageDown,
	ActionPageUp,
	ActionBackTab,
	ActionTab,
	ActionEsc,
	ActionResults,
	ActionProfiles,
	ActionGlobalHelp,
	ActionHelp,
	ActionSearch,
}

var DefaultMappings = map[string]string{
	DefaultBindingUndo:       ActionUndo,
	DefaultBindingRedo:       ActionRedo,
	DefaultBindingQuit:       ActionQuit,
	DefaultBindingSelect:     ActionSelect,
	DefaultBindingMulti:      ActionMulti,
	DefaultBindingMove:       ActionMove,
	DefaultBindingDelete:     ActionDelete,
	DefaultBindingDuplicate:  ActionDuplicate,
	DefaultBindingAdd1:       ActionAdd,
	DefaultBindingAdd2:       ActionAdd,
	DefaultBindingAdd3:       ActionAdd,
	DefaultBindingEdit1:      ActionEdit,
	DefaultBindingEdit2:      ActionEdit,
	DefaultBindingSave:       ActionSave,
	DefaultBindingEnd:        ActionEnd,
	DefaultBindingHome:       ActionHome,
	DefaultBindingDown:       ActionDown,
	DefaultBindingUp:         ActionUp,
	DefaultBindingLeft:       ActionLeft,
	DefaultBindingRight:      ActionRight,
	DefaultBindingPageDown:   ActionPageDown,
	DefaultBindingPageUp:     ActionPageUp,
	DefaultBindingBackTab:    ActionBackTab,
	DefaultBindingTab:        ActionTab,
	DefaultBindingEsc:        ActionEsc,
	DefaultBindingResults:    ActionResults,
	DefaultBindingProfiles:   ActionProfiles,
	DefaultBindingGlobalHelp: ActionGlobalHelp,
	DefaultBindingHelp:       ActionHelp,
	DefaultBindingSearch:     ActionSearch,
}

// For now, please keep all explanations under 80 chars.
const (
	ActionExplanationRedo       = "moves forward in the undo buffer"
	ActionExplanationUndo       = "moves backward in the undo buffer"
	ActionExplanationQuit       = "quit the application after a confirmation prompt"
	ActionExplanationSelect     = "toggle selecting of a single row in the transactions table"
	ActionExplanationMulti      = "select a range of items in the transactions table"
	ActionExplanationMove       = "moves all selected transactions to the highlighted row"
	ActionExplanationDelete     = "deletes all selected transactions or current profile"
	ActionExplanationDuplicate  = "duplicates all selected transactions"
	ActionExplanationAdd        = "adds a new transaction to the transactions table"
	ActionExplanationEdit       = "rename the current profile when profile list is focused"
	ActionExplanationSave       = "saves the current file"
	ActionExplanationEnd        = "context-specific movement to the end of the row/column/line/bounds"
	ActionExplanationHome       = "context-specific movement to the start of the row/column/line/bounds"
	ActionExplanationLeft       = "moves the cursor/focus left, varies depending on context"
	ActionExplanationRight      = "moves the cursor/focus right, varies depending on context"
	ActionExplanationDown       = "moves the cursor/focus down, varies depending on context"
	ActionExplanationUp         = "moves the cursor/focus up, varies depending on context"
	ActionExplanationPageDown   = "moves the cursor/focus a page down, varies depending on context"
	ActionExplanationPageUp     = "moves the cursor/focus a page up, varies depending on context"
	ActionExplanationBackTab    = "(shift+tab default) moves focus between elements, varies based on context"
	ActionExplanationTab        = "moves focus between elements, varies based on context"
	ActionExplanationEsc        = "escape the current context, press enough times and app will prompt to exit"
	ActionExplanationResults    = "takes you to the results page; press again to get some stats and refresh"
	ActionExplanationProfiles   = "immediately takes you to the profiles page"
	ActionExplanationGlobalHelp = "immediately takes you to the help page"
	ActionExplanationHelp       = "context-specific help, if available; otherwise, help page"
	ActionExplanationSearch     = "(not implemented yet!) search (via fuzzy find) in the current table"
)

var ActionExplanations = map[string]string{
	ActionRedo:       ActionExplanationRedo,
	ActionUndo:       ActionExplanationUndo,
	ActionQuit:       ActionExplanationQuit,
	ActionSelect:     ActionExplanationSelect,
	ActionMulti:      ActionExplanationMulti,
	ActionMove:       ActionExplanationMove,
	ActionDelete:     ActionExplanationDelete,
	ActionDuplicate:  ActionExplanationDuplicate,
	ActionAdd:        ActionExplanationAdd,
	ActionEdit:       ActionExplanationEdit,
	ActionSave:       ActionExplanationSave,
	ActionEnd:        ActionExplanationEnd,
	ActionHome:       ActionExplanationHome,
	ActionLeft:       ActionExplanationLeft,
	ActionRight:      ActionExplanationRight,
	ActionDown:       ActionExplanationDown,
	ActionUp:         ActionExplanationUp,
	ActionPageDown:   ActionExplanationPageDown,
	ActionPageUp:     ActionExplanationPageUp,
	ActionBackTab:    ActionExplanationBackTab,
	ActionTab:        ActionExplanationTab,
	ActionEsc:        ActionExplanationEsc,
	ActionResults:    ActionExplanationResults,
	ActionProfiles:   ActionExplanationProfiles,
	ActionGlobalHelp: ActionExplanationGlobalHelp,
	ActionHelp:       ActionExplanationHelp,
	ActionSearch:     ActionExplanationSearch,
}

const (
	DefaultBindingRedo       = "Ctrl+Y"
	DefaultBindingUndo       = "Ctrl+Z"
	DefaultBindingQuit       = "Ctrl+C"
	DefaultBindingSelect     = "Rune[ ]"
	DefaultBindingMulti      = "Ctrl+Space"
	DefaultBindingMove       = "Rune[m]"
	DefaultBindingDelete     = "Delete"
	DefaultBindingDuplicate  = "Ctrl+D"
	DefaultBindingAdd1       = "Rune[a]"
	DefaultBindingAdd2       = "Ctrl+N"
	DefaultBindingAdd3       = "Rune[n]"
	DefaultBindingEdit1      = "Rune[e]"
	DefaultBindingEdit2      = "Rune[r]"
	DefaultBindingSave       = "Ctrl+S"
	DefaultBindingEnd        = "End"
	DefaultBindingHome       = "Home"
	DefaultBindingLeft       = "Left"
	DefaultBindingRight      = "Right"
	DefaultBindingDown       = "Down"
	DefaultBindingUp         = "Up"
	DefaultBindingPageDown   = "PgDn"
	DefaultBindingPageUp     = "PgUp"
	DefaultBindingBackTab    = "Backtab"
	DefaultBindingTab        = "Tab"
	DefaultBindingEsc        = "Esc"
	DefaultBindingResults    = "F3"
	DefaultBindingProfiles   = "F2"
	DefaultBindingGlobalHelp = "F1"
	DefaultBindingHelp       = "Rune[?]"
	DefaultBindingSearch     = "Rune[/]"
)

const (
	ColumnDateIndex = iota
	ColumnBalanceIndex
	ColumnCumulativeIncomeIndex
	ColumnCumulativeExpensesIndex
	ColumnDayExpensesIndex
	ColumnDayIncomeIndex
	ColumnDayNetIndex
	ColumnDiffFromStartIndex
	ColumnDayTransactionNamesIndex
)

var ResultsColumns = []string{
	ColumnDate,
	ColumnBalance,
	ColumnCumulativeIncome,
	ColumnCumulativeExpenses,
	ColumnDayExpenses,
	ColumnDayIncome,
	ColumnDayNet,
	ColumnDiffFromStart,
	ColumnDayTransactionNames,
}

// ResultsColumnsIndexes should be the same length as the "columns" variable.
var ResultsColumnsIndexes = []int{
	ColumnDateIndex,
	ColumnBalanceIndex,
	ColumnCumulativeIncomeIndex,
	ColumnCumulativeExpensesIndex,
	ColumnDayExpensesIndex,
	ColumnDayIncomeIndex,
	ColumnDayNetIndex,
	ColumnDiffFromStartIndex,
	ColumnDayTransactionNamesIndex,
}

// Magic numbers that are used in multiple places.
const (
	DaysInMonth               = 31
	HoursInDay                = 24
	DefaultTransactionBalance = 500
)
