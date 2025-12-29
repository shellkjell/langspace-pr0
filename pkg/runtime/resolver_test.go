package runtime

import (
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/workspace"
)

func TestResolver_MethodCalls(t *testing.T) {
	ws := workspace.New()
	ctx := &ExecutionContext{
		Workspace: ws,
		Variables: map[string]interface{}{
			"text": "Hello, World!",
			"list": []interface{}{"a", "b", "c"},
			"data": map[string]interface{}{
				"key": "value",
			},
		},
		StepOutputs: map[string]interface{}{
			"analyze": "Analysis result",
		},
	}
	resolver := NewResolver(ctx)

	tests := []struct {
		name     string
		value    ast.Value
		expected interface{}
		wantErr  bool
	}{
		{
			name: "string_upper",
			value: ast.MethodCallValue{
				Object: ast.VariableValue{Name: "text"},
				Method: "upper",
			},
			expected: "HELLO, WORLD!",
		},
		{
			name: "string_lower",
			value: ast.MethodCallValue{
				Object: ast.VariableValue{Name: "text"},
				Method: "lower",
			},
			expected: "hello, world!",
		},
		{
			name: "string_split",
			value: ast.MethodCallValue{
				Object:    ast.VariableValue{Name: "text"},
				Method:    "split",
				Arguments: []ast.Value{ast.StringValue{Value: ", "}},
			},
			expected: []string{"Hello", "World!"},
		},
		{
			name: "array_join",
			value: ast.MethodCallValue{
				Object:    ast.VariableValue{Name: "list"},
				Method:    "join",
				Arguments: []ast.Value{ast.StringValue{Value: "-"}},
			},
			expected: "a-b-c",
		},
		{
			name: "array_len",
			value: ast.MethodCallValue{
				Object: ast.VariableValue{Name: "list"},
				Method: "len",
			},
			expected: 3.0,
		},
		{
			name: "chained_calls",
			value: ast.MethodCallValue{
				Object: ast.MethodCallValue{
					Object: ast.VariableValue{Name: "text"},
					Method: "upper",
				},
				Method:    "split",
				Arguments: []ast.Value{ast.StringValue{Value: ", "}},
			},
			expected: []string{"HELLO", "WORLD!"},
		},
		{
			name: "method_then_property",
			value: ast.MethodCallValue{
				Object: ast.MethodCallValue{
					Object: ast.VariableValue{Name: "text"},
					Method: "upper",
				},
				Method:    "len",
				Arguments: []ast.Value{},
			},
			expected: 13.0,
		},
		{
			name: "step_output_access",
			value: ast.MethodCallValue{
				Object: ast.FunctionCallValue{
					Function:  "step",
					Arguments: []ast.Value{ast.StringValue{Value: "analyze"}},
				},
				Method:    "output",
				Arguments: []ast.Value{},
			},
			expected: "Analysis result",
		},
		{
			name: "string_starts_with",
			value: ast.MethodCallValue{
				Object:    ast.VariableValue{Name: "text"},
				Method:    "starts_with",
				Arguments: []ast.Value{ast.StringValue{Value: "Hello"}},
			},
			expected: true,
		},
		{
			name: "array_first",
			value: ast.MethodCallValue{
				Object: ast.VariableValue{Name: "list"},
				Method: "first",
			},
			expected: "a",
		},
		{
			name: "object_has",
			value: ast.MethodCallValue{
				Object:    ast.VariableValue{Name: "data"},
				Method:    "has",
				Arguments: []ast.Value{ast.StringValue{Value: "key"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.Resolve(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// For slices, we need to compare elements
				if gotSlice, ok := got.([]string); ok {
					expectedSlice := tt.expected.([]string)
					if len(gotSlice) != len(expectedSlice) {
						t.Errorf("Resolve() = %v, want %v", got, tt.expected)
						return
					}
					for i := range gotSlice {
						if gotSlice[i] != expectedSlice[i] {
							t.Errorf("Resolve() = %v, want %v", got, tt.expected)
							return
						}
					}
				} else if got != tt.expected {
					t.Errorf("Resolve() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
