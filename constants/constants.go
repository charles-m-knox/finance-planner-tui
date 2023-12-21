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
)

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
	COLOR_COLUMN_AMOUNT    = "[yellow]"
	COLOR_COLUMN_ACTIVE    = "[gray]"
	COLOR_COLUMN_NAME      = "[blue]"
	COLOR_COLUMN_FREQUENCY = "[aqua]"
	COLOR_COLUMN_INTERVAL  = "[white]"
	COLOR_COLUMN_MONDAY    = "[red]"
	COLOR_COLUMN_TUESDAY   = "[orange]"
	COLOR_COLUMN_WEDNESDAY = "[yellow]"
	COLOR_COLUMN_THURSDAY  = "[green]"
	COLOR_COLUMN_FRIDAY    = "[blue]"
	COLOR_COLUMN_SATURDAY  = "[indigo]"
	COLOR_COLUMN_SUNDAY    = "[violet]"
	COLOR_COLUMN_STARTS    = "[aqua]"
	COLOR_COLUMN_ENDS      = "[blue]"
	COLOR_COLUMN_NOTE      = "[white]"
	COLOR_COLUMN_ID        = "[gray]"
	COLOR_COLUMN_CREATEDAT = "[blue]"
	COLOR_COLUMN_UPDATEDAT = "[blue]"

	COLOR_COLUMN_AMOUNT_POSITIVE = "[green]"
)
