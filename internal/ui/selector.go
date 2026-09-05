package ui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/d4rthvadr/node-cleaner/pkg/models"
	"github.com/dustin/go-humanize"
)

type SelectionModel struct {
	table         table.Model
	folders       []models.DependencyFolder
	selected      map[int]bool
	totalSelected int64
	totalSize     int64
}

func NewSelectionModel(folders []models.DependencyFolder) *SelectionModel {

	columns := []table.Column{
		{Title: "Select", Width: 8},
		{Title: "Size", Width: 12},
		{Title: "Last Accessed", Width: 20},
		{Title: "Path", Width: 50},
	}

	rows := make([]table.Row, len(folders))

	for i, folder := range folders {
		rows[i] = table.Row{
			"[ ]",
			humanize.Bytes(uint64(folder.Size)),
			humanize.Time(folder.AccessTime),
			folder.Path,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return &SelectionModel{
		table:    t,
		folders:  folders,
		selected: make(map[int]bool),
	}
}

func (m SelectionModel) Init() tea.Cmd {
	return nil
}

// Update provides minimal interaction to satisfy tea.Model and allow navigation.
func (m *SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case " ":
			idx := m.table.Cursor()
			m.selected[idx] = !m.selected[idx]

			// Update total
			if m.selected[idx] {
				m.totalSelected += m.folders[idx].Size
			} else {
				m.totalSelected -= m.folders[idx].Size
			}

			// Update row
			m.updateRow(idx)
			return m, nil
		case "enter":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table with a simple hint footer.
func (m SelectionModel) View() string {
	selectedCount := len(m.selected)
	for _, isSelected := range m.selected {
		if !isSelected {
			selectedCount--
		}
	}

	footer := "\n"
	footer += "Selected: " + humanize.Bytes(uint64(m.totalSelected))
	footer += " (" + humanize.Comma(int64(selectedCount)) + " folders)\n"
	footer += "\nControls: [Space] Toggle  [Enter] Confirm  [q/Esc] Cancel\n"

	return m.table.View() + footer
}

func (m *SelectionModel) updateRow(idx int) {
	// Get all current rows
	rows := make([]table.Row, len(m.folders))

	// Rebuild all rows with updated selection states
	for i, folder := range m.folders {
		checkmark := "[ ]"
		if m.selected[i] {
			checkmark = "[x]"
		}
		rows[i] = table.Row{
			checkmark,
			humanize.Bytes(uint64(folder.Size)),
			humanize.Time(folder.AccessTime),
			folder.Path,
		}
	}

	// Update the table with all rows
	m.table.SetRows(rows)
}

func (m *SelectionModel) GetSelectedFolders() []models.DependencyFolder {
	var selected []models.DependencyFolder
	for idx, isSelected := range m.selected {
		if isSelected {
			selected = append(selected, m.folders[idx])
		}
	}
	return selected

}
