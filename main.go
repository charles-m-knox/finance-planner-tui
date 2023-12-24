package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	c "finance-planner-tui/constants"
	"finance-planner-tui/lib"
	m "finance-planner-tui/models"
	"finance-planner-tui/translations"

	"github.com/adrg/xdg"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

//go:embed translations/*.yml
var AllTranslations embed.FS

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
}

// FP contains all shared data in a global. Avoid using globals where possible,
// but in the context of an application like this, things will get extremely
// messy without a global unless I spend a ton of time cleaning up and
// refactoring.
//
//nolint:gochecknoglobals
var FP FinancePlanner

var (
	config          m.Config
	selectedProfile *m.Profile
	previous        tview.Primitive

	// profilesPage items:
	profileList            *tview.List
	statusText             *tview.TextView
	transactionsTable      *tview.Table
	transactionsInputField *tview.InputField
	// results items:
	resultsTable       *tview.Table
	resultsForm        *tview.Form
	resultsDescription *tview.TextView
	resultsRightSide   *tview.Flex

	// The latest results are stored. For start & end dates that span huge
	// amounts of time, you may need to think critically about what can be stored
	// in this, and how garbage collection is a factor. Consider zeroing out
	// everything where necessary.
	latestResults *[]lib.Result

	// help page
	helpModal      *tview.TextView
	bottomHelpText *tview.TextView
	// prompt page
	promptBox *tview.Modal

	undoBuffer    [][]byte
	undoBufferPos int

	// flags
	configFile       string
	shouldMigrate    bool
	keyboardEchoMode bool
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
	helpModal = tview.NewTextView()
	helpModal.SetBorder(true)
	helpModal.SetText(getHelpText(config, FP.KeyBindings, FP.ActionBindings)).SetDynamicColors(true)
}

// returns the first configured keybinding for the provided action. returns
// "n/a" if none defined
func getBinding(action string) string {
	bindings, ok := FP.ActionBindings[action]
	if !ok || len(bindings) < 1 {
		return ""
	}

	return bindings[0]
}

func setBottomHelpText() {
	p, _ := FP.Pages.GetFrontPage()

	var sb strings.Builder

	if p == PageHelp {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gold] help [-:-:-:-]", getBinding(c.ActionGlobalHelp)))
	} else {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gray] help [-:-:-:-]", getBinding(c.ActionGlobalHelp)))
	}

	if p == PageProfiles {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gold] profiles & transactions [-:-:-:-]", getBinding(c.ActionProfiles)))
	} else {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gray] profiles & transactions [-:-:-:-]", getBinding(c.ActionProfiles)))
	}

	if p == PageResults {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gold] results [-:-:-:-]", getBinding(c.ActionResults)))
	} else {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gray] results[-:-:-:-]", getBinding(c.ActionResults)))
	}

	bottomHelpText.SetText(sb.String())
}

// attempts to load from the "file" path provided - if not successful,
// attempts to load from xdg config, then xdg home. Then it sets the global
// configFile to match the retrieved config
func loadConfig(file string) (m.Config, error) {
	conf := m.Config{}

	xdgConfig := path.Join(xdg.ConfigHome, "finance-planner-tui", "config.yml")
	xdgHome := path.Join(xdg.Home, "finance-planner-tui", "config.yml")

	specificFileGiven := true

	if file == "" {
		file = c.DefaultConfig
		specificFileGiven = false
	}

	b, err := os.ReadFile(file)
	if err == nil {
		err = yaml.Unmarshal(b, &conf)
		if err != nil {
			return conf, fmt.Errorf(
				"failed to read config from %v: %v",
				configFile,
				err.Error(),
			)
		}

		return conf, nil
	}

	// if a file was specified via the -f flag, but it doesn't exist, proceed to
	// set the configFile global var to it so that it will be written on next
	// save
	if specificFileGiven {
		configFile = file

		return conf, nil
	}

	b, err = os.ReadFile(xdgConfig)
	if err == nil {
		configFile = xdgConfig
		err = yaml.Unmarshal(b, &conf)

		if err != nil {
			return conf, fmt.Errorf(
				"failed to read config from %v: %v",
				xdgConfig,
				err.Error(),
			)
		}

		return conf, nil
	}

	b, err = os.ReadFile(xdgHome)
	if err == nil {
		configFile = xdgHome
		err = yaml.Unmarshal(b, &conf)

		if err != nil {
			return conf, fmt.Errorf(
				"failed to read config from %v: %v",
				xdgHome,
				err.Error(),
			)
		}

		return conf, nil
	}

	// if the config file doesn't exist, create it at xdgHome
	configFile = xdgHome

	return conf, nil
}

// converts a json file to yaml (one-off job for converting from legacy versions
// of this program)
func JSONtoYAML() {
	b, err := os.ReadFile("conf.json")
	if err != nil {
		log.Fatalf("failed to load conf.json")
	}

	nc := m.Config{
		Profiles: []m.Profile{
			{
				Name: "migrated",
				TX:   []lib.TX{},
			},
		},
	}

	err = json.Unmarshal(b, &nc.Profiles[0].TX)
	if err != nil {
		log.Fatalf("failed to unmarshal conf: %v", err.Error())
	}

	// update all uuids in the config
	for i := range nc.Profiles[0].TX {
		nc.Profiles[0].TX[i].ID = uuid.NewString()
	}

	out, err := yaml.Marshal(nc)
	if err != nil {
		log.Fatalf("failed to marshal nc: %v", err.Error())
	}

	err = os.WriteFile("migrated.yml", out, 0o644)
	if err != nil {
		log.Fatalf("failed to write migrated.yml: %v", err.Error())
	}
}

func getActiveProfileText(profile m.Profile) string {
	if selectedProfile != nil && selectedProfile.Name == profile.Name {
		return fmt.Sprintf("[white::bu]%v (open)%v", profile.Name, c.ResetStyle)
	}

	return profile.Name
}

// populateProfilesPage clears out the profile list and proceeds to populate it
// with the current profiles in the config, including handlers for changing
// the selectedProfile.
func populateProfilesPage() {
	profileList.Clear()

	for i := range config.Profiles {
		profile := &(config.Profiles[i])
		profileList.AddItem(getActiveProfileText(*profile), "", 0, func() {
			selectedProfile = profile
			populateProfilesPage()
			getTransactionsTable()
			FP.App.SetFocus(transactionsTable)
		})
	}
}

func sortTX() {
	if FP.SortTX == c.None || FP.SortTX == "" {
		return
	}

	FP.LastSelection = -1

	sort.SliceStable(
		selectedProfile.TX,
		func(i, j int) bool {
			tj := (selectedProfile.TX)[j]
			ti := (selectedProfile.TX)[i]

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

func getTransactionsTable() {
	transactionsTable.Clear()

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

	// transactionsTable.SetCell(0, 0, cellColumnOrder)
	transactionsTable.SetCell(0, 0, cellColumnAmount)
	transactionsTable.SetCell(0, 1, cellColumnActive)
	transactionsTable.SetCell(0, 2, cellColumnName)
	transactionsTable.SetCell(0, 3, cellColumnFrequency)
	transactionsTable.SetCell(0, 4, cellColumnInterval)
	transactionsTable.SetCell(0, 5, cellColumnMonday)
	transactionsTable.SetCell(0, 6, cellColumnTuesday)
	transactionsTable.SetCell(0, 7, cellColumnWednesday)
	transactionsTable.SetCell(0, 8, cellColumnThursday)
	transactionsTable.SetCell(0, 9, cellColumnFriday)
	transactionsTable.SetCell(0, 10, cellColumnSaturday)
	transactionsTable.SetCell(0, 11, cellColumnSunday)
	transactionsTable.SetCell(0, 12, cellColumnStarts)
	transactionsTable.SetCell(0, 13, cellColumnEnds)
	transactionsTable.SetCell(0, 14, cellColumnNote)
	// transactionsTable.SetCell(0, 15, cellColumnID)
	// transactionsTable.SetCell(0, 16, cellColumnCreatedAt)
	// transactionsTable.SetCell(0, 17, cellColumnUpdatedAt)

	if selectedProfile != nil {
		sortTX()
		// start by populating the table with the columns first
		for i, tx := range selectedProfile.TX {
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

			// transactionsTable.SetCell(i+1, 0, cellOrder)
			transactionsTable.SetCell(i+1, 0, cellAmount)
			transactionsTable.SetCell(i+1, 1, cellActive)
			transactionsTable.SetCell(i+1, 2, cellName)
			transactionsTable.SetCell(i+1, 3, cellFrequency)
			transactionsTable.SetCell(i+1, 4, cellInterval)
			transactionsTable.SetCell(i+1, 5, cellMonday)
			transactionsTable.SetCell(i+1, 6, cellTuesday)
			transactionsTable.SetCell(i+1, 7, cellWednesday)
			transactionsTable.SetCell(i+1, 8, cellThursday)
			transactionsTable.SetCell(i+1, 9, cellFriday)
			transactionsTable.SetCell(i+1, 10, cellSaturday)
			transactionsTable.SetCell(i+1, 11, cellSunday)
			transactionsTable.SetCell(i+1, 12, cellStarts)
			transactionsTable.SetCell(i+1, 13, cellEnds)
			transactionsTable.SetCell(i+1, 14, cellNote)
			// transactionsTable.SetCell(i+1, 15, cellID)
			// transactionsTable.SetCell(i+1, 16, cellCreatedAt)
			// transactionsTable.SetCell(i+1, 17, cellUpdatedAt)
		}

		transactionsTable.SetSelectedFunc(func(row, column int) {
			// get the current profile & transaction
			i := 0

			// based on the row, find the actual transaction definition
			// example: row 5 = TX 4 because of table's headers
			for i = range selectedProfile.TX {
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
			// 	transactionsInputField.SetDoneFunc(func(key tcell.Key) {
			// 		switch key {
			// 		case tcell.KeyEscape:
			// 			// don't save the changes
			// 			deactivateTransactionsInputField()
			// 			return
			// 		default:
			// 			d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
			// 			if err != nil || d < 1 {
			// 				activateTransactionsInputFieldNoAutocompleteReset("invalid order given:", fmt.Sprint(selectedProfile.TX[i].Order))
			// 				return
			// 			}

			// 			// update all selected values as well as the current one
			// 			for j := range selectedProfile.TX {
			// 				if selectedProfile.TX[j].Selected || j == i {
			// 					selectedProfile.TX[j].Order = int(d)
			// 					transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
			// 						"%v%v",
			// 						c.COLOR_COLUMN_ORDER,
			// 						selectedProfile.TX[j].Order,
			// 					))
			// 				}
			// 			}

			// 			modified()
			// 			deactivateTransactionsInputField()
			// 		}
			// 	})
			// 	activateTransactionsInputField("order:", fmt.Sprint(selectedProfile.TX[i].Order))
			case c.ColumnAmountIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnAmount)

					return
				}

				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()

						return
					default:
						a := lib.ParseDollarAmount(transactionsInputField.GetText(), false)
						isPositiveAmount := a >= 0
						amountColor := c.ColorColumnAmount
						if isPositiveAmount {
							amountColor = c.ColorColumnAmountPositive
						}

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Amount = int(a)
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									amountColor,
									lib.FormatAsCurrency(selectedProfile.TX[j].Amount),
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				renderedAmount := lib.FormatAsCurrency(selectedProfile.TX[i].Amount)
				if selectedProfile.TX[i].Amount >= 0 {
					renderedAmount = fmt.Sprintf("+%v", renderedAmount)
				}
				activateTransactionsInputField("amount (start with + or $+ for positive):", renderedAmount)
			case c.ColumnActiveIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnActive)

					return
				}

				newValue := !selectedProfile.TX[i].Active
				selectedProfile.TX[i].Active = !selectedProfile.TX[i].Active

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						activeText := "✔"
						if !newValue {
							activeText = " "
						}
						selectedProfile.TX[j].Active = newValue

						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnActive, activeText))
					}
				}

				modified()
			case c.ColumnNameIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnName)

					return
				}
				activateTransactionsInputField("edit name:", selectedProfile.TX[i].Name)
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						break
					default:
						selectedProfile.TX[i].Name = transactionsInputField.GetText()

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Name = selectedProfile.TX[i].Name
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnName, selectedProfile.TX[i].Name))
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
				activateTransactionsInputField("weekly|monthly|yearly:", selectedProfile.TX[i].Frequency)
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
						transactionsInputField.SetLabel("invalid value - can only be weekly, monthly, or yearly:")

						return
					}
					selectedProfile.TX[i].Frequency = validatedFrequency

					// update all selected values as well as the current one
					for j := range selectedProfile.TX {
						if selectedProfile.TX[j].Selected || j == i {
							selectedProfile.TX[j].Frequency = selectedProfile.TX[i].Frequency
							transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnFrequency, selectedProfile.TX[i].Frequency))
						}
					}

					modified()
				}
				transactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
					return fuzzy.Find(strings.TrimSpace(strings.ToUpper(currentText)), []string{
						c.MONTHLY,
						c.YEARLY,
						c.WEEKLY,
					})
				})
				transactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
					saveFunc(text)
					deactivateTransactionsInputField()

					return true
				})
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						break
					default:
						saveFunc(transactionsInputField.GetText())
					}
					deactivateTransactionsInputField()
				})
			case c.ColumnIntervalIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnInterval)

					return
				}
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()

						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 {
							activateTransactionsInputFieldNoAutocompleteReset(
								"invalid interval given:",
								strconv.Itoa(selectedProfile.TX[i].Interval),
							)

							return
						}

						selectedProfile.TX[i].Interval = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Interval = selectedProfile.TX[i].Interval
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnInterval,
									selectedProfile.TX[i].Interval,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				activateTransactionsInputField("interval:", strconv.Itoa(selectedProfile.TX[i].Interval))
			case c.ColumnMondayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnMonday)

					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayMondayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayMondayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayMondayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayMondayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnMonday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayMondayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnMonday, cellText))
					}
				}

				modified()
			case c.ColumnTuesdayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnTuesday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayTuesdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayTuesdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayTuesdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayTuesdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnTuesday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayTuesdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnTuesday, cellText))
					}
				}

				modified()
			case c.ColumnWednesdayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnWednesday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayWednesdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayWednesdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayWednesdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayWednesdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnWednesday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayWednesdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnWednesday, cellText))
					}
				}

				modified()
			case c.ColumnThursdayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnThursday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayThursdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayThursdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayThursdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayThursdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnThursday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayThursdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnThursday, cellText))
					}
				}

				modified()
			case c.ColumnFridayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnFriday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdayFridayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdayFridayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdayFridayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdayFridayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnFriday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayFridayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnFriday, cellText))
					}
				}

				modified()
			case c.ColumnSaturdayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnSaturday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdaySaturdayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdaySaturdayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdaySaturdayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdaySaturdayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnSaturday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdaySaturdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnSaturday, cellText))
					}
				}

				modified()
			case c.ColumnSundayIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnSunday)
					return
				}

				selectedProfile.TX[i].Weekdays = lib.ToggleDayFromWeekdays(selectedProfile.TX[i].Weekdays, c.WeekdaySundayInt)

				dayIsPresent := slices.Contains(selectedProfile.TX[i].Weekdays, c.WeekdaySundayInt)

				// update all selected values as well as the current one
				for j := range selectedProfile.TX {
					if selectedProfile.TX[j].Selected || j == i {
						dayIndex := slices.Index(selectedProfile.TX[j].Weekdays, c.WeekdaySundayInt)
						if dayIndex == -1 && dayIsPresent {
							selectedProfile.TX[j].Weekdays = append(selectedProfile.TX[j].Weekdays, c.WeekdaySundayInt)
						} else if dayIndex != -1 && !dayIsPresent {
							selectedProfile.TX[j].Weekdays = slices.Delete(selectedProfile.TX[j].Weekdays, dayIndex, dayIndex+1)
						}
						sort.Ints(selectedProfile.TX[j].Weekdays)

						cellText := fmt.Sprintf("%v✔", c.ColorColumnSunday)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdaySundayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnSunday, cellText))
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
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 || d > 31 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset(
								"invalid day given:",
								strconv.Itoa(selectedProfile.TX[i].StartsDay),
							)
							return
						}

						selectedProfile.TX[i].StartsDay = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsDay = selectedProfile.TX[i].StartsDay
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnStarts,
									selectedProfile.TX[j].GetStartDateString(),
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
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d > 12 || d < 0 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", strconv.Itoa(selectedProfile.TX[i].StartsMonth))
							return
						}

						selectedProfile.TX[i].StartsMonth = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsMonth = selectedProfile.TX[i].StartsMonth
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnStarts,
									selectedProfile.TX[j].GetStartDateString(),
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", strconv.Itoa(selectedProfile.TX[i].StartsDay))
						defer transactionsInputField.SetDoneFunc(dayFunc)
					}
				}

				yearFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", strconv.Itoa(selectedProfile.TX[i].StartsYear))
							return
						}

						selectedProfile.TX[i].StartsYear = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsYear = selectedProfile.TX[i].StartsYear
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnStarts,
									selectedProfile.TX[j].GetStartDateString(),
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", strconv.Itoa(selectedProfile.TX[i].StartsMonth))
						defer transactionsInputField.SetDoneFunc(monthFunc)
					}
				}

				transactionsInputField.SetDoneFunc(yearFunc)
				activateTransactionsInputField("year:", strconv.Itoa(selectedProfile.TX[i].StartsYear))
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
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 || d > 31 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid day given:", strconv.Itoa(selectedProfile.TX[i].EndsDay))
							return
						}

						selectedProfile.TX[i].EndsDay = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsDay = selectedProfile.TX[i].EndsDay
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnEnds,
									selectedProfile.TX[j].GetEndsDateString(),
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
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d > 12 || d < 0 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", strconv.Itoa(selectedProfile.TX[i].EndsMonth))
							return
						}

						selectedProfile.TX[i].EndsMonth = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsMonth = selectedProfile.TX[i].EndsMonth
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnEnds,
									selectedProfile.TX[j].GetEndsDateString(),
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", strconv.Itoa(selectedProfile.TX[i].EndsDay))
						defer transactionsInputField.SetDoneFunc(dayFunc)
					}
				}

				yearFunc := func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						// don't save the changes
						deactivateTransactionsInputField()
						return
					default:
						d, err := strconv.ParseInt(transactionsInputField.GetText(), 10, 64)
						if err != nil || d < 0 {
							// start over
							activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", strconv.Itoa(selectedProfile.TX[i].EndsYear))
							return
						}

						selectedProfile.TX[i].EndsYear = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsYear = selectedProfile.TX[i].EndsYear
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.ColorColumnEnds,
									selectedProfile.TX[j].GetEndsDateString(),
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", strconv.Itoa(selectedProfile.TX[i].EndsMonth))
						defer transactionsInputField.SetDoneFunc(monthFunc)
					}
				}

				transactionsInputField.SetDoneFunc(yearFunc)
				activateTransactionsInputField("year:", strconv.Itoa(selectedProfile.TX[i].EndsYear))
			case c.ColumnNoteIndex:
				if row == 0 {
					setTransactionsTableSort(c.ColumnNote)
					return
				}
				activateTransactionsInputField("edit note:", selectedProfile.TX[i].Note)
				transactionsInputField.SetDoneFunc(func(key tcell.Key) {
					switch key {
					case tcell.KeyEscape:
						break
					default:
						// save the changes
						selectedProfile.TX[i].Note = transactionsInputField.GetText()
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Note = selectedProfile.TX[i].Note
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.ColorColumnNote, selectedProfile.TX[j].Note))
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

	transactionsTable.SetTitle("Transactions")
	transactionsTable.SetBorders(false).
		SetSelectable(true, true). // set row & cells to be selectable
		SetSeparator(' ')
}

func resetTransactionsInputFieldAutocomplete() {
	transactionsInputField.SetAutocompleteFunc(func(currentText string) []string {
		return []string{}
	})
}

func deactivateTransactionsInputField() {
	transactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	transactionsInputField.SetLabel("[gray] editor appears here when editing")
	transactionsInputField.SetText("")

	if previous != nil {
		FP.App.SetFocus(previous)
	}
}

// focuses the transactions input field, updates its label, and sets
// its background color to something noticeable
func activateTransactionsInputField(msg, value string) {
	resetTransactionsInputFieldAutocomplete()

	transactionsInputField.SetFieldBackgroundColor(tcell.ColorDimGray)
	transactionsInputField.SetLabel(fmt.Sprintf("[lightgreen::b] %v[-:-:-:-]", msg))
	transactionsInputField.SetText(value)

	// don't mess with the previously stored focus if the text field is already
	// focused
	currentFocus := FP.App.GetFocus()
	if currentFocus == transactionsInputField {
		return
	}

	previous = currentFocus

	FP.App.SetFocus(transactionsInputField)
}

// focuses the transactions input field, updates its label, and sets
// its background color to something noticeable - in some cases, the
// resetTransactionsInputFieldAutocomplete cannot be called without risking
// an infinite loop, so this function does not call it
func activateTransactionsInputFieldNoAutocompleteReset(msg, value string) {
	transactionsInputField.SetFieldBackgroundColor(tcell.ColorDimGray)
	transactionsInputField.SetLabel(fmt.Sprintf("[lightgreen::b] %v[-:-:-:-]", msg))
	transactionsInputField.SetText(value)

	// don't mess with the previously stored focus if the text field is already
	// focused
	currentFocus := FP.App.GetFocus()
	if currentFocus == transactionsInputField {
		return
	}

	previous = currentFocus

	FP.App.SetFocus(transactionsInputField)
}

func setStatusNoChanges() {
	statusText.SetText("[gray] no changes")
}

// sets the selectedProfile & config to the value specified by the current undo
// buffer
//
// warning: naively assumes that the undoBufferPos has already been set to a
// valid value and updates the currently selected config & profile accordingly
func pushUndoBufferChange() {
	n := selectedProfile.Name

	err := yaml.Unmarshal(undoBuffer[undoBufferPos], &config)
	if err != nil {
		statusText.SetText("[red]config unmarshal failure")
	}

	// set the selectedProfile to the latest undoBuffer's config
	for i := range config.Profiles {
		if config.Profiles[i].Name == n {
			selectedProfile = &(config.Profiles[i])
		}
	}
}

// moves 1 step backward in the undoBuffer
func undo() {
	undoBufferLen := len(undoBuffer)
	newUndoBufferPos := undoBufferPos - 1

	if newUndoBufferPos < 0 {
		// nothing to undo - at beginning of undoBuffer
		statusText.SetText(fmt.Sprintf("[gray] nothing to undo [%v/%v]", undoBufferPos+1, undoBufferLen))
		return
	}

	undoBufferPos = newUndoBufferPos

	pushUndoBufferChange()

	statusText.SetText(fmt.Sprintf("[gray] undo: %v/%v", undoBufferPos+1, undoBufferLen))

	populateProfilesPage()
	getTransactionsTable()
	transactionsTable.Select(selectedProfile.SelectedRow, selectedProfile.SelectedColumn)
	FP.App.SetFocus(transactionsTable)
}

// moves 1 step forward in the undoBuffer
func redo() {
	undoBufferLen := len(undoBuffer)
	undoBufferLastPos := undoBufferLen - 1
	newUndoBufferPos := undoBufferPos + 1

	if newUndoBufferPos > undoBufferLastPos {
		// nothing to redo - at end of undoBuffer
		statusText.SetText(fmt.Sprintf("[gray] nothing to redo [%v/%v]", undoBufferPos+1, undoBufferLen))
		return
	}

	undoBufferPos = newUndoBufferPos

	pushUndoBufferChange()

	statusText.SetText(fmt.Sprintf("[gray] redo: [%v/%v]", undoBufferPos+1, undoBufferLen))

	populateProfilesPage()
	getTransactionsTable()
	transactionsTable.Select(selectedProfile.SelectedRow, selectedProfile.SelectedColumn)
	FP.App.SetFocus(transactionsTable)
}

// attempts to place the current config at undoBuffer[undoBufferPos+1]
// but only if there were actual changes.
//
// also updates the status text accordingly
func modified() {
	if selectedProfile == nil {
		return
	}

	selectedProfile.Modified = true
	cr, cc := transactionsTable.GetSelection()
	selectedProfile.SelectedColumn = cc
	selectedProfile.SelectedRow = cr

	// marshal to detect differences between this config and the latest
	// config in the undo buffer
	if len(undoBuffer) >= 1 {
		b, err := yaml.Marshal(config)
		if err != nil {
			statusText.SetText("[yellow] failed to marshal config")
		}

		// TODO: it's probably not necessary to render these as strings for
		// the comparison
		mb := fmt.Sprintf("%x", md5.Sum(b))
		mo := fmt.Sprintf("%x", md5.Sum(undoBuffer[undoBufferPos]))

		if mb == mo {
			// no difference between this config and previous one
			statusText.SetText(fmt.Sprintf("[gray] no change [%v/%v]", undoBufferPos+1, len(undoBuffer)))
			// setStatusNoChanges()
			return
		}
	}

	// if the undoBufferPos is not at the end of the undoBuffer, then all
	// values after undoBufferPos need to be deleted
	if undoBufferPos != len(undoBuffer)-1 {
		undoBuffer = slices.Delete(undoBuffer, undoBufferPos, len(undoBuffer))
	}

	// err := lib.ValidateTransactions(&selectedProfile.TX)
	// if err != nil {
	// 	statusText.SetText("[red] unable to auto-order")
	// }
	getTransactionsTable()

	// now that we've ensured that we are actually at the end of the buffer,
	// proceed to insert this config into the undoBuffer
	b, err := yaml.Marshal(config)
	if err != nil {
		statusText.SetText("[red] cannot marshal config")
	}

	undoBuffer = append(undoBuffer, b)
	undoBufferPos = len(undoBuffer) - 1

	pushUndoBufferChange()
	statusText.SetText(fmt.Sprintf("[white] %v*[gray][%v/%v]", configFile, undoBufferPos+1, len(undoBuffer)))
}

// returns a simple flex view with two columns:
// - a list of profiles (left side)
// - a quick summary of bills / stats for the highlighted profile (right side)
func getProfilesPage() *tview.Flex {
	profileList = tview.NewList()
	profileList.SetBorder(true)
	profileList.ShowSecondaryText(false).
		SetSelectedBackgroundColor(tcell.NewRGBColor(50, 50, 50)).
		SetSelectedTextColor(tcell.ColorWhite).
		SetTitle("Profiles")

	statusText = tview.NewTextView()
	statusText.SetBorder(true)
	statusText.SetDynamicColors(true)
	setStatusNoChanges()

	profilesLeftSide := tview.NewFlex().SetDirection(tview.FlexRow)
	profilesLeftSide.AddItem(profileList, 0, 1, true).
		AddItem(statusText, 3, 0, true)

	transactionsTable = tview.NewTable().SetFixed(1, 1)
	transactionsInputField = tview.NewInputField()

	transactionsTable.SetBorder(true)
	transactionsInputField.SetBorder(true)

	transactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	transactionsInputField.SetLabel("[gray] editor appears here when editing")

	populateProfilesPage()
	getTransactionsTable()

	transactionsPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(transactionsTable, 0, 1, false).
		AddItem(transactionsInputField, 3, 0, false)

	return tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(profilesLeftSide, 0, 1, true).
		AddItem(transactionsPage, 0, 10, false)
}

func setTransactionsTableSort(column string) {
	FP.SortTX = lib.GetNextSort(FP.SortTX, column)

	getTransactionsTable()
}

// completely rebuilds the results form, safe to run repeatedly
func updateResultsForm() {
	resultsForm.Clear(true)
	resultsForm.SetTitle("Parameters")

	if selectedProfile == nil {
		return
	}

	setSelectedProfileDefaults()

	resultsForm.
		AddInputField("Start Year:", selectedProfile.StartYear, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 {
				return false
			}
			return true
		}, func(text string) { selectedProfile.StartYear = text }).
		AddInputField("Start Month:", selectedProfile.StartMonth, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 12 {
				return false
			}
			return true
		}, func(text string) { selectedProfile.StartMonth = text }).
		AddInputField("Start Day:", selectedProfile.StartDay, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 31 {
				return false
			}
			return true
		}, func(text string) { selectedProfile.StartDay = text }).
		AddInputField("End Year:", selectedProfile.EndYear, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 {
				return false
			}
			return true
		}, func(text string) { selectedProfile.EndYear = text }).
		AddInputField("End Month:", selectedProfile.EndMonth, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 12 {
				return false
			}
			return true
		}, func(text string) { selectedProfile.EndMonth = text }).
		AddInputField("End Day:", selectedProfile.EndDay, 0, func(textToCheck string, lastChar rune) bool {
			i, err := strconv.ParseInt(textToCheck, 10, 64)
			if err != nil || i < 0 || i > 31 {
				return false
			}
			return true
		}, func(text string) { selectedProfile.EndDay = text }).
		AddInputField("Starting Balance:", selectedProfile.StartingBalance, 0, nil, func(text string) {
			selectedProfile.StartingBalance = lib.FormatAsCurrency(int(lib.ParseDollarAmount(text, true)))
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

	resultsForm.SetLabelColor(tcell.ColorViolet)
	resultsForm.SetFieldBackgroundColor(tcell.NewRGBColor(40, 40, 40))
	resultsForm.SetBorder(true)
}

func getResultsPage() *tview.Flex {
	resultsTable = tview.NewTable().SetFixed(1, 1)
	resultsDescription = tview.NewTextView()
	resultsForm = tview.NewForm()

	resultsTable.SetBorder(true)
	resultsDescription.SetBorder(true)
	resultsDescription.SetDynamicColors(true)

	updateResultsForm()

	resultsTable.SetTitle("Results")
	resultsTable.SetBorders(false).
		SetSelectable(true, false). // set row & cells to be selectable
		SetSeparator(' ')

	resultsRightSide = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(resultsTable, 0, 2, true).
		AddItem(resultsDescription, 0, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(resultsForm, 0, 1, true).
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

	selectedProfile.StartYear = strconv.Itoa(start.Year())
	selectedProfile.StartMonth = strconv.Itoa(int(start.Month()))
	selectedProfile.StartDay = strconv.Itoa(start.Day())

	selectedProfile.EndYear = strconv.Itoa(end.Year())
	selectedProfile.EndMonth = strconv.Itoa(int(end.Month()))
	selectedProfile.EndDay = strconv.Itoa(end.Day())
}

// sets sensible default values for the currently selected profile, if they are
// not defined. If there is no selectedProfile, this will do nothing
func setSelectedProfileDefaults() {
	if selectedProfile == nil {
		return
	}

	now := time.Now()
	yr := now.Add(time.Hour * 24 * 365)

	if selectedProfile.StartYear == "" {
		selectedProfile.StartYear = strconv.Itoa(now.Year())
	}

	if selectedProfile.StartMonth == "" {
		selectedProfile.StartMonth = strconv.Itoa(int(now.Month()))
	}

	if selectedProfile.StartDay == "" {
		selectedProfile.StartDay = strconv.Itoa(now.Day())
	}

	if selectedProfile.EndYear == "" {
		selectedProfile.EndYear = strconv.Itoa(yr.Year())
	}

	if selectedProfile.EndMonth == "" {
		selectedProfile.EndMonth = strconv.Itoa(int(yr.Month()))
	}

	if selectedProfile.EndDay == "" {
		selectedProfile.EndDay = strconv.Itoa(yr.Day())
	}

	if selectedProfile.StartingBalance == "" {
		selectedProfile.StartingBalance = lib.FormatAsCurrency(50000)
	}
}

func getResultsTable() {
	if FP.CalculatingResults {
		return
	}

	FP.CalculatingResults = true

	go func() {
		resultsTable.Clear()
		resultsDescription.Clear()
		resultsDescription.SetText("[gray] calculating results, please wait...[-]")

		setSelectedProfileDefaults()

		// get results
		results, err := lib.GenerateResultsFromDateStrings(
			&(selectedProfile.TX),
			int(lib.ParseDollarAmount(selectedProfile.StartingBalance, true)),
			lib.GetDateString(selectedProfile.StartYear, selectedProfile.StartMonth, selectedProfile.StartDay),
			lib.GetDateString(selectedProfile.EndYear, selectedProfile.EndMonth, selectedProfile.EndDay),
			func(status string) {
				if config.DisableResultsStatusMessages {
					return
				}
				if resultsDescription != nil {
					go func() {
						FP.App.QueueUpdateDraw(func() {
							resultsDescription.SetText(fmt.Sprintf("[gray]%v", status))
						})
					}()
				}
			},
		)
		if err != nil {
			// TODO: add better error display
			panic(err)
		}

		// this may help with garbage collection when working with bigger data
		if latestResults != nil {
			if *latestResults != nil {
				clear(*latestResults)
				(*latestResults) = nil
			}

			latestResults = nil
		}

		latestResults = &results

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

		resultsTable.SetCell(0, 0, hDate)
		resultsTable.SetCell(0, 1, hBalance)
		resultsTable.SetCell(0, 2, hCumulativeIncome)
		resultsTable.SetCell(0, 3, hCumulativeExpenses)
		resultsTable.SetCell(0, 4, hDayExpenses)
		resultsTable.SetCell(0, 5, hDayIncome)
		resultsTable.SetCell(0, 6, hDayNet)
		resultsTable.SetCell(0, 7, hDiffFromStart)
		resultsTable.SetCell(0, 7, hDiffFromStart)
		resultsTable.SetCell(0, 8, hDayTransactionNames)

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

			resultsTable.SetCell(i+1, 0, rDate)
			resultsTable.SetCell(i+1, 1, rBalance)
			resultsTable.SetCell(i+1, 2, rCumulativeIncome)
			resultsTable.SetCell(i+1, 3, rCumulativeExpenses)
			resultsTable.SetCell(i+1, 4, rDayExpenses)
			resultsTable.SetCell(i+1, 5, rDayIncome)
			resultsTable.SetCell(i+1, 6, rDayNet)
			resultsTable.SetCell(i+1, 7, rDiffFromStart)
			resultsTable.SetCell(i+1, 8, rDayTransactionNames)
		}

		resultsTable.SetSelectionChangedFunc(func(row, column int) {
			if row <= 0 {
				return
			}
			resultsDescription.Clear()
			// ensure there are enough results before trying to show something
			if len(*latestResults)-1 > row-1 {
				var sb strings.Builder
				for _, t := range (*latestResults)[row-1].DayTransactionNamesSlice {
					sb.WriteString(fmt.Sprintf("%v\n", t))
				}
				resultsDescription.SetText(sb.String())
			}
		})

		getResultsStats()

		FP.CalculatingResults = false

		FP.App.SetFocus(resultsTable)
	}()
}

// Populates the results description with basic statistics about the results,
// and queues an UpdateDraw
func getResultsStats() {
	go FP.App.QueueUpdateDraw(func() {
		if latestResults == nil {
			return
		}

		stats, err := lib.GetStats(*latestResults)
		if err != nil {
			resultsDescription.SetText(
				fmt.Sprintf("error getting stats: %v", err.Error()),
			)
		}

		resultsDescription.SetText(stats)
	})
}

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

	promptBox.ClearButtons().AddButtons(
		[]string{
			"I am sure, please exit",
			"No",
			"Cancel",
		},
	).SetText("Really quit?").SetDoneFunc(
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
		SetTextColor(tcell.ColorBlack) //.
	// SetButtonTextColor(tcell.ColorGray) //.
	// SetButtonBackgroundColor(tcell.NewRGBColor(100, 100, 100))

	FP.Pages.SwitchToPage(PagePrompt)
	promptBox.SetFocus(2)
	FP.App.SetFocus(promptBox)
}

// promptKBMode switches to the prompt page and shows a modal that informs the
// user that they are in keyboard echo mode. If KB echo mode is not enabled,
// this gracefully returns immediately and does nothing.
//
// Requires the first argument to be the translation map.
func promptKBMode(t map[string]string) {
	if !keyboardEchoMode {
		return
	}

	// temporarily turn off KB echo mode so that the user's keys are captured
	// properly until they can give consent to entering the mode
	keyboardEchoMode = false

	promptBox.ClearButtons().AddButtons(
		[]string{
			t["PromptKeyboardEchoModeButtonTurnOff"],
			t["PromptKeyboardEchoModeButtonExitNow"],
			t["PromptKeyboardEchoModeButtonContinue"],
		},
	).SetText(t["PromptKeyboardEchoModeText"]).SetDoneFunc(
		func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				keyboardEchoMode = false
				FP.Pages.SwitchToPage(PageProfiles)
			case 1:
				keyboardEchoMode = false
				FP.App.Stop()
			case 2:
				keyboardEchoMode = true
				FP.Pages.SwitchToPage(PageProfiles)
			default:
				keyboardEchoMode = false
				FP.App.Stop()
				return
			}
		},
	).SetBackgroundColor(tcell.ColorDimGray).
		SetTextColor(tcell.ColorWhite)

	FP.Pages.SwitchToPage(PagePrompt)
	promptBox.SetFocus(2)
	FP.App.SetFocus(promptBox)
}

// For an input keybinding (straight from event.Name()), an output action
// will be returned, for example - "Ctrl+Z" will return "undo".
func getDefaultKeybind(name string) string {
	m, ok := c.DefaultMappings[name]
	if !ok {
		return ""
	}

	return m
}

func actionRedo(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case transactionsInputField:
			return e
		case transactionsTable:
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
		case transactionsInputField:
			return e
		case transactionsTable:
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
		case transactionsInputField:
			return e
		case transactionsTable:
			// move all selected items to the currently selected row:
			// delete items, then re-add the items after the current
			// row, then highlight the correct row

			if FP.SortTX != c.None && FP.SortTX != "" {
				statusText.SetText(fmt.Sprintf("[orange]sort: %v", FP.SortTX))
				return nil
			}

			// but first, check if any items are selected at all
			anySelected := false

			for i := range selectedProfile.TX {
				if selectedProfile.TX[i].Selected {
					anySelected = true

					break
				}
			}

			if !anySelected {
				statusText.SetText("[gray]nothing to move")
				return nil
			}

			setTransactionsTableSort(c.None)

			// get the height & width of the transactions table
			cr, cc := transactionsTable.GetSelection()
			actual := cr - 1 // skip header

			// take note of the currently selected value (cannot be
			// a candidate for move/deletion since it is the target
			// for the move)
			txid := selectedProfile.TX[actual].ID

			// first delete the values from the slice and keep track of
			// them
			deleted := []lib.TX{}
			newTX := []lib.TX{}

			for i := range selectedProfile.TX {
				if selectedProfile.TX[i].ID == txid {
					// this is the target to move to
					selectedProfile.TX[i].Selected = false
					newTX = append(newTX, selectedProfile.TX[i])
				} else if selectedProfile.TX[i].Selected {
					selectedProfile.TX[i].Selected = true
					deleted = append(deleted, selectedProfile.TX[i])
				} else {
					selectedProfile.TX[i].Selected = false
					newTX = append(newTX, selectedProfile.TX[i])
				}
			}

			selectedProfile.TX = newTX

			// find the move target now that the slice has been shifted
			newPosition := 0

			for i := range selectedProfile.TX {
				if selectedProfile.TX[i].ID == txid {
					newPosition = i + 1

					break
				}
			}

			if newPosition >= len(selectedProfile.TX) {
				newPosition = len(selectedProfile.TX)
			} else if newPosition < 0 {
				newPosition = 0
			}

			FP.LastSelection = newPosition

			selectedProfile.TX = slices.Insert(selectedProfile.TX, newPosition, deleted...)

			modified()

			// re-render the table
			getTransactionsTable()

			// check that we aren't going to move the selection past the
			// final row
			newPosition++

			r := transactionsTable.GetRowCount()

			if newPosition >= r {
				newPosition = r - 1
			}

			transactionsTable.Select(newPosition, cc) // offset for headers
			FP.App.SetFocus(transactionsTable)
		default:
			FP.App.SetFocus(profileList)
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
		case transactionsInputField:
			return e
		case transactionsTable:
			cr, cc := transactionsTable.GetSelection()
			// get the height & width of the transactions table
			actual := cr - 1 // skip header
			if multiSelecting {
				// shift modifier is used to extend the selection
				// from the previously selected index to the current
				newSelectionValue := false
				// start by finding the currently highlighted TX
				for i := range selectedProfile.TX {
					if i == actual {
						newSelectionValue = !selectedProfile.TX[i].Selected

						break
					}
				}

				if FP.LastSelection == -1 {
					FP.LastSelection = actual
				}

				// now that we've determined what the selection value
				// should be, proceed to apply it to every value from
				// FP.LastSelection to the current index
				for i := range selectedProfile.TX {
					// last=5, current=10, select from 5-10 => last < i < actual
					// last=10, current=3, select from 3-10 => last > i > actual
					shouldModify := (FP.LastSelection < i && i <= actual) || (FP.LastSelection > i && i >= actual)
					if shouldModify {
						selectedProfile.TX[i].Selected = newSelectionValue
					}
				}
			} else {
				for i := range selectedProfile.TX {
					if i == actual {
						selectedProfile.TX[i].Selected = !selectedProfile.TX[i].Selected

						break
					}
				}
			}

			FP.LastSelection = actual

			modified()
			getTransactionsTable()
			transactionsTable.Select(cr, cc)
			FP.App.SetFocus(transactionsTable)

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
		case transactionsInputField:
			return e
		case transactionsTable:
			// duplicate the current transaction
			// get the height & width of the transactions table
			cr, cc := transactionsTable.GetSelection()
			actual := cr - 1 // skip header

			for i := len(selectedProfile.TX) - 1; i >= 0; i-- {
				if selectedProfile.TX[i].Selected || i == actual {
					selectedProfile.TX = slices.Delete(selectedProfile.TX, i, i+1)
				}
			}

			getTransactionsTable()
			transactionsTable.Select(cr, cc)
			FP.App.SetFocus(transactionsTable)
		case profileList:
			if len(config.Profiles) <= 1 {
				statusText.SetText("[gray] can't delete last profile")
				return nil
			}

			getPrompt := func() string {
				if selectedProfile == nil {
					return "no profile selected; please cancel this operation"
				}

				return fmt.Sprintf(
					"[gold::b]confirm deletion of profile %v by typing 'delete %v':%v",
					selectedProfile.Name,
					selectedProfile.Name,
					c.ResetStyle,
				)
			}

			transactionsInputField.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEscape:
					// don't save the changes
					deactivateTransactionsInputField()
					return
				default:
					// validate that the name is unique
					value := transactionsInputField.GetText()
					if strings.Index(value, "delete ") != 0 {
						transactionsInputField.SetLabel(getPrompt())
						return
					}

					profileName := strings.TrimPrefix(value, "delete ")
					if profileName != selectedProfile.Name {
						transactionsInputField.SetLabel(getPrompt())
						return
					}

					// proceed to delete the profile
					for i := range config.Profiles {
						if profileName == config.Profiles[i].Name {
							config.Profiles = slices.Delete(config.Profiles, i, i+1)
							return
						}
					}

					selectedProfile = &(config.Profiles[0])

					// config.Profiles = append(config.Profiles, newProfile)
					modified()
					deactivateTransactionsInputField()
					populateProfilesPage()
					getTransactionsTable()
					transactionsTable.Select(0, 0)
					FP.App.SetFocus(profileList)
				}
			})
			activateTransactionsInputField(getPrompt(), "")
		default:
			FP.App.SetFocus(profileList)
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
		case transactionsInputField:
			return e
		case transactionsTable:
			cr, cc := transactionsTable.GetSelection()
			actual := cr - 1 // skip header
			nt := []lib.TX{}

			FP.LastSelection = -1

			if !duplicating {
				// largestOrderHolder := []lib.TX{}
				// largestOrderHolder = append(largestOrderHolder, selectedProfile.TX...)
				// largestOrderHolder = append(largestOrderHolder, nt...)
				newTX := lib.GetNewTX()
				// newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1
				nt = append(nt, newTX)
			} else {
				// iterate through the list once to find how many selected
				// items there are
				numSelected := 0
				for i := range selectedProfile.TX {
					if selectedProfile.TX[i].Selected {
						numSelected++

						// we only care about knowing whether or not there
						// is more than 1 item selected
						if numSelected > 1 {
							break
						}
					}
				}

				for i := range selectedProfile.TX {
					isHighlightedRow := i == actual && numSelected <= 1
					isSelectedDuplicationCandidate := selectedProfile.TX[i].Selected && duplicating
					if isHighlightedRow || isSelectedDuplicationCandidate {
						// keep track of the highest order in a temporary
						// slice
						// largestOrderHolder := []lib.TX{}
						// largestOrderHolder = append(largestOrderHolder, selectedProfile.TX...)
						// largestOrderHolder = append(largestOrderHolder, nt...)

						newTX := lib.GetNewTX()
						// newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1

						newTX.Amount = selectedProfile.TX[i].Amount
						newTX.Active = selectedProfile.TX[i].Active
						newTX.Name = selectedProfile.TX[i].Name
						newTX.Note = selectedProfile.TX[i].Note
						newTX.RRule = selectedProfile.TX[i].RRule
						newTX.Frequency = selectedProfile.TX[i].Frequency
						newTX.Interval = selectedProfile.TX[i].Interval
						newTX.Weekdays = selectedProfile.TX[i].Weekdays
						newTX.StartsDay = selectedProfile.TX[i].StartsDay
						newTX.StartsMonth = selectedProfile.TX[i].StartsMonth
						newTX.StartsYear = selectedProfile.TX[i].StartsYear
						newTX.EndsDay = selectedProfile.TX[i].EndsDay
						newTX.EndsMonth = selectedProfile.TX[i].EndsMonth
						newTX.EndsYear = selectedProfile.TX[i].EndsYear

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

				if len(selectedProfile.TX) == 0 || actual > len(selectedProfile.TX)-1 {
					selectedProfile.TX = append(selectedProfile.TX, nt...)
				} else {
					selectedProfile.TX = slices.Insert(selectedProfile.TX, actual, nt...)
				}

				modified()
				getTransactionsTable()
				transactionsTable.Select(cr, cc)
				FP.App.SetFocus(transactionsTable)
			}

			return e
		case profileList:
			// add/duplicate new profile
			transactionsInputField.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEscape:
					// don't save the changes
					deactivateTransactionsInputField()
					return
				default:
					// validate that the name is unique
					newProfileName := transactionsInputField.GetText()
					for i := range config.Profiles {
						if newProfileName == config.Profiles[i].Name {
							transactionsInputField.SetLabel("profile name must be unique:")
							return
						}
					}

					newProfile := *selectedProfile
					newProfile.Name = newProfileName
					if !duplicating {
						newProfile = m.Profile{Name: newProfileName}
					}

					selectedProfile = &newProfile

					config.Profiles = append(config.Profiles, newProfile)
					modified()
					deactivateTransactionsInputField()
					populateProfilesPage()
					getTransactionsTable()
					transactionsTable.Select(0, 0)
					FP.App.SetFocus(profileList)
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
		case transactionsInputField:
			return e
		case profileList:
			// add/duplicate new profile
			transactionsInputField.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEscape:
					// don't save the changes
					deactivateTransactionsInputField()
					return
				default:
					// validate that the name is unique
					newProfileName := transactionsInputField.GetText()
					for i := range config.Profiles {
						if newProfileName == config.Profiles[i].Name {
							transactionsInputField.SetLabel("profile name must be unique:")
							return
						}
					}

					selectedProfile.Name = newProfileName
					modified()
					deactivateTransactionsInputField()
					populateProfilesPage()
					getTransactionsTable()
					transactionsTable.Select(0, 0)
					FP.App.SetFocus(profileList)
				}
			})
			activateTransactionsInputField(fmt.Sprintf("set new unique profile name for %v:", selectedProfile.Name), "")

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
	if config.Version == "" {
		config.Version = c.ConfigVersion
	}

	b, err := yaml.Marshal(config)
	if err != nil {
		statusText.SetText("failed to marshal")
		return nil
	}

	err = os.WriteFile(configFile, b, os.FileMode(0o644))
	if err != nil {
		statusText.SetText("failed to save")
		return nil
	}

	selectedProfile.Modified = false

	statusText.SetText("[gray] saved changes")

	return nil
}

func actionEnd(e *tcell.EventKey) *tcell.EventKey {
	pageName, _ := FP.Pages.GetFrontPage()
	switch pageName {
	case PageProfiles:
		switch FP.App.GetFocus() {
		case transactionsTable:
			c := transactionsTable.GetColumnCount() - 1
			cr, _ := transactionsTable.GetSelection()
			transactionsTable.Select(cr, c)
			FP.App.SetFocus(transactionsTable)

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
		case transactionsTable:
			cr, _ := transactionsTable.GetSelection()
			transactionsTable.Select(cr, 0)
			FP.App.SetFocus(transactionsTable)

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
	case transactionsInputField:
		return nil
	default:
		return e
	}
}

func actionUp(e *tcell.EventKey) *tcell.EventKey {
	switch FP.App.GetFocus() {
	case transactionsInputField:
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
		case profileList:
			FP.App.SetFocus(transactionsTable)
			return nil
		case transactionsTable:
			_, cc := transactionsTable.GetSelection()
			// focus the profile list when at column 0
			if cc == 0 {
				FP.App.SetFocus(profileList)
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
		case profileList:
			FP.App.SetFocus(transactionsTable)
			return nil
		case transactionsTable:
			c := transactionsTable.GetColumnCount() - 1
			_, cc := transactionsTable.GetSelection()
			// focus the profile list when at max column
			if cc == c {
				FP.App.SetFocus(profileList)
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
		case resultsDescription:
			return e
		case resultsTable:
			return e
		default:
			FP.App.SetFocus(resultsTable)
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
		case resultsDescription:
			return e
		case resultsTable:
			return e
		default:
			FP.App.SetFocus(resultsTable)
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
		case transactionsInputField:
			return nil
		case profileList:
			FP.App.SetFocus(transactionsTable)
		case transactionsTable:
			// get the height & width of the transactions table
			// r := transactionsTable.GetRowCount() - 1
			c := transactionsTable.GetColumnCount() - 1
			cr, cc := transactionsTable.GetSelection()
			nc := cc - 1
			nr := cr

			var focusTarget tview.Primitive
			focusTarget = transactionsTable

			if nc < 0 {
				nc = c
				nr--

				if nr < 0 {
					// nc = c
					nc = 0
					// nr = r
					nr = 0
				}
				// it's more intuitive to go back to the profileList
				// when backtabbing from the first column in the table
				focusTarget = profileList
			}

			transactionsTable.Select(nr, nc)
			FP.App.SetFocus(focusTarget)
		default:
			FP.App.SetFocus(profileList)
		}

		return nil
	case PageResults:
		switch FP.App.GetFocus() {
		case resultsTable:
			resultsForm.SetFocus(0)
			FP.App.SetFocus(resultsForm)

			return nil
		case resultsDescription:
			FP.App.SetFocus(resultsTable)
		case resultsForm:
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
		case transactionsInputField:
			return nil
		case profileList:
			FP.App.SetFocus(transactionsTable)
		case transactionsTable:
			// get the height & width of the transactions table
			r := transactionsTable.GetRowCount() - 1
			c := transactionsTable.GetColumnCount() - 1
			cr, cc := transactionsTable.GetSelection()
			nc := cc + 1
			nr := cr

			var focusTarget tview.Primitive

			focusTarget = transactionsTable

			if nc > c {
				nc = 0 // loop around
				nr++

				if nr > r {
					nc = 0
					nr = r
				}
				// it's more intuitive to go back to the profileList
				// when backtabbing from the first column in the table
				focusTarget = profileList
			}

			transactionsTable.Select(nr, nc)
			FP.App.SetFocus(focusTarget)
		default:
			FP.App.SetFocus(profileList)
		}

		return nil
	case PageResults:
		switch FP.App.GetFocus() {
		case resultsTable:
			FP.App.SetFocus(resultsDescription)
		case resultsDescription:
			resultsForm.SetFocus(0)
			FP.App.SetFocus(resultsForm)

			return nil
		case resultsForm:
			return e
		}

		return e
	}

	return e
}

func actionEsc(e *tcell.EventKey) *tcell.EventKey {
	currentFocus := FP.App.GetFocus()
	switch currentFocus {
	case transactionsInputField:
		return e
	case transactionsTable:
		// deselect the last selected index on the first press
		if FP.LastSelection != -1 {
			FP.LastSelection = -1

			getTransactionsTable()

			cr, cc := transactionsTable.GetSelection()

			transactionsTable.Select(cr, cc)
			FP.App.SetFocus(transactionsTable)

			return nil
		}

		anySelected := false

		for i := range selectedProfile.TX {
			if selectedProfile.TX[i].Selected {
				anySelected = true
				selectedProfile.TX[i].Selected = false
			}
		}

		if !anySelected {
			FP.App.SetFocus(profileList)
			return nil
		}

		modified()

		getTransactionsTable()

		cr, cc := transactionsTable.GetSelection()

		transactionsTable.Select(cr, cc)
		FP.App.SetFocus(transactionsTable)
	case resultsForm:
		FP.App.SetFocus(resultsTable)
		return nil
	case resultsTable:
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
	setBottomHelpText()

	if alreadyOnPage {
		getResultsTable()
		FP.App.SetFocus(resultsTable)
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
	setBottomHelpText()

	if alreadyOnPage {
		FP.App.SetFocus(profileList)
	}

	return nil
}

func actionGlobalHelp() *tcell.EventKey {
	FP.Pages.SwitchToPage(PageHelp)
	setBottomHelpText()

	return nil
}

func actionHelp(e *tcell.EventKey) *tcell.EventKey {
	switch FP.App.GetFocus() {
	case transactionsInputField:
		return e
	case resultsForm:
		return e
	default:
		FP.Pages.SwitchToPage(PageHelp)
		setBottomHelpText()

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

// capture is the primary input capture handler for the app, and should be used
// like: app.SetInputCapture(capture)
func capture(e *tcell.EventKey) *tcell.EventKey {
	n := e.Name()
	if keyboardEchoMode {
		statusText.SetDynamicColors(false).SetText(n)

		if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC {
			FP.App.Stop()
		}

		return nil
	}

	var final *tcell.EventKey
	final = e

	foundBinding := false

	for binding, actions := range config.Keybindings {
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

	undoBuffer = [][]byte{b}
	undoBufferPos = 0

	FP.LastSelection = -1
	FP.App = tview.NewApplication()

	FP.Pages = tview.NewPages()

	getHelpModal()

	promptBox = tview.NewModal()

	FP.Pages.AddPage(PageProfiles, getProfilesPage(), true, true).
		AddPage(PageResults, getResultsPage(), true, true).
		AddPage(PageHelp, helpModal, true, true).
		AddPage(PagePrompt, promptBox, true, true)

	FP.Pages.SwitchToPage(PageProfiles)

	bottomHelpText = tview.NewTextView()

	bottomHelpText.SetDynamicColors(true)
	setBottomHelpText()

	FP.Layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(FP.Pages, 0, 1, true).AddItem(bottomHelpText, 1, 0, false)

	FP.App.SetFocus(profileList)

	promptKBMode(t)

	FP.App.SetInputCapture(capture)
}

// parseFlags parses the command line flags, using t as the translation map.
func parseFlags(t map[string]string) {
	flag.StringVar(&configFile, t["FlagConfigFileFlag"], "", t["FlagConfigFileDesc"])
	flag.BoolVar(&shouldMigrate, t["FlagShouldMigrateFlag"], false, t["FlagShouldMigrateDesc"])
	flag.BoolVar(&keyboardEchoMode, t["FlagKeyboardEchoModeFlag"], false, t["FlagKeyboardEchoModeDesc"])
	flag.Parse()
}

func main() {
	var err error

	FP.T, err = translations.Load(AllTranslations)
	if err != nil {
		log.Fatalf("failed to load translations: %v", err.Error())
	}

	parseFlags(FP.T)

	if shouldMigrate {
		JSONtoYAML()
	}

	config, err = loadConfig(configFile)
	if err != nil {
		log.Fatalf("%v: %v", FP.T["ErrorFailedToLoadConfig"], err.Error())
	}

	if len(config.Profiles) > 0 {
		selectedProfile = &(config.Profiles[0])
	} else {
		n := m.Profile{
			TX:   []lib.TX{lib.GetNewTX()},
			Name: FP.T["DefaultNewProfileName"],
		}
		config.Profiles = append(config.Profiles, n)
		selectedProfile = &(config.Profiles[0])
	}

	bootstrap(FP.T, config)

	if err := FP.App.SetRoot(FP.Layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
