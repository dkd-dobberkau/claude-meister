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
