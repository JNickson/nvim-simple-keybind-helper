package main

import (
	"fmt"
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

func main() {
	columns := []table.Column{
		{Title: "Mode", Width: 8},
		{Title: "Keybind", Width: 16},
		{Title: "Action", Width: 80},
	}

	rows := []table.Row{
		// Visual
		{"visual", "y", "yank (copy) selection"},
		{"visual", ">", "indent selection right"},
		{"visual", "<", "indent selection left"},

		// Insert
		{"insert", "<C-h>", "delete previous character"},
		{"insert", "<C-w>", "delete all from the cursor back to  word boundary (space or punctuation)"},
		{"insert", "<C-c>", "exit insert mode"},
		{"insert", "<Esc>", "exit insert mode"},

		// Normal - Movement
		{"normal", "5h 20j 3k 4l", "move cursor left/down/up/right by amount"},
		{"normal", "h j k l", "move cursor left/down/up/right"},
		{"normal", "w", "move to next word"},
		{"normal", "b", "move to previous word"},
		{"normal", "gg", "go to top of file"},
		{"normal", "G", "go to bottom of file"},
		{"normal", "0", "go to beginning of line"},
		{"normal", "$", "go to end of line"},

		// Normal - Editing
		{"normal", "dd", "delete (cut) current line"},
		{"normal", "yy", "yank current line"},
		{"normal", "<S-P>", "paste clipboard"},
		{"normal", "p", "paste after cursor"},
		{"normal", "u", "undo last change"},
		{"normal", "<C-r>", "redo last undone change"},

		// Normal - Search
		{"normal", "/", "search forward"},
		{"normal", "?", "search backward"},
		{"normal", "n", "next search match (need to use / or ? before hand)"},
		{"normal", "N", "previous search match (need to use / or ? before hand)"},

		// Normal - LSP (if configured)
		{"normal", "<C-o>", "jump back from jump list (anything that moves the cursor counts as this)"},
		{"normal", "<C-i>", "jump forward in jump list (anything that moves the cursor counts as this)"},
		{"normal", "gd", "go to definition (LSP if attached)"},
		{"normal", "grr", "show references (LSP)"},
		{"normal", "K", "hover documentation (LSP or man page)"},

		// Your custom ones
		{"normal", "<leader>h", "open harpoon menu"},
		{"normal", "<leader>a", "append current file to harpoon"},
		{"normal", "<leader>fo", "format and organize imports"},
		{"normal", "<leader>ff", "find file in project"},
		{"normal", "di\"", "delete inside current double quotes"},
		{"normal", "da\"", "delete around current double quotes (including quotes)"},

		{"normal", "dw", "delete from cursor to start of next word"},
		{"normal", "db", "delete from cursor to start of previous word"},

		{"normal", "diw", "delete inner word (current word only)"},
		{"normal", "daw", "delete around word (word plus surrounding space)"},

		{"normal", "ciw", "change inner word (delete current word, and puts in insert mode)"},
		{"normal", "yiw", "yank inner word"},

		{"normal", "di(", "delete inside parentheses"},
		{"normal", "da(", "delete around parentheses"},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
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
