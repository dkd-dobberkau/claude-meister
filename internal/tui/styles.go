package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorRed    = lipgloss.Color("#FF5F56")
	ColorYellow = lipgloss.Color("#FFBD2E")
	ColorGreen  = lipgloss.Color("#27C93F")
	ColorBlue   = lipgloss.Color("#0A84FF")
	ColorGray   = lipgloss.Color("#666666")
	ColorWhite  = lipgloss.Color("#FFFFFF")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorBlue).
			Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	DirtyStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	CleanStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	BranchStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlue)
)
