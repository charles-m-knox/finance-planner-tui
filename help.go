package main

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
	"text/template"

	c "gitea.cmcode.dev/cmcode/finance-planner-tui/constants"
	m "gitea.cmcode.dev/cmcode/finance-planner-tui/models"

	"github.com/rivo/tview"
)

// merges the default keybindings with the user's customized keybindings.
//
// Example: "Ctrl+S": ["save"]
//
// Do not use outside of the context of documentation, because this will also
// modify things like Rune[x] to render properly within a dynamically colored
// textview. For example, Rune[x] will transform to Rune[x[].
func GetCombinedKeybindings(kb map[string][]string, def map[string]string) map[string][]string {
	r := make(map[string][]string)
	reg := regexp.MustCompile(`^Rune\[.\]$`)

	for k, v := range def {
		if reg.MatchString(k) {
			r[strings.Replace(k, "]", "[]", 1)] = []string{v}

			continue
		}

		r[k] = []string{v}
	}

	for k, v := range kb {
		if reg.MatchString(k) {
			r[strings.Replace(k, "]", "[]", 1)] = v

			continue
		}
		// delete the old keybinding and reformat it to show that it's customized
		formattedKeybinding := fmt.Sprintf("[gold::b]%v[-:-:-:-]", k)
		delete(r, k)
		r[formattedKeybinding] = v
	}

	return r
}

// merges the default keybindings with the user's customized keybindings, except
// unlike GetCombinedKeybindings, this will list every Action as the primary map
// key, and the keybindings are the map values. There may be multiple
// keybindings for a single action. In the event that there is a chained
// keybinding, such as Ctrl+X mapping to save+quit, the keybinding will be
// rendered lightgreen instead of gold (which is the norm for custom
// keybindings).
//
// Example: "save": []string{"[lightgreen]Ctrl+X[-]", "Ctrl+S"}
//
// Keybindings are inserted in order of priority - custom keybindings will be at
// the 0-based index of the slice, so that various UI elements can quickly
// render the last-defined keybinding (not all UI elements have the space to
// show every keybinding. Plus, the help file shows all defined keybindings).
//
// Do not use outside of the context of documentation, because this will also
// modify things like Rune[x] to render properly within a dynamically colored
// textview. For example, Rune[x] will transform to Rune[x[].
func GetAllBoundActions(kb map[string][]string, def map[string]string) map[string][]string {
	r := make(map[string][]string)
	reg := regexp.MustCompile(`^Rune\[.\]$`)

	// handle default actions first
	for binding, action := range def {
		fixedBinding := binding
		if reg.MatchString(fixedBinding) {
			fixedBinding = strings.Replace(fixedBinding, "]", "[]", 1)
		}

		r[action] = []string{fixedBinding}
	}

	// higlight custom key bindings next
	for binding, actions := range kb {
		color := "gold"
		if len(actions) > 1 {
			color = "#aaffee"
		}

		fixedBinding := binding
		if reg.MatchString(fixedBinding) {
			fixedBinding = strings.Replace(fixedBinding, "]", "[]", 1)
		}

		formattedBinding := fmt.Sprintf("[%v::b]%v[-:-:-:-]", color, fixedBinding)

		for _, action := range actions {
			r[action] = slices.Insert(r[action], 0, formattedBinding)
		}
	}

	return r
}

func getHelpText(conf m.Config, combinedKeybindings, combinedActions map[string][]string) string {
	type tmplDataShape struct {
		Conf                m.Config
		AllActions          []string
		DefaultKeybindings  map[string]string
		CombinedKeybindings map[string][]string
		CombinedActions     map[string][]string
		Explanations        map[string]string
	}

	tmplData := tmplDataShape{
		Conf:                conf,
		AllActions:          c.AllActions,
		DefaultKeybindings:  c.DefaultMappings,
		CombinedKeybindings: combinedKeybindings,
		CombinedActions:     combinedActions,
		Explanations:        c.ActionExplanations,
	}

	tmpl, err := template.New("help").Parse(FP.T["HelpTextTemplate"])
	if err != nil {
		log.Fatalf("failed to parse help text template: %v", err.Error())
	}

	var b bytes.Buffer

	err = tmpl.Execute(&b, tmplData)
	if err != nil {
		log.Fatalf("failed to render help text: %v", err.Error())
	}

	return b.String()
}

func getHelpModal() {
	FP.HelpTextView = tview.NewTextView()
	FP.HelpTextView.SetBorder(true)
	FP.HelpTextView.SetText(getHelpText(FP.Config, FP.KeyBindings, FP.ActionBindings)).SetDynamicColors(true)
}

// returns the first configured keybinding for the provided action. returns
// "n/a" if none defined.
func getBinding(action string) string {
	bindings, ok := FP.ActionBindings[action]
	if !ok || len(bindings) < 1 {
		return ""
	}

	return bindings[0]
}

// setBottomPageNavText renders something like "F1 help F2 profiles F3 results"
// at the bottom of the terminal.
func setBottomPageNavText() {
	p, _ := FP.Pages.GetFrontPage()

	pgs := [][]string{
		{PageHelp, FP.T["BottomPageNavTextHelp"], getBinding(c.ActionGlobalHelp)},
		{PageProfiles, FP.T["BottomPageNavTextProfiles"], getBinding(c.ActionProfiles)},
		{PageResults, FP.T["BottomPageNavTextResults"], getBinding(c.ActionResults)},
	}

	var sb strings.Builder

	for _, v := range pgs {
		color := "[gray]"
		if p == v[0] {
			color = "[gold]"
		}

		sb.WriteString(fmt.Sprintf("%v%v%v %v %v", v[2], c.Reset, color, v[1], c.Reset))
	}

	FP.BottomPageNavText.SetText(sb.String())
}
