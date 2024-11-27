package parser

import (
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []ast.Entity
		wantError bool
	}{
		{
			name:  "single_file_entity",
			input: `file "test.txt" path;`,
			want: []ast.Entity{
				&ast.FileEntity{
					Path:     "test.txt",
					Property: "path",
				},
			},
			wantError: false,
		},
		{
			name:  "multiple_entities",
			input: `file "test.txt" path; agent "gpt-4" model;`,
			want: []ast.Entity{
				&ast.FileEntity{
					Path:     "test.txt",
					Property: "path",
				},
				&ast.AgentEntity{
					Name:     "gpt-4",
					Property: "model",
				},
			},
			wantError: false,
		},
		{
			name:      "missing_string_property",
			input:     `file path;`,
			want:      nil,
			wantError: true,
		},
		{
			name:      "missing_semicolon",
			input:     `file "test.txt" path`,
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			got, err := p.Parse()

			if (err != nil) != tt.wantError {
				t.Errorf("Parser.Parse() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Errorf("Parser.Parse() got %v entities, want %v", len(got), len(tt.want))
					return
				}

				for i, entity := range got {
					if entity.Type() != tt.want[i].Type() {
						t.Errorf("Entity[%d].Type() = %v, want %v", i, entity.Type(), tt.want[i].Type())
					}
				}
			}
		})
	}
}

func TestParser_ParseFileEntity(t *testing.T) {
	input := `file "test.txt" path;`
	p := New(input)
	entities, err := p.Parse()

	if err != nil {
		t.Errorf("Parser.Parse() error = %v", err)
		return
	}

	if len(entities) != 1 {
		t.Errorf("Parser.Parse() got %v entities, want 1", len(entities))
		return
	}

	fileEntity, ok := entities[0].(*ast.FileEntity)
	if !ok {
		t.Errorf("Entity is not a FileEntity")
		return
	}

	if fileEntity.Path != "test.txt" {
		t.Errorf("FileEntity.Path = %v, want test.txt", fileEntity.Path)
	}

	if fileEntity.Property != "path" {
		t.Errorf("FileEntity.Property = %v, want path", fileEntity.Property)
	}
}

func TestParser_ParseAgentEntity(t *testing.T) {
	input := `agent "gpt-4" model;`
	p := New(input)
	entities, err := p.Parse()

	if err != nil {
		t.Errorf("Parser.Parse() error = %v", err)
		return
	}

	if len(entities) != 1 {
		t.Errorf("Parser.Parse() got %v entities, want 1", len(entities))
		return
	}

	agentEntity, ok := entities[0].(*ast.AgentEntity)
	if !ok {
		t.Errorf("Entity is not an AgentEntity")
		return
	}

	if agentEntity.Name != "gpt-4" {
		t.Errorf("AgentEntity.Name = %v, want gpt-4", agentEntity.Name)
	}

	if agentEntity.Property != "model" {
		t.Errorf("AgentEntity.Property = %v, want model", agentEntity.Property)
	}
}

func BenchmarkParser_Parse(b *testing.B) {
	input := `file "test.txt" path;
agent "gpt-4" model;`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(input)
		p.Parse()
	}
}
