package main

import (
	"bytes"
	"log"
	"text/template"

	c "finance-planner-tui/constants"
	m "finance-planner-tui/models"
)

const HelpTextTemplate = `[lightgreen::b]Finance Planner[-:-:-:-]

[gold]
                 _____ _
                |  ___(_)_ __   __ _ _ __   ___ ___
                | |_  | | '_ \ / _  | '_ \ / __/ _ \
                |  _| | | | | | (_| | | | | (_|  __/
                |_|__ |_|_| |_|\__,_|_| |_|\___\___|[lightgreen]
                |  _ \| | __ _ _ __  _ __   ___ _ __
                | |_) | |/ _  | '_ \| '_ \ / _ \ '__|
                |  __/| | (_| | | | | | | |  __/ |
                |_|   |_|\__,_|_| |_|_| |_|\___|_|
[-:-:-:-]


[lightgreen::b]General information[-:-:-:-]

[white]The purpose of this application is to allow you to define recurring bills
and income ([gold]transactions[white]), and then get a fairly accurate prediction
of where your money will be using the [blue]Results[white] page.

[lightgreen::b]More on Profiles[-:-:-:-]

[white]Profiles are shown on the left-hand side of the [blue]Profiles & Transactions[white] page.

- You may need to use the <tab> key to get to them.
- You can duplicate and rename profiles.
- [gold]Each profile must have a unique name.[white] Duplicate names will be refused.

You can create multiple [blue]profiles[white] to fulfill any purpose, such as:

- modeling a change in your financial plans (removing subscriptions,
  hypotheticals, etc)
- adding multiple family members

[lightgreen::b]More on Transactions[-:-:-:-]

[white]A [blue]transaction[white] is a recurring expense or income:

- If the transaction earns money, it is prefixed with a [lightgreen]+[white] (plus) sign.
- All transactions are assumed to be negative by default.

Each transaction has the following fields:

- Order:     You can define a custom integer sorting order for transactions.
             This field has no other purpose.
- Amount:    This is a positive or negative value as described above.
- Active:    This is a boolean value that determines whether the transaction should
             be included in calculations. This is useful for temporarily making
             changes without destroying anything.
- Name:      This is the human-readable name of the transaction for your eyes.
- Frequency: Transactions can occur [aqua]MONTHLY, [lightgreen]WEEKLY, or [gold]YEARLY.
             [white]This value must be exactly one of those three strings, but an auto-
             complete is provided to make it quicker.
- Interval:  The transaction occurs every [aqua]<interval>[white] WEEKS/MONTHS/YEARS.
- <Weekday>: The transaction only occurs on the checked days of the week, and
             will not occur if the defined recurrence pattern does not land on
             one of these days.
- Starts:    This is the starting date for the transaction's recurrence pattern.
             It is defined as [aqua]YYYY[white]-[lightgreen]MM[white]-[gold]DD[white].

             For simplicity when working with dates at the end of the month,
             you may want to consider putting setting the day value to 28, as
             some recurrence patterns may skip a 31.

             Months range from [aqua]1-12[white], and days range from [aqua]1-31[white].
             Years must be any positive value, and can be 0.
- Ends:      This is the last acceptable date for recurrence. Behavior is the
             exact same as the Starts field.
- Note:      A human-readable field for you to put arbitrary notes in.

[lightgreen::b]Keyboard Shortcuts:[-:-:-:-]

{{ range .AllActions }}
- {{- . }}: {{ if .Conf.Keybindings . }} {{ .Conf.Keybindings . }} {{ else }} {{ .DefaultKeybindings . }} {{ end }}
{{ end }}
`

// <tab>/<shift+tab>: cycle back and forth between panels/controls where
// appropriate

// <esc>: deselects the last selected mark, and then deselects the last
// selected items, then un-focuses panes until eventually exiting the application
// entirely

// <ctrl+s>: saves to config.yml in the current directory

// <ctrl+i>: shows statistics in the Results page's bottom text pane

// [lightgreen::b]Transactions page:[-:-:-:-]

// <space>: select the current transaction
// <ctrl+space>: toggle multi-select from the last previously selected item
// <>

func getHelpText(conf m.Config) (output string) {
	type tmplDataShape struct {
		Conf               m.Config
		AllActions         []string
		DefaultKeybindings map[string]string
	}

	tmplData := tmplDataShape{
		Conf:               conf,
		AllActions:         c.ALL_ACTIONS,
		DefaultKeybindings: c.DEFAULT_MAPPINGS,
	}

	tmpl, err := template.New("help").Parse(HelpTextTemplate)
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
