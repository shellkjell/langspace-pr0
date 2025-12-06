package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/shellkjell/langspace/pkg/parser"
	"github.com/shellkjell/langspace/pkg/workspace"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// run contains the main application logic, separated from main() for testability.
// This follows the pattern recommended by Mat Ryer and others for Go CLI apps.
func run(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("langspace", flag.ContinueOnError)
	inputFile := fs.String("file", "", "Input file to parse")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	input, err := readInput(*inputFile, stdin)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	// Create a new workspace
	ws := workspace.New()

	// Parse input
	p := parser.New(input)
	entities, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Add entities to workspace
	for _, entity := range entities {
		if err := ws.AddEntity(entity); err != nil {
			return fmt.Errorf("adding entity %q: %w", entity.Name(), err)
		}
	}

	// Print statistics
	printStats(stdout, ws.Stat(), len(entities))
	return nil
}

// readInput reads input from a file or stdin
func readInput(filePath string, stdin io.Reader) (string, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", filePath, err)
		}
		return string(data), nil
	}

	reader := bufio.NewReader(stdin)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	return string(data), nil
}

// printStats outputs workspace statistics
func printStats(w io.Writer, stats workspace.WorkspaceStats, entityCount int) {
	fmt.Fprintln(w, "Workspace statistics:")
	fmt.Fprintf(w, "  Number of entities: %d\n", stats.TotalEntities)
	fmt.Fprintf(w, "  Number of file entities: %d\n", stats.FileEntities)
	fmt.Fprintf(w, "  Number of agent entities: %d\n", stats.AgentEntities)
	fmt.Fprintf(w, "  Number of tool entities: %d\n", stats.ToolEntities)
	fmt.Fprintf(w, "  Number of intent entities: %d\n", stats.IntentEntities)
	fmt.Fprintf(w, "  Number of pipeline entities: %d\n", stats.PipelineEntities)
	fmt.Fprintf(w, "  Number of script entities: %d\n", stats.ScriptEntities)
	fmt.Fprintf(w, "  Number of relationships: %d\n", stats.TotalRelationships)
	fmt.Fprintf(w, "  Number of hooks: %d\n", stats.TotalHooks)
	fmt.Fprintf(w, "Successfully processed entities: %d\n", entityCount)
}
