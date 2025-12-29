package runtime

import (
	"context"
	"os"
	"testing"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/workspace"
)

func TestExecuteScript_Python(t *testing.T) {
	// Skip if python3 is not available
	if _, err := os.Stat("/usr/bin/python3"); os.IsNotExist(err) {
		t.Skip("python3 not found")
	}

	ws := workspace.New()
	rt := New(ws)

	script := ast.NewScriptEntity("test-python")
	script.SetProperty("language", ast.StringValue{Value: "python"})
	script.SetProperty("code", ast.StringValue{Value: "print('hello from python')"})

	res, err := rt.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !res.Success {
		t.Errorf("Execution failed: %v", res.Error)
	}

	output, ok := res.Output.(string)
	if !ok {
		t.Fatalf("Expected string output, got %T", res.Output)
	}

	if output != "hello from python\n" {
		t.Errorf("Expected 'hello from python\\n', got %q", output)
	}
}

func TestExecuteScript_Shell(t *testing.T) {
	ws := workspace.New()
	rt := New(ws)

	script := ast.NewScriptEntity("test-shell")
	script.SetProperty("language", ast.StringValue{Value: "bash"})
	script.SetProperty("code", ast.StringValue{Value: "echo 'hello from shell'"})

	res, err := rt.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !res.Success {
		t.Errorf("Execution failed: %v", res.Error)
	}

	output, ok := res.Output.(string)
	if !ok {
		t.Fatalf("Expected string output, got %T", res.Output)
	}

	if output != "hello from shell\n" {
		t.Errorf("Expected 'hello from shell\\n', got %q", output)
	}
}
