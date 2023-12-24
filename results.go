package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"
	"finance-planner-tui/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func resultsFormInputFieldYearValidator(textToCheck string, _ rune) bool {
	i, err := strconv.ParseInt(textToCheck, 10, 64)
	if err != nil || i < 0 {
		return false
	}

	return true
}

func resultsFormInputFieldMonthValidator(textToCheck string, _ rune) bool {
	i, err := strconv.ParseInt(textToCheck, 10, 64)
	if err != nil || i < 0 || i > 12 {
		return false
	}

	return true
}

func resultsFormInputFieldDayValidator(textToCheck string, _ rune) bool {
	i, err := strconv.ParseInt(textToCheck, 10, 64)
	if err != nil || i < 0 || i > 31 {
		return false
	}

	return true
}

func resultsFormInputFieldStartYearChanged(text string) {
	FP.SelectedProfile.StartYear = text
}

func getResultsFormLabel(m string) string {
	return fmt.Sprintf("%v:", m)
}

func resultsFormSubmit1Yr() {
	setResultsFormPreset(c.StartTodayPreset, c.OneYear)
	updateResultsForm()
	getResultsTable()
}

func resultsFormSubmit5Yr() {
	setResultsFormPreset(c.StartTodayPreset, c.FiveYear)
	updateResultsForm()
	getResultsTable()
}

// Completely rebuilds the results form, safe to run repeatedly.
func updateResultsForm() {
	FP.ResultsForm.Clear(true)
	FP.ResultsForm.SetTitle("Parameters")

	if FP.SelectedProfile == nil {
		return
	}

	setSelectedProfileDefaults()

	FP.ResultsForm.
		AddInputField(getResultsFormLabel(FP.T["ResultsFormStartYearLabel"]),
			FP.SelectedProfile.StartYear,
			0, resultsFormInputFieldYearValidator,
			resultsFormInputFieldStartYearChanged).
		AddInputField(getResultsFormLabel(FP.T["ResultsFormStartMonthLabel"]),
			FP.SelectedProfile.StartMonth,
			0, resultsFormInputFieldMonthValidator,
			func(text string) { FP.SelectedProfile.StartMonth = text }).
		AddInputField(getResultsFormLabel(FP.T["ResultsFormStartDayLabel"]),
			FP.SelectedProfile.StartDay,
			0, resultsFormInputFieldDayValidator,
			func(text string) { FP.SelectedProfile.StartDay = text }).
		AddInputField(getResultsFormLabel(FP.T["ResultsFormEndYearLabel"]),
			FP.SelectedProfile.EndYear,
			0, resultsFormInputFieldYearValidator,
			func(text string) { FP.SelectedProfile.EndYear = text }).
		AddInputField(getResultsFormLabel(FP.T["ResultsFormEndMonthLabel"]),
			FP.SelectedProfile.EndMonth,
			0, resultsFormInputFieldMonthValidator,
			func(text string) { FP.SelectedProfile.EndMonth = text }).
		AddInputField(getResultsFormLabel(FP.T["ResultsFormEndDayLabel"]),
			FP.SelectedProfile.EndDay,
			0, resultsFormInputFieldDayValidator,
			func(text string) { FP.SelectedProfile.EndDay = text }).
		AddInputField(getResultsFormLabel(FP.T["ResultsFormStartingBalanceLabel"]),
			FP.SelectedProfile.StartingBalance,
			0, nil,
			func(text string) {
				FP.SelectedProfile.StartingBalance = lib.FormatAsCurrency(int(lib.ParseDollarAmount(text, true)))
			}).
		AddButton(FP.T["ResultsFormSubmitButtonLabel"], getResultsTable).
		AddButton(FP.T["ResultsForm1yearButtonLabel"], resultsFormSubmit1Yr).
		AddButton(FP.T["ResultsForm5yearsButtonLabel"], resultsFormSubmit5Yr).
		AddButton(FP.T["ResultsFormStatsButtonLabel"], getResultsStats)

	FP.ResultsForm.SetLabelColor(tcell.ColorViolet)
	FP.ResultsForm.SetFieldBackgroundColor(tcell.NewRGBColor(40, 40, 40))
	FP.ResultsForm.SetBorder(true)
}

func getResultsPage() *tview.Flex {
	FP.ResultsTable = tview.NewTable().SetFixed(1, 1)

	FP.ResultsForm = tview.NewForm()

	FP.ResultsTable.SetBorder(true)
	updateResultsForm()

	FP.ResultsTable.SetTitle(FP.T["ResultsTableTitle"])
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
// TODO: implement other start date logic - currently only supports today.
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
// and queues an UpdateDraw.
func getResultsStats() {
	go FP.App.QueueUpdateDraw(func() {
		if FP.LatestResults == nil {
			return
		}

		stats, err := lib.GetStats(*(FP.LatestResults))
		if err != nil {
			FP.ResultsDescription.SetText(fmt.Sprintf(
				"%v%v: %v%v",
				FP.Colors["ResultsDescriptionError"],
				FP.T["ResultsStatsErrorGettingStats"],
				err.Error(),
				c.Reset,
			))
		}

		FP.ResultsDescription.SetText(fmt.Sprintf(
			"%v%v%v",
			FP.Colors["ResultsDescriptionStats"],
			stats,
			c.Reset,
		))
	})
}

// Returns a list, representing the ordered columns to be shown in
// the results table, alongside their configured colors.
func getResultsTableHeaders() []models.TableCell {
	return []models.TableCell{
		{Text: FP.T["ResultsColumnDate"], Color: FP.Colors["ResultsColumnDate"]},
		{Text: FP.T["ResultsColumnBalance"], Color: FP.Colors["ResultsColumnBalance"]},
		{Text: FP.T["ResultsColumnCumulativeIncome"], Color: FP.Colors["ResultsColumnCumulativeIncome"]},
		{Text: FP.T["ResultsColumnCumulativeExpenses"], Color: FP.Colors["ResultsColumnCumulativeExpenses"]},
		{Text: FP.T["ResultsColumnDayExpenses"], Color: FP.Colors["ResultsColumnDayExpenses"]},
		{Text: FP.T["ResultsColumnDayIncome"], Color: FP.Colors["ResultsColumnDayIncome"]},
		{Text: FP.T["ResultsColumnDayNet"], Color: FP.Colors["ResultsColumnDayNet"]},
		{Text: FP.T["ResultsColumnDiffFromStart"], Color: FP.Colors["ResultsColumnDiffFromStart"]},
		{Text: FP.T["ResultsColumnDayTransactionNames"], Color: FP.Colors["ResultsColumnDayTransactionNames"], Expand: 1},
	}
}

// Returns a list, representing the ordered columns to be shown in
// the results table, alongside their configured colors.
func getResultsTableCell(r lib.Result) []models.TableCell {
	return []models.TableCell{
		{Text: lib.FormatAsDate(r.Date), Color: FP.Colors["ResultsColumnDate"]},
		{Text: lib.FormatAsCurrency(r.Balance), Color: FP.Colors["ResultsColumnBalance"]},
		{Text: lib.FormatAsCurrency(r.CumulativeIncome), Color: FP.Colors["ResultsColumnCumulativeIncome"]},
		{Text: lib.FormatAsCurrency(r.CumulativeExpenses), Color: FP.Colors["ResultsColumnCumulativeExpenses"]},
		{Text: lib.FormatAsCurrency(r.DayExpenses), Color: FP.Colors["ResultsColumnDayExpenses"]},
		{Text: lib.FormatAsCurrency(r.DayIncome), Color: FP.Colors["ResultsColumnDayIncome"]},
		{Text: lib.FormatAsCurrency(r.DayNet), Color: FP.Colors["ResultsColumnDayNet"]},
		{Text: lib.FormatAsCurrency(r.DiffFromStart), Color: FP.Colors["ResultsColumnDiffFromStart"]},
		{Text: r.DayTransactionNames, Color: FP.Colors["ResultsColumnDayTransactionNames"], Expand: 1},
	}
}

// Constructs and sets the columns for the first row in the results table.
// Unsafe to run repeatedly and does not clear any existing fields/data.
func setResultsTableHeaders() {
	th := getResultsTableHeaders()

	for i := range th {
		cell := tview.NewTableCell(fmt.Sprintf("%v%v%v", th[i].Color, th[i].Text, c.Reset))
		if th[i].Expand > 0 {
			cell.SetExpansion(th[i].Expand)
		}

		FP.ResultsTable.SetCell(0, i, cell)
	}
}

// Constructs and sets the columns for the i'th row in the results table.
// Unsafe to run repeatedly and does not clear any existing fields/data.
func setResultsTableCellsForResult(i int, r lib.Result) {
	td := getResultsTableCell(r)

	for j := range td {
		cell := tview.NewTableCell(fmt.Sprintf("%v%v%v", td[j].Color, td[j].Text, c.Reset))
		if td[j].Expand > 0 {
			cell.SetExpansion(td[j].Expand)
		}

		FP.ResultsTable.SetCell(i, j, cell)
	}
}

// Executes a goroutine to asynchronously update the results table. Will do
// nothing if a goroutine has already been started.
func getResultsTable() {
	if FP.CalculatingResults {
		return
	}

	FP.CalculatingResults = true

	go func() {
		FP.ResultsTable.Clear()
		FP.ResultsDescription.Clear()

		FP.ResultsDescription.SetText(fmt.Sprintf("%v%v%v",
			FP.Colors["ResultsDescriptionPassive"],
			FP.T["ResultsTableStatusCalculatingPleaseWait"],
			c.Reset,
		))

		setSelectedProfileDefaults()

		bal := int(lib.ParseDollarAmount(FP.SelectedProfile.StartingBalance, true))

		st := lib.GetDateString(
			FP.SelectedProfile.StartYear,
			FP.SelectedProfile.StartMonth,
			FP.SelectedProfile.StartDay,
		)
		end := lib.GetDateString(
			FP.SelectedProfile.EndYear,
			FP.SelectedProfile.EndMonth,
			FP.SelectedProfile.EndDay,
		)

		statusHook := func(status string) {
			if FP.Config.DisableResultsStatusMessages || FP.ResultsDescription == nil {
				return
			}

			go FP.App.QueueUpdateDraw(func() {
				FP.ResultsDescription.SetText(fmt.Sprintf("%v%v%v",
					FP.Colors["ResultsDescriptionPassive"],
					status,
					c.Reset,
				))
			})
		}

		// get results
		results, err := lib.GenerateResultsFromDateStrings(&(FP.SelectedProfile.TX), bal, st, end, statusHook)
		if err != nil {
			FP.ResultsDescription.SetText(fmt.Sprintf("%v%v: %v%v",
				FP.Colors["ResultsDescriptionError"],
				FP.T["ResultsGenerationFailed"],
				err.Error(),
				c.Reset,
			))

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
		setResultsTableHeaders()

		// now add the remaining rows
		for i := range results {
			setResultsTableCellsForResult(i+1, results[i])
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
				FP.ResultsDescription.SetText(fmt.Sprintf("%v%v%v",
					FP.Colors["ResultsDescriptionPassive"],
					sb.String(),
					c.Reset,
				))
			}
		})

		getResultsStats()

		FP.CalculatingResults = false

		FP.App.SetFocus(FP.ResultsTable)
	}()
}
