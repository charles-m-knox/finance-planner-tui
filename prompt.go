package main

import "github.com/gdamore/tcell/v2"

// This file mainly contains functions for the hidden prompt page in the
// application.

func promptExit() {
	// check if we are already prompting
	currentPage, _ := FP.Pages.GetFrontPage()
	if currentPage == PagePrompt {
		return
	}

	// now check if the previous page is something other than the prompt already
	FP.PrevPage, _ = FP.Pages.GetFrontPage()
	if FP.PrevPage == PagePrompt {
		return
	}

	FP.PromptBox.ClearButtons().AddButtons(
		[]string{
			FP.T["PromptExitButtonExit"],
			FP.T["PromptExitButtonNo"],
			FP.T["PromptExitButtonCancel"],
		},
	).SetText(FP.T["PromptExitText"]).SetDoneFunc(
		func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				FP.App.Stop()
			case 1:
				fallthrough
			case 2:
				fallthrough
			default:
				FP.Pages.SwitchToPage(FP.PrevPage)
				return
			}
		},
	).SetBackgroundColor(tcell.ColorGoldenrod).
		SetTextColor(tcell.ColorBlack)

	FP.Pages.SwitchToPage(PagePrompt)
	FP.PromptBox.SetFocus(2)
	FP.App.SetFocus(FP.PromptBox)
}

// promptKBMode switches to the prompt page and shows a modal that informs the
// user that they are in keyboard echo mode. If KB echo mode is not enabled,
// this gracefully returns immediately and does nothing.
//
// Requires the first argument to be the translation map.
func promptKBMode(t map[string]string) {
	if !FP.FlagKeyboardEchoMode {
		return
	}

	// temporarily turn off KB echo mode so that the user's keys are captured
	// properly until they can give consent to entering the mode
	FP.FlagKeyboardEchoMode = false

	FP.PromptBox.ClearButtons().AddButtons(
		[]string{
			t["PromptKeyboardEchoModeButtonTurnOff"],
			t["PromptKeyboardEchoModeButtonExitNow"],
			t["PromptKeyboardEchoModeButtonContinue"],
		},
	).SetText(t["PromptKeyboardEchoModeText"]).SetDoneFunc(
		func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				FP.FlagKeyboardEchoMode = false
				FP.Pages.SwitchToPage(PageProfiles)
			case 1:
				FP.FlagKeyboardEchoMode = false
				FP.App.Stop()
			case 2:
				FP.FlagKeyboardEchoMode = true
				FP.Pages.SwitchToPage(PageProfiles)
			default:
				FP.FlagKeyboardEchoMode = false
				FP.App.Stop()
				return
			}
		},
	).SetBackgroundColor(tcell.ColorDimGray).
		SetTextColor(tcell.ColorWhite)

	FP.Pages.SwitchToPage(PagePrompt)
	FP.PromptBox.SetFocus(2)
	FP.App.SetFocus(FP.PromptBox)
}
