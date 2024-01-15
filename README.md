# finance-planner-tui

TODO:

- when sorting, the LastSelectedIndex does not seem to work
- fix no sorting for Note column
- weekday sorting only sorts for Monday
- when hitting Enter on a row, if there are other things selected, also select this row
- change all refs to c.Reset to tcell.ColorReset
- debug config.yml loading errors
- create xdg config dir when loading configs
- finish translations into english
- allow disabling mouse support so that things can be copied (config propery, or even through a shortcut?)
- `ctrl+F` and `/` for search
- Home and End keys should navigate to the top left & bottom right columns when already at the leftmost column/row
- write logs to xdg cache dir
- remind users that the Tab key is used for navigating through the results form
- update help file to show actual keybindings
- customize colors (later!)
- fix issue on mac with showing black on black in results page

## keybindings

- redo: ctrl+Y
- undo: ctrl+Z
- quit: ctrl+C
- select: space
- multi-select: ctrl+space
- move: m
- delete: delete
- duplicate: ctrl+D, ctrl+N
- add: a, n
- edit: e, r
- save: ctrl+s
- end: end
- home: home
- down: down
- up: up
- pagedown: PgDn
- pageup: PgUp
- backtab: backtab
- tab: tab
- escape: escape
- results: F3
- profiles: F2
- help: F1, ?
- search: / (not implemented yet)

unused:

- right: unused (default behavior)
- left: unused (default behavior)

## lint & formatters

```bash
go install github.com/daixiang0/gci@latest
```
