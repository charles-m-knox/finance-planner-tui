package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// completely rebuilds the results form, safe to run repeatedly
func updateResultsForm() {
	FP.ResultsForm.Clear(true)
	FP.ResultsForm.SetTitle("Parameters")

	if FP.SelectedProfile == nil {
		return
	}

	setSelectedProfileDefaults()

	FP.ResultsForm.
		AddInputField("Start Year:", FP.SelectedProfile.StartYear, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 {
				return false
			}
			return true
		}, func(text string) { FP.SelectedProfile.StartYear = text }).
		AddInputField("Start Month:", FP.SelectedProfile.StartMonth, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 12 {
				return false
			}
			return true
		}, func(text string) { FP.SelectedProfile.StartMonth = text }).
		AddInputField("Start Day:", FP.SelectedProfile.StartDay, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 31 {
				return false
			}
			return true
		}, func(text string) { FP.SelectedProfile.StartDay = text }).
		AddInputField("End Year:", FP.SelectedProfile.EndYear, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 {
				return false
			}
			return true
		}, func(text string) { FP.SelectedProfile.EndYear = text }).
		AddInputField("End Month:", FP.SelectedProfile.EndMonth, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 12 {
				return false
			}
			return true
		}, func(text string) { FP.SelectedProfile.EndMonth = text }).
		AddInputField("End Day:", FP.SelectedProfile.EndDay, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 31 {
				return false
			}
			return true
		}, func(text string) { FP.SelectedProfile.EndDay = text }).
		AddInputField("Starting Balance:", FP.SelectedProfile.StartingBalance, 0, nil, func(text string) {
			FP.SelectedProfile.StartingBalance = lib.FormatAsCurrency(int(lib.ParseDollarAmount(text, true)))
		}).
		AddButton("Submit", func() {
			getResultsTable()
		}).
		AddButton("1 year", func() {
			setResultsFormPreset(c.StartTodayPreset, c.OneYear)
			updateResultsForm()
			getResultsTable()
		}).
		AddButton("5 years", func() {
			setResultsFormPreset(c.StartTodayPreset, c.FiveYear)
			updateResultsForm()
			getResultsTable()
		}).
		AddButton("Stats", func() {
			getResultsStats()
		})

	FP.ResultsForm.SetLabelColor(tcell.ColorViolet)
	FP.ResultsForm.SetFieldBackgroundColor(tcell.NewRGBColor(40, 40, 40))
	FP.ResultsForm.SetBorder(true)
}

func getResultsPage() *tview.Flex {
	FP.ResultsTable = tview.NewTable().SetFixed(1, 1)

	FP.ResultsForm = tview.NewForm()

	FP.ResultsTable.SetBorder(true)
	updateResultsForm()

	FP.ResultsTable.SetTitle("Results")
	FP.ResultsTable.SetBorders(false).
		SetSelectable(true, false). // set row & cells to be selectable
		SetSeparator(' ')

	FP.ResultsDescription = tview.NewTextView().SetDynamicColors(true)
	FP.ResultsDescription.SetBorder(true)

	resultsRightSide := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(FP.ResultsTable, 0, 2, true).
		AddItem(FP.ResultsDescription, 0, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(FP.ResultsForm, 0, 1, true).
		AddItem(resultsRightSide, 0, 3, false)
}

// Allows a simple button press to set the start & end dates to various common
// use cases. For example, start from today and end 1 year or 5 years from now.
//
// TODO: implement other start date logic - currently only supports today
func setResultsFormPreset(startDate string, endDate string) {
	var start, end time.Time

	switch startDate {
	case c.StartTodayPreset:
		fallthrough
	default:
		start = time.Now()
	}

	switch endDate {
	case c.OneYear:
		end = start.Add(time.Hour * 24 * 365)
	case c.FiveYear:
		end = start.Add(time.Hour * 24 * 365 * 5)
	}

	FP.SelectedProfile.StartYear = strconv.Itoa(start.Year())
	FP.SelectedProfile.StartMonth = strconv.Itoa(int(start.Month()))
	FP.SelectedProfile.StartDay = strconv.Itoa(start.Day())

	FP.SelectedProfile.EndYear = strconv.Itoa(end.Year())
	FP.SelectedProfile.EndMonth = strconv.Itoa(int(end.Month()))
	FP.SelectedProfile.EndDay = strconv.Itoa(end.Day())
}

// Populates the results description with basic statistics about the results,
// and queues an UpdateDraw
func getResultsStats() {
	go FP.App.QueueUpdateDraw(func() {
		if FP.LatestResults == nil {
			return
		}

		stats, err := lib.GetStats(*(FP.LatestResults))
		if err != nil {
			resultsStatus(fmt.Sprintf(
				"%v: %v",
				FP.T["ResultsStatsErrorGettingStats"],
				err.Error(),
			))
		}

		resultsStatus(stats)
	})
}

// resultsStatus is a reusable function that produces text in the
// results textview on the bottom portion of the page.
func resultsStatus(m string) {
	FP.ResultsDescription.SetText(fmt.Sprintf("[-]%v[-]", m))
}

func getResultsTable() {
	if FP.CalculatingResults {
		return
	}

	FP.CalculatingResults = true

	go func() {
		FP.ResultsTable.Clear()
		FP.ResultsDescription.Clear()

		resultsStatus(FP.T["ResultsTableStatusCalculatingPleaseWait"])

		setSelectedProfileDefaults()

		// get results
		results, err := lib.GenerateResultsFromDateStrings(
			&(FP.SelectedProfile.TX),
			int(lib.ParseDollarAmount(FP.SelectedProfile.StartingBalance, true)),
			lib.GetDateString(FP.SelectedProfile.StartYear, FP.SelectedProfile.StartMonth, FP.SelectedProfile.StartDay),
			lib.GetDateString(FP.SelectedProfile.EndYear, FP.SelectedProfile.EndMonth, FP.SelectedProfile.EndDay),
			func(status string) {
				if FP.Config.DisableResultsStatusMessages || FP.ResultsDescription == nil {
					return
				}

				go FP.App.QueueUpdateDraw(func() {
					resultsStatus(status)
				})
			},
		)
		if err != nil {
			resultsStatus(fmt.Sprintf("%v: %v", FP.T["ResultsGenerationFailed"], err.Error()))
			return
		}

		// this may help with garbage collection when working with bigger data
		if FP.LatestResults != nil {
			if *(FP.LatestResults) != nil {
				clear(*(FP.LatestResults))
				(*(FP.LatestResults)) = nil
			}

			FP.LatestResults = nil
		}

		FP.LatestResults = &results

		// set up headers
		hDate := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDate, c.ColumnDate, c.ResetStyle))
		hBalance := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsBalance, c.ColumnBalance, c.ResetStyle))
		hCumulativeIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsCumulativeIncome, c.ColumnCumulativeIncome, c.ResetStyle))
		hCumulativeExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsCumulativeExpenses, c.ColumnCumulativeExpenses, c.ResetStyle))
		hDayExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayExpenses, c.ColumnDayExpenses, c.ResetStyle))
		hDayIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayIncome, c.ColumnDayIncome, c.ResetStyle))
		hDayNet := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayNet, c.ColumnDayNet, c.ResetStyle))
		hDiffFromStart := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDiffFromStart, c.ColumnDiffFromStart, c.ResetStyle))
		hDayTransactionNames := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayTransactionNames, c.ColumnDayTransactionNames, c.ResetStyle))

		FP.ResultsTable.SetCell(0, 0, hDate)
		FP.ResultsTable.SetCell(0, 1, hBalance)
		FP.ResultsTable.SetCell(0, 2, hCumulativeIncome)
		FP.ResultsTable.SetCell(0, 3, hCumulativeExpenses)
		FP.ResultsTable.SetCell(0, 4, hDayExpenses)
		FP.ResultsTable.SetCell(0, 5, hDayIncome)
		FP.ResultsTable.SetCell(0, 6, hDayNet)
		FP.ResultsTable.SetCell(0, 7, hDiffFromStart)
		FP.ResultsTable.SetCell(0, 7, hDiffFromStart)
		FP.ResultsTable.SetCell(0, 8, hDayTransactionNames)

		// now add the remaining rows
		for i := range results {
			rDate := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDate, lib.FormatAsDate(results[i].Date), c.ResetStyle))
			rBalance := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsBalance, lib.FormatAsCurrency(results[i].Balance), c.ResetStyle))
			rCumulativeIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsCumulativeIncome, lib.FormatAsCurrency(results[i].CumulativeIncome), c.ResetStyle))
			rCumulativeExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsCumulativeExpenses, lib.FormatAsCurrency(results[i].CumulativeExpenses), c.ResetStyle))
			rDayExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayExpenses, lib.FormatAsCurrency(results[i].DayExpenses), c.ResetStyle))
			rDayIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayIncome, lib.FormatAsCurrency(results[i].DayIncome), c.ResetStyle))
			rDayNet := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayNet, lib.FormatAsCurrency(results[i].DayNet), c.ResetStyle))
			rDiffFromStart := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDiffFromStart, lib.FormatAsCurrency(results[i].DiffFromStart), c.ResetStyle))
			rDayTransactionNames := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.ColorColumnResultsDayTransactionNames, results[i].DayTransactionNames, c.ResetStyle))

			rDayTransactionNames.SetExpansion(1)

			FP.ResultsTable.SetCell(i+1, 0, rDate)
			FP.ResultsTable.SetCell(i+1, 1, rBalance)
			FP.ResultsTable.SetCell(i+1, 2, rCumulativeIncome)
			FP.ResultsTable.SetCell(i+1, 3, rCumulativeExpenses)
			FP.ResultsTable.SetCell(i+1, 4, rDayExpenses)
			FP.ResultsTable.SetCell(i+1, 5, rDayIncome)
			FP.ResultsTable.SetCell(i+1, 6, rDayNet)
			FP.ResultsTable.SetCell(i+1, 7, rDiffFromStart)
			FP.ResultsTable.SetCell(i+1, 8, rDayTransactionNames)
		}

		FP.ResultsTable.SetSelectionChangedFunc(func(row, column int) {
			if row <= 0 {
				return
			}
			FP.ResultsDescription.Clear()
			// ensure there are enough results before trying to show something
			if len(*(FP.LatestResults))-1 > row-1 {
				var sb strings.Builder
				for _, t := range (*(FP.LatestResults))[row-1].DayTransactionNamesSlice {
					sb.WriteString(fmt.Sprintf("%v\n", t))
				}
				resultsStatus(sb.String())
			}
		})

		getResultsStats()

		FP.CalculatingResults = false

		FP.App.SetFocus(FP.ResultsTable)
	}()
}
