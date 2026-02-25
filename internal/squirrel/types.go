package squirrel

import "time"

// Project represents a single Claude Code project as reported by squirrel.
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

// Status represents the full output of `squirrel status --json`.
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
