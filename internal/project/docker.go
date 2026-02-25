package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerInfo holds the Docker/DDEV detection results for a project.
type DockerInfo struct {
	HasCompose  bool
	HasDDEV     bool
	ComposePath string
}

// DetectDocker checks a project directory for Docker Compose and DDEV configurations.
func DetectDocker(path string) DockerInfo {
	info := DockerInfo{}

	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		p := filepath.Join(path, name)
		if _, err := os.Stat(p); err == nil {
			info.HasCompose = true
			info.ComposePath = p
			break
		}
	}

	ddevDir := filepath.Join(path, ".ddev")
	if stat, err := os.Stat(ddevDir); err == nil && stat.IsDir() {
		info.HasDDEV = true
	}

	return info
}

// StopDocker stops Docker Compose and/or DDEV containers for a project.
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
