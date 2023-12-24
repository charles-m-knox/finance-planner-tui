package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"
	m "finance-planner-tui/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

func actionRedo(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.TransactionsTable:
			redo()
			return nil
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionUndo(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.TransactionsTable:
			undo()
			return nil
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionQuit() *tcell.EventKey {
	promptExit()
	return nil
}

func actionMove(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.TransactionsTable:
			// move all selected items to the currently selected row:
			// delete items, then re-add the items after the current
			// row, then highlight the correct row
			if FP.SortTX != c.None && FP.SortTX != "" {
				FP.ProfileStatusText.SetText(fmt.Sprintf("[orange]sort: %v", FP.SortTX))
				return nil
			}

			// but first, check if any items are selected at all
			anySelected := false

			for i := range FP.SelectedProfile.TX {
				if FP.SelectedProfile.TX[i].Selected {
					anySelected = true

					break
				}
			}

			if !anySelected {
				FP.ProfileStatusText.SetText("[gray]nothing to move")
				return nil
			}

			setTransactionsTableSort(c.None)

			// get the height & width of the transactions table
			cr, cc := FP.TransactionsTable.GetSelection()
			actual := cr - 1 // skip header

			// take note of the currently selected value (cannot be
			// a candidate for move/deletion since it is the target
			// for the move)
			txid := FP.SelectedProfile.TX[actual].ID

			// first delete the values from the slice and keep track of
			// them
			deleted := []lib.TX{}
			newTX := []lib.TX{}

			for i := range FP.SelectedProfile.TX {
				if FP.SelectedProfile.TX[i].ID == txid {
					// this is the target to move to
					FP.SelectedProfile.TX[i].Selected = false
					newTX = append(newTX, FP.SelectedProfile.TX[i])
				} else if FP.SelectedProfile.TX[i].Selected {
					FP.SelectedProfile.TX[i].Selected = true
					deleted = append(deleted, FP.SelectedProfile.TX[i])
				} else {
					FP.SelectedProfile.TX[i].Selected = false
					newTX = append(newTX, FP.SelectedProfile.TX[i])
				}
			}

			FP.SelectedProfile.TX = newTX

			// find the move target now that the slice has been shifted
			newPosition := 0

			for i := range FP.SelectedProfile.TX {
				if FP.SelectedProfile.TX[i].ID == txid {
					newPosition = i + 1

					break
				}
			}

			if newPosition >= len(FP.SelectedProfile.TX) {
				newPosition = len(FP.SelectedProfile.TX)
			} else if newPosition < 0 {
				newPosition = 0
			}

			FP.LastSelection = newPosition

			FP.SelectedProfile.TX = slices.Insert(FP.SelectedProfile.TX, newPosition, deleted...)

			modified()

			// re-render the table
			getTransactionsTable()

			// check that we aren't going to move the selection past the
			// final row
			newPosition++

			r := FP.TransactionsTable.GetRowCount()

			if newPosition >= r {
				newPosition = r - 1
			}

			FP.TransactionsTable.Select(newPosition, cc) // offset for headers
			FP.App.SetFocus(FP.TransactionsTable)
		default:
			FP.App.SetFocus(FP.ProfileList)
		}

		return nil
	case PageResults:
		return e
	default:
		return e
	}
}

func actionSelect(e *tcell.EventKey, multiSelecting bool) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.TransactionsTable:
			cr, cc := FP.TransactionsTable.GetSelection()
			// get the height & width of the transactions table
			actual := cr - 1 // skip header
			if multiSelecting {
				// shift modifier is used to extend the selection
				// from the previously selected index to the current
				newSelectionValue := false
				// start by finding the currently highlighted TX
				for i := range FP.SelectedProfile.TX {
					if i == actual {
						newSelectionValue = !FP.SelectedProfile.TX[i].Selected

						break
					}
				}

				if FP.LastSelection == -1 {
					FP.LastSelection = actual
				}

				// now that we've determined what the selection value
				// should be, proceed to apply it to every value from
				// FP.LastSelection to the current index
				for i := range FP.SelectedProfile.TX {
					// last=5, current=10, select from 5-10 => last < i < actual
					// last=10, current=3, select from 3-10 => last > i > actual
					shouldModify := (FP.LastSelection < i && i <= actual) || (FP.LastSelection > i && i >= actual)
					if shouldModify {
						FP.SelectedProfile.TX[i].Selected = newSelectionValue
					}
				}
			} else {
				for i := range FP.SelectedProfile.TX {
					if i == actual {
						FP.SelectedProfile.TX[i].Selected = !FP.SelectedProfile.TX[i].Selected

						break
					}
				}
			}

			FP.LastSelection = actual

			modified()
			getTransactionsTable()
			FP.TransactionsTable.Select(cr, cc)
			FP.App.SetFocus(FP.TransactionsTable)

			return e
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionDelete(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.TransactionsTable:
			// duplicate the current transaction
			// get the height & width of the transactions table
			cr, cc := FP.TransactionsTable.GetSelection()
			actual := cr - 1 // skip header

			for i := len(FP.SelectedProfile.TX) - 1; i >= 0; i-- {
				if FP.SelectedProfile.TX[i].Selected || i == actual {
					FP.SelectedProfile.TX = slices.Delete(FP.SelectedProfile.TX, i, i+1)
				}
			}

			getTransactionsTable()
			FP.TransactionsTable.Select(cr, cc)
			FP.App.SetFocus(FP.TransactionsTable)
		case FP.ProfileList:
			if len(FP.Config.Profiles) <= 1 {
				FP.ProfileStatusText.SetText("[gray] can't delete last profile")
				return nil
			}

			getPrompt := func() string {
				if FP.SelectedProfile == nil {
					return "no profile selected; please cancel this operation"
				}

				return fmt.Sprintf(
					"[gold::b]confirm deletion of profile %v by typing 'delete %v':%v",
					FP.SelectedProfile.Name,
					FP.SelectedProfile.Name,
					c.ResetStyle,
				)
			}

			FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEscape:
					// don't save the changes
					deactivateTransactionsInputField()
					return
				default:
					// validate that the name is unique
					value := FP.TransactionsInputField.GetText()
					if strings.Index(value, "delete ") != 0 {
						FP.TransactionsInputField.SetLabel(getPrompt())
						return
					}

					profileName := strings.TrimPrefix(value, "delete ")
					if profileName != FP.SelectedProfile.Name {
						FP.TransactionsInputField.SetLabel(getPrompt())
						return
					}

					// proceed to delete the profile
					for i := range FP.Config.Profiles {
						if profileName == FP.Config.Profiles[i].Name {
							FP.Config.Profiles = slices.Delete(FP.Config.Profiles, i, i+1)
							return
						}
					}

					FP.SelectedProfile = &(FP.Config.Profiles[0])

					// config.Profiles = append(config.Profiles, newProfile)
					modified()
					deactivateTransactionsInputField()
					populateProfilesPage()
					getTransactionsTable()
					FP.TransactionsTable.Select(0, 0)
					FP.App.SetFocus(FP.ProfileList)
				}
			})
			activateTransactionsInputField(getPrompt(), "")
		default:
			FP.App.SetFocus(FP.ProfileList)
		}

		return nil
	case PageResults:
		return e
	default:
		return e
	}
}

func actionAdd(e *tcell.EventKey, duplicating bool) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.TransactionsTable:
			cr, cc := FP.TransactionsTable.GetSelection()
			actual := cr - 1 // skip header
			nt := []lib.TX{}

			FP.LastSelection = -1

			if !duplicating {
				// largestOrderHolder := []lib.TX{}
				// largestOrderHolder = append(largestOrderHolder, FP.SelectedProfile.TX...)
				// largestOrderHolder = append(largestOrderHolder, nt...)
				newTX := lib.GetNewTX()
				// newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1
				nt = append(nt, newTX)
			} else {
				// iterate through the list once to find how many selected
				// items there are
				numSelected := 0
				for i := range FP.SelectedProfile.TX {
					if FP.SelectedProfile.TX[i].Selected {
						numSelected++

						// we only care about knowing whether or not there
						// is more than 1 item selected
						if numSelected > 1 {
							break
						}
					}
				}

				for i := range FP.SelectedProfile.TX {
					isHighlightedRow := i == actual && numSelected <= 1
					isSelectedDuplicationCandidate := FP.SelectedProfile.TX[i].Selected && duplicating
					if isHighlightedRow || isSelectedDuplicationCandidate {
						// keep track of the highest order in a temporary
						// slice
						// largestOrderHolder := []lib.TX{}
						// largestOrderHolder = append(largestOrderHolder, FP.SelectedProfile.TX...)
						// largestOrderHolder = append(largestOrderHolder, nt...)

						newTX := lib.GetNewTX()
						// newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1

						newTX.Amount = FP.SelectedProfile.TX[i].Amount
						newTX.Active = FP.SelectedProfile.TX[i].Active
						newTX.Name = FP.SelectedProfile.TX[i].Name
						newTX.Note = FP.SelectedProfile.TX[i].Note
						newTX.RRule = FP.SelectedProfile.TX[i].RRule
						newTX.Frequency = FP.SelectedProfile.TX[i].Frequency
						newTX.Interval = FP.SelectedProfile.TX[i].Interval
						newTX.Weekdays = FP.SelectedProfile.TX[i].Weekdays
						newTX.StartsDay = FP.SelectedProfile.TX[i].StartsDay
						newTX.StartsMonth = FP.SelectedProfile.TX[i].StartsMonth
						newTX.StartsYear = FP.SelectedProfile.TX[i].StartsYear
						newTX.EndsDay = FP.SelectedProfile.TX[i].EndsDay
						newTX.EndsMonth = FP.SelectedProfile.TX[i].EndsMonth
						newTX.EndsYear = FP.SelectedProfile.TX[i].EndsYear

						nt = append(nt, newTX)
					}
				}
			}

			if len(nt) > 0 {
				// handles the case of adding/duplicating when the cursor
				// is on the headers row
				if actual < 0 {
					actual = 0
				}

				if len(FP.SelectedProfile.TX) == 0 || actual > len(FP.SelectedProfile.TX)-1 {
					FP.SelectedProfile.TX = append(FP.SelectedProfile.TX, nt...)
				} else {
					FP.SelectedProfile.TX = slices.Insert(FP.SelectedProfile.TX, actual, nt...)
				}

				modified()
				getTransactionsTable()
				FP.TransactionsTable.Select(cr, cc)
				FP.App.SetFocus(FP.TransactionsTable)
			}

			return e
		case FP.ProfileList:
			// add/duplicate new profile
			FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEscape:
					// don't save the changes
					deactivateTransactionsInputField()
					return
				default:
					// validate that the name is unique
					newProfileName := FP.TransactionsInputField.GetText()
					for i := range FP.Config.Profiles {
						if newProfileName == FP.Config.Profiles[i].Name {
							FP.TransactionsInputField.SetLabel("profile name must be unique:")
							return
						}
					}

					newProfile := *FP.SelectedProfile
					newProfile.Name = newProfileName
					if !duplicating {
						newProfile = m.Profile{Name: newProfileName}
					}

					FP.SelectedProfile = &newProfile

					FP.Config.Profiles = append(FP.Config.Profiles, newProfile)
					modified()
					deactivateTransactionsInputField()
					populateProfilesPage()
					getTransactionsTable()
					FP.TransactionsTable.Select(0, 0)
					FP.App.SetFocus(FP.ProfileList)
				}
			})
			activateTransactionsInputField("set new unique profile name:", "")

			return nil
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionEdit(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return e
		case FP.ProfileList:
			// add/duplicate new profile
			FP.TransactionsInputField.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEscape:
					// don't save the changes
					deactivateTransactionsInputField()
					return
				default:
					// validate that the name is unique
					newProfileName := FP.TransactionsInputField.GetText()
					for i := range FP.Config.Profiles {
						if newProfileName == FP.Config.Profiles[i].Name {
							FP.TransactionsInputField.SetLabel("profile name must be unique:")
							return
						}
					}

					FP.SelectedProfile.Name = newProfileName
					modified()
					deactivateTransactionsInputField()
					populateProfilesPage()
					getTransactionsTable()
					FP.TransactionsTable.Select(0, 0)
					FP.App.SetFocus(FP.ProfileList)
				}
			})
			activateTransactionsInputField(fmt.Sprintf("set new unique profile name for %v:", FP.SelectedProfile.Name), "")

			return nil
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionSave() *tcell.EventKey {
	if FP.Config.Version == "" {
		FP.Config.Version = c.ConfigVersion
	}

	b, err := yaml.Marshal(FP.Config)
	if err != nil {
		FP.ProfileStatusText.SetText("failed to marshal")
		return nil
	}

	err = os.WriteFile(FP.FlagConfigFile, b, os.FileMode(0o644))
	if err != nil {
		FP.ProfileStatusText.SetText("failed to save")
		return nil
	}

	FP.SelectedProfile.Modified = false

	FP.ProfileStatusText.SetText("[gray] saved changes")

	return nil
}

func actionEnd(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsTable:
			c := FP.TransactionsTable.GetColumnCount() - 1
			cr, _ := FP.TransactionsTable.GetSelection()
			FP.TransactionsTable.Select(cr, c)
			FP.App.SetFocus(FP.TransactionsTable)

			return nil
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionHome(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsTable:
			cr, _ := FP.TransactionsTable.GetSelection()
			FP.TransactionsTable.Select(cr, 0)
			FP.App.SetFocus(FP.TransactionsTable)

			return nil
		default:
			return e
		}
	case PageResults:
		return e
	default:
		return e
	}
}

func actionDown(e *tcell.EventKey) *tcell.EventKey {
	switch FP.App.GetFocus() {
	case FP.TransactionsInputField:
		return nil
	default:
		return e
	}
}

func actionUp(e *tcell.EventKey) *tcell.EventKey {
	switch FP.App.GetFocus() {
	case FP.TransactionsInputField:
		return nil
	default:
		return e
	}
}

func actionLeft(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.ProfileList:
			FP.App.SetFocus(FP.TransactionsTable)
			return nil
		case FP.TransactionsTable:
			_, cc := FP.TransactionsTable.GetSelection()
			// focus the profile list when at column 0
			if cc == 0 {
				FP.App.SetFocus(FP.ProfileList)
				return nil
			}

			return e
		default:
			return e
		}
	default:
		return e
	}
}

func actionRight(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.ProfileList:
			FP.App.SetFocus(FP.TransactionsTable)
			return nil
		case FP.TransactionsTable:
			c := FP.TransactionsTable.GetColumnCount() - 1
			_, cc := FP.TransactionsTable.GetSelection()
			// focus the profile list when at max column
			if cc == c {
				FP.App.SetFocus(FP.ProfileList)
				return nil
			}

			return e
		default:
			return e
		}
	default:
		return e
	}
}

func actionPageDown(e *tcell.EventKey) *tcell.EventKey {
	f := FP.App.GetFocus()
	p, _ := FP.Pages.GetFrontPage()

	switch p {
	case PageResults:
		switch f {
		case FP.ResultsDescription:
			return e
		case FP.ResultsTable:
			return e
		default:
			FP.App.SetFocus(FP.ResultsTable)
			return nil
		}
	default:
		return e
	}
}

func actionPageUp(e *tcell.EventKey) *tcell.EventKey {
	f := FP.App.GetFocus()
	p, _ := FP.Pages.GetFrontPage()

	switch p {
	case PageResults:
		switch f {
		case FP.ResultsDescription:
			return e
		case FP.ResultsTable:
			return e
		default:
			FP.App.SetFocus(FP.ResultsTable)
			return nil
		}
	default:
		return e
	}
}

func actionBackTab(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return nil
		case FP.ProfileList:
			FP.App.SetFocus(FP.TransactionsTable)
		case FP.TransactionsTable:
			// get the height & width of the transactions table
			// r := FP.TransactionsTable.GetRowCount() - 1
			c := FP.TransactionsTable.GetColumnCount() - 1
			cr, cc := FP.TransactionsTable.GetSelection()
			nc := cc - 1
			nr := cr

			var focusTarget tview.Primitive
			focusTarget = FP.TransactionsTable

			if nc < 0 {
				nc = c
				nr--

				if nr < 0 {
					// nc = c
					nc = 0
					// nr = r
					nr = 0
				}
				// it's more intuitive to go back to the FP.ProfileList
				// when backtabbing from the first column in the table
				focusTarget = FP.ProfileList
			}

			FP.TransactionsTable.Select(nr, nc)
			FP.App.SetFocus(focusTarget)
		default:
			FP.App.SetFocus(FP.ProfileList)
		}

		return nil
	case PageResults:
		switch FP.App.GetFocus() {
		case FP.ResultsTable:
			FP.ResultsForm.SetFocus(0)
			FP.App.SetFocus(FP.ResultsForm)

			return nil
		case FP.ResultsDescription:
			FP.App.SetFocus(FP.ResultsTable)
		case FP.ResultsForm:
			return e
		}

		return e
	}

	return e
}

func actionTab(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case FP.TransactionsInputField:
			return nil
		case FP.ProfileList:
			FP.App.SetFocus(FP.TransactionsTable)
		case FP.TransactionsTable:
			// get the height & width of the transactions table
			r := FP.TransactionsTable.GetRowCount() - 1
			c := FP.TransactionsTable.GetColumnCount() - 1
			cr, cc := FP.TransactionsTable.GetSelection()
			nc := cc + 1
			nr := cr

			var focusTarget tview.Primitive

			focusTarget = FP.TransactionsTable

			if nc > c {
				nc = 0 // loop around
				nr++

				if nr > r {
					nc = 0
					nr = r
				}
				// it's more intuitive to go back to the FP.ProfileList
				// when backtabbing from the first column in the table
				focusTarget = FP.ProfileList
			}

			FP.TransactionsTable.Select(nr, nc)
			FP.App.SetFocus(focusTarget)
		default:
			FP.App.SetFocus(FP.ProfileList)
		}

		return nil
	case PageResults:
		switch FP.App.GetFocus() {
		case FP.ResultsTable:
			FP.App.SetFocus(FP.ResultsDescription)
		case FP.ResultsDescription:
			FP.ResultsForm.SetFocus(0)
			FP.App.SetFocus(FP.ResultsForm)

			return nil
		case FP.ResultsForm:
			return e
		}

		return e
	}

	return e
}

func actionEsc(e *tcell.EventKey) *tcell.EventKey {
	currentFocus := FP.App.GetFocus()
	switch currentFocus {
	case FP.TransactionsInputField:
		return e
	case FP.TransactionsTable:
		// deselect the last selected index on the first press
		if FP.LastSelection != -1 {
			FP.LastSelection = -1

			getTransactionsTable()

			cr, cc := FP.TransactionsTable.GetSelection()

			FP.TransactionsTable.Select(cr, cc)
			FP.App.SetFocus(FP.TransactionsTable)

			return nil
		}

		anySelected := false

		for i := range FP.SelectedProfile.TX {
			if FP.SelectedProfile.TX[i].Selected {
				anySelected = true
				FP.SelectedProfile.TX[i].Selected = false
			}
		}

		if !anySelected {
			FP.App.SetFocus(FP.ProfileList)
			return nil
		}

		modified()

		getTransactionsTable()

		cr, cc := FP.TransactionsTable.GetSelection()

		FP.TransactionsTable.Select(cr, cc)
		FP.App.SetFocus(FP.TransactionsTable)
	case FP.ResultsForm:
		FP.App.SetFocus(FP.ResultsTable)
		return nil
	case FP.ResultsTable:
		FP.Pages.SwitchToPage(PageProfiles)
		return nil
	default:
		promptExit()
		return nil
	}

	return e
}

func actionResults() *tcell.EventKey {
	// if the user is already on the results page, focus the
	// text view description instead
	p, _ := FP.Pages.GetFrontPage()
	alreadyOnPage := false

	if p == PageResults {
		alreadyOnPage = true
	}

	FP.Pages.SwitchToPage(PageResults)
	setBottomPageNavText()

	if alreadyOnPage {
		getResultsTable()
		FP.App.SetFocus(FP.ResultsTable)
	}

	return nil
}

func actionProfiles() *tcell.EventKey {
	p, _ := FP.Pages.GetFrontPage()
	alreadyOnPage := false

	if p == PageProfiles {
		alreadyOnPage = true
	}

	FP.Pages.SwitchToPage(PageProfiles)
	setBottomPageNavText()

	if alreadyOnPage {
		FP.App.SetFocus(FP.ProfileList)
	}

	return nil
}

func actionGlobalHelp() *tcell.EventKey {
	FP.Pages.SwitchToPage(PageHelp)
	setBottomPageNavText()

	return nil
}

func actionHelp(e *tcell.EventKey) *tcell.EventKey {
	switch FP.App.GetFocus() {
	case FP.TransactionsInputField:
		return e
	case FP.ResultsForm:
		return e
	default:
		FP.Pages.SwitchToPage(PageHelp)
		setBottomPageNavText()

		return e
	}
}

// action is the primary decision tree that is triggered when a key event
// is triggered. Please ensure that every case statement has a return or
// fallthrough, and note that the "nolint" for this function is required
// because there is really no way to make it any simpler without silliness.
//
//nolint:funlen,cyclop
func action(action string, e *tcell.EventKey) *tcell.EventKey {
	duplicating := false
	multiSelecting := false

	switch action {
	case c.ActionRedo:
		return actionRedo(e)
	case c.ActionUndo:
		return actionUndo(e)
	case c.ActionQuit:
		return actionQuit()
	case c.ActionMulti:
		multiSelecting = true

		fallthrough
	case c.ActionSelect:
		return actionSelect(e, multiSelecting)
	case c.ActionMove:
		return actionMove(e)
	case c.ActionDelete:
		return actionDelete(e)
	case c.ActionDuplicate:
		duplicating = true

		fallthrough
	case c.ActionAdd:
		return actionAdd(e, duplicating)
	case c.ActionEdit:
		return actionEdit(e)
	case c.ActionSave:
		return actionSave()
	case c.ActionEnd:
		return actionEnd(e)
	case c.ActionHome:
		return actionHome(e)
	case c.ActionDown:
		return actionDown(e)
	case c.ActionUp:
		return actionUp(e)
	case c.ActionLeft:
		return actionLeft(e)
	case c.ActionRight:
		return actionRight(e)
	case c.ActionPageDown:
		return actionPageDown(e)
	case c.ActionPageUp:
		return actionPageUp(e)
	case c.ActionBackTab:
		return actionBackTab(e)
	case c.ActionTab:
		return actionTab(e)
	case c.ActionEsc:
		return actionEsc(e)
	case c.ActionResults:
		return actionResults()
	case c.ActionProfiles:
		return actionProfiles()
	case c.ActionGlobalHelp:
		return actionGlobalHelp()
	case c.ActionHelp:
		return actionHelp(e)
	case c.ActionSearch:
		// searching not implemented yet
		fallthrough
	default:
		return e
	}
}
