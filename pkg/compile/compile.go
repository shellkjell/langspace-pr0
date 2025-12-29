// Package compile provides code generation for LangSpace workflows.
// It supports compiling LangSpace definitions to various target languages.
package compile

import (
	"fmt"

	"github.com/shellkjell/langspace/pkg/workspace"
)

// Target represents a compilation target.
type Target string

const (
	TargetPython     Target = "python"
	TargetTypeScript Target = "typescript"
)

// Output represents the result of compilation.
type Output struct {
	// Files maps relative paths to file contents.
	Files map[string]string
}

// Compiler defines the interface for code generators.
type Compiler interface {
	// Compile generates code for the given workspace.
	Compile(ws *workspace.Workspace) (*Output, error)

	// Target returns the compilation target.
	Target() Target
}

// CompileOptions holds options for compilation.
type CompileOptions struct {
	// OutputDir is the directory to write generated files.
	OutputDir string

	// EntryPoint is the name of the main entity to compile.
	// If empty, all entities are compiled.
	EntryPoint string

	// IncludeComments adds documentation comments to generated code.
	IncludeComments bool
}

// Registry holds registered compilers.
var registry = make(map[Target]Compiler)

// Register adds a compiler to the registry.
func Register(compiler Compiler) {
	registry[compiler.Target()] = compiler
}

// Get returns a compiler for the given target.
func Get(target Target) (Compiler, error) {
	c, ok := registry[target]
	if !ok {
		return nil, fmt.Errorf("unknown compilation target: %s", target)
	}
	return c, nil
}

// SupportedTargets returns a list of supported compilation targets.
func SupportedTargets() []Target {
	targets := make([]Target, 0, len(registry))
	for t := range registry {
		targets = append(targets, t)
	}
	return targets
}
