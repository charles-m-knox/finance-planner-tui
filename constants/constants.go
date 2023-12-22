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

	START_TODAY_PRESET = "today"
	ONE_YR             = "1year"
	FIVE_YR            = "5year"

	CONFIG_VERSION = "1"
)

const RESET_STYLE = "[-:-:-:-]"

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
	COLUMN_ORDER     = iota // int
	COLUMN_AMOUNT           // int in cents; 500 = $5.00
	COLUMN_ACTIVE           // bool true/false
	COLUMN_NAME             // editable string
	COLUMN_FREQUENCY        // dropdown, monthly/daily/weekly/yearly
	COLUMN_INTERVAL         // integer, occurs every x frequency
	COLUMN_MONDAY           // bool
	COLUMN_TUESDAY          // bool
	COLUMN_WEDNESDAY        // bool
	COLUMN_THURSDAY         // bool
	COLUMN_FRIDAY           // bool
	COLUMN_SATURDAY         // bool
	COLUMN_SUNDAY           // bool
	COLUMN_STARTS           // string
	COLUMN_ENDS             // string
	COLUMN_NOTE             // editable string
	COLUMN_ID               // non-editable strings
	COLUMN_CREATEDAT        // non-editable strings
	COLUMN_UPDATEDAT        // non-editable strings
)

const (
	COLOR_COLUMN_ORDER     = "[gray]"
	COLOR_COLUMN_AMOUNT    = "[gold]"
	COLOR_COLUMN_ACTIVE    = "[#8899dd]"
	COLOR_COLUMN_NAME      = "[blue]"
	COLOR_COLUMN_FREQUENCY = "[#70dd70]"
	COLOR_COLUMN_INTERVAL  = "[#de9a9a]"
	COLOR_COLUMN_MONDAY    = "[red]"
	COLOR_COLUMN_TUESDAY   = "[orange]"
	COLOR_COLUMN_WEDNESDAY = "[yellow]"
	COLOR_COLUMN_THURSDAY  = "[green]"
	COLOR_COLUMN_FRIDAY    = "[blue]"
	COLOR_COLUMN_SATURDAY  = "[indigo]"
	COLOR_COLUMN_SUNDAY    = "[violet]"
	COLOR_COLUMN_STARTS    = "[#aaffaa]"
	COLOR_COLUMN_ENDS      = "[#aaffee]"
	COLOR_COLUMN_NOTE      = "[white]"
	COLOR_COLUMN_ID        = "[gray]"
	COLOR_COLUMN_CREATEDAT = "[blue]"
	COLOR_COLUMN_UPDATEDAT = "[blue]"

	COLOR_INACTIVE = "[gray::i]"

	COLOR_COLUMN_AMOUNT_POSITIVE = "[lightgreen]"
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

	COLOR_COLUMN_RESULTS_DATE                = "[#8899dd]"
	COLOR_COLUMN_RESULTS_BALANCE             = "[white::b]"
	COLOR_COLUMN_RESULTS_CUMULATIVEINCOME    = "[lightgreen]"
	COLOR_COLUMN_RESULTS_CUMULATIVEEXPENSES  = "[gold]"
	COLOR_COLUMN_RESULTS_DAYEXPENSES         = "[orange]"
	COLOR_COLUMN_RESULTS_DAYINCOME           = "[lightgreen]"
	COLOR_COLUMN_RESULTS_DAYNET              = "[#cccccc]"
	COLOR_COLUMN_RESULTS_DIFFFROMSTART       = "[lightgoldenrodyellow]"
	COLOR_COLUMN_RESULTS_DAYTRANSACTIONNAMES = "[smoke]"
)

// actions that can be mapped to keybindings
const (
	ACTION_REDO        = "redo"
	ACTION_UNDO        = "undo"
	ACTION_QUIT        = "quit"
	ACTION_SELECT      = "select"
	ACTION_MULTI       = "multi"
	ACTION_MOVE        = "move"
	ACTION_DELETE      = "delete"
	ACTION_DUPLICATE   = "duplicate"
	ACTION_ADD         = "add"
	ACTION_EDIT        = "edit"
	ACTION_SAVE        = "save"
	ACTION_END         = "end"
	ACTION_HOME        = "home"
	ACTION_LEFT        = "left"
	ACTION_RIGHT       = "right"
	ACTION_DOWN        = "down"
	ACTION_UP          = "up"
	ACTION_PAGEDOWN    = "pagedown"
	ACTION_PAGEUP      = "pageup"
	ACTION_BACKTAB     = "backtab"
	ACTION_TAB         = "tab"
	ACTION_ESCAPE      = "escape"
	ACTION_RESULTS     = "results"
	ACTION_PROFILES    = "profiles"
	ACTION_GLOBAL_HELP = "globalhelp" // e.g. F1 key instead of ?
	ACTION_HELP        = "help"       // e.g. ? key that can also be used in input fields
	ACTION_SEARCH      = "search"
)

var ALL_ACTIONS = []string{
	ACTION_REDO,
	ACTION_UNDO,
	ACTION_QUIT,
	ACTION_SELECT,
	ACTION_MULTI,
	ACTION_MOVE,
	ACTION_DELETE,
	ACTION_DUPLICATE,
	ACTION_ADD,
	ACTION_EDIT,
	ACTION_SAVE,
	ACTION_END,
	ACTION_HOME,
	ACTION_LEFT,
	ACTION_RIGHT,
	ACTION_DOWN,
	ACTION_UP,
	ACTION_PAGEDOWN,
	ACTION_PAGEUP,
	ACTION_BACKTAB,
	ACTION_TAB,
	ACTION_ESCAPE,
	ACTION_RESULTS,
	ACTION_PROFILES,
	ACTION_GLOBAL_HELP,
	ACTION_HELP,
	ACTION_SEARCH,
}

var DEFAULT_MAPPINGS = map[string]string{
	DEFAULT_BINDING_UNDO:        ACTION_UNDO,
	DEFAULT_BINDING_REDO:        ACTION_REDO,
	DEFAULT_BINDING_UNDO:        ACTION_UNDO,
	DEFAULT_BINDING_QUIT:        ACTION_QUIT,
	DEFAULT_BINDING_SELECT:      ACTION_SELECT,
	DEFAULT_BINDING_MULTI:       ACTION_MULTI,
	DEFAULT_BINDING_MOVE:        ACTION_MOVE,
	DEFAULT_BINDING_DELETE:      ACTION_DELETE,
	DEFAULT_BINDING_DUPLICATE:   ACTION_DUPLICATE,
	DEFAULT_BINDING_ADD_1:       ACTION_ADD,
	DEFAULT_BINDING_ADD_2:       ACTION_ADD,
	DEFAULT_BINDING_ADD_3:       ACTION_ADD,
	DEFAULT_BINDING_EDIT_1:      ACTION_EDIT,
	DEFAULT_BINDING_EDIT_2:      ACTION_EDIT,
	DEFAULT_BINDING_SAVE:        ACTION_SAVE,
	DEFAULT_BINDING_END:         ACTION_END,
	DEFAULT_BINDING_HOME:        ACTION_HOME,
	DEFAULT_BINDING_DOWN:        ACTION_DOWN,
	DEFAULT_BINDING_UP:          ACTION_UP,
	DEFAULT_BINDING_LEFT:        ACTION_LEFT,
	DEFAULT_BINDING_RIGHT:       ACTION_RIGHT,
	DEFAULT_BINDING_PAGEDOWN:    ACTION_PAGEDOWN,
	DEFAULT_BINDING_PAGEUP:      ACTION_PAGEUP,
	DEFAULT_BINDING_BACKTAB:     ACTION_BACKTAB,
	DEFAULT_BINDING_TAB:         ACTION_TAB,
	DEFAULT_BINDING_ESCAPE:      ACTION_ESCAPE,
	DEFAULT_BINDING_RESULTS:     ACTION_RESULTS,
	DEFAULT_BINDING_PROFILES:    ACTION_PROFILES,
	DEFAULT_BINDING_GLOBAL_HELP: ACTION_GLOBAL_HELP,
	DEFAULT_BINDING_HELP:        ACTION_HELP,
	DEFAULT_BINDING_SEARCH:      ACTION_SEARCH,
}

var (
	DEFAULT_BINDING_REDO        = "Ctrl+Y"
	DEFAULT_BINDING_UNDO        = "Ctrl+Z"
	DEFAULT_BINDING_QUIT        = "Ctrl+C"
	DEFAULT_BINDING_SELECT      = "Rune[ ]"
	DEFAULT_BINDING_MULTI       = "Ctrl+Space"
	DEFAULT_BINDING_MOVE        = "Rune[m]"
	DEFAULT_BINDING_DELETE      = "Delete"
	DEFAULT_BINDING_DUPLICATE   = "Ctrl+D"
	DEFAULT_BINDING_ADD_1       = "Rune[a]"
	DEFAULT_BINDING_ADD_2       = "Ctrl+N"
	DEFAULT_BINDING_ADD_3       = "Rune[n]"
	DEFAULT_BINDING_EDIT_1      = "Rune[e]"
	DEFAULT_BINDING_EDIT_2      = "Rune[r]"
	DEFAULT_BINDING_SAVE        = "Ctrl+S"
	DEFAULT_BINDING_END         = "End"
	DEFAULT_BINDING_HOME        = "Home"
	DEFAULT_BINDING_LEFT        = "Left"
	DEFAULT_BINDING_RIGHT       = "Right"
	DEFAULT_BINDING_DOWN        = "Down"
	DEFAULT_BINDING_UP          = "Up"
	DEFAULT_BINDING_PAGEDOWN    = "PgDn"
	DEFAULT_BINDING_PAGEUP      = "PgUp"
	DEFAULT_BINDING_BACKTAB     = "Backtab"
	DEFAULT_BINDING_TAB         = "Tab"
	DEFAULT_BINDING_ESCAPE      = "Esc"
	DEFAULT_BINDING_RESULTS     = "F3"
	DEFAULT_BINDING_PROFILES    = "F2"
	DEFAULT_BINDING_GLOBAL_HELP = "F1"
	DEFAULT_BINDING_HELP        = "Rune[?]"
	DEFAULT_BINDING_SEARCH      = "Rune[/]"
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

// make ResultsColumnsIndexes the same length as the "columns" variable
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
