package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_WithStdin(t *testing.T) {
	input := `agent "test-agent" {
	model: "gpt-4"
	instruction: "Test instruction"
}`

	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}

	err := run([]string{}, stdin, stdout)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Number of entities: 1") {
		t.Errorf("expected output to contain entity count, got: %s", output)
	}
	if !strings.Contains(output, "Number of agent entities: 1") {
		t.Errorf("expected output to contain agent count, got: %s", output)
	}
}

func TestRun_WithFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ls")

	content := `file "config.json" {
	path: "./config.json"
}

agent "analyzer" {
	model: "claude-sonnet-4-20250514"
	instruction: "Analyze code"
}`

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	err := run([]string{"-file", tmpFile}, stdin, stdout)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Number of entities: 2") {
		t.Errorf("expected 2 entities, got: %s", output)
	}
	if !strings.Contains(output, "Number of file entities: 1") {
		t.Errorf("expected 1 file entity, got: %s", output)
	}
	if !strings.Contains(output, "Number of agent entities: 1") {
		t.Errorf("expected 1 agent entity, got: %s", output)
	}
}

func TestRun_ParseError(t *testing.T) {
	input := `agent "broken" {
	invalid syntax here
`

	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}

	err := run([]string{}, stdin, stdout)
	if err == nil {
		t.Error("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("expected parse error message, got: %v", err)
	}
}

func TestRun_FileNotFound(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	err := run([]string{"-file", "/nonexistent/file.ls"}, stdin, stdout)
	if err == nil {
		t.Error("expected file not found error, got nil")
	}
	if !strings.Contains(err.Error(), "reading input") {
		t.Errorf("expected reading input error, got: %v", err)
	}
}

func TestRun_InvalidFlags(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	err := run([]string{"-invalid-flag"}, stdin, stdout)
	if err == nil {
		t.Error("expected flag error, got nil")
	}
	if !strings.Contains(err.Error(), "parsing flags") {
		t.Errorf("expected parsing flags error, got: %v", err)
	}
}

func TestRun_EmptyInput(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	err := run([]string{}, stdin, stdout)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Number of entities: 0") {
		t.Errorf("expected 0 entities for empty input, got: %s", output)
	}
}

func TestRun_MultipleEntityTypes(t *testing.T) {
	input := `
file "readme.md" {
	contents: "Hello"
}

file "config.json" {
	path: "./config.json"
}

agent "reviewer" {
	model: "gpt-4"
	instruction: "Review code"
}

tool "linter" {
	command: "golangci-lint run"
}

intent "review-code" {
	use: agent("reviewer")
}

pipeline "ci-pipeline" {
	output: "results"
}

script "update-db" {
	language: "python"
	code: "print('hello')"
}
`

	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}

	err := run([]string{}, stdin, stdout)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Number of entities: 7") {
		t.Errorf("expected 7 entities, got: %s", output)
	}
	if !strings.Contains(output, "Number of file entities: 2") {
		t.Errorf("expected 2 file entities, got: %s", output)
	}
	if !strings.Contains(output, "Number of agent entities: 1") {
		t.Errorf("expected 1 agent entity, got: %s", output)
	}
	if !strings.Contains(output, "Number of tool entities: 1") {
		t.Errorf("expected 1 tool entity, got: %s", output)
	}
	if !strings.Contains(output, "Number of intent entities: 1") {
		t.Errorf("expected 1 intent entity, got: %s", output)
	}
	if !strings.Contains(output, "Number of pipeline entities: 1") {
		t.Errorf("expected 1 pipeline entity, got: %s", output)
	}
	if !strings.Contains(output, "Number of script entities: 1") {
		t.Errorf("expected 1 script entity, got: %s", output)
	}
}

func TestRun_WithComments(t *testing.T) {
	input := `
# This is a comment
agent "test" {
	model: "gpt-4"
	instruction: "Test"
}
`

	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}

	err := run([]string{}, stdin, stdout)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Number of entities: 1") {
		t.Errorf("expected 1 entity, got: %s", output)
	}
}
