package lib

import (
	"encoding/csv"
	"fmt"
	"log"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	c "finance-planner-tui/constants"

	"github.com/google/uuid"
	"github.com/teambition/rrule-go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type TX struct { // transaction
	// Order  int    `yaml:"order"`  // manual ordering
	Amount int    `yaml:"amount"` // in cents; 500 = $5.00
	Active bool   `yaml:"active"`
	Name   string `yaml:"name"`
	Note   string `yaml:"note"`
	// for examples of rrules:
	// https://github.com/teambition/rrule-go/blob/f71921a2b0a18e6e73c74dea155f3a549d71006d/rrule.go#L91
	// https://github.com/teambition/rrule-go/blob/master/rruleset_test.go
	// https://labix.org/python-dateutil/#head-88ab2bc809145fcf75c074817911575616ce7caf
	RRule string `yaml:"rrule"`
	// for when users don't want to use the rrules:
	Frequency   string    `yaml:"frequency"`
	Interval    int       `yaml:"interval"`
	Weekdays    []int     `yaml:"weekdays"` // monday starts on 0
	StartsDay   int       `yaml:"startsDay"`
	StartsMonth int       `yaml:"startsMonth"`
	StartsYear  int       `yaml:"startsYear"`
	EndsDay     int       `yaml:"endsDay"`
	EndsMonth   int       `yaml:"endsMonth"`
	EndsYear    int       `yaml:"endsYear"`
	ID          string    `yaml:"id"`
	CreatedAt   time.Time `yaml:"createdAt"`
	UpdatedAt   time.Time `yaml:"updatedAt"`
	Selected    bool      `yaml:"selected"` // when activated in the transactions table
}

// FormatAsDate takes an input time and formats it using the standard
// representation of a date in this application: "YYYY-MM-DD" (may not have
// padded zeroes).
func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()

	return fmt.Sprintf("%02d/%02d/%d", month, day, year)
}

// FormatAsCurrency converts an integer to a USD-formatted string. Input
// is assumed to be based in pennies, i.e., hundredths of a dollar - 100 would
// return "$1.00".
func FormatAsCurrency(a int) string {
	// convert to float and dump as currency string
	// TODO: print the integer and clip the last two digits instead of
	// using floats
	amt := float64(a)
	amt /= 100
	p := message.NewPrinter(language.English)

	return p.Sprintf("$%.2f", amt)
}

type PreCalculatedResult struct {
	Date                  time.Time
	DayTransactionNames   []string
	DayTransactionAmounts []int
}

type Result struct { // csv/table output row
	Record                   int
	Date                     time.Time
	Balance                  int
	CumulativeIncome         int
	CumulativeExpenses       int
	DayExpenses              int
	DayIncome                int
	DayNet                   int
	DayTransactionNames      string
	DiffFromStart            int
	DayTransactionNamesSlice []string
	ID                       string
	CreatedAt                string
	UpdatedAt                string
}

// GetNewTX returns an empty transaction with sensible defaults.
func GetNewTX() TX {
	now := time.Now()
	oneMonth := now.Add(time.Hour * c.HoursInDay * c.DaysInMonth)

	return TX{
		// Order:       0,
		Amount:      c.DefaultTransactionBalance,
		Active:      true,
		Name:        c.New,
		Frequency:   c.MONTHLY,
		Interval:    1,
		StartsDay:   now.Day(),
		StartsMonth: int(now.Month()),
		StartsYear:  now.Year(),
		EndsDay:     oneMonth.Day(),
		EndsMonth:   int(oneMonth.Month()),
		EndsYear:    oneMonth.Year(),
		ID:          uuid.NewString(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Note:        "",
		RRule:       "",
		Weekdays:    []int{},
		Selected:    false,
	}
}

// GetWeekdaysMap returns a map that can be used like this:
//
// m := GetWeekdaysMap()
//
// if m[rrule.MO.Day()] { /* do something * / }
//
// It is meant to be more efficient than repeatedly using tx.HasWeekday()
// to determine if a weekday is present in a given TX.
func (tx *TX) GetWeekdaysMap() map[int]bool {
	m := make(map[int]bool)
	for i := 0; i < 7; i++ {
		m[i] = false
	}

	for i := range tx.Weekdays {
		m[tx.Weekdays[i]] = true
	}

	return m
}

// GetWeekdaysCheckedMap returns a map that can be used like this:
//
// checkedGlyph := "X"
// uncheckedGlyph := " "
// m := GetWeekdaysCheckedMap(checkedGlyph)
//
// log.Println("occurs on mondays: %v", m[rrule.MO.Day()])
//
// It is meant to be more efficient than repeatedly using tx.HasWeekday()
// to determine if a weekday is present in a given TX.
func (tx *TX) GetWeekdaysCheckedMap(checked, unchecked string) map[int]string {
	m := make(map[int]string)
	for i := 0; i < 7; i++ {
		m[i] = unchecked
	}

	for i := range tx.Weekdays {
		m[tx.Weekdays[i]] = checked
	}

	return m
}

// HasWeekday checks if a recurring transaction definition contains
// the specified weekday as an rrule recurrence day of the week.
func (tx *TX) HasWeekday(weekday int) bool {
	for _, d := range tx.Weekdays {
		if weekday == d {
			return true
		}
	}

	return false
}

// MarkupText will italicize and gray-out the provide input string value if
// the value of tx.Active is false. Otherwise, it will simply return the input
// string value unaltered. In the future this may change, as this function may
// handle multiple different situations, depending on the state of tx.
func (tx *TX) MarkupText(input string) string {
	input = strings.ReplaceAll(input, "&", "&amp;")
	if !tx.Active {
		return fmt.Sprintf(`<i><span foreground="#AAAAAA">%v</span></i>`, input)
	}

	return input
}

// preserves the color of active currency values but italicizes values
// according to enabled/disabled
//
// TODO: refactor/improve this, it doesn't really work as intended but I'm
// lazy at the moment.
func (tx *TX) MarkupCurrency(input string) string {
	input = strings.ReplaceAll(input, "&", "&amp;")
	if !tx.Active {
		return fmt.Sprintf(`<i><span foreground="#CCCCCC">%v</span></i>`, input)
	}

	return input
}

func ToggleDayFromWeekdays(weekdays []int, weekday int) []int {
	if weekday < 0 || weekday > 6 {
		return weekdays
	}

	foundWeekday := false
	returnValue := []int{}

	for i := range weekdays {
		if weekdays[i] == weekday {
			foundWeekday = true
		} else {
			returnValue = append(returnValue, weekdays[i])
		}
	}

	if !foundWeekday {
		returnValue = append(returnValue, weekday)
	}

	sort.Ints(returnValue)

	return returnValue
}

func GetResults(tx []TX, startDate time.Time, endDate time.Time, startBalance int, statusHook func(status string)) ([]Result, error) {
	if startDate.After(endDate) {
		return []Result{}, fmt.Errorf("start date is after end date: %v vs %v", startDate, endDate)
	}

	// start by quickly generating an index of every single date from startDate to endDate
	dates := make(map[int64]Result)
	preCalculatedDates := make(map[int64]PreCalculatedResult)

	r, err := rrule.NewRRule(
		rrule.ROption{
			Freq:    rrule.DAILY,
			Dtstart: startDate,
			Until:   endDate,
		},
	)
	if err != nil {
		return []Result{}, fmt.Errorf("failed to construct rrule for results date window: %v", err.Error())
	}

	allDates := r.All()

	statusHook("preparing dates...")

	for i, dt := range allDates {
		dtInt := dt.Unix()
		dates[dtInt] = Result{
			Record: i,
			Date:   dt,
		}
		preCalculatedDates[dtInt] = PreCalculatedResult{
			Date: dt,
		}
	}

	emptyDate := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)

	// iterate over every TX definition, starting with its start date
	txLen := len(tx)

	statusHook(fmt.Sprintf("recurrences... [%v/%v]", 0, txLen))

	for i, txi := range tx {
		if !txi.Active {
			continue
		}

		if i%1000 == 0 {
			// to avoid unnecessary slowdown, only update every 1000 iterations
			statusHook(fmt.Sprintf("recurrences... [%v/%v]", i+1, txLen))
		}

		var allOccurrences []time.Time

		if txi.RRule != "" {
			s, err := rrule.StrToRRuleSet(txi.RRule)
			if err != nil {
				return []Result{}, fmt.Errorf(
					"failed to process rrule for tx %v: %v",
					txi.Name,
					err.Error(),
				)
			}

			allOccurrences = s.Between(
				startDate,
				endDate,
				true,
			)
		} else {
			txiStartsDate := time.Date(txi.StartsYear, time.Month(txi.StartsMonth), txi.StartsDay, 0, 0, 0, 0, time.UTC)
			txiEndsDate := time.Date(txi.EndsYear, time.Month(txi.EndsMonth), txi.EndsDay, 0, 0, 0, 0, time.UTC)
			// input validation: if the end date for the transaction definition is after
			// the final end date, then just use the ending date.
			// also, if the transaction definition's end date is unset (equal to emptyDate),
			// then default to the ending date as well
			if txiEndsDate.After(endDate) || txiEndsDate == emptyDate {
				txiEndsDate = endDate
			}
			// input validation: if the transaction definition's start date is
			// unset (equal to emptyDate), then default to the start date
			if txiStartsDate == emptyDate {
				txiStartsDate = startDate
			}
			// convert the user input frequency to a value that rrule lib
			// will accept
			freq := rrule.DAILY
			if txi.Frequency == rrule.YEARLY.String() {
				freq = rrule.YEARLY
			} else if txi.Frequency == rrule.MONTHLY.String() {
				freq = rrule.MONTHLY
			}
			// convert the user input weekdays into a value that rrule lib will
			// accept
			weekdays := []rrule.Weekday{}
			for _, weekday := range txi.Weekdays {
				switch weekday {
				case rrule.MO.Day():
					weekdays = append(weekdays, rrule.MO)
				case rrule.TU.Day():
					weekdays = append(weekdays, rrule.TU)
				case rrule.WE.Day():
					weekdays = append(weekdays, rrule.WE)
				case rrule.TH.Day():
					weekdays = append(weekdays, rrule.TH)
				case rrule.FR.Day():
					weekdays = append(weekdays, rrule.FR)
				case rrule.SA.Day():
					weekdays = append(weekdays, rrule.SA)
				case rrule.SU.Day():
					weekdays = append(weekdays, rrule.SU)
				default:
					break
				}
			}
			// create the rule based on the input parameters from the user
			s, err := rrule.NewRRule(rrule.ROption{
				Freq:      freq,
				Interval:  txi.Interval,
				Dtstart:   txiStartsDate,
				Until:     txiEndsDate,
				Byweekday: weekdays,
			})
			if err != nil {
				return []Result{}, fmt.Errorf(
					"failed to construct rrule for tx %v: %v",
					txi.Name,
					err.Error(),
				)
			}

			allOccurrences = s.Between(startDate, endDate, true)
		}

		for _, dt := range allOccurrences {
			dtInt := dt.Unix()
			newResult := preCalculatedDates[dtInt]
			newResult.Date = dt
			newResult.DayTransactionAmounts = append(newResult.DayTransactionAmounts, txi.Amount)
			newResult.DayTransactionNames = append(newResult.DayTransactionNames, txi.Name)
			preCalculatedDates[dtInt] = newResult
		}
	}

	results := []Result{}
	for _, result := range dates {
		results = append(results, result)
	}

	resultsLen := len(results)
	statusHook(fmt.Sprintf("sorting dates... [%v]", resultsLen))
	sort.SliceStable(
		results,
		func(i, j int) bool {
			return results[j].Date.After(results[i].Date)
		},
	)

	// now that it's sorted, we can roll out the calculations
	currentBalance := startBalance
	diff := 0
	cumulativeIncome := 0
	cumulativeExpenses := 0

	statusHook(fmt.Sprintf("calculating... [%v/%v]", 0, resultsLen))

	for i := range results {
		if i%1000 == 0 {
			// to avoid unnecessary slowdown, only update every 1000 iterations
			statusHook(fmt.Sprintf("calculating... [%v/%v]", i+1, resultsLen))
		}

		resultsDateInt := results[i].Date.Unix()
		numDayTransactionAmounts := len(preCalculatedDates[resultsDateInt].DayTransactionAmounts)
		numDdayTransactionNames := len(preCalculatedDates[resultsDateInt].DayTransactionNames)

		// if for some reason not all transaction names and amounts match up,
		// exit now
		if numDayTransactionAmounts != numDdayTransactionNames {
			return results, fmt.Errorf(
				"there was a different number of transaction amounts versus transaction names for date %v",
				resultsDateInt,
			)
		}

		for j := range preCalculatedDates[resultsDateInt].DayTransactionAmounts {
			// determine if the amount is an expense or income
			amt := preCalculatedDates[resultsDateInt].DayTransactionAmounts[j]
			if amt >= 0 {
				results[i].DayIncome += amt
				cumulativeIncome += amt
			} else {
				results[i].DayExpenses += amt
				cumulativeExpenses += amt
			}

			// basically just doing a join on a slice of strings, should
			// use the proper method for this in the future
			name := preCalculatedDates[resultsDateInt].DayTransactionNames[j]
			if results[i].DayTransactionNames == "" {
				results[i].DayTransactionNames = name
			} else {
				results[i].DayTransactionNames += fmt.Sprintf("; %v", name)
			}

			results[i].DayTransactionNamesSlice = append(results[i].DayTransactionNamesSlice, name)

			results[i].DayNet += amt
			diff += amt
			currentBalance += amt
		}

		results[i].Balance = currentBalance
		results[i].CumulativeIncome = cumulativeIncome
		results[i].CumulativeExpenses = cumulativeExpenses
		results[i].DiffFromStart = diff
	}

	statusHook(fmt.Sprintf("done [%v/%v]", resultsLen, resultsLen))

	return results, nil
}

// GetNowDateString returns a string corresponding to the current YYYY-MM-DD
// value, but does not necessarily include 0-padded values.
func GetNowDateString() string {
	now := time.Now()

	return fmt.Sprintf("%04v-%02v-%02v", now.Year(), int(now.Month()), now.Day())
}

// GetDefaultEndDateString returns a string corresponding to the current YYYY-MM-DD
// value plus 1 year in the future, but does not necessarily include 0-padded
// values.
func GetDefaultEndDateString() string {
	now := time.Now()

	return fmt.Sprintf("%04v-%02v-%02v", now.Year()+1, int(now.Month()), now.Day())
}

// GetDateString formats a string as YYYY-MM-DD with zero-padding.
func GetDateString(y, m, d any) string {
	return fmt.Sprintf("%04v-%02v-%02v", y, m, d)
}

// GetStartDateString returns a formatted date string for the transaction's
// start date.
func (tx *TX) GetStartDateString() string {
	return GetDateString(tx.StartsYear, tx.StartsMonth, tx.StartsDay)
}

// GetEndsDateString returns a formatted date string for the transaction's end
// date.
func (tx *TX) GetEndsDateString() string {
	return GetDateString(tx.EndsYear, tx.EndsMonth, tx.EndsDay)
}

// ParseYearMonthDateString takes an input value such as 2020-01-01 and returns
// three integer values - year, month, day. Returns 0, 0, 0 if invalid input
// is received.
func ParseYearMonthDateString(input string) (int, int, int) {
	vals := strings.Split(input, "-")
	if len(vals) != 3 {
		return 0, 0, 0
	}

	yr, _ := strconv.ParseInt(vals[0], 10, 64)
	mo, _ := strconv.ParseInt(vals[1], 10, 64)
	day, _ := strconv.ParseInt(vals[2], 10, 64)

	return int(yr), int(mo), int(day)
}

// ParseDollarAmount takes an input currency-formatted string, such as $100.00,
// and returns an integer corresponding to the underlying value, such as 10000.
// Generally in this application, values are assumed to be negative (i.e.
// recurring bills), so if assumePositive is set to true, the returned value
// will be positive, but otherwise it will default to negative.
func ParseDollarAmount(input string, assumePositive bool) int64 {
	cents := int64(0)
	multiplier := int64(-1)
	r := regexp.MustCompile(`[^\d.]*`)
	s := r.ReplaceAllString(input, "")

	// all values are assumed negative, unless it starts with a + character
	if strings.Index(input, "+") == 0 || strings.Index(input, "$+") == 0 || assumePositive {
		multiplier = int64(1)
	}

	// in the event that the user is entering the starting balance,
	// they may want to set a negative starting balance. So basically just the
	// reverse from above logic, since the user will have to be typing a
	// negative sign in front.
	if assumePositive && (strings.Index(input, "$-") == 0 || strings.Index(input, "-") == 0) {
		multiplier = int64(-1)
	}

	// check if the user entered a period
	ss := strings.Split(s, ".")

	if len(ss) == 2 {
		cents, _ = strconv.ParseInt(ss[1], 10, 64)
		// if the user types e.g. 10.2, they meant $10.20
		// but not if the value started with a 0
		if strings.Index(ss[1], "0") != 0 && cents < 10 {
			cents *= 10
		}
		// if they put in too many numbers, zero it out
		if cents >= 100 {
			cents = 0
		}
	}

	var whole int64
	whole, _ = strconv.ParseInt(ss[0], 10, 64)

	// account for the negative case when re-combining the two values
	if whole < 0 {
		return multiplier * (whole*100 - cents)
	}

	return multiplier * (whole*100 + cents)
}

// RemoveTXAtIndex is a quick helper function to remove a transaction from
// a slice. There are more generic ways to do this, and it's fairly trivial,
// but it's nice to have a dedicated helper function for it.
func RemoveTXAtIndex(txs []TX, i int) []TX {
	return append(txs[:i], txs[i+1:]...)
}

// RemoveTXByID manipulates an input TX slice by removing a TX with the provided
// id.
func RemoveTXByID(txs *[]TX, id string) {
	for i := range *txs {
		tx := (*txs)[i]

		if tx.ID != id {
			continue
		}

		*txs = RemoveTXAtIndex(*txs, i)

		break
	}
}

// GetTXByID finds the index of a TX for the provided id, returning an error
// and -1 if not present.
func GetTXByID(txs *[]TX, id string) (int, error) {
	for i := range *txs {
		tx := (*txs)[i]

		if tx.ID != id {
			continue
		}

		return i, nil
	}

	return -1, fmt.Errorf("not present")
}

// GenerateResultsFromDateStrings takes an input start and end date (either can
// be the default '0-0-0' values, in which case it uses today for the start,
// and a year from now for the end), and calculates all of the calculable
// transactions for the provided range.
func GenerateResultsFromDateStrings(
	txs *[]TX,
	bal int,
	startDt string,
	endDt string,
	statusHook func(status string),
) ([]Result, error) {
	now := time.Now()
	stYr, stMo, stDay := ParseYearMonthDateString(startDt)
	endYr, endMo, endDay := ParseYearMonthDateString(endDt)

	if startDt == "0-0-0" || startDt == "--" || startDt == "" {
		stYr = now.Year()
		stMo = int(now.Month())
		stDay = now.Day()
	}

	if endDt == "0-0-0" || endDt == "--" || endDt == "" {
		endYr = now.Year() + 1
		endMo = int(now.Month())
		endDay = now.Day()
	}

	res, err := GetResults(
		*txs,
		time.Date(stYr, time.Month(stMo), stDay, 0, 0, 0, 0, time.UTC),
		time.Date(endYr, time.Month(endMo), endDay, 0, 0, 0, 0, time.UTC),
		bal,
		statusHook,
	)
	if err != nil {
		return []Result{}, fmt.Errorf("failed to get results: %v", err.Error())
	}

	return res, nil
}

// GetStats spits out some quick calculations about the provided set of results.
// Calculations include, for example, yearly+monthly+daily income/expenses, as
// well as some other things. Users may want to copy this information to the
// clipboard.
func GetStats(results []Result) (string, error) {
	count := len(results)
	i := 365

	if count > i {
		b := new(strings.Builder)
		b.WriteString("Here are some statistics about your finances.\n\n")

		dailySpendingAvg := results[i].CumulativeExpenses / i
		dailyIncomeAvg := results[i].CumulativeIncome / i

		b.WriteString(fmt.Sprintf(
			"Daily spending: %v\nDaily income: %v\nDaily net: %v",
			FormatAsCurrency(dailySpendingAvg),
			FormatAsCurrency(dailyIncomeAvg),
			FormatAsCurrency(dailySpendingAvg+dailyIncomeAvg),
		))

		moSpendingAvg := results[i].CumulativeExpenses / 12
		moIncomeAvg := results[i].CumulativeIncome / 12

		b.WriteString(fmt.Sprintf(
			"\nMonthly spending: %v\nMonthly income: %v\nMonthly net: %v",
			FormatAsCurrency(moSpendingAvg),
			FormatAsCurrency(moIncomeAvg),
			FormatAsCurrency(moSpendingAvg+moIncomeAvg),
		))

		yrSpendingAvg := results[i].CumulativeExpenses
		yrIncomeAvg := results[i].CumulativeIncome

		b.WriteString(fmt.Sprintf(
			"\nYearly spending: %v\nYearly income: %v\nYearly net: %v",
			FormatAsCurrency(yrSpendingAvg),
			FormatAsCurrency(yrIncomeAvg),
			FormatAsCurrency(yrSpendingAvg+yrIncomeAvg),
		))

		return b.String(), nil
	}

	return "", fmt.Errorf(
		"You need at least one year between your start date and end date to get statistics about your finances.",
	)
}

func GetResultsCSVString(results *[]Result) string {
	b := new(strings.Builder)
	w := csv.NewWriter(b)

	for _, r := range *results {
		var record []string
		record = append(record, FormatAsDate(r.Date))
		record = append(record, FormatAsCurrency(r.Balance))
		record = append(record, FormatAsCurrency(r.CumulativeIncome))
		record = append(record, FormatAsCurrency(r.CumulativeExpenses))
		record = append(record, FormatAsCurrency(r.DayExpenses))
		record = append(record, FormatAsCurrency(r.DayIncome))
		record = append(record, FormatAsCurrency(r.DayNet))
		record = append(record, FormatAsCurrency(r.DiffFromStart))
		record = append(record, r.DayTransactionNames)
		_ = w.Write(record)
	}

	w.Flush()

	return b.String()
}

func GetUser() *user.User {
	user, err := user.Current()
	if err != nil {
		log.Printf("failed to get the user's home directory: %v", err.Error())
	}

	return user
}

// GetCSVString produces a simple semi-colon-separated value string.
func GetCSVString(input []string) string {
	result := new(strings.Builder)
	if len(input) > 0 {
		result.WriteString(fmt.Sprintf("(%v) ", len(input)))
	}

	for _, name := range input {
		result.WriteString(fmt.Sprintf(`%v; `, name))
	}

	return result.String()
}

// GetNextSort takes the current sort, which is typically something like
// OrderAsc, OrderDesc, or None, and attempts to do some basic string parsing
// to figure out what the next sort should be. The cycle is None -> Asc -> Desc.
// Note that if the `next` argument is a different column than the `current`
// argument (after stripping away Asc/Desc), the resulting sort will always be
// the `next` column with Asc ordering.
func GetNextSort(current, next string) string {
	if next == c.None {
		return c.None
	}

	if current == c.None {
		return fmt.Sprintf("%v%v", next, c.Asc)
	}

	base := strings.TrimSuffix(current, c.Desc)
	base = strings.TrimSuffix(base, c.Asc)

	if strings.HasSuffix(current, c.Desc) {
		if base != next {
			return fmt.Sprintf("%v%v", next, c.Asc)
		}

		return c.None
	}

	if strings.HasSuffix(current, c.Asc) {
		if base != next {
			return fmt.Sprintf("%v%v", next, c.Asc)
		}

		return fmt.Sprintf("%v%v", base, c.Desc)
	}

	return fmt.Sprintf("%v%v", next, c.Asc)
}

// ValidateTransactions asserts that every TX definition has a unique Order
// field.
// TODO: Stretch goal - assert that every TX definition is numerically ordered
// without any gaps. These gaps can occur when deleting/cloning/adding.
// One possible approach might be to just simply iterate through the list,
// starting from Order=1, and then alter any values that are not immediately
// accessible by iterating to the next integer.
// func ValidateTransactions(tx *[]TX) error {
// 	missingFirst := false
// 	hasDuplicate := false
// 	outOfSequence := false
// 	sequenceFixes := make(map[int]int)
// 	uniques := make(map[int]int)
// 	msg := new(strings.Builder)

// 	// start by sorting the TX by Order
// 	sort.SliceStable(*tx, func(i, j int) bool {
// 		return (*tx)[j].Order > (*tx)[i].Order
// 	})

// 	// iterate through the list and assert that i is incremental
// 	prev := -1
// 	for i, t := range *tx {
// 		// assert that the first TX has a value of Order=1
// 		if i == 0 {
// 			if t.Order != 1 {
// 				missingFirst = true
// 				msg.WriteString("First value did not have Order=1. ")
// 				(*tx)[i].Order = 1
// 			}
// 		}

// 		// assert that each TX sequentially follows the next
// 		if t.Order != prev+2 {
// 			sequenceFixes[i] = prev + 2
// 			outOfSequence = true
// 			msg.WriteString(
// 				fmt.Sprintf(
// 					"Index %v had out of sequence order=%v instead of %v. ",
// 					i,
// 					t.Order,
// 					sequenceFixes[i],
// 				),
// 			)
// 			(*tx)[i].Order = sequenceFixes[i]
// 		}

// 		// assert that there are no duplicate Order values
// 		uniques[t.Order] += 1
// 		if uniques[t.Order] > 1 {
// 			newOrder := i + 1
// 			hasDuplicate = true
// 			msg.WriteString(
// 				fmt.Sprintf(
// 					"Order=%v duplicated %v times, setting to %v. ",
// 					t.Order,
// 					uniques[t.Order],
// 					newOrder,
// 				),
// 			)
// 			(*tx)[i].Order = newOrder
// 		}

// 		prev += 1
// 	}

// 	if hasDuplicate || outOfSequence || missingFirst {
// 		return fmt.Errorf("%v", msg.String())
// 	}

// 	return nil
// }

// GetLargestOrder returns the highest "Order" present in a list of
// transactions.
// func GetLargestOrder(txs []TX) int {
// 	m := 0

// 	for i := range txs {
// 		if txs[i].Order > m {
// 			m = txs[i].Order
// 		}
// 	}

// 	return m
// }

// GetNowStr is a simple function that returns the current time in
// HH:MM:SS (24 hr) format.
func GetNowStr() string {
	return time.Now().Format("15:04:05")
}
