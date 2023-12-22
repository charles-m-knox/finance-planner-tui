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

const HelpText = `[lightgreen::b]Finance Planner[-:-:-:-]

[gold]
                 _____ _
                |  ___(_)_ __   __ _ _ __   ___ ___
                | |_  | | '_ \ / _  | '_ \ / __/ _ \
                |  _| | | | | | (_| | | | | (_|  __/
                |_|__ |_|_| |_|\__,_|_| |_|\___\___|[lightgreen]
                |  _ \| | __ _ _ __  _ __   ___ _ __
                | |_) | |/ _  | '_ \| '_ \ / _ \ '__|
                |  __/| | (_| | | | | | | |  __/ |
                |_|   |_|\__,_|_| |_|_| |_|\___|_|
[-:-:-:-]


[lightgreen::b]General information[-:-:-:-]

[white]The purpose of this application is to allow you to define recurring bills
and income ([gold]transactions[white]), and then get a fairly accurate prediction
of where your money will be using the [blue]Results[white] page.

[lightgreen::b]More on Profiles[-:-:-:-]

[white]Profiles are shown on the left-hand side of the [blue]Profiles & Transactions[white] page.

- You may need to use the <tab> key to get to them.
- You can duplicate and rename profiles.
- [gold]Each profile must have a unique name.[white] Duplicate names will be refused.

You can create multiple [blue]profiles[white] to fulfill any purpose, such as:

- modeling a change in your financial plans (removing subscriptions,
  hypotheticals, etc)
- adding multiple family members

[lightgreen::b]More on Transactions[-:-:-:-]

[white]A [blue]transaction[white] is a recurring expense or income:

- If the transaction earns money, it is prefixed with a [lightgreen]+[white] (plus) sign.
- All transactions are assumed to be negative by default.

Each transaction has the following fields:

- Order:     You can define a custom integer sorting order for transactions.
             This field has no other purpose.
- Amount:    This is a positive or negative value as described above.
- Active:    This is a boolean value that determines whether the transaction should
             be included in calculations. This is useful for temporarily making
             changes without destroying anything.
- Name:      This is the human-readable name of the transaction for your eyes.
- Frequency: Transactions can occur [aqua]MONTHLY, [lightgreen]WEEKLY, or [gold]YEARLY.
             [white]This value must be exactly one of those three strings, but an auto-
             complete is provided to make it quicker.
- Interval:  The transaction occurs every [aqua]<interval>[white] WEEKS/MONTHS/YEARS.
- <Weekday>: The transaction only occurs on the checked days of the week, and
             will not occur if the defined recurrence pattern does not land on
             one of these days.
- Starts:    This is the starting date for the transaction's recurrence pattern.
             It is defined as [aqua]YYYY[white]-[lightgreen]MM[white]-[gold]DD[white].

             For simplicity when working with dates at the end of the month,
             you may want to consider putting setting the day value to 28, as
             some recurrence patterns may skip a 31.

             Months range from [aqua]1-12[white], and days range from [aqua]1-31[white].
             Years must be any positive value, and can be 0.
- Ends:      This is the last acceptable date for recurrence. Behavior is the
             exact same as the Starts field.
- Note:      A human-readable field for you to put arbitrary notes in.

[lightgreen::b]General Keyboard Shortcuts:[-:-:-:-]

<tab>/<shift+tab>: cycle back and forth between panels/controls where
appropriate

<esc>: deselects the last selected mark, and then deselects the last
selected items, then un-focuses panes until eventually exiting the application
entirely

<ctrl+s>: saves to config.yml in the current directory

<ctrl+i>: shows statistics in the Results page's bottom text pane

[lightgreen::b]Transactions page:[-:-:-:-]

<space>: select the current transaction
<ctrl+space>: toggle multi-select from the last previously selected item
<>
`
