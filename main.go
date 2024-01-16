package main

import (
	"embed"
	"flag"
	"log"

	c "gitea.cmcode.dev/cmcode/finance-planner-tui/constants"
	"gitea.cmcode.dev/cmcode/finance-planner-tui/lib"
	m "gitea.cmcode.dev/cmcode/finance-planner-tui/models"
	"gitea.cmcode.dev/cmcode/finance-planner-tui/themes"
	"gitea.cmcode.dev/cmcode/finance-planner-tui/translations"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

//go:embed translations/*.yml
var AllTranslations embed.FS

//go:embed themes/*.yml
var AllThemes embed.FS

//go:embed example.yml
var ExampleConfig embed.FS

const (
	// PageProfiles is not shown to the user ever, and is only used in the code.
	// Its primary purpose is for use in switch/case statements to determine the
	// current page.
	PageProfiles = "Profiles"
	// PageResults is not shown to the user ever, and is only used in the code.
	// Its primary purpose is for use in switch/case statements to determine the
	// current page.
	PageResults = "Results"
	// PageHelp is not shown to the user ever, and is only used in the code. Its
	// primary purpose is for use in switch/case statements to determine the
	// current page.
	PageHelp = "Help"
	// PagePrompt is not shown to the user ever, and is only used in the code.
	// Its primary purpose is for use in switch/case statements to determine the
	// current page.
	PagePrompt = "Prompt"
)

type FinancePlanner struct {
	// The tview/tcell terminal application.
	App *tview.Application

	// The currently loaded configuration. The contents of this will be saved
	// to disk.
	Config m.Config

	// A pointer to the currently selected profile, which is a member of the
	// currently loaded config.
	SelectedProfile *m.Profile

	// The primary primitive that the app uses as its root in the terminal.
	Layout *tview.Flex

	// Translations that are loaded at runtime.
	T map[string]string

	// The previously focused primitive.
	Previous tview.Primitive

	// The previously shown page (via the primary pages primitive).
	PrevPage string

	// The primary page-switching primitive.
	Pages *tview.Pages

	// SortTX is the name of the column that is used for sorting data in the
	// transactions table, followed by Asc/Desc.
	SortTX string

	// True when calculating results. Used in async operations. Use with
	// care.
	CalculatingResults bool

	// The last-selected index. Set to -1 to safely reset. When set, the
	// transactions table will highlight this row to show where the last
	// selected item was (useful for multi-selecting)
	LastSelection int

	// All activated key bindings. Composed of the user's key bindings merged on
	// top of the default key bindings, as one would expect. It is
	// possible for unsupported keyboard shortcuts to be present in this map.
	//
	// usage example: KeyBindings["Ctrl+Z"] = ["undo", "save"].
	KeyBindings map[string][]string

	// All activated action bindings. Composed of the user's configured actions
	// merged on top of the default actions, as one would expect. It is
	// possible for unsupported actions to be present in this map.
	//
	// usage example: ActionBindings["save"] = ["Ctrl+S", "[gold]Ctrl+X"].
	ActionBindings map[string][]string

	// Shows the gigantic help text on the help page.
	HelpTextView *tview.TextView

	// Always shown on every page - renders the keyboard shortcuts for
	// all supported pages.
	BottomPageNavText *tview.TextView

	// This is the text that is shown below the profile list, and contains
	// status and error messages, the name of the config (if it fits), and other
	// things.
	ProfileStatusText *tview.TextView

	// Contains the list of profiles that are contained within the current
	// config. This is a list that the user can navigate and upon hitting the
	// enter key, the selected profile will be loaded into the transactions
	// table, results page, etc.
	ProfileList *tview.List

	TransactionsTable      *tview.Table
	TransactionsInputField *tview.InputField

	// This is the text that is shown below the results table, and contains
	// status messages, stats about the results, and any other errors that might
	// come up.
	ResultsDescription *tview.TextView
	ResultsTable       *tview.Table
	ResultsForm        *tview.Form

	// The latest results are stored. For start & end dates that span huge
	// amounts of time, you may need to think critically about what can be
	// stored in this, and how garbage collection is a factor. Consider zeroing
	// out everything where necessary.
	LatestResults *[]lib.Result

	// There is a hidden fourth page that only shows a modal, typically shown
	// only for exiting or keyboard echo mode.
	PromptBox *tview.Modal

	// The undo buffer contains yaml-serialized byte slices. Each member of the
	// slice is the entire serialized yaml config at a specific point in time.
	// Moving back and forth throughout the undo buffer works as you'd expect,
	// see the undo(), redo(), and modified() functions.
	UndoBuffer [][]byte

	// The undo buffer's position is tracked globally via this variable.
	UndoBufferPos int

	// The name of the configuration file. This will get populated if set by
	// a flag at runtime, and determines the name of the file that this program
	// will save configuration changes to. The value can be an absolute or a
	// relative path. See the loadConfig function.
	FlagConfigFile string

	// This is a hidden flag. It loads a config json file from a previous
	// version of this software and attempts to convert it to the latest config.
	FlagShouldMigrate bool

	// If this flag is set to true, the application will only show the user the
	// keyboard keys that they press. They will of course be prompted to proceed
	// before being fully immersed into this restricted mode.
	FlagKeyboardEchoMode bool

	// Allows the colors of the application to be changed. Themes are included
	// as an embedded file when compiling, and will be parsed at runtime.
	//
	// If this is set to a value ending in .yml or .yaml, this application will
	// attempt to load the file at runtime. Conversely, if it does not end in
	// both of those two suffixes, this application will load from the included
	// themes/$FlagTheme.yml files. Values left undefined in the current theme
	// will fallback to the default theme's values.
	FlagTheme string

	// All default & custom colors are stored in here at runtime. Themes can be
	// loaded via FlagTheme.
	Colors map[string]string

	// For an input string such as "AmountAsc", this will return a predefined sort
	// function that can be executed.
	TransactionsSortMap map[string]TxSortFunc

	// An index of the days of the week. This is needed so that we can create a
	// direct mapping between the translation table's weekday entries and the
	// acceptable rrule.Weekday values (which regard Monday as the start of
	// the week, instead of Sunday, which is what the Go standard lib does).
	WeekdaysMap map[string]int

	// All of the columns that will be shown in the transactions table. Loaded
	// once at runtime with values from translation table.
	TransactionsTableHeaders []m.TableCell
}

// FP contains all shared data in a global. Avoid using globals where possible,
// but in the context of an application like this, things will get extremely
// messy without a global unless I spend a ton of time cleaning up and
// refactoring.
//
//nolint:gochecknoglobals
var FP FinancePlanner

// For an input keybinding (straight from event.Name()), an output action
// will be returned, for example - "Ctrl+Z" will return "undo".
func getDefaultKeybind(name string) string {
	m, ok := c.DefaultMappings[name]
	if !ok {
		return ""
	}

	return m
}

// capture is the primary input capture handler for the app, and should be used
// like: app.SetInputCapture(capture)
func capture(e *tcell.EventKey) *tcell.EventKey {
	n := e.Name()
	if FP.FlagKeyboardEchoMode {
		FP.ProfileStatusText.SetDynamicColors(false).SetText(n)

		if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC {
			FP.App.Stop()
		}

		return nil
	}

	var final *tcell.EventKey
	final = e

	foundBinding := false

	for binding, actions := range FP.Config.Keybindings {
		if n != binding {
			continue
		}

		foundBinding = true

		for i := range actions {
			final = action(actions[i], final)
		}
	}

	if !foundBinding {
		// execute default action
		return action(getDefaultKeybind(n), e)
	}

	return final
}

// bootstrap is the initialization function for the app, including initializing
// globals. This function should only ever be run once.
//
// t is the translation map, and conf is the freshly loaded config.
func bootstrap(t map[string]string, conf m.Config) {
	b, err := yaml.Marshal(conf)
	if err != nil {
		log.Fatalf("%v: %v", t["ErrorFailedToMarshalInitialConfig"], err.Error())
	}

	FP.KeyBindings = GetCombinedKeybindings(conf.Keybindings, c.DefaultMappings)
	FP.ActionBindings = GetAllBoundActions(conf.Keybindings, c.DefaultMappings)

	initializeUndo(b, conf.DisableGzipCompressionInUndoBuffer)

	FP.LastSelection = -1
	FP.App = tview.NewApplication()

	FP.Pages = tview.NewPages()

	getHelpModal()

	FP.PromptBox = tview.NewModal()

	FP.Pages.AddPage(PageProfiles, getProfilesPage(), true, true).
		AddPage(PageResults, getResultsPage(), true, true).
		AddPage(PageHelp, FP.HelpTextView, true, true).
		AddPage(PagePrompt, FP.PromptBox, true, true)

	FP.Pages.SwitchToPage(PageProfiles)

	FP.BottomPageNavText = tview.NewTextView()

	FP.BottomPageNavText.SetDynamicColors(true)
	setBottomPageNavText()

	FP.Layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(FP.Pages, 0, 1, true).AddItem(FP.BottomPageNavText, 1, 0, false)

	FP.App.SetFocus(FP.ProfileList)

	promptKBMode(t)

	FP.App.SetInputCapture(capture)
}

// parseFlags parses the command line flags, using t as the translation map.
func parseFlags(t map[string]string) {
	flag.StringVar(&FP.FlagConfigFile, t["FlagConfigFileFlag"], "", t["FlagConfigFileDesc"])
	flag.BoolVar(&FP.FlagShouldMigrate, t["FlagShouldMigrateFlag"], false, t["FlagShouldMigrateDesc"])
	flag.BoolVar(&FP.FlagKeyboardEchoMode, t["FlagKeyboardEchoModeFlag"], false, t["FlagKeyboardEchoModeDesc"])
	flag.StringVar(&FP.FlagTheme, t["FlagThemeFlag"], "", t[" FlagThemeDesc"])
	flag.Parse()
}

func main() {
	var err error

	FP.T, err = translations.Load(AllTranslations)
	if err != nil {
		log.Fatalf("failed to load translations: %v", err.Error())
	}

	parseFlags(FP.T)

	if FP.FlagShouldMigrate {
		JSONtoYAML()
	}

	FP.Config, FP.FlagConfigFile, err = loadConfig(FP.FlagConfigFile, FP.T, ExampleConfig)
	if err != nil {
		log.Fatalf("%v: %v", FP.T["ErrorFailedToLoadConfig"], err.Error())
	}

	processConfig(&FP.Config)

	theme := FP.Config.Theme
	if FP.FlagTheme != "" {
		theme = FP.FlagTheme
	}

	FP.Colors, err = themes.Load(AllThemes, theme)
	if err != nil {
		log.Fatalf("%v: %v", FP.T["ErrorFailedToLoadThemes"], err.Error())
	}

	if len(FP.Config.Profiles) > 0 {
		FP.SelectedProfile = &(FP.Config.Profiles[0])
	} else {
		n := m.Profile{
			TX:   []lib.TX{lib.GetNewTX()},
			Name: FP.T["DefaultNewProfileName"],
		}
		FP.Config.Profiles = append(FP.Config.Profiles, n)
		FP.SelectedProfile = &(FP.Config.Profiles[0])
	}

	bootstrap(FP.T, FP.Config)

	if err := FP.App.SetRoot(FP.Layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
