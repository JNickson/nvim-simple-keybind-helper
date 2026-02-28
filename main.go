package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var searchStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("205")).
	Bold(true)

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241"))

type model struct {
	table      table.Model
	allRows    []table.Row
	search     textinput.Model
	searchMode bool
}

type columnConfig struct {
	Title string `json:"title"`
	Width int    `json:"width"`
}

type rowConfig struct {
	Mode    string `json:"mode"`
	Keybind string `json:"keybind"`
	Action  string `json:"action"`
}

type appConfig struct {
	Columns []columnConfig `json:"columns"`
	Rows    []rowConfig    `json:"rows"`
	Height  int            `json:"height"`
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) filterRows(query string) {
	query = strings.ToLower(query)

	if query == "" {
		m.table.SetRows(m.allRows)
		return
	}

	var filtered []table.Row

	for _, row := range m.allRows {
		for _, col := range row {
			if strings.Contains(strings.ToLower(col), query) {
				filtered = append(filtered, row)
				break // stop checking this row once matched
			}
		}
	}

	m.table.SetRows(filtered)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:

		// Always allow ctrl+c
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// 🔎 SEARCH MODE
		if m.searchMode {

			switch msg.String() {

			case "enter":
				// Commit search and exit search mode
				m.searchMode = false
				m.search.Blur()
				return m, nil

			case "esc":
				// Cancel search completely
				m.searchMode = false
				m.search.Blur()
				m.search.SetValue("")
				m.filterRows("")
				return m, nil
			}

			m.search, cmd = m.search.Update(msg)
			m.filterRows(m.search.Value())
			return m, cmd
		}

		// 🧭 NORMAL MODE KEYS
		switch msg.String() {

		case "/":
			m.searchMode = true
			m.search.Focus()
			m.search.SetValue("")
			return m, nil

		case "q":
			return m, tea.Quit
		}
	}

	// Let table handle j/k navigation
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	tableView := baseStyle.Render(m.table.View())

	// Search visualiser
	var searchLine string
	if m.searchMode {
		searchLine = searchStyle.Render("Search: ") + m.search.View()
	} else if m.search.Value() != "" {
		searchLine = searchStyle.Render("Filtered by: ") + m.search.Value()
	}

	// Help footer (always shown)
	helpLine := helpStyle.Render("Press '/' to search | j/k to move | q to quit")

	// Compose view
	if searchLine != "" {
		return tableView + "\n\n" + searchLine + "\n" + helpLine + "\n"
	}

	return tableView + "\n\n" + helpLine + "\n"
}

func defaultConfig() appConfig {
	return appConfig{
		Columns: []columnConfig{
			{Title: "Mode", Width: 8},
			{Title: "Keybind", Width: 16},
			{Title: "Action", Width: 80},
		},
		Rows: []rowConfig{
			// Visual
			{Mode: "visual", Keybind: "y", Action: "yank (copy) selection"},
			{Mode: "visual", Keybind: ">", Action: "indent selection right"},
			{Mode: "visual", Keybind: "<", Action: "indent selection left"},

			// Insert
			{Mode: "insert", Keybind: "<C-h>", Action: "delete previous character"},
			{Mode: "insert", Keybind: "<C-w>", Action: "delete all from the cursor back to  word boundary (space or punctuation)"},
			{Mode: "insert", Keybind: "<C-c>", Action: "exit insert mode"},
			{Mode: "insert", Keybind: "<Esc>", Action: "exit insert mode"},

			// Normal - Movement
			{Mode: "normal", Keybind: "5h 20j 3k 4l", Action: "move cursor left/down/up/right by amount"},
			{Mode: "normal", Keybind: "h j k l", Action: "move cursor left/down/up/right"},
			{Mode: "normal", Keybind: "w", Action: "move to next word"},
			{Mode: "normal", Keybind: "b", Action: "move to previous word"},
			{Mode: "normal", Keybind: "gg", Action: "go to top of file"},
			{Mode: "normal", Keybind: "G", Action: "go to bottom of file"},
			{Mode: "normal", Keybind: "0", Action: "go to beginning of line"},
			{Mode: "normal", Keybind: "$", Action: "go to end of line"},
			{Mode: "normal", Keybind: "<C-f>", Action: "page down and centre, we added zz"},
			{Mode: "normal", Keybind: "<C-b>", Action: "page up and centre, we added zz"},

			// Normal - Editing
			{Mode: "normal", Keybind: "dd", Action: "delete (cut) current line"},
			{Mode: "normal", Keybind: "yy", Action: "yank current line"},
			{Mode: "normal", Keybind: "<S-P>", Action: "paste clipboard"},
			{Mode: "normal", Keybind: "p", Action: "paste after cursor"},
			{Mode: "normal", Keybind: "u", Action: "undo last change"},
			{Mode: "normal", Keybind: "<C-r>", Action: "redo last undone change"},

			// Normal - Search
			{Mode: "normal", Keybind: "/", Action: "search forward"},
			{Mode: "normal", Keybind: "?", Action: "search backward"},
			{Mode: "normal", Keybind: "n", Action: "next search match (need to use / or ? before hand)"},
			{Mode: "normal", Keybind: "N", Action: "previous search match (need to use / or ? before hand)"},

			// Normal - LSP (if configured)
			{Mode: "normal", Keybind: "<C-o>", Action: "jump back from jump list (anything that moves the cursor counts as this)"},
			{Mode: "normal", Keybind: "<C-i>", Action: "jump forward in jump list (anything that moves the cursor counts as this)"},
			{Mode: "normal", Keybind: "gd", Action: "go to definition (LSP if attached)"},
			{Mode: "normal", Keybind: "grr", Action: "show references (LSP)"},
			{Mode: "normal", Keybind: "K", Action: "hover documentation (LSP or man page)"},

			// Your custom ones
			{Mode: "normal", Keybind: "<leader>h", Action: "open harpoon menu"},
			{Mode: "normal", Keybind: "<leader>a", Action: "append current file to harpoon"},
			{Mode: "normal", Keybind: "<leader>fo", Action: "format and organize imports"},
			{Mode: "normal", Keybind: "<leader>ff", Action: "find file in project"},
			{Mode: "normal", Keybind: "di\"", Action: "delete inside current double quotes"},
			{Mode: "normal", Keybind: "da\"", Action: "delete around current double quotes (including quotes)"},

			{Mode: "normal", Keybind: "dw", Action: "delete from cursor to start of next word"},
			{Mode: "normal", Keybind: "db", Action: "delete from cursor to start of previous word"},

			{Mode: "normal", Keybind: "diw", Action: "delete inner word (current word only)"},
			{Mode: "normal", Keybind: "daw", Action: "delete around word (word plus surrounding space)"},

			{Mode: "normal", Keybind: "ciw", Action: "change inner word (delete current word, and puts in insert mode)"},
			{Mode: "normal", Keybind: "yiw", Action: "yank inner word"},

			{Mode: "normal", Keybind: "di(", Action: "delete inside parentheses"},
			{Mode: "normal", Keybind: "da(", Action: "delete around parentheses"},
		},
		Height: 7,
	}
}

func resolveConfigPath(args []string, envValue string) (string, error) {
	flags := flag.NewFlagSet("nvim-simple-keybind-helper", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	configPathFlag := flags.String("config", "", "path to JSON config file")
	if err := flags.Parse(args); err != nil {
		return "", err
	}

	if path := strings.TrimSpace(*configPathFlag); path != "" {
		return path, nil
	}

	return strings.TrimSpace(envValue), nil
}

func loadConfig(path string) (appConfig, error) {
	defaultCfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return appConfig{}, err
	}

	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{}, err
	}

	if len(cfg.Columns) == 0 {
		cfg.Columns = defaultCfg.Columns
	}

	if cfg.Height <= 0 {
		cfg.Height = defaultCfg.Height
	}

	if cfg.Rows == nil {
		cfg.Rows = []rowConfig{}
	}

	return cfg, nil
}

func configColumnsToTableColumns(columns []columnConfig) []table.Column {
	tableColumns := make([]table.Column, 0, len(columns))

	for _, column := range columns {
		tableColumns = append(tableColumns, table.Column{Title: column.Title, Width: column.Width})
	}

	return tableColumns
}

func configRowsToTableRows(rows []rowConfig) []table.Row {
	tableRows := make([]table.Row, 0, len(rows))

	for _, row := range rows {
		tableRows = append(tableRows, table.Row{row.Mode, row.Keybind, row.Action})
	}

	return tableRows
}

func main() {
	cfg := defaultConfig()

	configPath, err := resolveConfigPath(os.Args[1:], os.Getenv("NVIM_HELPER_CONFIG"))
	if err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if configPath != "" {
		loadedCfg, err := loadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config from %s: %v\n", configPath, err)
			os.Exit(1)
		}

		cfg = loadedCfg
	}

	columns := configColumnsToTableColumns(cfg.Columns)
	rows := configRowsToTableRows(cfg.Rows)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(cfg.Height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))

	t.SetStyles(s)

	search := textinput.New()
	search.Placeholder = "Search actions..."
	search.Width = 30

	m := model{
		table:   t,
		allRows: rows,
		search:  search,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
