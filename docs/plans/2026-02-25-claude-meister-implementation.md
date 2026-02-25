# claude-meister Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI/TUI tool that reads squirrel output and helps clean up forgotten Claude Code projects (git cleanup, docker stop, archive, delete).

**Architecture:** Hybrid CLI+TUI. Cobra for commands, Bubble Tea for interactive TUI. Squirrel JSON as data source. go-git for git ops. Config via YAML.

**Tech Stack:** Go, cobra, bubbletea, lipgloss, bubbles, go-git, yaml.v3

---

## Phase 1: Project Skeleton + Squirrel Adapter

### Task 1: Initialize Go module and install dependencies

**Files:**
- Create: `go.mod`
- Create: `main.go`

**Step 1: Initialize Go module**

Run:
```bash
cd /Users/olivier/Versioncontrol/local/claude-meister
go mod init github.com/olivercodes/claude-meister
```
Expected: `go.mod` created

**Step 2: Install dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/go-git/go-git/v5@latest
go get gopkg.in/yaml.v3@latest
```

**Step 3: Create main.go entrypoint**

```go
// main.go
package main

import (
	"fmt"
	"os"

	"github.com/olivercodes/claude-meister/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 4: Create root command**

```go
// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-meister",
	Short: "Clean up forgotten Claude Code projects",
	Long:  "claude-meister uses squirrel to find forgotten projects and helps you clean them up: commit/discard git changes, stop Docker containers, archive or delete projects.",
}

func Execute() error {
	return rootCmd.Execute()
}
```

**Step 5: Verify it compiles and runs**

Run: `go run . --help`
Expected: Help text with "Clean up forgotten Claude Code projects"

**Step 6: Commit**

```bash
git add main.go cmd/root.go go.mod go.sum
git commit -m "feat: initialize Go module with cobra root command"
```

---

### Task 2: Squirrel adapter - data types and JSON parsing

**Files:**
- Create: `internal/squirrel/types.go`
- Create: `internal/squirrel/parse.go`
- Create: `internal/squirrel/parse_test.go`

**Step 1: Define squirrel data types**

```go
// internal/squirrel/types.go
package squirrel

import "time"

type Project struct {
	Path             string    `json:"path"`
	ShortName        string    `json:"shortName"`
	PromptCount      int       `json:"promptCount"`
	LastActivity     time.Time `json:"lastActivity"`
	FirstActivity    time.Time `json:"firstActivity"`
	LastPrompt       string    `json:"lastPrompt"`
	GitDirty         bool      `json:"gitDirty"`
	GitBranch        string    `json:"gitBranch"`
	UncommittedFiles int       `json:"uncommittedFiles"`
	DaysSinceActive  int       `json:"daysSinceActive"`
	IsOpenWork       bool      `json:"isOpenWork"`
	Score            float64   `json:"score"`
}

type Status struct {
	OpenWork       []Project `json:"openWork"`
	RecentActivity []Project `json:"recentActivity"`
	Sleeping       []Project `json:"sleeping"`
	Acknowledged   []Project `json:"acknowledged"`
}

// AllProjects returns all projects across all categories.
func (s *Status) AllProjects() []Project {
	var all []Project
	all = append(all, s.OpenWork...)
	all = append(all, s.RecentActivity...)
	all = append(all, s.Sleeping...)
	return all
}

// FindByName finds a project by short name across all categories.
func (s *Status) FindByName(name string) *Project {
	for i := range s.OpenWork {
		if s.OpenWork[i].ShortName == name {
			return &s.OpenWork[i]
		}
	}
	for i := range s.RecentActivity {
		if s.RecentActivity[i].ShortName == name {
			return &s.RecentActivity[i]
		}
	}
	for i := range s.Sleeping {
		if s.Sleeping[i].ShortName == name {
			return &s.Sleeping[i]
		}
	}
	for i := range s.Acknowledged {
		if s.Acknowledged[i].ShortName == name {
			return &s.Acknowledged[i]
		}
	}
	return nil
}
```

**Step 2: Write the failing test for JSON parsing**

```go
// internal/squirrel/parse_test.go
package squirrel

import (
	"testing"
)

const testJSON = `{
  "openWork": [
    {
      "path": "/Users/test/project-a",
      "shortName": "project-a",
      "promptCount": 42,
      "lastActivity": "2026-02-25T10:00:00.000+01:00",
      "firstActivity": "2026-02-20T08:00:00.000+01:00",
      "lastPrompt": "fix the bug",
      "gitDirty": true,
      "gitBranch": "main",
      "uncommittedFiles": 3,
      "daysSinceActive": 0,
      "isOpenWork": true,
      "score": 95.5
    }
  ],
  "recentActivity": [
    {
      "path": "/Users/test/project-b",
      "shortName": "project-b",
      "promptCount": 10,
      "lastActivity": "2026-02-24T10:00:00.000+01:00",
      "firstActivity": "2026-02-24T08:00:00.000+01:00",
      "lastPrompt": "hello",
      "gitDirty": false,
      "gitBranch": "feature/x",
      "uncommittedFiles": 0,
      "daysSinceActive": 1,
      "isOpenWork": false,
      "score": 50.0
    }
  ],
  "sleeping": [],
  "acknowledged": []
}`

func TestParseStatus(t *testing.T) {
	status, err := ParseStatus([]byte(testJSON))
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}

	if len(status.OpenWork) != 1 {
		t.Fatalf("expected 1 open work project, got %d", len(status.OpenWork))
	}

	p := status.OpenWork[0]
	if p.ShortName != "project-a" {
		t.Errorf("expected shortName 'project-a', got %q", p.ShortName)
	}
	if p.PromptCount != 42 {
		t.Errorf("expected promptCount 42, got %d", p.PromptCount)
	}
	if !p.GitDirty {
		t.Error("expected gitDirty to be true")
	}
	if p.UncommittedFiles != 3 {
		t.Errorf("expected 3 uncommitted files, got %d", p.UncommittedFiles)
	}
	if p.Score != 95.5 {
		t.Errorf("expected score 95.5, got %f", p.Score)
	}
}

func TestParseStatus_InvalidJSON(t *testing.T) {
	_, err := ParseStatus([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAllProjects(t *testing.T) {
	status, _ := ParseStatus([]byte(testJSON))
	all := status.AllProjects()
	if len(all) != 2 {
		t.Fatalf("expected 2 total projects, got %d", len(all))
	}
}

func TestFindByName(t *testing.T) {
	status, _ := ParseStatus([]byte(testJSON))

	p := status.FindByName("project-a")
	if p == nil {
		t.Fatal("expected to find project-a")
	}
	if p.Path != "/Users/test/project-a" {
		t.Errorf("wrong path: %s", p.Path)
	}

	missing := status.FindByName("nonexistent")
	if missing != nil {
		t.Error("expected nil for nonexistent project")
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/squirrel/ -v`
Expected: FAIL (ParseStatus undefined)

**Step 4: Implement ParseStatus**

```go
// internal/squirrel/parse.go
package squirrel

import "encoding/json"

// ParseStatus parses the JSON output of `squirrel status --json`.
func ParseStatus(data []byte) (*Status, error) {
	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/squirrel/ -v`
Expected: PASS (all 4 tests)

**Step 6: Commit**

```bash
git add internal/squirrel/
git commit -m "feat: add squirrel types and JSON parser with tests"
```

---

### Task 3: Squirrel adapter - run squirrel CLI

**Files:**
- Create: `internal/squirrel/runner.go`
- Create: `internal/squirrel/runner_test.go`

**Step 1: Write the failing test**

```go
// internal/squirrel/runner_test.go
package squirrel

import (
	"testing"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner(30, "deep")
	if r.Days != 30 {
		t.Errorf("expected days=30, got %d", r.Days)
	}
	if r.Depth != "deep" {
		t.Errorf("expected depth=deep, got %s", r.Depth)
	}
}

func TestBuildArgs(t *testing.T) {
	r := NewRunner(14, "medium")
	args := r.buildArgs()
	expected := []string{"status", "--json", "--days", "14", "--depth", "medium"}
	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i, a := range args {
		if a != expected[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, expected[i], a)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/squirrel/ -v -run TestNewRunner`
Expected: FAIL

**Step 3: Implement Runner**

```go
// internal/squirrel/runner.go
package squirrel

import (
	"fmt"
	"os/exec"
)

// Runner executes the squirrel CLI and returns parsed results.
type Runner struct {
	Days  int
	Depth string
}

// NewRunner creates a Runner with the given parameters.
func NewRunner(days int, depth string) *Runner {
	return &Runner{Days: days, Depth: depth}
}

func (r *Runner) buildArgs() []string {
	return []string{
		"status", "--json",
		"--days", fmt.Sprintf("%d", r.Days),
		"--depth", r.Depth,
	}
}

// Run executes squirrel and returns the parsed status.
func (r *Runner) Run() (*Status, error) {
	cmd := exec.Command("squirrel", r.buildArgs()...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("squirrel failed: %w", err)
	}
	return ParseStatus(out)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/squirrel/ -v`
Expected: PASS (all tests including new ones)

**Step 5: Commit**

```bash
git add internal/squirrel/runner.go internal/squirrel/runner_test.go
git commit -m "feat: add squirrel CLI runner"
```

---

### Task 4: Config package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SquirrelDays != 30 {
		t.Errorf("expected days=30, got %d", cfg.SquirrelDays)
	}
	if cfg.SquirrelDepth != "deep" {
		t.Errorf("expected depth=deep, got %q", cfg.SquirrelDepth)
	}
	if cfg.ArchivePath == "" {
		t.Error("expected non-empty archive path")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("missing config file should return defaults, got error: %v", err)
	}
	if cfg.SquirrelDays != 30 {
		t.Errorf("expected default days=30, got %d", cfg.SquirrelDays)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := []byte(`archive_path: /tmp/my-archive
squirrel_days: 7
squirrel_depth: quick
ignore_paths:
  - /Users/test/important
`)
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.ArchivePath != "/tmp/my-archive" {
		t.Errorf("expected archive path /tmp/my-archive, got %q", cfg.ArchivePath)
	}
	if cfg.SquirrelDays != 7 {
		t.Errorf("expected days=7, got %d", cfg.SquirrelDays)
	}
	if len(cfg.IgnorePaths) != 1 || cfg.IgnorePaths[0] != "/Users/test/important" {
		t.Errorf("unexpected ignore paths: %v", cfg.IgnorePaths)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL

**Step 3: Implement config**

```go
// internal/config/config.go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ArchivePath   string   `yaml:"archive_path"`
	SquirrelDays  int      `yaml:"squirrel_days"`
	SquirrelDepth string   `yaml:"squirrel_depth"`
	IgnorePaths   []string `yaml:"ignore_paths"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ArchivePath:   filepath.Join(home, "Archive", "projects"),
		SquirrelDays:  30,
		SquirrelDepth: "deep",
		IgnorePaths:   nil,
	}
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude-meister", "config.yaml")
}

// Load reads config from a YAML file. Returns defaults if file doesn't exist.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: PASS (all 3 tests)

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with YAML loading and defaults"
```

---

### Task 5: Scan command - basic TUI with project table

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/keys.go`
- Create: `internal/tui/model.go`
- Create: `cmd/scan.go`

**Step 1: Create TUI styles**

```go
// internal/tui/styles.go
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
```

**Step 2: Create key bindings**

```go
// internal/tui/keys.go
package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Clean   key.Binding
	Archive key.Binding
	Delete  key.Binding
	Docker  key.Binding
	Tab     key.Binding
	Help    key.Binding
	Quit    key.Binding
}

var Keys = KeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
	Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Clean:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clean")),
	Archive: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
	Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Docker:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "docker-stop")),
	Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next category")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
```

**Step 3: Create the main TUI model**

```go
// internal/tui/model.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olivercodes/claude-meister/internal/squirrel"
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

type Model struct {
	status   *squirrel.Status
	table    table.Model
	category category
	view     view
	selected *squirrel.Project
	width    int
	height   int
	message  string
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

	case tea.KeyMsg:
		switch {
		case msg.String() == "q" || msg.String() == "ctrl+c":
			return m, tea.Quit

		case msg.String() == "tab":
			m.category = (m.category + 1) % 4
			m.table = m.buildTable()
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
	b.WriteString(HelpStyle.Render("  [c]lean  [a]rchive  [d]elete  [D]ocker-stop  [esc] back"))

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
```

**Step 4: Create scan command**

```go
// cmd/scan.go
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/olivercodes/claude-meister/internal/tui"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Interactive TUI overview of all projects",
	Long:  "Opens an interactive terminal UI showing all projects from squirrel, organized by category. Navigate with arrow keys, press Enter for details.",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	fmt.Fprintln(os.Stderr, "Scanning projects with squirrel...")
	status, err := runner.Run()
	if err != nil {
		return fmt.Errorf("running squirrel: %w", err)
	}

	model := tui.NewModel(status)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
```

**Step 5: Verify it compiles**

Run: `go build .`
Expected: Compiles without errors

**Step 6: Manual test**

Run: `go run . scan`
Expected: TUI opens with project table, navigable with arrow keys, Tab switches categories, q quits

**Step 7: Commit**

```bash
git add internal/tui/ cmd/scan.go
git commit -m "feat: add scan command with interactive TUI project table"
```

---

## Phase 2: Detail View + Clean Command

### Task 6: Project package - git status helper

**Files:**
- Create: `internal/project/git.go`
- Create: `internal/project/git_test.go`

**Step 1: Write the failing test**

```go
// internal/project/git_test.go
package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repo for testing.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("setup %v failed: %s %v", args, out, err)
		}
	}

	// Create initial commit
	f := filepath.Join(dir, "README.md")
	os.WriteFile(f, []byte("# test"), 0644)
	c := exec.Command("git", "add", ".")
	c.Dir = dir
	c.Run()
	c = exec.Command("git", "commit", "-m", "initial")
	c.Dir = dir
	c.Run()

	return dir
}

func TestGitStatus_Clean(t *testing.T) {
	dir := setupTestRepo(t)
	status, err := GitStatus(dir)
	if err != nil {
		t.Fatalf("GitStatus failed: %v", err)
	}
	if status.Dirty {
		t.Error("expected clean repo")
	}
	if len(status.ModifiedFiles) != 0 {
		t.Errorf("expected 0 modified files, got %d", len(status.ModifiedFiles))
	}
	if len(status.UntrackedFiles) != 0 {
		t.Errorf("expected 0 untracked files, got %d", len(status.UntrackedFiles))
	}
}

func TestGitStatus_Dirty(t *testing.T) {
	dir := setupTestRepo(t)

	// Create untracked file
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)
	// Modify tracked file
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# changed"), 0644)

	status, err := GitStatus(dir)
	if err != nil {
		t.Fatalf("GitStatus failed: %v", err)
	}
	if !status.Dirty {
		t.Error("expected dirty repo")
	}
	if len(status.UntrackedFiles) != 1 {
		t.Errorf("expected 1 untracked file, got %d", len(status.UntrackedFiles))
	}
	if len(status.ModifiedFiles) != 1 {
		t.Errorf("expected 1 modified file, got %d", len(status.ModifiedFiles))
	}
}

func TestGitStatus_Branch(t *testing.T) {
	dir := setupTestRepo(t)
	status, err := GitStatus(dir)
	if err != nil {
		t.Fatalf("GitStatus failed: %v", err)
	}
	if status.Branch == "" {
		t.Error("expected non-empty branch name")
	}
}
```

**Step 2: Run tests to verify failure**

Run: `go test ./internal/project/ -v`
Expected: FAIL

**Step 3: Implement GitStatus**

```go
// internal/project/git.go
package project

import (
	"fmt"
	"os/exec"
	"strings"
)

type GitStatusResult struct {
	Branch         string
	Dirty          bool
	ModifiedFiles  []string
	UntrackedFiles []string
	StagedFiles    []string
	StashCount     int
}

// GitStatus returns the git status of a project directory.
func GitStatus(path string) (*GitStatusResult, error) {
	result := &GitStatusResult{}

	// Get branch name
	branch, err := gitCmd(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("getting branch: %w", err)
	}
	result.Branch = strings.TrimSpace(branch)

	// Get porcelain status
	status, err := gitCmd(path, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("getting status: %w", err)
	}

	for _, line := range strings.Split(status, "\n") {
		if line == "" {
			continue
		}
		if len(line) < 3 {
			continue
		}
		xy := line[:2]
		file := strings.TrimSpace(line[2:])

		switch {
		case strings.HasPrefix(xy, "??"):
			result.UntrackedFiles = append(result.UntrackedFiles, file)
		case xy[0] != ' ' && xy[0] != '?':
			result.StagedFiles = append(result.StagedFiles, file)
		case xy[1] != ' ':
			result.ModifiedFiles = append(result.ModifiedFiles, file)
		}
	}

	result.Dirty = len(result.ModifiedFiles) > 0 || len(result.UntrackedFiles) > 0 || len(result.StagedFiles) > 0

	// Get stash count
	stash, _ := gitCmd(path, "stash", "list")
	if stash != "" {
		result.StashCount = len(strings.Split(strings.TrimSpace(stash), "\n"))
	}

	return result, nil
}

// GitCommitAll stages and commits all changes.
func GitCommitAll(path, message string) error {
	if _, err := gitCmd(path, "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	if _, err := gitCmd(path, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// GitStash stashes all changes.
func GitStash(path string) error {
	_, err := gitCmd(path, "stash", "push", "-m", "claude-meister auto-stash")
	return err
}

// GitDiscardAll discards all uncommitted changes.
func GitDiscardAll(path string) error {
	if _, err := gitCmd(path, "checkout", "."); err != nil {
		return fmt.Errorf("git checkout: %w", err)
	}
	if _, err := gitCmd(path, "clean", "-fd"); err != nil {
		return fmt.Errorf("git clean: %w", err)
	}
	return nil
}

func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(out))
	}
	return string(out), nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/project/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/project/
git commit -m "feat: add git status and operations helpers with tests"
```

---

### Task 7: Clean command (CLI mode)

**Files:**
- Create: `cmd/clean.go`

**Step 1: Implement clean command**

```go
// cmd/clean.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/project"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var cleanDryRun bool

var cleanCmd = &cobra.Command{
	Use:   "clean <project-name>",
	Short: "Clean up git state of a project",
	Long:  "Shows the git status of a project and offers options to commit, stash, or discard changes.",
	Args:  cobra.ExactArgs(1),
	RunE:  runClean,
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be done without doing it")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	status, err := runner.Run()
	if err != nil {
		return fmt.Errorf("running squirrel: %w", err)
	}

	p := status.FindByName(projectName)
	if p == nil {
		return fmt.Errorf("project %q not found in squirrel results", projectName)
	}

	gitStatus, err := project.GitStatus(p.Path)
	if err != nil {
		return fmt.Errorf("getting git status: %w", err)
	}

	// Display status
	fmt.Printf("\n  Project:  %s\n", p.ShortName)
	fmt.Printf("  Path:     %s\n", p.Path)
	fmt.Printf("  Branch:   %s\n", gitStatus.Branch)

	if !gitStatus.Dirty {
		fmt.Println("\n  Status: clean - nothing to do!")
		return nil
	}

	fmt.Printf("\n  Modified:  %d files\n", len(gitStatus.ModifiedFiles))
	for _, f := range gitStatus.ModifiedFiles {
		fmt.Printf("    M  %s\n", f)
	}
	fmt.Printf("  Untracked: %d files\n", len(gitStatus.UntrackedFiles))
	for _, f := range gitStatus.UntrackedFiles {
		fmt.Printf("    ?  %s\n", f)
	}
	fmt.Printf("  Staged:    %d files\n", len(gitStatus.StagedFiles))
	for _, f := range gitStatus.StagedFiles {
		fmt.Printf("    A  %s\n", f)
	}
	if gitStatus.StashCount > 0 {
		fmt.Printf("  Stashes:   %d\n", gitStatus.StashCount)
	}

	// Ask what to do
	fmt.Println("\n  What do you want to do?")
	fmt.Println("  [1] Commit all changes")
	fmt.Println("  [2] Stash changes")
	fmt.Println("  [3] Discard all changes")
	fmt.Println("  [4] Skip (do nothing)")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n  Choice: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if cleanDryRun {
		fmt.Println("\n  [dry-run] Would execute the selected action")
		return nil
	}

	switch choice {
	case "1":
		fmt.Print("  Commit message: ")
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = fmt.Sprintf("chore: claude-meister cleanup of %s", p.ShortName)
		}
		if err := project.GitCommitAll(p.Path, msg); err != nil {
			return fmt.Errorf("committing: %w", err)
		}
		fmt.Println("  Done! Changes committed.")

	case "2":
		if err := project.GitStash(p.Path); err != nil {
			return fmt.Errorf("stashing: %w", err)
		}
		fmt.Println("  Done! Changes stashed.")

	case "3":
		fmt.Printf("  Are you sure? This will discard ALL changes in %s [y/N]: ", p.ShortName)
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			fmt.Println("  Aborted.")
			return nil
		}
		if err := project.GitDiscardAll(p.Path); err != nil {
			return fmt.Errorf("discarding: %w", err)
		}
		fmt.Println("  Done! All changes discarded.")

	case "4":
		fmt.Println("  Skipped.")

	default:
		fmt.Println("  Invalid choice, skipped.")
	}

	return nil
}
```

**Step 2: Verify it compiles**

Run: `go build .`
Expected: Compiles

**Step 3: Manual test**

Run: `go run . clean --help`
Expected: Help text for clean command

**Step 4: Commit**

```bash
git add cmd/clean.go
git commit -m "feat: add clean command for interactive git cleanup"
```

---

### Task 8: Wire detail view actions in TUI

**Files:**
- Modify: `internal/tui/model.go` (add action dispatch from detail view)

**Step 1: Add action messages and detail-view key handling**

Add to `internal/tui/model.go` after the existing code. Modify the `Update` method to handle action keys in detail view, and add a confirmation flow. The TUI should dispatch to `project.GitCommitAll`, `project.GitStash`, `project.GitDiscardAll` when keys are pressed in the detail view.

Key additions:
- In detail view: `c` triggers commit, `s` triggers stash, `x` triggers discard
- Each action shows a confirmation prompt before executing
- After action completes, show result message and return to list view

Implementation approach:
- Add `confirming bool` and `confirmAction string` fields to Model
- When action key pressed: set confirming=true, show "Press Y to confirm"
- When Y pressed in confirming state: execute action, show result

```go
// Add these fields to the Model struct:
//   confirming    bool
//   confirmAction string
//   confirmLabel  string

// Add this to the detail-view key handling in Update():
// case msg.String() == "c" && m.view == viewDetail:
//     m.confirming = true
//     m.confirmAction = "commit"
//     m.confirmLabel = "Commit all changes?"
// case msg.String() == "y" && m.confirming:
//     // execute action based on m.confirmAction
//     // set m.message to result
//     // return to list view
```

Full implementation: integrate into the existing `Update` method, adding the confirmation flow and calling `project.GitCommitAll()` / `project.GitStash()` / `project.GitDiscardAll()` from within tea.Cmd functions that return result messages.

**Step 2: Verify it compiles**

Run: `go build .`

**Step 3: Manual test**

Run: `go run . scan` → navigate to a project → press c/s/x → verify confirmation flow

**Step 4: Commit**

```bash
git add internal/tui/model.go
git commit -m "feat: add action dispatch from TUI detail view"
```

---

## Phase 3: Archive + Delete Commands

### Task 9: Archive functionality

**Files:**
- Create: `internal/project/archive.go`
- Create: `internal/project/archive_test.go`
- Create: `cmd/archive.go`

**Step 1: Write failing test**

```go
// internal/project/archive_test.go
package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveProject(t *testing.T) {
	// Create source project
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "main.go"), []byte("package main"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "file.go"), []byte("package sub"), 0644)

	archiveBase := t.TempDir()

	dest, err := ArchiveProject(src, archiveBase)
	if err != nil {
		t.Fatalf("ArchiveProject failed: %v", err)
	}

	// Source should be gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source directory should be removed after archive")
	}

	// Dest should contain the files
	if _, err := os.Stat(filepath.Join(dest, "main.go")); err != nil {
		t.Error("main.go should exist in archive")
	}
	if _, err := os.Stat(filepath.Join(dest, "sub", "file.go")); err != nil {
		t.Error("sub/file.go should exist in archive")
	}
}
```

**Step 2: Run test → FAIL**

Run: `go test ./internal/project/ -v -run TestArchive`

**Step 3: Implement**

```go
// internal/project/archive.go
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchiveProject moves a project directory to the archive.
// Returns the destination path.
func ArchiveProject(projectPath, archiveBase string) (string, error) {
	// Create archive directory: archiveBase/YYYY-MM/projectName
	now := time.Now()
	monthDir := filepath.Join(archiveBase, now.Format("2006-01"))
	projectName := filepath.Base(projectPath)
	dest := filepath.Join(monthDir, projectName)

	if err := os.MkdirAll(monthDir, 0755); err != nil {
		return "", fmt.Errorf("creating archive dir: %w", err)
	}

	// Check dest doesn't already exist
	if _, err := os.Stat(dest); err == nil {
		dest = fmt.Sprintf("%s-%s", dest, now.Format("150405"))
	}

	if err := os.Rename(projectPath, dest); err != nil {
		return "", fmt.Errorf("moving project: %w", err)
	}

	return dest, nil
}

// DeleteProject removes a project directory entirely.
func DeleteProject(projectPath string) error {
	return os.RemoveAll(projectPath)
}
```

**Step 4: Run test → PASS**

**Step 5: Create archive command**

```go
// cmd/archive.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/project"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <project-name>",
	Short: "Archive a project to the archive directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	status, err := runner.Run()
	if err != nil {
		return err
	}

	p := status.FindByName(projectName)
	if p == nil {
		return fmt.Errorf("project %q not found", projectName)
	}

	fmt.Printf("\n  Project:  %s\n", p.ShortName)
	fmt.Printf("  Path:     %s\n", p.Path)
	fmt.Printf("  Archive:  %s\n\n", cfg.ArchivePath)

	if p.GitDirty {
		fmt.Printf("  WARNING: Project has %d uncommitted files!\n", p.UncommittedFiles)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("  Move project to archive? [y/N]: ")
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
		fmt.Println("  Aborted.")
		return nil
	}

	dest, err := project.ArchiveProject(p.Path, cfg.ArchivePath)
	if err != nil {
		return fmt.Errorf("archiving: %w", err)
	}

	fmt.Printf("  Archived to: %s\n", dest)

	// Acknowledge in squirrel
	ackCmd := fmt.Sprintf("squirrel ack %s", p.ShortName)
	fmt.Printf("  Run '%s' to remove from squirrel list\n", ackCmd)

	return nil
}
```

**Step 6: Commit**

```bash
git add internal/project/archive.go internal/project/archive_test.go cmd/archive.go
git commit -m "feat: add archive command and project archival"
```

---

### Task 10: Delete command

**Files:**
- Create: `cmd/delete.go`

**Step 1: Implement delete command**

```go
// cmd/delete.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/project"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <project-name>",
	Short: "Permanently delete a project",
	Long:  "Permanently deletes a project directory. Requires typing the project name to confirm. Offers to archive first.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	status, err := runner.Run()
	if err != nil {
		return err
	}

	p := status.FindByName(projectName)
	if p == nil {
		return fmt.Errorf("project %q not found", projectName)
	}

	fmt.Printf("\n  Project:  %s\n", p.ShortName)
	fmt.Printf("  Path:     %s\n", p.Path)

	if p.GitDirty {
		fmt.Printf("  WARNING: %d uncommitted files!\n", p.UncommittedFiles)
	}

	reader := bufio.NewReader(os.Stdin)

	// Offer archive first
	fmt.Printf("\n  Would you like to archive instead of delete? [Y/n]: ")
	archiveChoice, _ := reader.ReadString('\n')
	archiveChoice = strings.TrimSpace(strings.ToLower(archiveChoice))
	if archiveChoice == "" || archiveChoice == "y" {
		dest, err := project.ArchiveProject(p.Path, cfg.ArchivePath)
		if err != nil {
			return fmt.Errorf("archiving: %w", err)
		}
		fmt.Printf("  Archived to: %s\n", dest)
		return nil
	}

	// Double confirmation for delete
	fmt.Printf("\n  This will PERMANENTLY delete %s\n", p.Path)
	fmt.Printf("  Type the project name to confirm: ")
	confirmation, _ := reader.ReadString('\n')
	if strings.TrimSpace(confirmation) != p.ShortName {
		fmt.Println("  Names don't match. Aborted.")
		return nil
	}

	if err := project.DeleteProject(p.Path); err != nil {
		return fmt.Errorf("deleting: %w", err)
	}

	fmt.Println("  Deleted.")
	return nil
}
```

**Step 2: Verify compilation**

Run: `go build .`

**Step 3: Commit**

```bash
git add cmd/delete.go
git commit -m "feat: add delete command with double confirmation"
```

---

## Phase 4: Docker Stop Command

### Task 11: Docker detection and stop

**Files:**
- Create: `internal/project/docker.go`
- Create: `internal/project/docker_test.go`
- Create: `cmd/dockerstop.go`

**Step 1: Write failing test**

```go
// internal/project/docker_test.go
package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDocker_ComposeFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services:"), 0644)

	info := DetectDocker(dir)
	if !info.HasCompose {
		t.Error("expected HasCompose to be true")
	}
}

func TestDetectDocker_DDEV(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".ddev"), 0755)

	info := DetectDocker(dir)
	if !info.HasDDEV {
		t.Error("expected HasDDEV to be true")
	}
}

func TestDetectDocker_None(t *testing.T) {
	dir := t.TempDir()
	info := DetectDocker(dir)
	if info.HasCompose || info.HasDDEV {
		t.Error("expected no docker detection in empty dir")
	}
}
```

**Step 2: Run test → FAIL**

**Step 3: Implement**

```go
// internal/project/docker.go
package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DockerInfo struct {
	HasCompose  bool
	HasDDEV     bool
	ComposePath string
}

// DetectDocker checks if a project uses Docker or DDEV.
func DetectDocker(path string) DockerInfo {
	info := DockerInfo{}

	// Check for docker-compose files
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		p := filepath.Join(path, name)
		if _, err := os.Stat(p); err == nil {
			info.HasCompose = true
			info.ComposePath = p
			break
		}
	}

	// Check for DDEV
	ddevDir := filepath.Join(path, ".ddev")
	if stat, err := os.Stat(ddevDir); err == nil && stat.IsDir() {
		info.HasDDEV = true
	}

	return info
}

// StopDocker stops Docker containers for a project.
func StopDocker(path string) (string, error) {
	info := DetectDocker(path)
	var results []string

	if info.HasDDEV {
		cmd := exec.Command("ddev", "stop")
		cmd.Dir = path
		out, err := cmd.CombinedOutput()
		if err != nil {
			results = append(results, fmt.Sprintf("DDEV stop failed: %s", string(out)))
		} else {
			results = append(results, "DDEV stopped")
		}
	}

	if info.HasCompose {
		cmd := exec.Command("docker", "compose", "down")
		cmd.Dir = path
		out, err := cmd.CombinedOutput()
		if err != nil {
			results = append(results, fmt.Sprintf("docker compose down failed: %s", string(out)))
		} else {
			results = append(results, "Docker containers stopped")
		}
	}

	if len(results) == 0 {
		return "No Docker/DDEV configuration found", nil
	}
	return strings.Join(results, "\n"), nil
}
```

**Step 4: Run tests → PASS**

**Step 5: Create docker-stop command**

```go
// cmd/dockerstop.go
package cmd

import (
	"fmt"

	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/project"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var dockerStopCmd = &cobra.Command{
	Use:   "docker-stop",
	Short: "Stop Docker/DDEV containers of all forgotten projects",
	RunE:  runDockerStop,
}

func init() {
	rootCmd.AddCommand(dockerStopCmd)
}

func runDockerStop(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	fmt.Println("Scanning projects...")
	status, err := runner.Run()
	if err != nil {
		return err
	}

	allProjects := status.AllProjects()
	found := 0

	for _, p := range allProjects {
		info := project.DetectDocker(p.Path)
		if !info.HasCompose && !info.HasDDEV {
			continue
		}
		found++

		dockerType := "Docker Compose"
		if info.HasDDEV {
			dockerType = "DDEV"
		}
		fmt.Printf("\n  %s (%s)\n", p.ShortName, dockerType)
		fmt.Printf("  Path: %s\n", p.Path)

		result, err := project.StopDocker(p.Path)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  %s\n", result)
		}
	}

	if found == 0 {
		fmt.Println("No projects with Docker/DDEV configuration found.")
	} else {
		fmt.Printf("\nProcessed %d projects with Docker/DDEV.\n", found)
	}
	return nil
}
```

**Step 6: Commit**

```bash
git add internal/project/docker.go internal/project/docker_test.go cmd/dockerstop.go
git commit -m "feat: add docker-stop command to stop containers of forgotten projects"
```

---

## Phase 5: Config Command + Polish

### Task 12: Config command

**Files:**
- Create: `cmd/config.go`

**Step 1: Implement config command**

```go
// cmd/config.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or initialize configuration",
	RunE:  runConfig,
}

var configInitFlag bool

func init() {
	configCmd.Flags().BoolVar(&configInitFlag, "init", false, "Create default config file")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfgPath := config.ConfigPath()

	if configInitFlag {
		if _, err := os.Stat(cfgPath); err == nil {
			return fmt.Errorf("config already exists at %s", cfgPath)
		}
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
			return err
		}
		cfg := config.DefaultConfig()
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			return err
		}
		fmt.Printf("Config created at %s\n", cfgPath)
		return nil
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	fmt.Printf("Config file: %s\n\n", cfgPath)
	data, _ := yaml.Marshal(cfg)
	fmt.Println(string(data))
	return nil
}
```

**Step 2: Commit**

```bash
git add cmd/config.go
git commit -m "feat: add config command to show/initialize configuration"
```

---

### Task 13: Add version flag and build metadata

**Files:**
- Modify: `cmd/root.go`

**Step 1: Add version to root command**

Add to `cmd/root.go`:

```go
var version = "dev"

func init() {
	rootCmd.Version = version
}
```

**Step 2: Create Makefile**

```makefile
# Makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build install test clean

build:
	go build -ldflags "-X github.com/olivercodes/claude-meister/cmd.version=$(VERSION)" -o claude-meister .

install: build
	cp claude-meister $(GOPATH)/bin/ || cp claude-meister ~/go/bin/

test:
	go test ./... -v

clean:
	rm -f claude-meister
```

**Step 3: Verify**

Run: `make build && ./claude-meister --version`
Expected: `claude-meister version dev`

Run: `make test`
Expected: All tests pass

**Step 4: Commit**

```bash
git add cmd/root.go Makefile
git commit -m "feat: add version flag and Makefile"
```

---

### Task 14: Final polish - README and .gitignore

**Files:**
- Create: `.gitignore`

**Step 1: Create .gitignore**

```
claude-meister
*.exe
dist/
```

**Step 2: Verify full build and test**

Run:
```bash
make clean && make build && make test
./claude-meister --help
./claude-meister scan
```

Expected: Everything works, TUI shows projects, all tests pass.

**Step 3: Final commit**

```bash
git add .gitignore
git commit -m "chore: add .gitignore"
```

---

## Summary of all tasks

| Task | Description | Phase |
|------|-------------|-------|
| 1 | Go module + cobra root command | 1 |
| 2 | Squirrel types + JSON parser + tests | 1 |
| 3 | Squirrel CLI runner | 1 |
| 4 | Config package with YAML + tests | 1 |
| 5 | Scan command with TUI table view | 1 |
| 6 | Git status helper + tests | 2 |
| 7 | Clean command (CLI mode) | 2 |
| 8 | TUI detail view with actions | 2 |
| 9 | Archive functionality + command + tests | 3 |
| 10 | Delete command with double confirm | 3 |
| 11 | Docker detection + stop command + tests | 4 |
| 12 | Config command (show/init) | 5 |
| 13 | Version flag + Makefile | 5 |
| 14 | .gitignore + final verification | 5 |
