package main

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"
	"finance-planner-tui/models"

	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"github.com/teambition/rrule-go"
)

// When an input field loses focus, its auto-complete function needs to be
// set to an empty slice, otherwise it may continue showing the auto-complete
// dropdown after focus has moved elsewhere.
func resetTransactionsInputFieldAutocomplete() {
	FP.TransactionsInputField.SetAutocompleteFunc(
		func(currentText string) []string {
			return []string{}
		},
	)
}

// When the transactions input field loses focus, either by direct user action
// or some other event demanding focus elsewhere, this function should be
// executed.
func deactivateTransactionsInputField() {
	FP.TransactionsInputField.SetFieldBackgroundColor(
		tcell.ColorNames[FP.Colors["TransactionsInputFieldBlurredBackground"]],
	)

	FP.TransactionsInputField.SetLabel(fmt.Sprintf("%v%v%v",
		FP.Colors["TransactionsInputFieldPassive"],
		FP.T["TransactionsInputFieldPlaceholderLabel"],
		c.Reset,
	))

	FP.TransactionsInputField.SetText("")

	if FP.Previous == nil {
		return
	}

	FP.App.SetFocus(FP.Previous)
}

// Focuses the transactions input field, updates its label, and sets
// its background color to something noticeable.
func activateTransactionsInputField(msg, value string) {
	resetTransactionsInputFieldAutocomplete()
	activateTransactionsInputFieldNoAutocompleteReset(msg, value)
}

// Focuses the transactions input field, updates its label, and sets
// its background color to something noticeable - in some cases, the
// resetTransactionsInputFieldAutocomplete cannot be called without risking
// an infinite loop, so this function does not call it.
func activateTransactionsInputFieldNoAutocompleteReset(msg, value string) {
	FP.TransactionsInputField.SetFieldBackgroundColor(
		tcell.ColorNames[FP.Colors["TransactionsInputFieldFocusedBackground"]],
	)

	FP.TransactionsInputField.SetLabel(fmt.Sprintf("%v%v%v",
		FP.Colors["TransactionsInputFieldActive"],
		msg,
		c.Reset,
	))

	FP.TransactionsInputField.SetText(value)

	// don't mess with the previously stored focus if the text field is already
	// focused
	currentFocus := FP.App.GetFocus()
	if currentFocus == FP.TransactionsInputField {
		return
	}

	FP.Previous = currentFocus

	FP.App.SetFocus(FP.TransactionsInputField)
}

// Cycles through the available sortable configurations for the current set of
// transactions, then proceeds to update the transactions table.
func setTransactionsTableSort(column string) {
	FP.SortTX = lib.GetNextSort(FP.SortTX, column)

	getTransactionsTable()
}

func getWeekdaysMap() map[string]int {
	return map[string]int{
		FP.T["WeekdayMonday"]:    rrule.MO.Day(),
		FP.T["WeekdayTuesday"]:   rrule.TU.Day(),
		FP.T["WeekdayWednesday"]: rrule.WE.Day(),
		FP.T["WeekdayThursday"]:  rrule.TH.Day(),
		FP.T["WeekdayFriday"]:    rrule.FR.Day(),
		FP.T["WeekdaySaturday"]:  rrule.SA.Day(),
		FP.T["WeekdaySunday"]:    rrule.SU.Day(),
	}
}

type (
	TxSortFunc        func(ti, tj lib.TX) bool
	TxSortChooserFunc func(bool) TxSortFunc
)

// weekday sort functions

func sortWeekday(weekdays map[string]int, day string, asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		tiw := slices.Index(ti.Weekdays, weekdays[day]) != -1
		tjw := slices.Index(tj.Weekdays, weekdays[day]) != -1

		if asc {
			if tiw == tjw {
				return ti.ID > tj.ID
			}

			return tiw
		}

		if tiw == tjw {
			return ti.ID < tj.ID
		}

		return tjw
	}
}

// numeric sort functions

func sortAmount(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		if asc {
			if ti.Amount == tj.Amount {
				return ti.ID > tj.ID
			}

			return ti.Amount > tj.Amount
		}

		if ti.Amount == tj.Amount {
			return ti.ID < tj.ID
		}

		return ti.Amount < tj.Amount
	}
}

func sortFrequency(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		if asc {
			if ti.Frequency == tj.Frequency {
				return ti.ID > tj.ID
			}

			return ti.Frequency > tj.Frequency
		}

		if ti.Frequency == tj.Frequency {
			return ti.ID < tj.ID
		}

		return ti.Frequency < tj.Frequency
	}
}

func sortInterval(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		if asc {
			if ti.Interval == tj.Interval {
				return ti.ID > tj.ID
			}

			return ti.Interval > tj.Interval
		}

		if ti.Interval == tj.Interval {
			return ti.ID < tj.ID
		}

		return ti.Interval < tj.Interval
	}
}

// string sort functions

func sortNote(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		til := strings.ToLower(ti.Note)
		tjl := strings.ToLower(tj.Note)

		if asc {
			if til == tjl {
				return ti.ID > tj.ID
			}

			return til > tjl
		}

		if til == tjl {
			return ti.ID < tj.ID
		}

		return til < tjl
	}
}

func sortName(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		til := strings.ToLower(ti.Name)
		tjl := strings.ToLower(tj.Name)

		if asc {
			if til == tjl {
				return ti.ID > tj.ID
			}

			return til > tjl
		}

		if til == tjl {
			return ti.ID < tj.ID
		}

		return til < tjl
	}
}

func sortID(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		til := strings.ToLower(ti.ID)
		tjl := strings.ToLower(tj.ID)

		if asc {
			if til == tjl {
				return ti.ID > tj.ID
			}

			return til > tjl
		}

		if til == tjl {
			return ti.ID < tj.ID
		}

		return til < tjl
	}
}

// string-typed date sorting functions

func sortStarts(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		tis := ti.GetStartDateString()
		tjs := tj.GetStartDateString()

		if asc {
			if tis == tjs {
				return ti.ID > tj.ID
			}

			return tis > tjs
		}

		if tis == tjs {
			return ti.ID < tj.ID
		}

		return tis < tjs
	}
}

func sortEnds(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		tis := ti.GetEndsDateString()
		tjs := tj.GetEndsDateString()

		if asc {
			if tis == tjs {
				return ti.ID > tj.ID
			}

			return tis > tjs
		}

		if tis == tjs {
			return ti.ID < tj.ID
		}

		return tis < tjs
	}
}

// strongly typed date sorting functions

// TODO: validate that this works as expected
func sortCreatedAt(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		if asc {
			return ti.CreatedAt.After(tj.CreatedAt)
		}

		return ti.CreatedAt.Before(tj.CreatedAt)
	}
}

// TODO: validate that this works as expected
func sortUpdatedAt(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		if asc {
			return ti.UpdatedAt.After(tj.UpdatedAt)
		}

		return ti.UpdatedAt.Before(tj.UpdatedAt)
	}
}

// boolean sort functions

func sortActive(asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		if asc {
			if ti.Active == tj.Active {
				return ti.ID > tj.ID
			}

			return ti.Active
		}

		if ti.Active == tj.Active {
			return ti.ID < tj.ID
		}

		return tj.Active
	}
}

type TransactionsColumn struct {
	Name     string
	SortFunc TxSortChooserFunc
	// If true, the SortFunc will receive true/false
	// to determine if it should sort by ascending/descending.
	Ascending bool
}

// Returns an ordered list of the columns that will be shown in the transactions
// table, as well as their sort functions.
func getTransactionsColumns() map[string]TransactionsColumn {
	mo := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnMonday"], b) }
	tu := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnTuesday"], b) }
	we := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnWednesday"], b) }
	th := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnThursday"], b) }
	fr := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnFriday"], b) }
	sa := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnSaturday"], b) }
	su := func(b bool) TxSortFunc { return sortWeekday(FP.WeekdaysMap, FP.T["TransactionsColumnSunday"], b) }

	return map[string]TransactionsColumn{
		FP.T["TransactionsColumnAmount"]:    {SortFunc: sortAmount},
		FP.T["TransactionsColumnActive"]:    {SortFunc: sortActive},
		FP.T["TransactionsColumnName"]:      {SortFunc: sortName},
		FP.T["TransactionsColumnFrequency"]: {SortFunc: sortFrequency},
		FP.T["TransactionsColumnInterval"]:  {SortFunc: sortInterval},
		FP.T["TransactionsColumnMonday"]:    {SortFunc: mo},
		FP.T["TransactionsColumnTuesday"]:   {SortFunc: tu},
		FP.T["TransactionsColumnWednesday"]: {SortFunc: we},
		FP.T["TransactionsColumnThursday"]:  {SortFunc: th},
		FP.T["TransactionsColumnFriday"]:    {SortFunc: fr},
		FP.T["TransactionsColumnSaturday"]:  {SortFunc: sa},
		FP.T["TransactionsColumnSunday"]:    {SortFunc: su},
		FP.T["TransactionsColumnStarts"]:    {SortFunc: sortStarts},
		FP.T["TransactionsColumnEnds"]:      {SortFunc: sortEnds},
		FP.T["TransactionsColumnNote"]:      {SortFunc: sortNote},
	}
}

// Returns the possible sortable directions for all columns, which is simply
// Asc and Desc, but loaded from the translations table. If the value is true,
// it means that it is an ascending sort; false if descending.
func getSortableDirections() map[string]bool {
	return map[string]bool{
		FP.T["TransactionsColumnSortAsc"]:  true,
		FP.T["TransactionsColumnSortDesc"]: false,
	}
}

// For an input string such as "AmountAsc", this will return a predefined sort
// function that can be executed.
func getTransactionsSortMap() map[string]TxSortFunc {
	m := make(map[string]TxSortFunc)
	dirs := getSortableDirections()

	for col, def := range getTransactionsColumns() {
		for dir, asc := range dirs {
			m[fmt.Sprintf("%v%v", col, dir)] = def.SortFunc(asc)
		}
	}

	return m
}

// Sorts all transactions by the current sort column.
func sortTX(sortMap map[string]TxSortFunc) {
	if FP.SortTX == c.None || FP.SortTX == "" {
		return
	}

	FP.LastSelection = -1

	sort.SliceStable(
		FP.SelectedProfile.TX,
		func(i, j int) bool {
			ti := (FP.SelectedProfile.TX)[i]
			tj := (FP.SelectedProfile.TX)[j]

			return sortMap[FP.SortTX](ti, tj)
		},
	)
}

// Returns both the glyph (second return value) that should be shown for the
// current sort column as well as the currently sorted column name itself (first
// return value).
func getSort(currentSort string) (string, string) {
	s, g := "", ""
	if strings.HasSuffix(currentSort, FP.T["TransactionsColumnSortAsc"]) {
		s = strings.Split(currentSort, FP.T["TransactionsColumnSortAsc"])[0]
		g = "↑"
	} else if strings.HasSuffix(currentSort, FP.T["TransactionsColumnSortDesc"]) {
		s = strings.Split(currentSort, FP.T["TransactionsColumnSortDesc"])[0]
		g = "↓"
	}

	return s, g
}

// Returns a list, representing the ordered columns to be shown in
// the transactions table, alongside their configured colors.
func getTransactionsTableHeaders() []models.TableCell {
	return []models.TableCell{
		{Text: FP.T["TransactionsColumnAmount"], Color: FP.Colors["TransactionsColumnAmount"]},
		{Text: FP.T["TransactionsColumnActive"], Color: FP.Colors["TransactionsColumnActive"]},
		{Text: FP.T["TransactionsColumnName"], Color: FP.Colors["TransactionsColumnName"], Expand: 1},
		{Text: FP.T["TransactionsColumnFrequency"], Color: FP.Colors["TransactionsColumnFrequency"]},
		{Text: FP.T["TransactionsColumnInterval"], Color: FP.Colors["TransactionsColumnInterval"]},
		{Text: FP.T["TransactionsColumnMonday"], Color: FP.Colors["TransactionsColumnMonday"]},
		{Text: FP.T["TransactionsColumnTuesday"], Color: FP.Colors["TransactionsColumnTuesday"]},
		{Text: FP.T["TransactionsColumnWednesday"], Color: FP.Colors["TransactionsColumnWednesday"]},
		{Text: FP.T["TransactionsColumnThursday"], Color: FP.Colors["TransactionsColumnThursday"]},
		{Text: FP.T["TransactionsColumnFriday"], Color: FP.Colors["TransactionsColumnFriday"]},
		{Text: FP.T["TransactionsColumnSaturday"], Color: FP.Colors["TransactionsColumnSaturday"]},
		{Text: FP.T["TransactionsColumnSunday"], Color: FP.Colors["TransactionsColumnSunday"]},
		{Text: FP.T["TransactionsColumnStarts"], Color: FP.Colors["TransactionsColumnStarts"]},
		{Text: FP.T["TransactionsColumnEnds"], Color: FP.Colors["TransactionsColumnEnds"]},
		{Text: FP.T["TransactionsColumnNote"], Color: FP.Colors["TransactionsColumnNote"], Expand: 1},
	}
}

// Returns a list, representing the ordered columns to be shown in
// the transactions table, alongside their configured colors.
func getTransactionsTableCell(tx lib.TX) []models.TableCell {
	cAmount := FP.Colors["TransactionsColumnAmount"]
	cActive := FP.Colors["TransactionsColumnActive"]
	cName := FP.Colors["TransactionsColumnName"]
	cFrequency := FP.Colors["TransactionsColumnFrequency"]
	cInterval := FP.Colors["TransactionsColumnInterval"]
	cMonday := FP.Colors["TransactionsColumnMonday"]
	cTuesday := FP.Colors["TransactionsColumnTuesday"]
	cWednesday := FP.Colors["TransactionsColumnWednesday"]
	cThursday := FP.Colors["TransactionsColumnThursday"]
	cFriday := FP.Colors["TransactionsColumnFriday"]
	cSaturday := FP.Colors["TransactionsColumnSaturday"]
	cSunday := FP.Colors["TransactionsColumnSunday"]
	cStarts := FP.Colors["TransactionsColumnStarts"]
	cEnds := FP.Colors["TransactionsColumnEnds"]
	cNote := FP.Colors["TransactionsColumnNote"]

	active := FP.T["CheckedGlyph"]

	w := tx.GetWeekdaysCheckedMap(FP.T["CheckedGlyph"], FP.T["UncheckedGlyph"])

	if !tx.Active {
		active = ""
		cAmount = FP.T["TransactionsInactive"]
		cActive = FP.T["TransactionsInactive"]
		cName = FP.T["TransactionsInactive"]
		cFrequency = FP.T["TransactionsInactive"]
		cInterval = FP.T["TransactionsInactive"]
		cMonday = FP.T["TransactionsInactive"]
		cTuesday = FP.T["TransactionsInactive"]
		cWednesday = FP.T["TransactionsInactive"]
		cThursday = FP.T["TransactionsInactive"]
		cFriday = FP.T["TransactionsInactive"]
		cSaturday = FP.T["TransactionsInactive"]
		cSunday = FP.T["TransactionsInactive"]
		cStarts = FP.T["TransactionsInactive"]
		cEnds = FP.T["TransactionsInactive"]
		cNote = FP.T["TransactionsInactive"]
	} else { //nolint:gocritic // <-- intentionally structured like this
		if tx.Amount >= 0 {
			cAmount = FP.Colors["TransactionsAmountPositive"]
		}
	}

	cells := []models.TableCell{
		{Text: lib.FormatAsCurrency(tx.Amount), Color: cAmount, Align: tview.AlignCenter},
		{Text: active, Color: cActive, Align: tview.AlignCenter},
		{Text: tx.Name, Color: cName, Expand: 1, Align: tview.AlignLeft},
		{Text: tx.Frequency, Color: cFrequency, Align: tview.AlignCenter},
		{Text: strconv.Itoa(tx.Interval), Color: cInterval, Align: tview.AlignCenter},
		{Text: w[rrule.MO.Day()], Color: cMonday, Align: tview.AlignCenter},
		{Text: w[rrule.TU.Day()], Color: cTuesday, Align: tview.AlignCenter},
		{Text: w[rrule.WE.Day()], Color: cWednesday, Align: tview.AlignCenter},
		{Text: w[rrule.TH.Day()], Color: cThursday, Align: tview.AlignCenter},
		{Text: w[rrule.FR.Day()], Color: cFriday, Align: tview.AlignCenter},
		{Text: w[rrule.SA.Day()], Color: cSaturday, Align: tview.AlignCenter},
		{Text: w[rrule.SU.Day()], Color: cSunday, Align: tview.AlignCenter},
		{Text: tx.GetStartDateString(), Color: cStarts, Align: tview.AlignCenter},
		{Text: tx.GetEndsDateString(), Color: cEnds, Align: tview.AlignCenter},
		{Text: tx.Note, Color: cNote, Expand: 1, Align: tview.AlignLeft},
	}

	return cells
}

// Constructs and sets the columns for the first row in the transactions table.
// Unsafe to run repeatedly and does not clear any existing fields/data.
func setTransactionsTableHeaders(currentSort, sortGlyph string) {
	th := getTransactionsTableHeaders()

	for i := range th {
		g := ""
		if currentSort == th[i].Text {
			g = sortGlyph
		}

		cell := tview.NewTableCell(fmt.Sprintf("%v%v%v%v",
			th[i].Color,
			g,
			th[i].Text,
			c.Reset,
		))
		if th[i].Expand > 0 {
			cell.SetExpansion(th[i].Expand)
		}

		FP.TransactionsTable.SetCell(0, i, cell)
	}
}

// Constructs and sets the columns for the i'th row in the transactions table.
// Unsafe to run repeatedly and does not clear any existing fields/data.
func setTransactionsTableCellsForTransaction(i int, tx lib.TX, isLastSelection bool) {
	td := getTransactionsTableCell(tx)

	bg := tcell.ColorReset

	// needs to be this exact if-elseif chain because a switch statement won't
	// reasonably suffice
	//nolint:gocritic
	if isLastSelection && tx.Selected {
		bg = tcell.GetColor(FP.Colors["TransactionsRowSelectedAndLastSelectedColor"])
	} else if isLastSelection {
		bg = tcell.GetColor(FP.Colors["TransactionsRowLastSelectedColor"])
	} else if tx.Selected {
		bg = tcell.GetColor(FP.Colors["TransactionsRowSelectedColor"])
	}

	for j := range td {
		cell := tview.NewTableCell(fmt.Sprintf("%v%v%v",
			td[j].Color,
			td[j].Text,
			c.Reset,
		)).SetBackgroundColor(bg).SetAlign(td[j].Align)
		if td[j].Expand > 0 {
			cell.SetExpansion(td[j].Expand)
		}

		FP.TransactionsTable.SetCell(i, j, cell)
	}
}

// Creates the transactions table, based on the currently selected profile.
// Heads up: This DOES modify the existing profile's transaction (mainly applies
// sorting).
func getTransactionsTable() {
	FP.TransactionsTable.Clear()

	currentSort, sortGlyph := getSort(FP.SortTX)
	setTransactionsTableHeaders(currentSort, sortGlyph)

	FP.TransactionsTable.SetTitle(FP.T["TransactionsTableTitle"])
	FP.TransactionsTable.SetBorders(false).
		SetSelectable(true, true).
		SetSeparator(' ')

	if FP.SelectedProfile == nil {
		return
	}

	sortTX(FP.TransactionsSortMap)

	// start by populating the table with the columns first
	for i := range FP.SelectedProfile.TX {
		setTransactionsTableCellsForTransaction(i+1, FP.SelectedProfile.TX[i], FP.LastSelection == i)
	}

	FP.TransactionsTable.SetSelectedFunc(transactionsTableSelectionChanged)
}

// This is basically a callback function that is executed when the transactions
// table's selection is changed.
func transactionsTableSelectionChanged(row, column int) {
	// get the current profile & transaction
	i := 0

	// based on the row, find the actual transaction definition
	// example: row 5 = TX 4 because of table's headers
	for i = range FP.SelectedProfile.TX {
		txi := row - 1
		if i == txi {
			i = txi

			break
		}
	}

	switch column {
	// case c.COLUMN_ORDER:
	// 	if row == 0 {
	// 		setTransactionsTableSort(c.ColumnOrder)
	// 		return
	// 	}
	// 	FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
	// 		switch key {
	// 		case tcell.KeyEscape:
	// 			// don't save the changes
	// 			deactivateTransactionsInputField()
	// 			return
	// 		default:
	// 			d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
	// 			if err != nil || d < 1 {
	// 				activateTransactionsInputFieldNoAutocompleteReset("invalid order given:", fmt.Sprint(FP.SelectedProfile.TX[i].Order))
	// 				return
	// 			}

	// 			// update all selected values as well as the current one
	// 			for j := range FP.SelectedProfile.TX {
	// 				if FP.SelectedProfile.TX[j].Selected || j == i {
	// 					FP.SelectedProfile.TX[j].Order = int(d)
	// 					FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
	// 						"%v%v",
	// 						c.COLOR_COLUMN_ORDER,
	// 						FP.SelectedProfile.TX[j].Order,
	// 					))
	// 				}
	// 			}

	// 			modified()
	// 			deactivateTransactionsInputField()
	// 		}
	// 	})
	// 	activateTransactionsInputField("order:", fmt.Sprint(FP.SelectedProfile.TX[i].Order))
	case c.ColumnAmountIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnAmount)

			return
		}

		FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				// don't save the changes
				deactivateTransactionsInputField()

				return
			default:
				a := lib.ParseDollarAmount(FP.TransactionsInputField.GetText(), false)
				isPositiveAmount := a >= 0
				amountColor := c.ColorColumnAmount
				if isPositiveAmount {
					amountColor = c.ColorColumnAmountPositive
				}

				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].Amount = int(a)
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							amountColor,
							lib.FormatAsCurrency(FP.SelectedProfile.TX[j].Amount),
						))
					}
				}

				modified()
				deactivateTransactionsInputField()
			}
		})
		renderedAmount := lib.FormatAsCurrency(FP.SelectedProfile.TX[i].Amount)
		if FP.SelectedProfile.TX[i].Amount >= 0 {
			renderedAmount = fmt.Sprintf("+%v", renderedAmount)
		}
		activateTransactionsInputField("amount (start with + or $+ for positive):", renderedAmount)
	case c.ColumnActiveIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnActive)

			return
		}

		newValue := !FP.SelectedProfile.TX[i].Active
		FP.SelectedProfile.TX[i].Active = !FP.SelectedProfile.TX[i].Active

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				activeText := "✔"
				if !newValue {
					activeText = " "
				}
				FP.SelectedProfile.TX[j].Active = newValue

				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnActive, activeText))
			}
		}

		modified()
	case c.ColumnNameIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnName)

			return
		}
		activateTransactionsInputField("edit name:", FP.SelectedProfile.TX[i].Name)
		FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				break
			default:
				FP.SelectedProfile.TX[i].Name = FP.TransactionsInputField.GetText()

				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].Name = FP.SelectedProfile.TX[i].Name
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnName, FP.SelectedProfile.TX[i].Name))
					}
				}

				modified()
			}
			deactivateTransactionsInputField()
		})
	case c.ColumnFrequencyIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnFrequency)

			return
		}
		activateTransactionsInputField("weekly|monthly|yearly:", FP.SelectedProfile.TX[i].Frequency)
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
				FP.TransactionsInputField.SetLabel("invalid value - can only be weekly, monthly, or yearly:")

				return
			}
			FP.SelectedProfile.TX[i].Frequency = validatedFrequency

			// update all selected values as well as the current one
			for j := range FP.SelectedProfile.TX {
				if FP.SelectedProfile.TX[j].Selected || j == i {
					FP.SelectedProfile.TX[j].Frequency = FP.SelectedProfile.TX[i].Frequency
					FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnFrequency, FP.SelectedProfile.TX[i].Frequency))
				}
			}

			modified()
		}
		FP.TransactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
			return fuzzy.Find(strings.TrimSpace(strings.ToUpper(currentText)), []string{
				c.MONTHLY,
				c.YEARLY,
				c.WEEKLY,
			})
		})
		FP.TransactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
			saveFunc(text)
			deactivateTransactionsInputField()

			return true
		})
		FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				break
			default:
				saveFunc(FP.TransactionsInputField.GetText())
			}
			deactivateTransactionsInputField()
		})
	case c.ColumnIntervalIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnInterval)

			return
		}
		FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				// don't save the changes
				deactivateTransactionsInputField()

				return
			default:
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d < 0 {
					activateTransactionsInputFieldNoAutocompleteReset(
						"invalid interval given:",
						strconv.Itoa(FP.SelectedProfile.TX[i].Interval),
					)

					return
				}

				FP.SelectedProfile.TX[i].Interval = int(d)

				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].Interval = FP.SelectedProfile.TX[i].Interval
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnInterval,
							FP.SelectedProfile.TX[i].Interval,
						))
					}
				}

				modified()
				deactivateTransactionsInputField()
			}
		})
		activateTransactionsInputField("interval:", strconv.Itoa(FP.SelectedProfile.TX[i].Interval))
	case c.ColumnMondayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnMonday)

			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayMondayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayMondayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayMondayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayMondayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnMonday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdayMondayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnMonday, cellText))
			}
		}

		modified()
	case c.ColumnTuesdayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnTuesday)
			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayTuesdayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayTuesdayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayTuesdayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayTuesdayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnTuesday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdayTuesdayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnTuesday, cellText))
			}
		}

		modified()
	case c.ColumnWednesdayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnWednesday)
			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayWednesdayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayWednesdayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayWednesdayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayWednesdayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnWednesday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdayWednesdayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnWednesday, cellText))
			}
		}

		modified()
	case c.ColumnThursdayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnThursday)
			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayThursdayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayThursdayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayThursdayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayThursdayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnThursday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdayThursdayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnThursday, cellText))
			}
		}

		modified()
	case c.ColumnFridayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnFriday)
			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayFridayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdayFridayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayFridayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdayFridayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnFriday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdayFridayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnFriday, cellText))
			}
		}

		modified()
	case c.ColumnSaturdayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnSaturday)
			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdaySaturdayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdaySaturdayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdaySaturdayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdaySaturdayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnSaturday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdaySaturdayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnSaturday, cellText))
			}
		}

		modified()
	case c.ColumnSundayIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnSunday)
			return
		}

		FP.SelectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(FP.SelectedProfile.TX[i].Weekdays, c.WeekdaySundayInt)

		dayIsPresent := slices.Contains(FP.SelectedProfile.TX[i].Weekdays, c.WeekdaySundayInt)

		// update all selected values as well as the current one
		for j := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[j].Selected || j == i {
				dayIndex := slices.Index(FP.SelectedProfile.TX[j].Weekdays, c.WeekdaySundayInt)
				if dayIndex == -1 && dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = append(FP.SelectedProfile.TX[j].Weekdays, c.WeekdaySundayInt)
				} else if dayIndex != -1 && !dayIsPresent {
					FP.SelectedProfile.TX[j].Weekdays = slices.Delete(FP.SelectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
				}
				sort.Ints(FP.SelectedProfile.TX[j].Weekdays)

				cellText := fmt.Sprintf("%v✔", c.ColorColumnSunday)
				if !FP.SelectedProfile.TX[j].HasWeekday(c.WeekdaySundayInt) {
					cellText = "[white] "
				}
				FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnSunday, cellText))
			}
		}

		modified()
	case c.ColumnStartsIndex:
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
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d < 0 || d > 31 {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset(
						"invalid day given:",
						strconv.Itoa(FP.SelectedProfile.TX[i].StartsDay),
					)
					return
				}

				FP.SelectedProfile.TX[i].StartsDay = int(d)

				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].StartsDay = FP.SelectedProfile.TX[i].StartsDay
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnStarts,
							FP.SelectedProfile.TX[j].GetStartDateString(),
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
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d > 12 || d < 0 {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", strconv.Itoa(FP.SelectedProfile.TX[i].StartsMonth))
					return
				}

				FP.SelectedProfile.TX[i].StartsMonth = int(d)

				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].StartsMonth = FP.SelectedProfile.TX[i].StartsMonth
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnStarts,
							FP.SelectedProfile.TX[j].GetStartDateString(),
						))
					}
				}

				modified()
				deactivateTransactionsInputField()
				activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", strconv.Itoa(FP.SelectedProfile.TX[i].StartsDay))
				defer FP.TransactionsInputField.SetDoneFunc(dayFunc)
			}
		}

		yearFunc := func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				// don't save the changes
				deactivateTransactionsInputField()
				return
			default:
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d < 0 {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", strconv.Itoa(FP.SelectedProfile.TX[i].StartsYear))
					return
				}

				FP.SelectedProfile.TX[i].StartsYear = int(d)

				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].StartsYear = FP.SelectedProfile.TX[i].StartsYear
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnStarts,
							FP.SelectedProfile.TX[j].GetStartDateString(),
						))
					}
				}

				modified()
				deactivateTransactionsInputField()
				activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", strconv.Itoa(FP.SelectedProfile.TX[i].StartsMonth))
				defer FP.TransactionsInputField.SetDoneFunc(monthFunc)
			}
		}

		FP.TransactionsInputField.SetDoneFunc(yearFunc)
		activateTransactionsInputField("year:", strconv.Itoa(FP.SelectedProfile.TX[i].StartsYear))
	case c.ColumnEndsIndex:
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
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d < 0 || d > 31 {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset("invalid day given:", strconv.Itoa(FP.SelectedProfile.TX[i].EndsDay))
					return
				}

				FP.SelectedProfile.TX[i].EndsDay = int(d)
				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].EndsDay = FP.SelectedProfile.TX[i].EndsDay
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnEnds,
							FP.SelectedProfile.TX[j].GetEndsDateString(),
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
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d > 12 || d < 0 {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", strconv.Itoa(FP.SelectedProfile.TX[i].EndsMonth))
					return
				}

				FP.SelectedProfile.TX[i].EndsMonth = int(d)
				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].EndsMonth = FP.SelectedProfile.TX[i].EndsMonth
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnEnds,
							FP.SelectedProfile.TX[j].GetEndsDateString(),
						))
					}
				}
				modified()
				deactivateTransactionsInputField()
				activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", strconv.Itoa(FP.SelectedProfile.TX[i].EndsDay))
				defer FP.TransactionsInputField.SetDoneFunc(dayFunc)
			}
		}

		yearFunc := func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				// don't save the changes
				deactivateTransactionsInputField()
				return
			default:
				d, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				if err != nil || d < 0 {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", strconv.Itoa(FP.SelectedProfile.TX[i].EndsYear))
					return
				}

				FP.SelectedProfile.TX[i].EndsYear = int(d)
				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].EndsYear = FP.SelectedProfile.TX[i].EndsYear
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
							"%v%v",
							c.ColorColumnEnds,
							FP.SelectedProfile.TX[j].GetEndsDateString(),
						))
					}
				}
				modified()
				deactivateTransactionsInputField()
				activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", strconv.Itoa(FP.SelectedProfile.TX[i].EndsMonth))
				defer FP.TransactionsInputField.SetDoneFunc(monthFunc)
			}
		}

		FP.TransactionsInputField.SetDoneFunc(yearFunc)
		activateTransactionsInputField("year:", strconv.Itoa(FP.SelectedProfile.TX[i].EndsYear))
	case c.ColumnNoteIndex:
		if row == 0 {
			setTransactionsTableSort(c.ColumnNote)
			return
		}
		activateTransactionsInputField("edit note:", FP.SelectedProfile.TX[i].Note)
		FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				break
			default:
				// save the changes
				FP.SelectedProfile.TX[i].Note = FP.TransactionsInputField.GetText()
				// update all selected values as well as the current one
				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						FP.SelectedProfile.TX[j].Note = FP.SelectedProfile.TX[i].Note
						FP.TransactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnNote, FP.SelectedProfile.TX[j].Note))
					}
				}

				modified()
			}
			deactivateTransactionsInputField()
		})
	case c.ColumnIDIndex:
		// pass for now
	case c.ColumnCreatedAtIndex:
		// pass for now
	case c.ColumnUpdatedAtIndex:
		// pass for now
	default:
		break
	}
}
