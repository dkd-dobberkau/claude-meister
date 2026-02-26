package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dkd-dobberkau/claude-meister/internal/project"
	"github.com/dkd-dobberkau/claude-meister/internal/squirrel"
)

type view int

const (
	viewList view = iota
	viewDetail
)

// Category tabs
type category int

const (
	catOpenWork category = iota
	catRecent
	catSleeping
	catAll
)

func (c category) String() string {
	switch c {
	case catOpenWork:
		return "Open Work"
	case catRecent:
		return "Recent"
	case catSleeping:
		return "Sleeping"
	case catAll:
		return "All"
	}
	return ""
}

// actionResultMsg is returned by async git action commands.
type actionResultMsg struct {
	success bool
	message string
}

type Model struct {
	status        *squirrel.Status
	table         table.Model
	category      category
	view          view
	selected      *squirrel.Project
	width         int
	height        int
	message       string
	confirming    bool
	confirmAction string
	confirmLabel  string
}

func NewModel(status *squirrel.Status) Model {
	m := Model{
		status:   status,
		category: catOpenWork,
		view:     viewList,
	}
	m.table = m.buildTable()
	return m
}

func commitAction(path string) tea.Cmd {
	return func() tea.Msg {
		err := project.GitCommitAll(path, "chore: claude-meister cleanup")
		if err != nil {
			return actionResultMsg{success: false, message: fmt.Sprintf("Commit failed: %v", err)}
		}
		return actionResultMsg{success: true, message: "Changes committed successfully"}
	}
}

func stashAction(path string) tea.Cmd {
	return func() tea.Msg {
		err := project.GitStash(path)
		if err != nil {
			return actionResultMsg{success: false, message: fmt.Sprintf("Stash failed: %v", err)}
		}
		return actionResultMsg{success: true, message: "Changes stashed successfully"}
	}
}

func discardAction(path string) tea.Cmd {
	return func() tea.Msg {
		err := project.GitDiscardAll(path)
		if err != nil {
			return actionResultMsg{success: false, message: fmt.Sprintf("Discard failed: %v", err)}
		}
		return actionResultMsg{success: true, message: "All changes discarded"}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table = m.buildTable()
		return m, nil

	case actionResultMsg:
		m.message = msg.message
		m.confirming = false
		m.confirmAction = ""
		m.confirmLabel = ""
		m.view = viewList
		m.selected = nil
		return m, nil

	case tea.KeyMsg:
		// If a message is displayed, clear it on any key press
		if m.message != "" {
			m.message = ""
			return m, nil
		}

		// Handle confirmation prompt
		if m.confirming && m.selected != nil {
			switch msg.String() {
			case "y":
				path := m.selected.Path
				switch m.confirmAction {
				case "commit":
					return m, commitAction(path)
				case "stash":
					return m, stashAction(path)
				case "discard":
					return m, discardAction(path)
				}
			case "n", "esc":
				m.confirming = false
				m.confirmAction = ""
				m.confirmLabel = ""
			}
			return m, nil
		}

		switch {
		case msg.String() == "q" || msg.String() == "ctrl+c":
			return m, tea.Quit

		case msg.String() == "tab":
			if m.view == viewList {
				m.category = (m.category + 1) % 4
				m.table = m.buildTable()
			}
			return m, nil

		case msg.String() == "enter":
			if m.view == viewList {
				row := m.table.SelectedRow()
				if row != nil {
					p := m.status.FindByName(row[0])
					if p != nil {
						m.selected = p
						m.view = viewDetail
					}
				}
			}
			return m, nil

		case msg.String() == "esc":
			if m.view == viewDetail {
				m.view = viewList
				m.selected = nil
			}
			return m, nil

		// Detail view action keys
		case msg.String() == "c":
			if m.view == viewDetail {
				m.confirming = true
				m.confirmAction = "commit"
				m.confirmLabel = "Commit all changes?"
			}
			return m, nil

		case msg.String() == "s":
			if m.view == viewDetail {
				m.confirming = true
				m.confirmAction = "stash"
				m.confirmLabel = "Stash all changes?"
			}
			return m, nil

		case msg.String() == "x":
			if m.view == viewDetail {
				m.confirming = true
				m.confirmAction = "discard"
				m.confirmLabel = "Discard ALL changes? This cannot be undone!"
			}
			return m, nil
		}
	}

	if m.view == viewList {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.view == viewDetail && m.selected != nil {
		return m.renderDetail()
	}
	return m.renderList()
}

func (m Model) renderList() string {
	var b strings.Builder

	// Title
	title := TitleStyle.Render(" claude-meister ")
	b.WriteString(title + "\n\n")

	// Category tabs with counts
	tabs := []struct {
		cat   category
		count int
	}{
		{catOpenWork, len(m.status.OpenWork)},
		{catRecent, len(m.status.RecentActivity)},
		{catSleeping, len(m.status.Sleeping)},
		{catAll, len(m.status.AllProjects())},
	}

	var tabParts []string
	for _, t := range tabs {
		label := fmt.Sprintf(" %s (%d) ", t.cat.String(), t.count)
		if t.cat == m.category {
			tabParts = append(tabParts, TabActiveStyle.Render(label))
		} else {
			tabParts = append(tabParts, TabInactiveStyle.Render(label))
		}
	}
	b.WriteString(strings.Join(tabParts, " | ") + "\n\n")

	// Table
	b.WriteString(m.table.View() + "\n\n")

	// Message banner
	if m.message != "" {
		b.WriteString(fmt.Sprintf("  %s\n\n", m.message))
	}

	// Help
	help := HelpStyle.Render("[c]lean  [a]rchive  [d]elete  [D]ocker-stop  [tab] category  [q]uit  [?]help")
	b.WriteString(help)

	return b.String()
}

func (m Model) renderDetail() string {
	p := m.selected
	var b strings.Builder

	title := TitleStyle.Render(fmt.Sprintf(" %s ", p.ShortName))
	b.WriteString(title + "\n\n")

	b.WriteString(fmt.Sprintf("  Path:     %s\n", p.Path))
	b.WriteString(fmt.Sprintf("  Branch:   %s\n", BranchStyle.Render(p.GitBranch)))

	if p.GitDirty {
		b.WriteString(fmt.Sprintf("  Status:   %s\n", DirtyStyle.Render(fmt.Sprintf("%d uncommitted files", p.UncommittedFiles))))
	} else {
		b.WriteString(fmt.Sprintf("  Status:   %s\n", CleanStyle.Render("clean")))
	}

	b.WriteString(fmt.Sprintf("  Prompts:  %d\n", p.PromptCount))
	b.WriteString(fmt.Sprintf("  Idle:     %d days\n", p.DaysSinceActive))
	b.WriteString(fmt.Sprintf("  Score:    %.0f\n", p.Score))

	if p.LastPrompt != "" {
		prompt := p.LastPrompt
		if len(prompt) > 60 {
			prompt = prompt[:57] + "..."
		}
		b.WriteString(fmt.Sprintf("  Last:     %q\n", prompt))
	}

	b.WriteString("\n")

	if m.confirming {
		b.WriteString(DirtyStyle.Render(fmt.Sprintf("  %s", m.confirmLabel)))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("  [y]es  [n]o  [esc] cancel"))
	} else if m.message != "" {
		b.WriteString(fmt.Sprintf("  %s\n", m.message))
		b.WriteString(HelpStyle.Render("  Press any key to continue"))
	} else {
		b.WriteString(HelpStyle.Render("  [c]ommit  [s]tash  [x] discard  [esc] back"))
	}

	return b.String()
}

func (m Model) currentProjects() []squirrel.Project {
	switch m.category {
	case catOpenWork:
		return m.status.OpenWork
	case catRecent:
		return m.status.RecentActivity
	case catSleeping:
		return m.status.Sleeping
	case catAll:
		return m.status.AllProjects()
	}
	return nil
}

func (m Model) buildTable() table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 28},
		{Title: "Branch", Width: 22},
		{Title: "Dirty", Width: 7},
		{Title: "Days", Width: 6},
		{Title: "Score", Width: 7},
	}

	projects := m.currentProjects()
	rows := make([]table.Row, len(projects))
	for i, p := range projects {
		dirty := ""
		if p.GitDirty {
			dirty = fmt.Sprintf("* %d", p.UncommittedFiles)
		}
		branch := p.GitBranch
		if len(branch) > 20 {
			branch = branch[:17] + "..."
		}
		rows[i] = table.Row{
			p.ShortName,
			branch,
			dirty,
			fmt.Sprintf("%d", p.DaysSinceActive),
			fmt.Sprintf("%.0f", p.Score),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+1, 20)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}
