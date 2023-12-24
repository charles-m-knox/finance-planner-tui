package main

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"

	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
)

func resetTransactionsInputFieldAutocomplete() {
	FP.TransactionsInputField.SetAutocompleteFunc(func(currentText string) []string {
		return []string{}
	})
}

func deactivateTransactionsInputField() {
	FP.TransactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	FP.TransactionsInputField.SetLabel("[gray] editor appears here when editing")
	FP.TransactionsInputField.SetText("")

	if FP.Previous != nil {
		FP.App.SetFocus(FP.Previous)
	}
}

// focuses the transactions input field, updates its label, and sets
// its background color to something noticeable
func activateTransactionsInputField(msg, value string) {
	resetTransactionsInputFieldAutocomplete()

	FP.TransactionsInputField.SetFieldBackgroundColor(tcell.ColorDimGray)
	FP.TransactionsInputField.SetLabel(fmt.Sprintf("[lightgreen::b] %v[-:-:-:-]", msg))
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

// focuses the transactions input field, updates its label, and sets
// its background color to something noticeable - in some cases, the
// resetTransactionsInputFieldAutocomplete cannot be called without risking
// an infinite loop, so this function does not call it
func activateTransactionsInputFieldNoAutocompleteReset(msg, value string) {
	FP.TransactionsInputField.SetFieldBackgroundColor(tcell.ColorDimGray)
	FP.TransactionsInputField.SetLabel(fmt.Sprintf("[lightgreen::b] %v[-:-:-:-]", msg))
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

// Sorts all transactions by the current sort column.
func sortTX() {
	if FP.SortTX == c.None || FP.SortTX == "" {
		return
	}

	FP.LastSelection = -1

	sort.SliceStable(
		FP.SelectedProfile.TX,
		func(i, j int) bool {
			tj := (FP.SelectedProfile.TX)[j]
			ti := (FP.SelectedProfile.TX)[i]

			switch FP.SortTX {
			// invisible order column (default when no sort is set)
			// case c.None:
			// return tj.Order > ti.Order

			// Order
			// case fmt.Sprintf("%v%v", c.ColumnOrder, c.Asc):
			// 	return ti.Order > tj.Order
			// case fmt.Sprintf("%v%v", c.ColumnOrder, c.Desc):
			// 	return ti.Order < tj.Order

			// active
			case fmt.Sprintf("%v%v", c.ColumnActive, c.Asc):
				if ti.Active == tj.Active {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return ti.Active
			case fmt.Sprintf("%v%v", c.ColumnActive, c.Desc):
				if ti.Active == tj.Active {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tj.Active

			// weekdays
			case fmt.Sprintf("%v%v", c.WeekdayMonday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayMondayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayMondayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayMonday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayMondayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayMondayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayTuesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayTuesdayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayTuesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayTuesdayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayWednesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayWednesdayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayWednesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayWednesdayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayThursday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayThursdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayThursdayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayThursday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayThursdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayThursdayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayFriday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayFridayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayFridayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayFriday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayFridayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayFridayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

			case fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySaturdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySaturdayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySaturdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySaturdayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

			case fmt.Sprintf("%v%v", c.WeekdaySunday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySundayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySundayInt) != -1
				if tiw == tjw {
					return ti.ID > tj.ID
					// return ti.Order > tj.Order
				}

				return tiw

			case fmt.Sprintf("%v%v", c.WeekdaySunday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySundayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySundayInt) != -1
				if tiw == tjw {
					return ti.ID < tj.ID
					// return ti.Order < tj.Order
				}

				return tjw

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
				return ti.GetStartDateString() > tj.GetStartDateString()
			case fmt.Sprintf("%v%v", c.ColumnStarts, c.Desc):
				return ti.GetStartDateString() < tj.GetStartDateString()

			case fmt.Sprintf("%v%v", c.ColumnEnds, c.Asc):
				return ti.GetEndsDateString() > tj.GetEndsDateString()
			case fmt.Sprintf("%v%v", c.ColumnEnds, c.Desc):
				return ti.GetEndsDateString() < tj.GetEndsDateString()

			default:
				return false
			}
		},
	)
}

// Creates the transactions table, based on the currently selected profile.
// Heads up: This DOES modify the existing profile's transaction (mainly applies
// sorting).
func getTransactionsTable() {
	FP.TransactionsTable.Clear()

	lastSelectedColor := tcell.NewRGBColor(30, 30, 30)
	selectedAndLastSelectedColor := tcell.NewRGBColor(70, 70, 70)
	selectedColor := tcell.NewRGBColor(50, 50, 50)

	// determine the current sort
	// currentSort := strings.TrimSuffix(strings.TrimSuffix(FP.SortTX, c.Asc), c.Desc)
	currentSort := ""
	currentSortDir := ""

	if strings.HasSuffix(FP.SortTX, c.Asc) {
		currentSort = strings.Split(FP.SortTX, c.Asc)[0]
		// currentSortDir = c.Asc
		currentSortDir = "↑"
	} else if strings.HasSuffix(FP.SortTX, c.Desc) {
		currentSort = strings.Split(FP.SortTX, c.Desc)[0]
		currentSortDir = "↓"
	}

	// cellColumnOrderText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ORDER, c.ColumnOrder)
	// if currentSort == c.ColumnOrder {
	// 	cellColumnOrderText = fmt.Sprintf("%v%v", currentSortDir, cellColumnOrderText)
	// }

	cellColumnAmountText := fmt.Sprintf("%v%v", c.ColorColumnAmount, c.ColumnAmount)

	if currentSort == c.ColumnAmount {
		cellColumnAmountText = fmt.Sprintf("%v%v", currentSortDir, cellColumnAmountText)
	}

	cellColumnActiveText := fmt.Sprintf("%v%v", c.ColorColumnActive, c.ColumnActive)

	if currentSort == c.ColumnActive {
		cellColumnActiveText = fmt.Sprintf("%v%v", currentSortDir, cellColumnActiveText)
	}

	cellColumnNameText := fmt.Sprintf("%v%v", c.ColorColumnName, c.ColumnName)

	if currentSort == c.ColumnName {
		cellColumnNameText = fmt.Sprintf("%v%v", currentSortDir, cellColumnNameText)
	}

	cellColumnFrequencyText := fmt.Sprintf("%v%v", c.ColorColumnFrequency, c.ColumnFrequency)

	if currentSort == c.ColumnFrequency {
		cellColumnFrequencyText = fmt.Sprintf("%v%v", currentSortDir, cellColumnFrequencyText)
	}

	cellColumnIntervalText := fmt.Sprintf("%v%v", c.ColorColumnInterval, c.ColumnInterval)

	if currentSort == c.ColumnInterval {
		cellColumnIntervalText = fmt.Sprintf("%v%v", currentSortDir, cellColumnIntervalText)
	}

	cellColumnMondayText := fmt.Sprintf("%v%v", c.ColorColumnMonday, c.ColumnMonday)

	if currentSort == c.ColumnMonday {
		cellColumnMondayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnMondayText)
	}

	cellColumnTuesdayText := fmt.Sprintf("%v%v", c.ColorColumnTuesday, c.ColumnTuesday)

	if currentSort == c.ColumnTuesday {
		cellColumnTuesdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnTuesdayText)
	}

	cellColumnWednesdayText := fmt.Sprintf("%v%v", c.ColorColumnWednesday, c.ColumnWednesday)

	if currentSort == c.ColumnWednesday {
		cellColumnWednesdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnWednesdayText)
	}

	cellColumnThursdayText := fmt.Sprintf("%v%v", c.ColorColumnThursday, c.ColumnThursday)

	if currentSort == c.ColumnThursday {
		cellColumnThursdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnThursdayText)
	}

	cellColumnFridayText := fmt.Sprintf("%v%v", c.ColorColumnFriday, c.ColumnFriday)

	if currentSort == c.ColumnFriday {
		cellColumnFridayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnFridayText)
	}

	cellColumnSaturdayText := fmt.Sprintf("%v%v", c.ColorColumnSaturday, c.ColumnSaturday)

	if currentSort == c.ColumnSaturday {
		cellColumnSaturdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnSaturdayText)
	}

	cellColumnSundayText := fmt.Sprintf("%v%v", c.ColorColumnSunday, c.ColumnSunday)

	if currentSort == c.ColumnSunday {
		cellColumnSundayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnSundayText)
	}

	cellColumnStartsText := fmt.Sprintf("%v%v", c.ColorColumnStarts, c.ColumnStarts)

	if currentSort == c.ColumnStarts {
		cellColumnStartsText = fmt.Sprintf("%v%v", currentSortDir, cellColumnStartsText)
	}

	cellColumnEndsText := fmt.Sprintf("%v%v", c.ColorColumnEnds, c.ColumnEnds)

	if currentSort == c.ColumnEnds {
		cellColumnEndsText = fmt.Sprintf("%v%v", currentSortDir, cellColumnEndsText)
	}

	cellColumnNoteText := fmt.Sprintf("%v%v", c.ColorColumnNote, c.ColumnNote)

	if currentSort == c.ColumnNote {
		cellColumnNoteText = fmt.Sprintf("%v%v", currentSortDir, cellColumnNoteText)
	}

	// cellColumnOrder := tview.NewTableCell(cellColumnOrderText)
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

	// FP.TransactionsTable.SetCell(0, 0, cellColumnOrder)
	FP.TransactionsTable.SetCell(0, 0, cellColumnAmount)
	FP.TransactionsTable.SetCell(0, 1, cellColumnActive)
	FP.TransactionsTable.SetCell(0, 2, cellColumnName)
	FP.TransactionsTable.SetCell(0, 3, cellColumnFrequency)
	FP.TransactionsTable.SetCell(0, 4, cellColumnInterval)
	FP.TransactionsTable.SetCell(0, 5, cellColumnMonday)
	FP.TransactionsTable.SetCell(0, 6, cellColumnTuesday)
	FP.TransactionsTable.SetCell(0, 7, cellColumnWednesday)
	FP.TransactionsTable.SetCell(0, 8, cellColumnThursday)
	FP.TransactionsTable.SetCell(0, 9, cellColumnFriday)
	FP.TransactionsTable.SetCell(0, 10, cellColumnSaturday)
	FP.TransactionsTable.SetCell(0, 11, cellColumnSunday)
	FP.TransactionsTable.SetCell(0, 12, cellColumnStarts)
	FP.TransactionsTable.SetCell(0, 13, cellColumnEnds)
	FP.TransactionsTable.SetCell(0, 14, cellColumnNote)
	// FP.TransactionsTable.SetCell(0, 15, cellColumnID)
	// FP.TransactionsTable.SetCell(0, 16, cellColumnCreatedAt)
	// FP.TransactionsTable.SetCell(0, 17, cellColumnUpdatedAt)

	if FP.SelectedProfile != nil {
		sortTX()
		// start by populating the table with the columns first
		for i, tx := range FP.SelectedProfile.TX {
			isPositiveAmount := tx.Amount >= 0
			amountColor := c.ColorColumnAmount

			if isPositiveAmount {
				amountColor = c.ColorColumnAmountPositive
			}

			nameColor := c.ColorColumnName
			noteColor := c.ColorColumnNote

			if !tx.Active {
				amountColor = c.ColorInactive
				nameColor = c.ColorInactive
				noteColor = c.ColorInactive
			}

			// cellOrder := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ORDER, tx.Order)).SetAlign(tview.AlignCenter)
			cellAmount := tview.NewTableCell(fmt.Sprintf("%v%v", amountColor, lib.FormatAsCurrency(tx.Amount))).SetAlign(tview.AlignCenter)

			activeText := "✔"
			if !tx.Active {
				activeText = " "
			}

			cellActive := tview.NewTableCell(fmt.Sprintf("%v%v", c.ColorColumnActive, activeText)).SetAlign(tview.AlignCenter)
			cellName := tview.NewTableCell(fmt.Sprintf("%v%v", nameColor, tx.Name)).SetAlign(tview.AlignLeft)
			cellFrequency := tview.NewTableCell(fmt.Sprintf("%v%v", c.ColorColumnFrequency, tx.Frequency)).SetAlign(tview.AlignCenter)
			cellInterval := tview.NewTableCell(fmt.Sprintf("%v%v", c.ColorColumnInterval, tx.Interval)).SetAlign(tview.AlignCenter)

			mondayText := fmt.Sprintf("%v✔", c.ColorColumnMonday)

			if !tx.HasWeekday(c.WeekdayMondayInt) {
				mondayText = "[white] "
			}

			tuesdayText := fmt.Sprintf("%v✔", c.ColorColumnTuesday)

			if !tx.HasWeekday(c.WeekdayTuesdayInt) {
				tuesdayText = "[white] "
			}

			wednesdayText := fmt.Sprintf("%v✔", c.ColorColumnWednesday)

			if !tx.HasWeekday(c.WeekdayWednesdayInt) {
				wednesdayText = "[white] "
			}

			thursdayText := fmt.Sprintf("%v✔", c.ColorColumnThursday)

			if !tx.HasWeekday(c.WeekdayThursdayInt) {
				thursdayText = "[white] "
			}

			fridayText := fmt.Sprintf("%v✔", c.ColorColumnFriday)

			if !tx.HasWeekday(c.WeekdayFridayInt) {
				fridayText = "[white] "
			}

			saturdayText := fmt.Sprintf("%v✔", c.ColorColumnSaturday)

			if !tx.HasWeekday(c.WeekdaySaturdayInt) {
				saturdayText = "[white] "
			}

			sundayText := fmt.Sprintf("%v✔", c.ColorColumnSunday)

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

			cellStarts := tview.NewTableCell(fmt.Sprintf("%v%v", c.ColorColumnStarts, tx.GetStartDateString())).SetAlign(tview.AlignCenter)
			cellEnds := tview.NewTableCell(fmt.Sprintf("%v%v", c.ColorColumnEnds, tx.GetEndsDateString())).SetAlign(tview.AlignCenter)

			cellNote := tview.NewTableCell(fmt.Sprintf("%v%v", noteColor, tx.Note))

			// cellID := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ID, tx.ID))
			// cellCreatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",fmt.Sprintf("%v", tx.CreatedAt)))
			// cellUpdatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",fmt.Sprintf("%v", tx.UpdatedAt)))

			cellName.SetExpansion(1)
			cellNote.SetExpansion(1)

			if FP.LastSelection == i {
				if tx.Selected {
					// cellOrder.SetBackgroundColor(selectedAndLastSelectedColor)
					cellAmount.SetBackgroundColor(selectedAndLastSelectedColor)
					cellActive.SetBackgroundColor(selectedAndLastSelectedColor)
					cellName.SetBackgroundColor(selectedAndLastSelectedColor)
					cellFrequency.SetBackgroundColor(selectedAndLastSelectedColor)
					cellInterval.SetBackgroundColor(selectedAndLastSelectedColor)
					cellMonday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellTuesday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellWednesday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellThursday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellFriday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellSaturday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellSunday.SetBackgroundColor(selectedAndLastSelectedColor)
					cellStarts.SetBackgroundColor(selectedAndLastSelectedColor)
					cellEnds.SetBackgroundColor(selectedAndLastSelectedColor)
					cellNote.SetBackgroundColor(selectedAndLastSelectedColor)
					// cellID.SetBackgroundColor(selectedAndLastSelectedColor)
					// cellCreatedAt.SetBackgroundColor(selectedAndLastSelectedColor)
					// cellUpdatedAt.SetBackgroundColor(selectedAndLastSelectedColor)
				} else {
					// cellOrder.SetBackgroundColor(lastSelectedColor)
					cellAmount.SetBackgroundColor(lastSelectedColor)
					cellActive.SetBackgroundColor(lastSelectedColor)
					cellName.SetBackgroundColor(lastSelectedColor)
					cellFrequency.SetBackgroundColor(lastSelectedColor)
					cellInterval.SetBackgroundColor(lastSelectedColor)
					cellMonday.SetBackgroundColor(lastSelectedColor)
					cellTuesday.SetBackgroundColor(lastSelectedColor)
					cellWednesday.SetBackgroundColor(lastSelectedColor)
					cellThursday.SetBackgroundColor(lastSelectedColor)
					cellFriday.SetBackgroundColor(lastSelectedColor)
					cellSaturday.SetBackgroundColor(lastSelectedColor)
					cellSunday.SetBackgroundColor(lastSelectedColor)
					cellStarts.SetBackgroundColor(lastSelectedColor)
					cellEnds.SetBackgroundColor(lastSelectedColor)
					cellNote.SetBackgroundColor(lastSelectedColor)
					// cellID.SetBackgroundColor(lastSelectedColor)
					// cellCreatedAt.SetBackgroundColor(lastSelectedColor)
					// cellUpdatedAt.SetBackgroundColor(lastSelectedColor)
				}
			} else if tx.Selected {
				// cellOrder.SetBackgroundColor(selectedColor)
				cellAmount.SetBackgroundColor(selectedColor)
				cellActive.SetBackgroundColor(selectedColor)
				cellName.SetBackgroundColor(selectedColor)
				cellFrequency.SetBackgroundColor(selectedColor)
				cellInterval.SetBackgroundColor(selectedColor)
				cellMonday.SetBackgroundColor(selectedColor)
				cellTuesday.SetBackgroundColor(selectedColor)
				cellWednesday.SetBackgroundColor(selectedColor)
				cellThursday.SetBackgroundColor(selectedColor)
				cellFriday.SetBackgroundColor(selectedColor)
				cellSaturday.SetBackgroundColor(selectedColor)
				cellSunday.SetBackgroundColor(selectedColor)
				cellStarts.SetBackgroundColor(selectedColor)
				cellEnds.SetBackgroundColor(selectedColor)
				cellNote.SetBackgroundColor(selectedColor)
				// cellID.SetBackgroundColor(selectedColor)
				// cellCreatedAt.SetBackgroundColor(selectedColor)
				// cellUpdatedAt.SetBackgroundColor(selectedColor)
			}

			// FP.TransactionsTable.SetCell(i+1, 0, cellOrder)
			FP.TransactionsTable.SetCell(i+1, 0, cellAmount)
			FP.TransactionsTable.SetCell(i+1, 1, cellActive)
			FP.TransactionsTable.SetCell(i+1, 2, cellName)
			FP.TransactionsTable.SetCell(i+1, 3, cellFrequency)
			FP.TransactionsTable.SetCell(i+1, 4, cellInterval)
			FP.TransactionsTable.SetCell(i+1, 5, cellMonday)
			FP.TransactionsTable.SetCell(i+1, 6, cellTuesday)
			FP.TransactionsTable.SetCell(i+1, 7, cellWednesday)
			FP.TransactionsTable.SetCell(i+1, 8, cellThursday)
			FP.TransactionsTable.SetCell(i+1, 9, cellFriday)
			FP.TransactionsTable.SetCell(i+1, 10, cellSaturday)
			FP.TransactionsTable.SetCell(i+1, 11, cellSunday)
			FP.TransactionsTable.SetCell(i+1, 12, cellStarts)
			FP.TransactionsTable.SetCell(i+1, 13, cellEnds)
			FP.TransactionsTable.SetCell(i+1, 14, cellNote)
			// FP.TransactionsTable.SetCell(i+1, 15, cellID)
			// FP.TransactionsTable.SetCell(i+1, 16, cellCreatedAt)
			// FP.TransactionsTable.SetCell(i+1, 17, cellUpdatedAt)
		}

		FP.TransactionsTable.SetSelectedFunc(func(row, column int) {
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
		})
	}

	FP.TransactionsTable.SetTitle("Transactions")
	FP.TransactionsTable.SetBorders(false).
		SetSelectable(true, true). // set row & cells to be selectable
		SetSeparator(' ')
}

func setTransactionsTableSort(column string) {
	FP.SortTX = lib.GetNextSort(FP.SortTX, column)

	getTransactionsTable()
}
