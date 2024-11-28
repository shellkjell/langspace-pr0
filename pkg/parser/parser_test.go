package parser

import (
	"strings"
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
)

func TestParser_Parse_Basic(t *testing.T) {
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
			name:      "empty_input",
			input:     "",
			want:      nil,
			wantError: false,
		},
		{
			name:      "whitespace_only",
			input:     "   \n\t  \n  ",
			want:      nil,
			wantError: false,
		},
	}

	runParserTests(t, tests)
}

func TestParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []ast.Entity
		wantError bool
	}{
		{
			name:      "invalid_entity_type",
			input:     `unknown "test" prop;`,
			want:      nil,
			wantError: true,
		},
		{
			name:      "missing_property",
			input:     `file "test.txt";`,
			want:      nil,
			wantError: true,
		},
		{
			name:      "multiple_properties",
			input:     `file "test.txt" path extra;`,
			want:      nil,
			wantError: true,
		},
		{
			name:      "invalid_file_path",
			input:     `file path path;`,
			want:      nil,
			wantError: true,
		},
		{
			name:      "missing_agent_name",
			input:     `agent model;`,
			want:      nil,
			wantError: true,
		},
	}

	runParserTests(t, tests)
}

func TestParser_Parse_MultilineContent(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []ast.Entity
		wantError bool
	}{
		{
			name:  "multiline_file_content",
			input: `file "script.sh" contents ` + "```" + `
#!/bin/bash
echo 'Hello World'
exit 0
` + "```" + `;`,
			want: []ast.Entity{
				&ast.FileEntity{
					Property: "contents",
					Path:     "#!/bin/bash\necho 'Hello World'\nexit 0\n",
				},
			},
			wantError: false,
		},
		{
			name:  "multiline_with_backticks",
			input: `file "code.go" contents ` + "```" + `
func main() {
    fmt.Println("Hello world")
}
` + "```" + `;`,
			want: []ast.Entity{
				&ast.FileEntity{
					Property: "contents",
					Path:     "func main() {\n    fmt.Println(\"Hello world\")\n}\n",
				},
			},
			wantError: false,
		},
	}

	runParserTests(t, tests)
}

func TestParser_Parse_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []ast.Entity
		wantError bool
	}{
		{
			name: "multiple_files_with_content",
			input: strings.TrimSpace(`file "test1.sh" contents ` + "```" + `
#!/bin/bash
echo "test1"
` + "```" + `;
file "test2.sh" contents ` + "```" + `
#!/bin/bash
echo "test2"
` + "```" + `;`),
			want: []ast.Entity{
				&ast.FileEntity{
					Property: "contents",
					Path:     "#!/bin/bash\necho \"test1\"\n",
				},
				&ast.FileEntity{
					Property: "contents",
					Path:     "#!/bin/bash\necho \"test2\"\n",
				},
			},
			wantError: false,
		},
		{
			name: "mixed_entities_with_whitespace",
			input: strings.TrimSpace(`
file "main.go" path;
agent "gpt-4" model;
file "test.txt" contents;`),
			want: []ast.Entity{
				&ast.FileEntity{
					Path:     "main.go",
					Property: "path",
				},
				&ast.AgentEntity{
					Name:     "gpt-4",
					Property: "model",
				},
				&ast.FileEntity{
					Path:     "test.txt",
					Property: "contents",
				},
			},
			wantError: false,
		},
	}

	runParserTests(t, tests)
}

func runParserTests(t *testing.T, tests []struct {
	name      string
	input     string
	want      []ast.Entity
	wantError bool
}) {
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

					switch e := entity.(type) {
					case *ast.FileEntity:
						want := tt.want[i].(*ast.FileEntity)
						if e.Property != want.Property || e.Path != want.Path {
							t.Errorf("FileEntity[%d] = {Property: %q, Path: %q}, want {Property: %q, Path: %q}",
								i, e.Property, e.Path, want.Property, want.Path)
						}
					case *ast.AgentEntity:
						want := tt.want[i].(*ast.AgentEntity)
						if e.Name != want.Name || e.Property != want.Property {
							t.Errorf("AgentEntity[%d] = {Name: %q, Property: %q}, want {Name: %q, Property: %q}",
								i, e.Name, e.Property, want.Name, want.Property)
						}
					}
				}
			}
		})
	}
}

func BenchmarkParser_Parse(b *testing.B) {
	input := `file "test.txt" path; agent "gpt-4" model;`
	p := New(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Parse()
	}
}
