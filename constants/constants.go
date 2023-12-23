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

	DEFAULT_CONFIG = "config.yml"

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
	COLOR_COLUMN_ACTIVE    = "[white]"
	COLOR_COLUMN_NAME      = "[#8899dd]"
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

// for now, please keep all explanations under 80 chars
const (
	ACTION_EXPLANATION_REDO        = "moves forward in the undo buffer"
	ACTION_EXPLANATION_UNDO        = "moves backward in the undo buffer"
	ACTION_EXPLANATION_QUIT        = "quit the application after a confirmation prompt"
	ACTION_EXPLANATION_SELECT      = "toggle selecting of a single row in the transactions table"
	ACTION_EXPLANATION_MULTI       = "select a range of items in the transactions table"
	ACTION_EXPLANATION_MOVE        = "moves all selected transactions to the highlighted row"
	ACTION_EXPLANATION_DELETE      = "deletes all selected transactions or current profile"
	ACTION_EXPLANATION_DUPLICATE   = "duplicates all selected transactions"
	ACTION_EXPLANATION_ADD         = "adds a new transaction to the transactions table"
	ACTION_EXPLANATION_EDIT        = "rename the current profile when profile list is focused"
	ACTION_EXPLANATION_SAVE        = "saves the current file"
	ACTION_EXPLANATION_END         = "context-specific movement to the end of the row/column/line/bounds"
	ACTION_EXPLANATION_HOME        = "context-specific movement to the start of the row/column/line/bounds"
	ACTION_EXPLANATION_LEFT        = "moves the cursor/focus left, varies depending on context"
	ACTION_EXPLANATION_RIGHT       = "moves the cursor/focus right, varies depending on context"
	ACTION_EXPLANATION_DOWN        = "moves the cursor/focus down, varies depending on context"
	ACTION_EXPLANATION_UP          = "moves the cursor/focus up, varies depending on context"
	ACTION_EXPLANATION_PAGEDOWN    = "moves the cursor/focus a page down, varies depending on context"
	ACTION_EXPLANATION_PAGEUP      = "moves the cursor/focus a page up, varies depending on context"
	ACTION_EXPLANATION_BACKTAB     = "(shift+tab default) moves focus between elements, varies based on context"
	ACTION_EXPLANATION_TAB         = "moves focus between elements, varies based on context"
	ACTION_EXPLANATION_ESCAPE      = "escape the current context, press enough times and app will prompt to exit"
	ACTION_EXPLANATION_RESULTS     = "takes you to the results page; press again to get some stats and refresh"
	ACTION_EXPLANATION_PROFILES    = "immediately takes you to the profiles page"
	ACTION_EXPLANATION_GLOBAL_HELP = "immediately takes you to the help page"
	ACTION_EXPLANATION_HELP        = "context-specific help, if available; otherwise, help page"
	ACTION_EXPLANATION_SEARCH      = "(not implemented yet!) search (via fuzzy find) in the current table"
)

var ACTION_EXPLANATIONS = map[string]string{
	ACTION_REDO:        ACTION_EXPLANATION_REDO,
	ACTION_UNDO:        ACTION_EXPLANATION_UNDO,
	ACTION_QUIT:        ACTION_EXPLANATION_QUIT,
	ACTION_SELECT:      ACTION_EXPLANATION_SELECT,
	ACTION_MULTI:       ACTION_EXPLANATION_MULTI,
	ACTION_MOVE:        ACTION_EXPLANATION_MOVE,
	ACTION_DELETE:      ACTION_EXPLANATION_DELETE,
	ACTION_DUPLICATE:   ACTION_EXPLANATION_DUPLICATE,
	ACTION_ADD:         ACTION_EXPLANATION_ADD,
	ACTION_EDIT:        ACTION_EXPLANATION_EDIT,
	ACTION_SAVE:        ACTION_EXPLANATION_SAVE,
	ACTION_END:         ACTION_EXPLANATION_END,
	ACTION_HOME:        ACTION_EXPLANATION_HOME,
	ACTION_LEFT:        ACTION_EXPLANATION_LEFT,
	ACTION_RIGHT:       ACTION_EXPLANATION_RIGHT,
	ACTION_DOWN:        ACTION_EXPLANATION_DOWN,
	ACTION_UP:          ACTION_EXPLANATION_UP,
	ACTION_PAGEDOWN:    ACTION_EXPLANATION_PAGEDOWN,
	ACTION_PAGEUP:      ACTION_EXPLANATION_PAGEUP,
	ACTION_BACKTAB:     ACTION_EXPLANATION_BACKTAB,
	ACTION_TAB:         ACTION_EXPLANATION_TAB,
	ACTION_ESCAPE:      ACTION_EXPLANATION_ESCAPE,
	ACTION_RESULTS:     ACTION_EXPLANATION_RESULTS,
	ACTION_PROFILES:    ACTION_EXPLANATION_PROFILES,
	ACTION_GLOBAL_HELP: ACTION_EXPLANATION_GLOBAL_HELP,
	ACTION_HELP:        ACTION_EXPLANATION_HELP,
	ACTION_SEARCH:      ACTION_EXPLANATION_SEARCH,
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

const HelpTextTemplate = `[lightgreen::b]Finance Planner[-:-:-:-]

[gold]
                 _____ _
                |  ___(_)_ __   __ _ _ __   ___ ___
                | |_  | | '_ \ / _  | '_ \ / __/ _ \
                |  _| | | | | | (_| | | | | (_|  __/
                |[lightgreen]_[gold]|[lightgreen]__[gold] |[lightgreen]_[gold]|_| |_|\__,_|_| |_|\___\___|[lightgreen]
                |  _ \| | __ _ _ __  _ __   ___ _ __
                | |_) | |/ _  | '_ \| '_ \ / _ \ '__|
                |  __/| | (_| | | | | | | |  __/ |
                |_|   |_|\__,_|_| |_|_| |_|\___|_|
[-:-:-:-]


[lightgreen::b]General information[-:-:-:-]

The purpose of this application is to allow you to define recurring bills
and income ([gold]transactions[-]), and then get a fairly accurate prediction
of where your money will be using the [#8899dd]Results[white] page.

[lightgreen::b]Profiles[-:-:-:-]

Profiles are shown on the left-hand side of the [#8899dd]Profiles & Transactions[white] page.

- You may need to use the <tab> key to get to them.
- You can duplicate and rename profiles.
- [gold]Each profile must have a unique name.[-] Duplicate names will be refused.

You can create multiple [#8899dd]profiles[-] to fulfill any purpose, such as:

- modeling a change in your financial plans (removing subscriptions,
  hypotheticals, etc)
- adding multiple family members

[lightgreen::b]Transactions[-:-:-:-]

A [#8899dd]transaction[-] is a recurring expense or income:

- If the transaction earns money, it is prefixed with a [lightgreen]+[-] (plus) sign.
- All transactions are assumed to be negative by default.

Each transaction has the following fields:

- [::b]Order[-]:     You can define a custom integer sorting order for transactions.
             This field has no other purpose.
- [::b]Amount[-]:    This is a positive or negative value as described above.
- [::b]Active[-]:    This is a boolean value that determines whether the transaction should
             be included in calculations. This is useful for temporarily making
             changes without destroying anything.
- [::b]Name[-]:      This is the human-readable name of the transaction for your eyes.
- [::b]Frequency[-]: Transactions can occur [#8899dd]MONTHLY[-], [lightgreen]WEEKLY[-], or [gold]YEARLY.
             [-]This value must be exactly one of those three strings, but an auto-
             complete is provided to make it quicker.
- [::b]Interval[-]:  The transaction occurs every [#8899dd]<interval>[white] WEEKS/MONTHS/YEARS.
- [::b]<[-]Weekday>: The transaction only occurs on the checked days of the week, and
             will not occur if the defined recurrence pattern does not land on
             one of these days.
- [::b]Starts[-]:    This is the starting date for the transaction's recurrence pattern.
             It is defined as [#8899dd]YYYY[white]-[lightgreen]MM[white]-[gold]DD[white].

             For simplicity when working with dates at the end of the month,
             you may want to consider putting setting the day value to 28, as
             some recurrence patterns may skip a 31.

             Months range from [#8899dd]1-12[white], and days range from [#8899dd]1-31[white].
             Years must be any positive value, and can be 0.
- [::b]Ends[-]:      This is the last acceptable date for recurrence. Behavior is the
             exact same as the Starts field.
- [::b]Note[-]:      A human-readable field for you to put arbitrary notes in.

[lightgreen::b]Results[-:-:-:-]

The results page allows you to see a projection of your finances into the
future. It shows the following:

- A form on the left containing start & end dates, and the starting balance
  for the projection to start with
- A table containing one day per row, with each of the transactions that
  occurred on that day, as well as other numbers such as the total expenses,
  running balance since the first day of the projection, etc.

The same hotkey that opens the results page can be pressed multiple times to
re-submit the results form and will also show some useful statistics about
your finances.

[lightgreen::b]Keyboard Shortcuts: Current & Default[-:-:-:-]

Custom keybindings are shown in [gold::b]gold[-:-:-:-]:
{{ range $k, $v := .CombinedKeybindings }}
- [::b]{{ $k -}}[-:-:-:-]: {{ range $v -}}{{- . }} {{ end -}}
{{ end }}

[lightgreen::b]Keyboard Shortcuts: All Actions Explained[-:-:-:-]
{{ range $k, $v := .Explanations }}
- [::b]{{ $k -}}[-:-:-:-]: {{ $v }}
{{- end }}

[lightgreen::b]Keyboard Shortcuts: All Actions' Mappings[-:-:-:-]

For the sake of debugging any custom configuration changes you've made, actions
and all of the ways they can be executed are shown below. Custom bindings are
shown in [gold]gold[-], and bindings that are used as part of a chain of actions
are shown in [#aaffee]a light blue color[-].
{{ range $k, $v := .CombinedActions }}
- [::b]{{ $k -}}[-:-:-:-]: {{ range $v -}}{{- . }} {{ end -}}
{{ end }}

[lightgreen::b]Keyboard Shortcuts: How to configure[-:-:-:-]

Keyboard shortcuts can be bound to [::b]actions[-:-:-:-]. In your config.yml file,
they can be specified with a top-level "keybindings" object, as shown in this
example:

  ---
  keybindings:
    "Ctrl+X":
      - quit
    "Ctrl+V":
      - pagedown
    "Alt+V":
      - pageup
    "Rune[r]":
      - results

Not all keyboard combinations will be actionable due to limitations of
terminal emulators. To figure out which keybindings are acceptable, run this
program with the [::b]"-kb"[-:-:-:-] flag to enter keyboard echo mode (more info will be
given on startup). [::b]To unset a key, set its action value to 'none'[-:-:-:-].

Astute readers will also note that more than one action can be provided per
keybinding. For example, you may want to trigger multiple add actions at the
same time by just pressing one set of keys. [yellow]You should be warned[-], however, that
the number of permutations granted by chaining actions is massive, and can lead
to completely untested scenarios, so expect bugs when using more than 1 action
per key binding.
`

// <tab>/<shift+tab>: cycle back and forth between panels/controls where
// appropriate

// <esc>: deselects the last selected mark, and then deselects the last
// selected items, then un-focuses panes until eventually exiting the application
// entirely

// <ctrl+s>: saves to config.yml in the current directory

// <ctrl+i>: shows statistics in the Results page's bottom text pane

// [lightgreen::b]Transactions page:[-:-:-:-]

// <space>: select the current transaction
// <ctrl+space>: toggle multi-select from the last previously selected item
// <>
