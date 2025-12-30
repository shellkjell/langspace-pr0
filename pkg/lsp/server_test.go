package lsp

import (
	"encoding/json"
	"testing"
)

func TestServer_HandleDefinition(t *testing.T) {
	s := NewServer()
	uri := "file:///test.ls"
	content := `agent "researcher" { model: "gpt-4" }
pipeline "main" {
    step "search" { use: researcher }
}`
	s.files[uri] = content

	// Index
	_ = s.reindex()

	// Test Go to Definition for "researcher" at line 3, col 25
	params := map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
		"position":     map[string]int{"line": 2, "character": 25},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := s.handleDefinition(paramsJSON)
	if err != nil {
		t.Fatalf("handleDefinition failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected definition result, got nil")
	}

	resMap := result.(map[string]interface{})
	if resMap["uri"] != uri {
		t.Errorf("expected uri %s, got %v", uri, resMap["uri"])
	}

	rangeMap := resMap["range"].(map[string]interface{})
	start := rangeMap["start"].(map[string]int)
	if start["line"] != 0 {
		t.Errorf("expected line 0, got %d", start["line"])
	}
}

func TestIsIdentChar(t *testing.T) {
	tests := []struct {
		char byte
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'_', true},
		{'-', true},
		{'.', true},
		{' ', false},
		{'(', false},
		{'"', false},
	}

	for _, tt := range tests {
		if got := isIdentChar(tt.char); got != tt.want {
			t.Errorf("isIdentChar(%c) = %v, want %v", tt.char, got, tt.want)
		}
	}
}
