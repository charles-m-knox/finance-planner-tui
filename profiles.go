package main

import (
	"fmt"
	"strconv"
	"time"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"
	m "finance-planner-tui/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setStatusNoChanges() {
	FP.ProfileStatusText.SetText(fmt.Sprintf("[gray] %v", FP.T["ProfilesPageStatusTextNoChanges"]))
}

func getActiveProfileText(profile m.Profile) string {
	if FP.SelectedProfile != nil && FP.SelectedProfile.Name == profile.Name {
		return fmt.Sprintf("[white::bu]%v %v%v", profile.Name, FP.T["ProfilesPageProfileOpenMarker"], c.ResetStyle)
	}

	return profile.Name
}

// populateProfilesPage clears out the profile list and proceeds to populate it
// with the current profiles in the config, including handlers for changing
// the FP.SelectedProfile.
func populateProfilesPage() {
	FP.ProfileList.Clear()

	for i := range FP.Config.Profiles {
		profile := &(FP.Config.Profiles[i])
		FP.ProfileList.AddItem(getActiveProfileText(*profile), "", 0, func() {
			FP.SelectedProfile = profile
			populateProfilesPage()
			getTransactionsTable()
			FP.App.SetFocus(FP.TransactionsTable)
		})
	}
}

// returns a simple flex view with two columns:
// - a list of profiles (left side)
// - a quick summary of bills / stats for the highlighted profile (right side)
func getProfilesPage() *tview.Flex {
	FP.ProfileList = tview.NewList()
	FP.ProfileList.SetBorder(true)
	FP.ProfileList.ShowSecondaryText(false).
		SetSelectedBackgroundColor(tcell.NewRGBColor(50, 50, 50)).
		SetSelectedTextColor(tcell.ColorWhite).
		SetTitle(FP.T["ProfilesPageTitle"])

	FP.ProfileStatusText = tview.NewTextView()
	FP.ProfileStatusText.SetBorder(true)
	FP.ProfileStatusText.SetDynamicColors(true)
	setStatusNoChanges()

	profilesLeftSide := tview.NewFlex().SetDirection(tview.FlexRow)
	profilesLeftSide.AddItem(FP.ProfileList, 0, 1, true).
		AddItem(FP.ProfileStatusText, 3, 0, true)

	FP.TransactionsTable = tview.NewTable().SetFixed(1, 1)
	FP.TransactionsInputField = tview.NewInputField()

	FP.TransactionsTable.SetBorder(true)
	FP.TransactionsInputField.SetBorder(true)

	FP.TransactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	FP.TransactionsInputField.SetLabel(fmt.Sprintf("[gray] %v%v", FP.T["ProfilesPageInputFieldAppearsHere"], c.ResetStyle))

	populateProfilesPage()
	getTransactionsTable()

	transactionsPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(FP.TransactionsTable, 0, 1, false).
		AddItem(FP.TransactionsInputField, 3, 0, false)

	return tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(profilesLeftSide, 0, 1, true).
		AddItem(transactionsPage, 0, 10, false)
}

// sets sensible default values for the currently selected profile, if they are
// not defined. If there is no FP.SelectedProfile, this will do nothing
func setSelectedProfileDefaults() {
	if FP.SelectedProfile == nil {
		return
	}

	now := time.Now()
	yr := now.Add(time.Hour * 24 * 365)

	if FP.SelectedProfile.StartYear == "" {
		FP.SelectedProfile.StartYear = strconv.Itoa(now.Year())
	}

	if FP.SelectedProfile.StartMonth == "" {
		FP.SelectedProfile.StartMonth = strconv.Itoa(int(now.Month()))
	}

	if FP.SelectedProfile.StartDay == "" {
		FP.SelectedProfile.StartDay = strconv.Itoa(now.Day())
	}

	if FP.SelectedProfile.EndYear == "" {
		FP.SelectedProfile.EndYear = strconv.Itoa(yr.Year())
	}

	if FP.SelectedProfile.EndMonth == "" {
		FP.SelectedProfile.EndMonth = strconv.Itoa(int(yr.Month()))
	}

	if FP.SelectedProfile.EndDay == "" {
		FP.SelectedProfile.EndDay = strconv.Itoa(yr.Day())
	}

	if FP.SelectedProfile.StartingBalance == "" {
		FP.SelectedProfile.StartingBalance = lib.FormatAsCurrency(50000)
	}
}
