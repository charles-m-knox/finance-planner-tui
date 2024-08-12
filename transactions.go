package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	lib "git.cmcode.dev/cmcode/finance-planner-lib"

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
		func(_ string) []string {
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
		Reset,
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
		Reset,
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

func sortWeekday(day int, asc bool) TxSortFunc {
	return func(ti, tj lib.TX) bool {
		tiw := ti.Weekdays[day]
		tjw := tj.Weekdays[day]

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

// func sortID(asc bool) TxSortFunc {
// 	return func(ti, tj lib.TX) bool {
// 		til := strings.ToLower(ti.ID)
// 		tjl := strings.ToLower(tj.ID)

// 		if asc {
// 			if til == tjl {
// 				return ti.ID > tj.ID
// 			}

// 			return til > tjl
// 		}

// 		if til == tjl {
// 			return ti.ID < tj.ID
// 		}

// 		return til < tjl
// 	}
// }

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

// // TODO: validate that this works as expected.
// func sortCreatedAt(asc bool) TxSortFunc {
// 	return func(ti, tj lib.TX) bool {
// 		if asc {
// 			return ti.CreatedAt.After(tj.CreatedAt)
// 		}

// 		return ti.CreatedAt.Before(tj.CreatedAt)
// 	}
// }

// // TODO: validate that this works as expected.
// func sortUpdatedAt(asc bool) TxSortFunc {
// 	return func(ti, tj lib.TX) bool {
// 		if asc {
// 			return ti.UpdatedAt.After(tj.UpdatedAt)
// 		}

// 		return ti.UpdatedAt.Before(tj.UpdatedAt)
// 	}
// }

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
	mo := func(b bool) TxSortFunc { return sortWeekday(rrule.MO.Day(), b) }
	tu := func(b bool) TxSortFunc { return sortWeekday(rrule.TU.Day(), b) }
	we := func(b bool) TxSortFunc { return sortWeekday(rrule.WE.Day(), b) }
	th := func(b bool) TxSortFunc { return sortWeekday(rrule.TH.Day(), b) }
	fr := func(b bool) TxSortFunc { return sortWeekday(rrule.FR.Day(), b) }
	sa := func(b bool) TxSortFunc { return sortWeekday(rrule.SA.Day(), b) }
	su := func(b bool) TxSortFunc { return sortWeekday(rrule.SU.Day(), b) }

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
	if FP.SortTX == None || FP.SortTX == "" {
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
func getTransactionsTableHeaders() []TableCell {
	return []TableCell{
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
func getTransactionsTableCell(tx lib.TX) []TableCell {
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
		cAmount = FP.Colors["TransactionsInactive"]
		cActive = FP.Colors["TransactionsInactive"]
		cName = FP.Colors["TransactionsInactive"]
		cFrequency = FP.Colors["TransactionsInactive"]
		cInterval = FP.Colors["TransactionsInactive"]
		cMonday = FP.Colors["TransactionsInactive"]
		cTuesday = FP.Colors["TransactionsInactive"]
		cWednesday = FP.Colors["TransactionsInactive"]
		cThursday = FP.Colors["TransactionsInactive"]
		cFriday = FP.Colors["TransactionsInactive"]
		cSaturday = FP.Colors["TransactionsInactive"]
		cSunday = FP.Colors["TransactionsInactive"]
		cStarts = FP.Colors["TransactionsInactive"]
		cEnds = FP.Colors["TransactionsInactive"]
		cNote = FP.Colors["TransactionsInactive"]
	} else { //nolint:gocritic // <-- intentionally structured like this
		if tx.Amount >= 0 {
			cAmount = FP.Colors["TransactionsAmountPositive"]
		}
	}

	cells := []TableCell{
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
func setTransactionsTableHeaders(th []TableCell, currentSort, sortGlyph string) {
	for i := range th {
		g := ""
		if currentSort == th[i].Text {
			g = sortGlyph
		}

		cell := tview.NewTableCell(fmt.Sprintf("%v%v%v%v",
			th[i].Color,
			g,
			th[i].Text,
			Reset,
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
			Reset,
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

	FP.TransactionsTableHeaders = getTransactionsTableHeaders()

	setTransactionsTableHeaders(FP.TransactionsTableHeaders, currentSort, sortGlyph)

	FP.TransactionsTable.SetTitle(FP.T["TransactionsTableTitle"])
	FP.TransactionsTable.SetBorders(false).
		SetSelectable(true, true).
		SetSeparator(' ')

	if FP.SelectedProfile == nil {
		return
	}

	sortTX(FP.TransactionsSortMap)

	for i := range FP.SelectedProfile.TX {
		setTransactionsTableCellsForTransaction(i+1, FP.SelectedProfile.TX[i], FP.LastSelection == i)
	}

	FP.TransactionsTable.SetSelectedFunc(transactionsTableSelectedFunc)
}

// First, prompt for the year. Then, prompt for month. Then, prompt for day.
//
// If start is false, this will modify the end date; otherwise, it modifies the
// start date.
func txChangeDate(i int, start bool) {
	const (
		Y = "Y"
		M = "M"
		D = "D"
	)

	// helps get rid of some boilerplate (but not all)
	fnc := func(yrMoDay string, f func()) func(key tcell.Key) {
		return func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				// don't save the changes
				deactivateTransactionsInputField()

				return
			default:
				v, err := strconv.ParseInt(FP.TransactionsInputField.GetText(), 10, 64)
				valid := false

				// This is a dynamic method of setting the value of the selected
				// TX's start/end yr/mo/day.
				var field *int

				switch yrMoDay {
				case Y:
					valid = v >= 0

					if start {
						field = &(FP.SelectedProfile.TX[i].StartsYear)
					} else {
						field = &(FP.SelectedProfile.TX[i].EndsYear)
					}
				case M:
					valid = v >= 0 && v <= 12

					if start {
						field = &(FP.SelectedProfile.TX[i].StartsMonth)
					} else {
						field = &(FP.SelectedProfile.TX[i].EndsMonth)
					}
				case D:
					valid = v >= 0 && v <= 31

					if start {
						field = &(FP.SelectedProfile.TX[i].StartsDay)
					} else {
						field = &(FP.SelectedProfile.TX[i].EndsDay)
					}
				}

				if err != nil || !valid {
					// start over
					activateTransactionsInputFieldNoAutocompleteReset(
						FP.T[fmt.Sprintf("TransactionsInputFieldInvalidDateGivenLabel%v", yrMoDay)],
						strconv.Itoa(*field),
					)

					return
				}

				// FP.SelectedProfile.TX[i].StartsYear = int(d)
				*field = int(v)

				for j := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[j].Selected || j == i {
						var jField *int

						switch yrMoDay {
						case Y:
							if start {
								jField = &(FP.SelectedProfile.TX[j].StartsYear)
							} else {
								jField = &(FP.SelectedProfile.TX[j].EndsYear)
							}
						case M:
							if start {
								jField = &(FP.SelectedProfile.TX[j].StartsMonth)
							} else {
								jField = &(FP.SelectedProfile.TX[j].EndsMonth)
							}
						case D:
							if start {
								jField = &(FP.SelectedProfile.TX[j].StartsDay)
							} else {
								jField = &(FP.SelectedProfile.TX[j].EndsDay)
							}
						}

						*jField = *field
					}
				}

				modified()
				deactivateTransactionsInputField()

				f()
			}
		}
	}

	dayFunc := func() {
		FP.TransactionsInputField.SetDoneFunc(func(_ /* key */ tcell.Key) {})
	}

	monthFunc := func() {
		FP.TransactionsInputField.SetDoneFunc(fnc(D, dayFunc))

		var m string

		if start {
			m = strconv.Itoa(FP.SelectedProfile.TX[i].StartsDay)
		} else {
			m = strconv.Itoa(FP.SelectedProfile.TX[i].EndsDay)
		}

		activateTransactionsInputFieldNoAutocompleteReset(fmt.Sprintf(
			"%v:", FP.T["TransactionsInputFieldDayPromptLabel"],
		), m)
	}

	yearFunc := func() {
		FP.TransactionsInputField.SetDoneFunc(fnc(M, monthFunc))

		var m string

		if start {
			m = strconv.Itoa(FP.SelectedProfile.TX[i].StartsMonth)
		} else {
			m = strconv.Itoa(FP.SelectedProfile.TX[i].EndsMonth)
		}

		activateTransactionsInputFieldNoAutocompleteReset(fmt.Sprintf(
			"%v:", FP.T["TransactionsInputFieldMonthPromptLabel"],
		), m)
	}

	var m string

	if start {
		m = strconv.Itoa(FP.SelectedProfile.TX[i].StartsYear)
	} else {
		m = strconv.Itoa(FP.SelectedProfile.TX[i].EndsYear)
	}

	activateTransactionsInputField(fmt.Sprintf("%v:", FP.T["TransactionsInputFieldYearPromptLabel"]), m)

	FP.TransactionsInputField.SetDoneFunc(fnc(Y, yearFunc))
}

// TODO: translate these.
// TODO: map colors, if any are used.
func txChangeFrequency(i int) {
	activateTransactionsInputField("weekly|monthly|yearly:", FP.SelectedProfile.TX[i].Frequency)

	saveFunc := func(newValue string) {
		validatedFrequency := strings.TrimSpace(strings.ToUpper(newValue))
		switch validatedFrequency {
		case WEEKLY:
			fallthrough
		case MONTHLY:
			fallthrough
		case YEARLY:
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
			}
		}

		modified()
	}

	FP.TransactionsInputField.SetAutocompleteFunc(func(currentText string) []string {
		return fuzzy.Find(strings.TrimSpace(strings.ToUpper(currentText)), []string{
			MONTHLY,
			YEARLY,
			WEEKLY,
		})
	})

	FP.TransactionsInputField.SetAutocompletedFunc(func(text string, _ /* index */, _ /* source */ int) bool {
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
}

// See txChangeDoneFunc.
type TxChangeDoneFunc func(i int, newVal string, key tcell.Key) bool

// txChangeDoneFunc is a generalized function. When the user is finished
// entering data into the transactions input field, only four keys are supported
// by tview - enter, backtab, tab, and escape. In most cases, the
// tab+backtab+enter keys are all used as "affirmative" keys and the escape key
// is treated as the "cancel" signal.
//
// When the above conditions are desirable, this function is meant to be passed
// as the transactions input field "done" handler.
func txChangeDoneFunc(i int, f func(ii int, newVal string) bool) func(tcell.Key) {
	return func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			break
		default:
			if f(i, FP.TransactionsInputField.GetText()) {
				modified()
			} else {
				return // don't drop focus - the user entered invalid input
			}
		}

		deactivateTransactionsInputField()
	}
}

// Updates all selected values as well as the current one. The "i" parameter
// is the current TX to modify, used as FP.SelectedProfile.TX[i]. Pass in
// a currency-formatted string, straight from the transactions input field.
//
// "amt" must be a string.
func txSetAmount(i int, amt string) bool {
	a := int(lib.ParseDollarAmount(amt, false))

	for j := range FP.SelectedProfile.TX {
		if FP.SelectedProfile.TX[j].Selected || j == i {
			FP.SelectedProfile.TX[j].Amount = a
		}
	}

	return true
}

// txChangeAmount is the callback that gets executed when the user changes the
// Amount field in the transactions table. The "i" parameter is the index of the
// currently selected transaction, for example:
//
// FP.SelectedProfile.TX[i].Amount.
func txChangeAmount(i int) {
	FP.TransactionsInputField.SetDoneFunc(txChangeDoneFunc(i, txSetAmount))

	renderedAmount := lib.FormatAsCurrency(FP.SelectedProfile.TX[i].Amount)
	if FP.SelectedProfile.TX[i].Amount >= 0 {
		renderedAmount = fmt.Sprintf("+%v", renderedAmount)
	}

	activateTransactionsInputField(
		fmt.Sprintf("%v:", FP.T["TransactionsInputFieldEditAmountLabel"]),
		renderedAmount,
	)
}

func txSetActive(i int) bool {
	FP.SelectedProfile.TX[i].Active = !FP.SelectedProfile.TX[i].Active

	for j := range FP.SelectedProfile.TX {
		if !FP.SelectedProfile.TX[j].Selected {
			continue
		}

		FP.SelectedProfile.TX[j].Active = FP.SelectedProfile.TX[i].Active
	}

	return true
}

func txSetWeekday(i int, w int) bool {
	FP.SelectedProfile.TX[i].Weekdays[w] = !FP.SelectedProfile.TX[i].Weekdays[w]

	for j := range FP.SelectedProfile.TX {
		if !FP.SelectedProfile.TX[j].Selected {
			continue
		}

		FP.SelectedProfile.TX[j].Weekdays[w] = FP.SelectedProfile.TX[i].Weekdays[w]
	}

	return true
}

func txSetName(i int, name string) bool {
	for j := range FP.SelectedProfile.TX {
		if FP.SelectedProfile.TX[j].Selected || j == i {
			FP.SelectedProfile.TX[j].Name = name
		}
	}

	return true
}

func txChangeName(i int) {
	FP.TransactionsInputField.SetDoneFunc(txChangeDoneFunc(i, txSetName))

	activateTransactionsInputField(
		fmt.Sprintf("%v:", FP.T["TransactionsInputFieldEditNameLabel"]),
		FP.SelectedProfile.TX[i].Name,
	)
}

func txSetNote(i int, note string) bool {
	for j := range FP.SelectedProfile.TX {
		if FP.SelectedProfile.TX[j].Selected || j == i {
			FP.SelectedProfile.TX[j].Note = note
		}
	}

	return true
}

func txChangeNote(i int) {
	FP.TransactionsInputField.SetDoneFunc(txChangeDoneFunc(i, txSetNote))

	activateTransactionsInputField(
		fmt.Sprintf("%v:", FP.T["TransactionsInputFieldEditNoteLabel"]),
		FP.SelectedProfile.TX[i].Note,
	)
}

func txSetInterval(i int, interval string) bool {
	d, err := strconv.ParseInt(interval, 10, 64)
	if err != nil || d < 0 {
		activateTransactionsInputFieldNoAutocompleteReset(
			fmt.Sprintf("%v:", FP.T["TransactionsInputFieldInvalidIntervalGivenLabel"]),
			strconv.Itoa(FP.SelectedProfile.TX[i].Interval),
		)

		return false
	}

	FP.SelectedProfile.TX[i].Interval = int(d)

	for j := range FP.SelectedProfile.TX {
		if !FP.SelectedProfile.TX[j].Selected {
			continue
		}

		FP.SelectedProfile.TX[j].Interval = FP.SelectedProfile.TX[i].Interval
	}

	return true
}

func txChangeInterval(i int) {
	FP.TransactionsInputField.SetDoneFunc(txChangeDoneFunc(i, txSetInterval))

	activateTransactionsInputField(
		fmt.Sprintf("%v:", FP.T["TransactionsInputFieldEditIntervalLabel"]),
		strconv.Itoa(FP.SelectedProfile.TX[i].Interval),
	)
}

// This is basically a callback function that is executed when the user hits
// the enter key on a cell in the transactions table.
//
//nolint:funlen,cyclop
func transactionsTableSelectedFunc(row, column int) {
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

	field := FP.TransactionsTableHeaders[column].Text

	if row == 0 {
		setTransactionsTableSort(field)

		return
	}

	// Some actions do not contain a call to run modified() because they
	// don't use the transactions input field.
	var isModified bool

	switch field {
	case FP.T["TransactionsColumnAmount"]:
		txChangeAmount(i)
	case FP.T["TransactionsColumnActive"]:
		isModified = txSetActive(i)
	case FP.T["TransactionsColumnName"]:
		txChangeName(i)
	case FP.T["TransactionsColumnFrequency"]:
		txChangeFrequency(i)
	case FP.T["TransactionsColumnInterval"]:
		txChangeInterval(i)
	case FP.T["TransactionsColumnMonday"]:
		isModified = txSetWeekday(i, rrule.MO.Day())
	case FP.T["TransactionsColumnTuesday"]:
		isModified = txSetWeekday(i, rrule.TU.Day())
	case FP.T["TransactionsColumnWednesday"]:
		isModified = txSetWeekday(i, rrule.WE.Day())
	case FP.T["TransactionsColumnThursday"]:
		isModified = txSetWeekday(i, rrule.TH.Day())
	case FP.T["TransactionsColumnFriday"]:
		isModified = txSetWeekday(i, rrule.FR.Day())
	case FP.T["TransactionsColumnSaturday"]:
		isModified = txSetWeekday(i, rrule.SA.Day())
	case FP.T["TransactionsColumnSunday"]:
		isModified = txSetWeekday(i, rrule.SU.Day())
	case FP.T["TransactionsColumnStarts"]:
		txChangeDate(i, true)
	case FP.T["TransactionsColumnEnds"]:
		txChangeDate(i, false)
	case FP.T["TransactionsColumnNote"]:
		txChangeNote(i)
	default:
		break
	}

	if isModified {
		modified()
	}
}
