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
