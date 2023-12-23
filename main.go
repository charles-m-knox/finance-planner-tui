package main

import (
	"bytes"
	"crypto/md5"
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

	"github.com/adrg/xdg"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

const (
	// new
	PAGE_PROFILES = "Profiles"
	PAGE_RESULTS  = "Results"
	PAGE_HELP     = "Help"
	PAGE_PROMPT   = "Prompt"
)

var (
	app                         *tview.Application
	layout                      *tview.Flex
	config                      m.Config
	selectedProfile             *m.Profile
	previousPage                string
	pages                       *tview.Pages
	currentlyFocusedBox         string
	profilesPage                *tview.Flex
	transactionsPage            *tview.Flex
	resultsPage                 *tview.Flex
	transactionsTableSortColumn string
	lastSelectedIndex           int

	// usage example: allKeyBindings["Ctrl+Z"] = ["undo", "save"]
	allKeyBindings map[string][]string
	// usage example: allBoundActions["save"] = ["Ctrl+S", "[gold]Ctrl+X"]
	allBoundActions map[string][]string

	// the previously focused primitive
	previous tview.Primitive
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
	latestResults      *[]lib.Result

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

func init() {
	flag.StringVar(&configFile, "f", "", "the file to load from and save to")
	flag.BoolVar(&shouldMigrate, "migrate", false, "whether or not to migrate a file named conf.json in the current directory from a previous config version to the latest version and save it as migrated.yml")
	flag.BoolVar(&keyboardEchoMode, "kb", false, "if true, all keyboard inputs except escape and ctrl+c will be echoed to the bottom left text field of the page. Use this mode to test your preferred keyboard mappings - take the exact text and put it into your keybindings configuration.")
	flag.Parse()

	if shouldMigrate {
		JSONtoYAML()
	}
}

// merges the default keybindings with the user's customized keybindings.
//
// Example: "Ctrl+S": ["save"]
//
// Do not
// use outside of the context of documentation, because this will also modify
// things like Rune[x] to render properly within a dynamically colored textview.
// For example, Rune[x] will transform to Rune[x[].
func GetCombinedKeybindings(kb map[string][]string, def map[string]string) (r map[string][]string) {
	r = make(map[string][]string)
	reg := regexp.MustCompile(`^Rune\[.\]$`)
	for k, v := range def {
		if reg.MatchString(k) {
			r[strings.Replace(k, "]", "[]", 1)] = []string{v}
			continue
		}
		r[k] = []string{v}
	}

	for k, v := range kb {
		// higlight custom key bindings
		formattedActions := []string{}
		for _, action := range v {
			formattedActions = append(formattedActions, fmt.Sprintf("[gold::b]%v[-:-:-:-]", action))
		}
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
func GetAllBoundActions(kb map[string][]string, def map[string]string) (r map[string][]string) {
	r = make(map[string][]string)
	reg := regexp.MustCompile(`^Rune\[.\]$`)

	// handle default actions first
	for binding, action := range def {
		fixedBinding := string(binding)
		if reg.MatchString(fixedBinding) {
			fixedBinding = strings.Replace(fixedBinding, "]", "[]", 1)
		}
		// if reg.MatchString(binding) {
		// 	r[strings.Replace(binding, "]", "[]", 1)] = []string{action}
		// 	continue
		// }
		r[action] = []string{fixedBinding}
	}

	// higlight custom key bindings next
	for binding, actions := range kb {

		// formattedActions := []string{}

		color := "gold"
		if len(actions) > 1 {
			color = "#aaffee"
		}

		// for _, action := range actions {
		// 	formattedActions = append(formattedActions, fmt.Sprintf("[%v::b]%v[-:-:-:-]", color, action))
		// }

		fixedBinding := string(binding)
		if reg.MatchString(fixedBinding) {
			fixedBinding = strings.Replace(fixedBinding, "]", "[]", 1)
		}
		formattedBinding := fmt.Sprintf("[%v::b]%v[-:-:-:-]", color, fixedBinding)

		for _, action := range actions {
			r[action] = slices.Insert(r[action], 0, formattedBinding)
		}
		// for existingActions, existingBindings := range r {
		// 	if
		// }

		// delete the old keybinding and reformat it to show that it's customized
		// formattedKeybinding := fmt.Sprintf("[gold::b]%v[-:-:-:-]", binding)
		// delete(r, binding)

		// r[formattedKeybinding] = actions
	}

	return r
}

func getHelpText(conf m.Config, combinedKeybindings, combinedActions map[string][]string) (output string) {
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
		AllActions:          c.ALL_ACTIONS,
		DefaultKeybindings:  c.DEFAULT_MAPPINGS,
		CombinedKeybindings: combinedKeybindings,
		CombinedActions:     combinedActions,
		Explanations:        c.ACTION_EXPLANATIONS,
	}

	tmpl, err := template.New("help").Parse(c.HelpTextTemplate)
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
	helpModal.SetText(getHelpText(config, allKeyBindings, allBoundActions)).SetDynamicColors(true)
}

// returns the first configured keybinding for the provided action. returns
// "n/a" if none defined
func getBinding(action string) string {
	bindings, ok := allBoundActions[action]
	if !ok || len(bindings) < 1 {
		return ""
	}

	return bindings[0]
}

func setBottomHelpText() {
	p, _ := pages.GetFrontPage()

	var sb strings.Builder

	if p == PAGE_HELP {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gold] help [-:-:-:-]", getBinding(c.ACTION_GLOBAL_HELP)))
	} else {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gray] help [-:-:-:-]", getBinding(c.ACTION_GLOBAL_HELP)))
	}

	if p == PAGE_PROFILES {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gold] profiles & transactions [-:-:-:-]", getBinding(c.ACTION_PROFILES)))
	} else {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gray] profiles & transactions [-:-:-:-]", getBinding(c.ACTION_PROFILES)))
	}

	if p == PAGE_RESULTS {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gold] results [-:-:-:-]", getBinding(c.ACTION_RESULTS)))
	} else {
		sb.WriteString(fmt.Sprintf("%v[-:-:-:-][gray] results[-:-:-:-]", getBinding(c.ACTION_RESULTS)))
	}

	bottomHelpText.SetText(sb.String())
}

// attempts to load from the "file" path provided - if not successful,
// attempts to load from xdg config, then xdg home. Then it sets the global
// configFile to match the retrieved config
func loadConfig(file string) (conf m.Config, err error) {
	xdgConfig := path.Join(xdg.ConfigHome, "finance-planner-tui", "config.yml")
	xdgHome := path.Join(xdg.Home, "finance-planner-tui", "config.yml")

	specificFileGiven := true
	if file == "" {
		file = c.DEFAULT_CONFIG
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

	// return c, fmt.Errorf(
	// 	"failed to read config from %v, %v, and %v: %v",
	// 	configFile,
	// 	xdgConfig,
	// 	xdgHome,
	// 	err.Error(),
	// )
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
	err = os.WriteFile("migrated.yml", out, 0o644)
	if err != nil {
		log.Fatalf("failed to write migrated.yml: %v", err.Error())
	}
}

func getNowStr() string {
	return time.Now().Format("15:04:05")
}

func getActiveProfileText(profile m.Profile) string {
	profileText := fmt.Sprintf("%v", profile.Name)
	if selectedProfile != nil && selectedProfile.Name == profile.Name {
		profileText = fmt.Sprintf("[white::bu]%v (open)%v", profile.Name, c.RESET_STYLE)
	}

	return profileText
}

func populateProfilesPage(doProfile, doBills bool) {
	if doProfile {
		profileList.Clear()
	}
	// if doBills {
	// 	billsSummary.Clear()
	// }

	// var sb strings.Builder
	for i := range config.Profiles {
		profile := &(config.Profiles[i])
		if doProfile {
			profileList.AddItem(getActiveProfileText(*profile), "", 0, func() {
				selectedProfile = profile
				populateProfilesPage(true, true)
				getTransactionsTable()
				app.SetFocus(transactionsTable)
			})
		}

		if selectedProfile == nil || selectedProfile.Name != profile.Name {
			continue
		}

		if !doBills {
			continue
		}

		// for i, tx := range profile.TX {
		// 	// only show top 25 transactions for profile
		// 	if i > 25 {
		// 		break
		// 	}
		// 	if tx.Amount < 0 {
		// 		fmt.Fprintf(&sb, "[yellow]%v [gray]| [yellow]%v\n", tx.Name, lib.FormatAsCurrency(tx.Amount))
		// 	} else {
		// 		fmt.Fprintf(&sb, "[green]%v [gray]| [green]%v\n", tx.Name, lib.FormatAsCurrency(tx.Amount))
		// 	}
		// }
	}

	// billsSummary.SetText(sb.String())
}

func sortTX() {
	sort.SliceStable(
		selectedProfile.TX,
		func(i, j int) bool {
			tj := (selectedProfile.TX)[j]
			ti := (selectedProfile.TX)[i]

			switch transactionsTableSortColumn {

			// invisible order column (default when no sort is set)
			case c.None:
				return tj.Order > ti.Order

			// Order
			case fmt.Sprintf("%v%v", c.ColumnOrder, c.Asc):
				return ti.Order > tj.Order
			case fmt.Sprintf("%v%v", c.ColumnOrder, c.Desc):
				return ti.Order < tj.Order

			// active
			case fmt.Sprintf("%v%v", c.ColumnActive, c.Asc):
				if ti.Active == tj.Active {
					return ti.Order > tj.Order
				}
				return ti.Active
			case fmt.Sprintf("%v%v", c.ColumnActive, c.Desc):
				if ti.Active == tj.Active {
					return ti.Order < tj.Order
				}
				return tj.Active

			// weekdays
			case fmt.Sprintf("%v%v", c.WeekdayMonday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayMondayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayMondayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayMonday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayMondayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayMondayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
				}
				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayTuesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayTuesdayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayTuesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayTuesdayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
				}
				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayWednesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayWednesdayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayWednesdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayWednesdayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
				}
				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayThursday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayThursdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayThursdayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayThursday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayThursdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayThursdayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
				}
				return tjw

			case fmt.Sprintf("%v%v", c.WeekdayFriday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayFridayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayFridayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdayFriday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdayFridayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdayFridayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
				}
				return tjw

			case fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySaturdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySaturdayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySaturdayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySaturdayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
				}
				return tjw

			case fmt.Sprintf("%v%v", c.WeekdaySunday, c.Asc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySundayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySundayInt) != -1
				if tiw == tjw {
					return ti.Order > tj.Order
				}
				return tiw

			case fmt.Sprintf("%v%v", c.WeekdaySunday, c.Desc):
				tiw := slices.Index(ti.Weekdays, c.WeekdaySundayInt) != -1
				tjw := slices.Index(tj.Weekdays, c.WeekdaySundayInt) != -1
				if tiw == tjw {
					return ti.Order < tj.Order
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
				ist := fmt.Sprintf("%v-%v-%v", tj.StartsYear, tj.StartsMonth, tj.StartsDay)
				jst := fmt.Sprintf("%v-%v-%v", ti.StartsYear, ti.StartsMonth, ti.StartsDay)
				return ist > jst
			case fmt.Sprintf("%v%v", c.ColumnStarts, c.Desc):
				ist := fmt.Sprintf("%v-%v-%v", tj.StartsYear, tj.StartsMonth, tj.StartsDay)
				jst := fmt.Sprintf("%v-%v-%v", ti.StartsYear, ti.StartsMonth, ti.StartsDay)
				return ist < jst

			case fmt.Sprintf("%v%v", c.ColumnEnds, c.Asc):
				jend := fmt.Sprintf("%v-%v-%v", tj.EndsYear, tj.EndsMonth, tj.EndsDay)
				iend := fmt.Sprintf("%v-%v-%v", ti.EndsYear, ti.EndsMonth, ti.EndsDay)
				return iend > jend
			case fmt.Sprintf("%v%v", c.ColumnEnds, c.Desc):
				jend := fmt.Sprintf("%v-%v-%v", tj.EndsYear, tj.EndsMonth, tj.EndsDay)
				iend := fmt.Sprintf("%v-%v-%v", ti.EndsYear, ti.EndsMonth, ti.EndsDay)
				return iend < jend

			default:
				return false
				// return txs[j].Date.After(txs[i].Date)
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
	// currentSort := strings.TrimSuffix(strings.TrimSuffix(transactionsTableSortColumn, c.Asc), c.Desc)
	currentSort := ""
	currentSortDir := ""
	if strings.HasSuffix(transactionsTableSortColumn, c.Asc) {
		currentSort = strings.Split(transactionsTableSortColumn, c.Asc)[0]
		// currentSortDir = c.Asc
		currentSortDir = "↑"
	} else if strings.HasSuffix(transactionsTableSortColumn, c.Desc) {
		currentSort = strings.Split(transactionsTableSortColumn, c.Desc)[0]
		currentSortDir = "↓"
	}

	cellColumnOrderText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ORDER, c.ColumnOrder)
	if currentSort == c.ColumnOrder {
		cellColumnOrderText = fmt.Sprintf("%v%v", currentSortDir, cellColumnOrderText)
	}
	cellColumnAmountText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_AMOUNT, c.ColumnAmount)
	if currentSort == c.ColumnAmount {
		cellColumnAmountText = fmt.Sprintf("%v%v", currentSortDir, cellColumnAmountText)
	}
	cellColumnActiveText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ACTIVE, c.ColumnActive)
	if currentSort == c.ColumnActive {
		cellColumnActiveText = fmt.Sprintf("%v%v", currentSortDir, cellColumnActiveText)
	}
	cellColumnNameText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_NAME, c.ColumnName)
	if currentSort == c.ColumnName {
		cellColumnNameText = fmt.Sprintf("%v%v", currentSortDir, cellColumnNameText)
	}
	cellColumnFrequencyText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_FREQUENCY, c.ColumnFrequency)
	if currentSort == c.ColumnFrequency {
		cellColumnFrequencyText = fmt.Sprintf("%v%v", currentSortDir, cellColumnFrequencyText)
	}
	cellColumnIntervalText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_INTERVAL, c.ColumnInterval)
	if currentSort == c.ColumnInterval {
		cellColumnIntervalText = fmt.Sprintf("%v%v", currentSortDir, cellColumnIntervalText)
	}
	cellColumnMondayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_MONDAY, c.ColumnMonday)
	if currentSort == c.ColumnMonday {
		cellColumnMondayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnMondayText)
	}
	cellColumnTuesdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_TUESDAY, c.ColumnTuesday)
	if currentSort == c.ColumnTuesday {
		cellColumnTuesdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnTuesdayText)
	}
	cellColumnWednesdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_WEDNESDAY, c.ColumnWednesday)
	if currentSort == c.ColumnWednesday {
		cellColumnWednesdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnWednesdayText)
	}
	cellColumnThursdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_THURSDAY, c.ColumnThursday)
	if currentSort == c.ColumnThursday {
		cellColumnThursdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnThursdayText)
	}
	cellColumnFridayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_FRIDAY, c.ColumnFriday)
	if currentSort == c.ColumnFriday {
		cellColumnFridayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnFridayText)
	}
	cellColumnSaturdayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_SATURDAY, c.ColumnSaturday)
	if currentSort == c.ColumnSaturday {
		cellColumnSaturdayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnSaturdayText)
	}
	cellColumnSundayText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_SUNDAY, c.ColumnSunday)
	if currentSort == c.ColumnSunday {
		cellColumnSundayText = fmt.Sprintf("%v%v", currentSortDir, cellColumnSundayText)
	}
	cellColumnStartsText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_STARTS, c.ColumnStarts)
	if currentSort == c.ColumnStarts {
		cellColumnStartsText = fmt.Sprintf("%v%v", currentSortDir, cellColumnStartsText)
	}
	cellColumnEndsText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_ENDS, c.ColumnEnds)
	if currentSort == c.ColumnEnds {
		cellColumnEndsText = fmt.Sprintf("%v%v", currentSortDir, cellColumnEndsText)
	}
	cellColumnNoteText := fmt.Sprintf("%v%v", c.COLOR_COLUMN_NOTE, c.ColumnNote)
	if currentSort == c.ColumnNote {
		cellColumnNoteText = fmt.Sprintf("%v%v", currentSortDir, cellColumnNoteText)
	}

	cellColumnOrder := tview.NewTableCell(cellColumnOrderText)
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

	transactionsTable.SetCell(0, 0, cellColumnOrder)
	transactionsTable.SetCell(0, 1, cellColumnAmount)
	transactionsTable.SetCell(0, 2, cellColumnActive)
	transactionsTable.SetCell(0, 3, cellColumnName)
	transactionsTable.SetCell(0, 4, cellColumnFrequency)
	transactionsTable.SetCell(0, 5, cellColumnInterval)
	transactionsTable.SetCell(0, 6, cellColumnMonday)
	transactionsTable.SetCell(0, 7, cellColumnTuesday)
	transactionsTable.SetCell(0, 8, cellColumnWednesday)
	transactionsTable.SetCell(0, 9, cellColumnThursday)
	transactionsTable.SetCell(0, 10, cellColumnFriday)
	transactionsTable.SetCell(0, 11, cellColumnSaturday)
	transactionsTable.SetCell(0, 12, cellColumnSunday)
	transactionsTable.SetCell(0, 13, cellColumnStarts)
	transactionsTable.SetCell(0, 14, cellColumnEnds)
	transactionsTable.SetCell(0, 15, cellColumnNote)
	// transactionsTable.SetCell(0, 16, cellColumnID)
	// transactionsTable.SetCell(0, 17, cellColumnCreatedAt)
	// transactionsTable.SetCell(0, 18, cellColumnUpdatedAt)

	if selectedProfile != nil {
		if transactionsTableSortColumn != "" {
			sortTX()
		}
		// start by populating the table with the columns first
		for i, tx := range selectedProfile.TX {
			isPositiveAmount := tx.Amount >= 0
			amountColor := c.COLOR_COLUMN_AMOUNT
			if isPositiveAmount {
				amountColor = c.COLOR_COLUMN_AMOUNT_POSITIVE
			}

			nameColor := c.COLOR_COLUMN_NAME
			noteColor := c.COLOR_COLUMN_NOTE
			if !tx.Active {
				amountColor = c.COLOR_INACTIVE
				nameColor = c.COLOR_INACTIVE
				noteColor = c.COLOR_INACTIVE
			}

			cellOrder := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ORDER, tx.Order)).SetAlign(tview.AlignCenter)
			cellAmount := tview.NewTableCell(fmt.Sprintf("%v%v", amountColor, lib.FormatAsCurrency(tx.Amount))).SetAlign(tview.AlignCenter)

			activeText := "✔"
			if !tx.Active {
				activeText = " "
			}

			cellActive := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ACTIVE, activeText)).SetAlign(tview.AlignCenter)
			cellName := tview.NewTableCell(fmt.Sprintf("%v%v", nameColor, tx.Name)).SetAlign(tview.AlignLeft)
			cellFrequency := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_FREQUENCY, tx.Frequency)).SetAlign(tview.AlignCenter)
			cellInterval := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_INTERVAL, tx.Interval)).SetAlign(tview.AlignCenter)

			mondayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_MONDAY)
			if !tx.HasWeekday(c.WeekdayMondayInt) {
				mondayText = "[white] "
			}
			tuesdayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_TUESDAY)
			if !tx.HasWeekday(c.WeekdayTuesdayInt) {
				tuesdayText = "[white] "
			}
			wednesdayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_WEDNESDAY)
			if !tx.HasWeekday(c.WeekdayWednesdayInt) {
				wednesdayText = "[white] "
			}
			thursdayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_THURSDAY)
			if !tx.HasWeekday(c.WeekdayThursdayInt) {
				thursdayText = "[white] "
			}
			fridayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_FRIDAY)
			if !tx.HasWeekday(c.WeekdayFridayInt) {
				fridayText = "[white] "
			}
			saturdayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_SATURDAY)
			if !tx.HasWeekday(c.WeekdaySaturdayInt) {
				saturdayText = "[white] "
			}
			sundayText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_SUNDAY)
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

			cellStarts := tview.NewTableCell(fmt.Sprintf("%v%v-%v-%v", c.COLOR_COLUMN_STARTS, tx.StartsYear, tx.StartsMonth, tx.StartsDay)).SetAlign(tview.AlignCenter)
			cellEnds := tview.NewTableCell(fmt.Sprintf("%v%v-%v-%v", c.COLOR_COLUMN_ENDS, tx.EndsYear, tx.EndsMonth, tx.EndsDay)).SetAlign(tview.AlignCenter)

			cellNote := tview.NewTableCell(fmt.Sprintf("%v%v", noteColor, tx.Note))

			// cellID := tview.NewTableCell(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ID, tx.ID))
			// cellCreatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",fmt.Sprintf("%v", tx.CreatedAt)))
			// cellUpdatedAt := tview.NewTableCell(fmt.Sprintf("[yellow]%v",fmt.Sprintf("%v", tx.UpdatedAt)))

			cellName.SetExpansion(1)
			cellNote.SetExpansion(1)

			if lastSelectedIndex == i {
				if tx.Selected {
					cellOrder.SetBackgroundColor(selectedAndLastSelectedColor)
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
					cellOrder.SetBackgroundColor(lastSelectedColor)
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
				cellOrder.SetBackgroundColor(selectedColor)
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

			transactionsTable.SetCell(i+1, 0, cellOrder)
			transactionsTable.SetCell(i+1, 1, cellAmount)
			transactionsTable.SetCell(i+1, 2, cellActive)
			transactionsTable.SetCell(i+1, 3, cellName)
			transactionsTable.SetCell(i+1, 4, cellFrequency)
			transactionsTable.SetCell(i+1, 5, cellInterval)
			transactionsTable.SetCell(i+1, 6, cellMonday)
			transactionsTable.SetCell(i+1, 7, cellTuesday)
			transactionsTable.SetCell(i+1, 8, cellWednesday)
			transactionsTable.SetCell(i+1, 9, cellThursday)
			transactionsTable.SetCell(i+1, 10, cellFriday)
			transactionsTable.SetCell(i+1, 11, cellSaturday)
			transactionsTable.SetCell(i+1, 12, cellSunday)
			transactionsTable.SetCell(i+1, 13, cellStarts)
			transactionsTable.SetCell(i+1, 14, cellEnds)
			transactionsTable.SetCell(i+1, 15, cellNote)
			// transactionsTable.SetCell(i+1, 16, cellID)
			// transactionsTable.SetCell(i+1, 17, cellCreatedAt)
			// transactionsTable.SetCell(i+1, 18, cellUpdatedAt)
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
			case c.COLUMN_ORDER:
				if row == 0 {
					setTransactionsTableSort(c.ColumnOrder)
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
						if err != nil || d < 1 {
							activateTransactionsInputFieldNoAutocompleteReset("invalid order given:", fmt.Sprint(selectedProfile.TX[i].Order))
							return
						}

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Order = int(d)
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.COLOR_COLUMN_ORDER,
									selectedProfile.TX[j].Order,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				activateTransactionsInputField("order:", fmt.Sprint(selectedProfile.TX[i].Order))
			case c.COLUMN_AMOUNT:
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
						amountColor := c.COLOR_COLUMN_AMOUNT
						if isPositiveAmount {
							amountColor = c.COLOR_COLUMN_AMOUNT_POSITIVE
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
			case c.COLUMN_ACTIVE:
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

						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_ACTIVE, activeText))
					}
				}

				modified()
			case c.COLUMN_NAME:
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
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_NAME, selectedProfile.TX[i].Name))
							}
						}

						modified()
					}
					deactivateTransactionsInputField()
				})
			case c.COLUMN_FREQUENCY:
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
							transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_FREQUENCY, selectedProfile.TX[i].Frequency))
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
			case c.COLUMN_INTERVAL:
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid interval given:", fmt.Sprint(selectedProfile.TX[i].Interval))
							return
						}

						selectedProfile.TX[i].Interval = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].Interval = selectedProfile.TX[i].Interval
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v",
									c.COLOR_COLUMN_INTERVAL,
									selectedProfile.TX[i].Interval,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
					}
				})
				activateTransactionsInputField("interval:", fmt.Sprint(selectedProfile.TX[i].Interval))
			case c.COLUMN_MONDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_MONDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayMondayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_MONDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_TUESDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_TUESDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayTuesdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_TUESDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_WEDNESDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_WEDNESDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayWednesdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_WEDNESDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_THURSDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_THURSDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayThursdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_THURSDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_FRIDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_FRIDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdayFridayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_FRIDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_SATURDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_SATURDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdaySaturdayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_SATURDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_SUNDAY:
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

						cellText := fmt.Sprintf("%v✔", c.COLOR_COLUMN_SUNDAY)
						if !selectedProfile.TX[j].HasWeekday(c.WeekdaySundayInt) {
							cellText = "[white] "
						}
						transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_SUNDAY, cellText))
					}
				}

				modified()
			case c.COLUMN_STARTS:
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid day given:", fmt.Sprint(selectedProfile.TX[i].StartsDay))
							return
						}

						selectedProfile.TX[i].StartsDay = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsDay = selectedProfile.TX[i].StartsDay
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_STARTS,
									selectedProfile.TX[j].StartsYear,
									selectedProfile.TX[j].StartsMonth,
									selectedProfile.TX[j].StartsDay,
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", fmt.Sprint(selectedProfile.TX[i].StartsMonth))
							return
						}

						selectedProfile.TX[i].StartsMonth = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsMonth = selectedProfile.TX[i].StartsMonth
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_STARTS,
									selectedProfile.TX[j].StartsYear,
									selectedProfile.TX[j].StartsMonth,
									selectedProfile.TX[j].StartsDay,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", fmt.Sprint(selectedProfile.TX[i].StartsDay))
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", fmt.Sprint(selectedProfile.TX[i].StartsYear))
							return
						}

						selectedProfile.TX[i].StartsYear = int(d)

						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].StartsYear = selectedProfile.TX[i].StartsYear
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_STARTS,
									selectedProfile.TX[j].StartsYear,
									selectedProfile.TX[j].StartsMonth,
									selectedProfile.TX[j].StartsDay,
								))
							}
						}

						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", fmt.Sprint(selectedProfile.TX[i].StartsMonth))
						defer transactionsInputField.SetDoneFunc(monthFunc)
					}
				}

				transactionsInputField.SetDoneFunc(yearFunc)
				activateTransactionsInputField("year:", fmt.Sprint(selectedProfile.TX[i].StartsYear))
			case c.COLUMN_ENDS:
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid day given:", fmt.Sprint(selectedProfile.TX[i].EndsDay))
							return
						}

						selectedProfile.TX[i].EndsDay = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsDay = selectedProfile.TX[i].EndsDay
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_ENDS,
									selectedProfile.TX[j].EndsYear,
									selectedProfile.TX[j].EndsMonth,
									selectedProfile.TX[j].EndsDay,
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid month given:", fmt.Sprint(selectedProfile.TX[i].EndsMonth))
							return
						}

						selectedProfile.TX[i].EndsMonth = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsMonth = selectedProfile.TX[i].EndsMonth
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_ENDS,
									selectedProfile.TX[j].EndsYear,
									selectedProfile.TX[j].EndsMonth,
									selectedProfile.TX[j].EndsDay,
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("day (1-31):", fmt.Sprint(selectedProfile.TX[i].EndsDay))
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
							activateTransactionsInputFieldNoAutocompleteReset("invalid year given:", fmt.Sprint(selectedProfile.TX[i].EndsYear))
							return
						}

						selectedProfile.TX[i].EndsYear = int(d)
						// update all selected values as well as the current one
						for j := range selectedProfile.TX {
							if selectedProfile.TX[j].Selected || j == i {
								selectedProfile.TX[j].EndsYear = selectedProfile.TX[i].EndsYear
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf(
									"%v%v-%v-%v",
									c.COLOR_COLUMN_ENDS,
									selectedProfile.TX[j].EndsYear,
									selectedProfile.TX[j].EndsMonth,
									selectedProfile.TX[j].EndsDay,
								))
							}
						}
						modified()
						deactivateTransactionsInputField()
						activateTransactionsInputFieldNoAutocompleteReset("month (1-12):", fmt.Sprint(selectedProfile.TX[i].EndsMonth))
						defer transactionsInputField.SetDoneFunc(monthFunc)
					}
				}

				transactionsInputField.SetDoneFunc(yearFunc)
				activateTransactionsInputField("year:", fmt.Sprint(selectedProfile.TX[i].EndsYear))
			case c.COLUMN_NOTE:
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
								transactionsTable.GetCell(j+1, column).SetText(fmt.Sprintf("%v%v", c.COLOR_COLUMN_NOTE, selectedProfile.TX[j].Note))
							}
						}

						modified()
					}
					deactivateTransactionsInputField()
				})
			case c.COLUMN_ID:
				// pass for now
			case c.COLUMN_CREATEDAT:
				// pass for now
			case c.COLUMN_UPDATEDAT:
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
	// transactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
	// 	return true
	// })
	transactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
		return []string{}
	})
}

func deactivateTransactionsInputField() {
	// transactionsInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
	// 	return []string{}
	// })
	// transactionsInputField.SetAutocompletedFunc(func(text string, index, source int) bool {
	// 	return true
	// })
	transactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	transactionsInputField.SetLabel("[gray] editor appears here when editing")
	transactionsInputField.SetText("")

	if previous != nil {
		app.SetFocus(previous)
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
	currentFocus := app.GetFocus()
	if currentFocus == transactionsInputField {
		return
	}

	previous = currentFocus
	app.SetFocus(transactionsInputField)
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
	currentFocus := app.GetFocus()
	if currentFocus == transactionsInputField {
		return
	}

	previous = currentFocus
	app.SetFocus(transactionsInputField)
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

	populateProfilesPage(true, true)
	getTransactionsTable()
	transactionsTable.Select(selectedProfile.SelectedRow, selectedProfile.SelectedColumn)
	app.SetFocus(transactionsTable)
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

	populateProfilesPage(true, true)
	getTransactionsTable()
	transactionsTable.Select(selectedProfile.SelectedRow, selectedProfile.SelectedColumn)
	app.SetFocus(transactionsTable)
}

// attempts to place the current config at undoBuffer[undoBufferPos+1]
// but only if there were actual changes.
//
// also updates the status text accordingly
func modified() {
	if selectedProfile != nil {
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

		err := lib.ValidateTransactions(&selectedProfile.TX)
		if err != nil {
			statusText.SetText("[red] unable to auto-order")
		}
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
		// // set the selectedProfile to the latest undoBuffer's config
		// n := selectedProfile.Name
		// for i := range undoBuffer[undoBufferPos].Profiles {
		// 	if undoBuffer[undoBufferPos].Profiles[i].Name == n {
		// 		selectedProfile = &(undoBuffer[undoBufferPos].Profiles[i])
		// 	}
		// }

		statusText.SetText(fmt.Sprintf("[white] %v*[gray][%v/%v]", configFile, undoBufferPos+1, len(undoBuffer)))
	}
}

// returns a simple flex view with two columns:
// - a list of profiles (left side)
// - a quick summary of bills / stats for the highlighted profile (right side)
func getProfilesFlex() {
	profileList = tview.NewList()
	profileList.SetBorder(true)
	profileList.ShowSecondaryText(false).
		SetSelectedBackgroundColor(tcell.NewRGBColor(50, 50, 50)).
		SetSelectedTextColor(tcell.ColorWhite).
		SetTitle("Profiles")

	// TODO: doesn't work as well as expected; an infinite loop will occur in
	// some cases, and in other cases, we can't set the text of the profileList
	// items
	// profileList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
	// 	selectedProfile = &(config.Profiles[index])
	// 	// populateProfilesPage(true, true)
	// 	getTransactionsTable()
	// 	// app.SetFocus(transactionsTable)
	// })

	statusText = tview.NewTextView()
	statusText.SetBorder(true)
	statusText.SetDynamicColors(true)
	setStatusNoChanges()

	profilesLeftSide := tview.NewFlex().SetDirection(tview.FlexRow)
	profilesLeftSide.AddItem(profileList, 0, 1, true).
		AddItem(statusText, 3, 0, true)

	transactionsPage = tview.NewFlex().SetDirection(tview.FlexRow)
	transactionsTable = tview.NewTable().SetFixed(1, 1)
	transactionsInputField = tview.NewInputField()

	transactionsTable.SetBorder(true)
	transactionsInputField.SetBorder(true)

	transactionsInputField.SetFieldBackgroundColor(tcell.ColorBlack)
	transactionsInputField.SetLabel("[gray] editor appears here when editing")

	populateProfilesPage(true, true)
	getTransactionsTable()

	transactionsPage.AddItem(transactionsTable, 0, 1, false).
		AddItem(transactionsInputField, 3, 0, false)

	profilesPage = tview.NewFlex().SetDirection(tview.FlexColumn)
	profilesPage.AddItem(profilesLeftSide, 0, 1, true).
		AddItem(transactionsPage, 0, 10, false)
}

func setTransactionsTableSort(column string) {
	transactionsTableSortColumn = lib.GetNextSort(transactionsTableSortColumn, column)
	defer getTransactionsTable()
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
			setResultsFormPreset(c.START_TODAY_PRESET, c.ONE_YR)
			updateResultsForm()
			getResultsTable()
		}).
		AddButton("5 years", func() {
			setResultsFormPreset(c.START_TODAY_PRESET, c.FIVE_YR)
			updateResultsForm()
			getResultsTable()
		})

	resultsForm.SetLabelColor(tcell.ColorViolet)
	resultsForm.SetFieldBackgroundColor(tcell.NewRGBColor(40, 40, 40))
	resultsForm.SetBorder(true)
}

func getResultsFlex() {
	resultsTable = tview.NewTable().SetFixed(1, 1)
	resultsTable.SetBorder(true)
	resultsDescription = tview.NewTextView()
	resultsDescription.SetBorder(true)
	resultsDescription.SetDynamicColors(true)

	resultsForm = tview.NewForm()
	updateResultsForm()

	resultsTable.SetTitle("Results")
	resultsTable.SetBorders(false).
		SetSelectable(true, false). // set row & cells to be selectable
		SetSeparator(' ')

	resultsRightSide = tview.NewFlex().SetDirection(tview.FlexRow)
	resultsRightSide.AddItem(resultsTable, 0, 2, true).AddItem(resultsDescription, 0, 1, false)

	resultsPage = tview.NewFlex().SetDirection(tview.FlexColumn)
	resultsPage.AddItem(resultsForm, 0, 1, true).AddItem(resultsRightSide, 0, 3, false)
}

// Allows a simple button press to set the start & end dates to various common
// use cases. For example, start from today and end 1 year or 5 years from now.
//
// TODO: implement other start date logic - currently only supports today
func setResultsFormPreset(startDate string, endDate string) {
	var start time.Time
	switch startDate {
	case c.START_TODAY_PRESET:
		fallthrough
	default:
		start = time.Now()
		break
	}

	var end time.Time
	switch endDate {
	case c.ONE_YR:
		end = start.Add(time.Hour * 24 * 365)
	case c.FIVE_YR:
		end = start.Add(time.Hour * 24 * 365 * 5)
	}

	selectedProfile.StartYear = fmt.Sprint(start.Year())
	selectedProfile.StartMonth = fmt.Sprint(int(start.Month()))
	selectedProfile.StartDay = fmt.Sprint(start.Day())

	selectedProfile.EndYear = fmt.Sprint(end.Year())
	selectedProfile.EndMonth = fmt.Sprint(int(end.Month()))
	selectedProfile.EndDay = fmt.Sprint(end.Day())
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
		selectedProfile.StartYear = fmt.Sprint(now.Year())
	}
	if selectedProfile.StartMonth == "" {
		selectedProfile.StartMonth = fmt.Sprint(int(now.Month()))
	}
	if selectedProfile.StartDay == "" {
		selectedProfile.StartDay = fmt.Sprint(now.Day())
	}

	if selectedProfile.EndYear == "" {
		selectedProfile.EndYear = fmt.Sprint(yr.Year())
	}
	if selectedProfile.EndMonth == "" {
		selectedProfile.EndMonth = fmt.Sprint(int(yr.Month()))
	}
	if selectedProfile.EndDay == "" {
		selectedProfile.EndDay = fmt.Sprint(yr.Day())
	}

	if selectedProfile.StartingBalance == "" {
		selectedProfile.StartingBalance = lib.FormatAsCurrency(50000)
	}
}

func getResultsTable() {
	resultsTable.Clear()

	setSelectedProfileDefaults()

	// get results
	results, err := lib.GenerateResultsFromDateStrings(
		&(selectedProfile.TX),
		int(lib.ParseDollarAmount(selectedProfile.StartingBalance, true)),
		fmt.Sprintf(
			"%v-%v-%v",
			selectedProfile.StartYear,
			selectedProfile.StartMonth,
			selectedProfile.StartDay,
		),
		fmt.Sprintf(
			"%v-%v-%v",
			selectedProfile.EndYear,
			selectedProfile.EndMonth,
			selectedProfile.EndDay,
		),
	)
	if err != nil {
		// TODO: add better error display
		panic(err)
	}

	// this may help with garbage collection
	latestResults = nil

	latestResults = &results

	// set up headers
	hDate := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DATE, c.ColumnDate, c.RESET_STYLE))
	hBalance := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_BALANCE, c.ColumnBalance, c.RESET_STYLE))
	hCumulativeIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_CUMULATIVEINCOME, c.ColumnCumulativeIncome, c.RESET_STYLE))
	hCumulativeExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_CUMULATIVEEXPENSES, c.ColumnCumulativeExpenses, c.RESET_STYLE))
	hDayExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYEXPENSES, c.ColumnDayExpenses, c.RESET_STYLE))
	hDayIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYINCOME, c.ColumnDayIncome, c.RESET_STYLE))
	hDayNet := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYNET, c.ColumnDayNet, c.RESET_STYLE))
	hDiffFromStart := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DIFFFROMSTART, c.ColumnDiffFromStart, c.RESET_STYLE))
	hDayTransactionNames := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYTRANSACTIONNAMES, c.ColumnDayTransactionNames, c.RESET_STYLE))

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
		rDate := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DATE, lib.FormatAsDate(results[i].Date), c.RESET_STYLE))
		rBalance := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_BALANCE, lib.FormatAsCurrency(results[i].Balance), c.RESET_STYLE))
		rCumulativeIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_CUMULATIVEINCOME, lib.FormatAsCurrency(results[i].CumulativeIncome), c.RESET_STYLE))
		rCumulativeExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_CUMULATIVEEXPENSES, lib.FormatAsCurrency(results[i].CumulativeExpenses), c.RESET_STYLE))
		rDayExpenses := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYEXPENSES, lib.FormatAsCurrency(results[i].DayExpenses), c.RESET_STYLE))
		rDayIncome := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYINCOME, lib.FormatAsCurrency(results[i].DayIncome), c.RESET_STYLE))
		rDayNet := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYNET, lib.FormatAsCurrency(results[i].DayNet), c.RESET_STYLE))
		rDiffFromStart := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DIFFFROMSTART, lib.FormatAsCurrency(results[i].DiffFromStart), c.RESET_STYLE))
		rDayTransactionNames := tview.NewTableCell(fmt.Sprintf("%v%v%v", c.COLOR_COLUMN_RESULTS_DAYTRANSACTIONNAMES, results[i].DayTransactionNames, c.RESET_STYLE))

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
				sb.Write([]byte(fmt.Sprintf("%v\n", t)))
			}
			resultsDescription.SetText(sb.String())
		}
	})

	app.SetFocus(resultsTable)
}

func promptExit() {
	// check if we are already prompting
	currentPage, _ := pages.GetFrontPage()
	if currentPage == PAGE_PROMPT {
		return
	}

	// now check if the previous page is something other than the prompt already
	previousPage, _ = pages.GetFrontPage()
	if previousPage == PAGE_PROMPT {
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
				app.Stop()
			case 1:
				fallthrough
			case 2:
				fallthrough
			default:
				pages.SwitchToPage(previousPage)
				return
			}
		},
	).SetBackgroundColor(tcell.ColorGoldenrod).
		SetTextColor(tcell.ColorBlack) //.
	// SetButtonTextColor(tcell.ColorGray) //.
	// SetButtonBackgroundColor(tcell.NewRGBColor(100, 100, 100))

	pages.SwitchToPage(PAGE_PROMPT)
	promptBox.SetFocus(2)
	app.SetFocus(promptBox)
}

// switches to the prompt page and shows a modal that informs the user that they
// are in keyboard echo mode. If KB echo mode is not enabled, this gracefully
// returns immediately and does nothing.
func promptKBMode() {
	if !keyboardEchoMode {
		return
	}

	// temporarily turn off KB echo mode so that the user's keys are captured
	// properly until they can give consent to entering the mode
	keyboardEchoMode = false

	promptBox.ClearButtons().AddButtons(
		[]string{
			"Turn off",
			"Exit Now",
			"Continue",
		},
	).SetText(
		"You have started this application in keyboard echo mode. If you choose to continue, all of your key presses will be shown in the bottom left status bar. You can press Escape or Ctrl+C at any time to quit.\n\nThese keyboard shortcuts are directly to be entered into your configuration for customization as shown, such as 'Rune[R[]'. If you wish to later remap the Escape key, use the string 'Esc', and for Ctrl+C, use the string 'Ctrl+C' in your configuration. Note that when in keyboard echoing mode, Ctrl+C and Escape are permanently mapped and cannot be changed.\n\nThis was triggered by passing the -kb flag to this application on startup.", // the Rune[R[] is not a typo; it's required in order to show brackets in tcell
	).SetDoneFunc(
		func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				keyboardEchoMode = false
				pages.SwitchToPage(PAGE_PROFILES)
			case 1:
				keyboardEchoMode = false
				app.Stop()
			case 2:
				keyboardEchoMode = true
				pages.SwitchToPage(PAGE_PROFILES)
			default:
				keyboardEchoMode = false
				app.Stop()
				return
			}
		},
	).SetBackgroundColor(tcell.ColorDimGray).
		SetTextColor(tcell.ColorWhite)

	pages.SwitchToPage(PAGE_PROMPT)
	promptBox.SetFocus(2)
	app.SetFocus(promptBox)
}

// For an input keybinding (straight from event.Name()), an output action
// will be returned, for example - "Ctrl+Z" will return "undo".
func getDefaultKeybind(name string) string {
	m, ok := c.DEFAULT_MAPPINGS[name]
	if !ok {
		return ""
	}

	return m
}

func action(action string, e *tcell.EventKey) *tcell.EventKey {
	duplicating := false
	multiSelecting := false
	switch action {
	case c.ACTION_REDO:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsInputField:
				return e
			case transactionsTable:
				redo()
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_UNDO:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsInputField:
				return e
			case transactionsTable:
				undo()
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_QUIT:
		promptExit()
		return nil
	case c.ACTION_MULTI:
		multiSelecting = true
		fallthrough
	case c.ACTION_SELECT:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
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

					if lastSelectedIndex == -1 {
						lastSelectedIndex = actual
					}

					// now that we've determined what the selection value
					// should be, proceed to apply it to every value from
					// lastSelectedIndex to the current index
					for i := range selectedProfile.TX {
						// last=5, current=10, select from 5-10 => last < i < actual
						// last=10, current=3, select from 3-10 => last > i > actual
						shouldModify := (lastSelectedIndex < i && i <= actual) || (lastSelectedIndex > i && i >= actual)
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
				lastSelectedIndex = actual
				modified()
				getTransactionsTable()
				transactionsTable.Select(cr, cc)
				app.SetFocus(transactionsTable)
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_MOVE:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsInputField:
				return e
			case transactionsTable:
				// move all selected items to the currently selected row:
				// delete items, then re-add the items after the current
				// row, then highlight the correct row

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

				selectedProfile.TX = slices.Insert(selectedProfile.TX, newPosition, deleted...)

				modified()

				// re-render the table
				getTransactionsTable()

				transactionsTable.Select(newPosition+1, cc) // offset for headers
				app.SetFocus(transactionsTable)
			default:
				app.SetFocus(profileList)
			}
			return nil
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_DELETE:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
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
				// for i := range selectedProfile.TX {
				// 	if i == actual {
				// 		selectedProfile.TX = slices.Delete(selectedProfile.TX, i, i+1)
				// 	}
				// }
				getTransactionsTable()
				transactionsTable.Select(cr, cc)
				app.SetFocus(transactionsTable)
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
						c.RESET_STYLE,
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
						populateProfilesPage(true, true)
						getTransactionsTable()
						transactionsTable.Select(0, 0)
						app.SetFocus(profileList)
					}
				})
				activateTransactionsInputField(getPrompt(), "")
			default:
				app.SetFocus(profileList)
			}
			return nil
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_DUPLICATE:
		duplicating = true
		fallthrough
	case c.ACTION_ADD:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsInputField:
				return e
			case transactionsTable:
				// duplicate the current transaction
				// get the height & width of the transactions table
				cr, cc := transactionsTable.GetSelection()
				actual := cr - 1 // skip header
				nt := []lib.TX{}
				// iterate through the list once to find how many selected
				// items there are
				numSelected := 0
				for i := range selectedProfile.TX {
					if selectedProfile.TX[i].Selected {
						numSelected += 1
					}
				}
				for i := range selectedProfile.TX {
					isHighlightedRow := i == actual && numSelected <= 1
					isSelectedDuplicationCandidate := selectedProfile.TX[i].Selected && duplicating
					if isHighlightedRow || isSelectedDuplicationCandidate {
						// keep track of the highest order in a temporary
						// slice
						largestOrderHolder := []lib.TX{}
						largestOrderHolder = append(largestOrderHolder, selectedProfile.TX...)
						largestOrderHolder = append(largestOrderHolder, nt...)

						newTX := lib.GetNewTX()
						newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1

						if duplicating {
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
						}

						nt = append(nt, newTX)
					}
				}

				// edge case: if we are not duplicating, and none are
				// selected, then it means that the user is currently trying
				// to add a new transaction, and they probably have the
				// cursor on the table's headers row
				if !duplicating && numSelected == 0 && len(nt) == 0 {
					largestOrderHolder := []lib.TX{}
					largestOrderHolder = append(largestOrderHolder, selectedProfile.TX...)
					largestOrderHolder = append(largestOrderHolder, nt...)
					newTX := lib.GetNewTX()
					newTX.Order = lib.GetLargestOrder(largestOrderHolder) + 1
					nt = append(nt, newTX)
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
					app.SetFocus(transactionsTable)
				}
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

						newProfile := m.Profile(*selectedProfile)
						newProfile.Name = newProfileName
						if !duplicating {
							newProfile = m.Profile{Name: newProfileName}
						}

						selectedProfile = &newProfile

						config.Profiles = append(config.Profiles, newProfile)
						modified()
						deactivateTransactionsInputField()
						populateProfilesPage(true, true)
						getTransactionsTable()
						transactionsTable.Select(0, 0)
						// app.SetFocus(transactionsTable)
						app.SetFocus(profileList)
					}
				})
				activateTransactionsInputField("set new unique profile name:", "")
				return nil
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_EDIT:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
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
						populateProfilesPage(true, true)
						getTransactionsTable()
						transactionsTable.Select(0, 0)
						app.SetFocus(profileList)
					}
				})
				activateTransactionsInputField(fmt.Sprintf("set new unique profile name for %v:", selectedProfile.Name), "")
				return nil
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_SAVE:
		if config.Version == "" {
			config.Version = c.CONFIG_VERSION
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
	case c.ACTION_END:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsTable:
				c := transactionsTable.GetColumnCount() - 1
				cr, _ := transactionsTable.GetSelection()
				transactionsTable.Select(cr, c)
				app.SetFocus(transactionsTable)
				return nil
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_HOME:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsTable:
				cr, _ := transactionsTable.GetSelection()
				transactionsTable.Select(cr, 0)
				app.SetFocus(transactionsTable)
				return nil
			default:
				return e
			}
		case PAGE_RESULTS:
			return e
		}
	case c.ACTION_DOWN:
		switch app.GetFocus() {
		case transactionsInputField:
			return nil
		default:
			return e
		}
	case c.ACTION_UP:
		switch app.GetFocus() {
		case transactionsInputField:
			return nil
		default:
			return e
		}
	case c.ACTION_LEFT:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case profileList:
				app.SetFocus(transactionsTable)
				return nil
			case transactionsTable:
				_, cc := transactionsTable.GetSelection()
				// focus the profile list when at column 0
				if cc == 0 {
					app.SetFocus(profileList)
					return nil
				}
				return e
			default:
				return e
			}
		default:
			return e
		}
	case c.ACTION_RIGHT:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case profileList:
				app.SetFocus(transactionsTable)
				return nil
			case transactionsTable:
				c := transactionsTable.GetColumnCount() - 1
				_, cc := transactionsTable.GetSelection()
				// focus the profile list when at max column
				if cc == c {
					app.SetFocus(profileList)
					return nil
				}
			default:
				return e
			}
		default:
			return e
		}
	case c.ACTION_PAGEDOWN:
		f := app.GetFocus()
		p, _ := pages.GetFrontPage()
		switch p {
		case PAGE_RESULTS:
			switch f {
			case resultsDescription:
				return e
			case resultsTable:
				return e
			default:
				app.SetFocus(resultsTable)
				return nil
			}
		}
	case c.ACTION_PAGEUP:
		f := app.GetFocus()
		p, _ := pages.GetFrontPage()
		switch p {
		case PAGE_RESULTS:
			switch f {
			case resultsDescription:
				return e
			case resultsTable:
				return e
			default:
				app.SetFocus(resultsTable)
				return nil
			}
		}
	case c.ACTION_BACKTAB:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsInputField:
				return nil
			case profileList:
				app.SetFocus(transactionsTable)
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
					nr -= 1
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
				app.SetFocus(focusTarget)
			default:
				app.SetFocus(profileList)
			}
			return nil
		case PAGE_RESULTS:
			switch app.GetFocus() {
			case resultsTable:
				resultsForm.SetFocus(0)
				app.SetFocus(resultsForm)
				return nil
			case resultsDescription:
				app.SetFocus(resultsTable)
			case resultsForm:
				return e
			}
			return e
		}
	case c.ACTION_TAB:
		pageName, _ := pages.GetFrontPage()
		switch pageName {
		case PAGE_PROFILES:
			switch app.GetFocus() {
			case transactionsInputField:
				return nil
			case profileList:
				app.SetFocus(transactionsTable)
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
					nr += 1
					if nr > r {
						nc = 0
						nr = r
					}
					// it's more intuitive to go back to the profileList
					// when backtabbing from the first column in the table
					focusTarget = profileList
				}
				transactionsTable.Select(nr, nc)
				app.SetFocus(focusTarget)
			default:
				app.SetFocus(profileList)
			}
			return nil
		case PAGE_RESULTS:
			switch app.GetFocus() {
			case resultsTable:
				app.SetFocus(resultsDescription)
			case resultsDescription:
				resultsForm.SetFocus(0)
				app.SetFocus(resultsForm)
				return nil
			case resultsForm:
				return e
			}
			return e
		}
	case c.ACTION_ESCAPE:
		currentFocus := app.GetFocus()
		switch currentFocus {
		case transactionsInputField:
			return e
		case transactionsTable:
			// deselect the last selected index on the first press
			if lastSelectedIndex != -1 {
				lastSelectedIndex = -1
				getTransactionsTable()
				cr, cc := transactionsTable.GetSelection()
				transactionsTable.Select(cr, cc)
				app.SetFocus(transactionsTable)
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
				app.SetFocus(profileList)
				return nil
			}
			modified()
			getTransactionsTable()
			cr, cc := transactionsTable.GetSelection()
			transactionsTable.Select(cr, cc)
			app.SetFocus(transactionsTable)
		case resultsForm:
			app.SetFocus(resultsTable)
			return nil
		case resultsTable:
			pages.SwitchToPage(PAGE_PROFILES)
			return nil
		default:
			promptExit()
			return nil
		}
	case c.ACTION_RESULTS:
		// if the user is already on the results page, focus the
		// text view description instead
		p, _ := pages.GetFrontPage()
		alreadyOnPage := false
		if p == PAGE_RESULTS {
			alreadyOnPage = true
		}
		pages.SwitchToPage(PAGE_RESULTS)
		setBottomHelpText()

		if alreadyOnPage {
			getResultsTable()
			if latestResults != nil {
				stats, err := lib.GetStats(*latestResults)
				if err != nil {
					return nil
				}
				resultsDescription.SetText(stats)
				return nil
			}
			app.SetFocus(resultsTable)
		}
		return nil
	case c.ACTION_PROFILES:
		p, _ := pages.GetFrontPage()
		alreadyOnPage := false
		if p == PAGE_PROFILES {
			alreadyOnPage = true
		}
		pages.SwitchToPage(PAGE_PROFILES)
		setBottomHelpText()
		if alreadyOnPage {
			app.SetFocus(profileList)
		}
		return nil
	case c.ACTION_GLOBAL_HELP:
		pages.SwitchToPage(PAGE_HELP)
		setBottomHelpText()
		return nil
	case c.ACTION_HELP:
		switch app.GetFocus() {
		case transactionsInputField:
		case resultsForm:
			return e
		default:
			pages.SwitchToPage(PAGE_HELP)
			setBottomHelpText()
		}
	case c.ACTION_SEARCH:
		// searching not implemented yet
		fallthrough
	default:
		return e
	}

	return e
}

func bindings() {
	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		n := e.Name()
		if keyboardEchoMode {
			statusText.SetDynamicColors(false).SetText(n)
			if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC {
				app.Stop()
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
	})
}

func main() {
	var err error

	config, err = loadConfig(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err.Error())
	}

	if len(config.Profiles) > 0 {
		selectedProfile = &(config.Profiles[0])
	} else {
		n := m.Profile{
			TX:   []lib.TX{lib.GetNewTX()},
			Name: "New Profile Name",
		}
		config.Profiles = append(config.Profiles, n)
		selectedProfile = &(config.Profiles[0])
	}

	b, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("failed to marshal config for loading into undo buffer: %v", err.Error())
	}

	allKeyBindings = GetCombinedKeybindings(config.Keybindings, c.DEFAULT_MAPPINGS)
	allBoundActions = GetAllBoundActions(config.Keybindings, c.DEFAULT_MAPPINGS)

	undoBuffer = [][]byte{b}
	undoBufferPos = 0

	lastSelectedIndex = -1
	app = tview.NewApplication()
	pages = tview.NewPages()
	getProfilesFlex()
	getResultsFlex()
	getHelpModal()

	promptBox = tview.NewModal()

	pages.AddPage(PAGE_PROFILES, profilesPage, true, true).
		AddPage(PAGE_RESULTS, resultsPage, true, true).
		AddPage(PAGE_HELP, helpModal, true, true).
		AddPage(PAGE_PROMPT, promptBox, true, true)

	pages.SwitchToPage(PAGE_PROFILES)

	bottomHelpText = tview.NewTextView()
	bottomHelpText.SetDynamicColors(true)
	setBottomHelpText()

	layout = tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(pages, 0, 1, true).AddItem(bottomHelpText, 1, 0, false)

	app.SetFocus(profileList)

	promptKBMode()

	bindings()

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
