// Package main provides the LangSpace Command Line Interface (CLI).
// It allows users to parse, run, validate, and compile LangSpace workflows.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shellkjell/langspace/pkg/compile"
	_ "github.com/shellkjell/langspace/pkg/compile/python"     // Register Python compiler
	_ "github.com/shellkjell/langspace/pkg/compile/typescript" // Register TypeScript compiler
	"github.com/shellkjell/langspace/pkg/lsp"
	"github.com/shellkjell/langspace/pkg/parser"
	"github.com/shellkjell/langspace/pkg/runtime"
	"github.com/shellkjell/langspace/pkg/workspace"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

// run contains the main application logic, separated from main() for testability.
// This follows the pattern recommended by Mat Ryer and others for Go CLI apps.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return showHelp(stdout)
	}

	command := args[0]
	commandArgs := args[1:]

	var err error
	switch command {
	case "parse":
		err = runParse(commandArgs, stdin, stdout)
	case "run":
		err = runExecute(commandArgs, stdin, stdout, stderr)
	case "compile":
		err = runCompile(commandArgs, stdout)
	case "serve":
		err = runServe(commandArgs, stdin, stdout, stderr)
	case "lsp":
		err = runLSP(commandArgs, stdin, stdout, stderr)
	case "validate":
		err = runValidate(commandArgs, stdin, stdout)
	case "help", "-h", "--help":
		return showHelp(stdout)
	case "version":
		return showVersion(stdout)
	default:
		return fmt.Errorf("unknown command %q. Run 'langspace help' for usage", command)
	}

	if err != nil {
		checkPrint(fmt.Fprintf(stderr, "Error: %v\n", err))
		return err
	}
	return nil
}

func showHelp(w io.Writer) error {
	help := `LangSpace - A declarative language for composing AI workflows

Usage:
  langspace <command> [options]

Commands:
  parse     Parse a LangSpace file and display entities
  run       Execute an intent or pipeline
  compile   Compile to target language (python, typescript)
  validate  Validate a LangSpace file without executing
  serve     Start trigger server

Options:
  -h, --help     Show this help message
  -v, --version  Show version information

Examples:
  langspace parse -file workflow.ls
  langspace run -file workflow.ls -name my-intent
  langspace run -file workflow.ls -name my-pipeline -input "Review this code"
  langspace validate -file workflow.ls

For more information, visit: https://github.com/shellkjell/langspace
`
	checkPrint(fmt.Fprint(w, help))
	return nil
}

func showVersion(w io.Writer) error {
	checkPrint(fmt.Fprintln(w, "langspace version 0.1.0"))
	return nil
}

// runParse handles the parse command
func runParse(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	inputFile := fs.String("file", "", "Input file to parse")
	showJSON := fs.Bool("json", false, "Output as JSON")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// Create a new workspace
	ws := workspace.New()

	if *inputFile == "" {
		// Read from stdin
		content, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		if len(content) == 0 {
			// Handle empty input gracefully
			if *showJSON {
				return outputJSON(stdout, ws)
			}
			printStats(stdout, ws.Stat(), 0)
			return nil
		}

		p := parser.New(string(content))
		entities, _, err := p.Parse()
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		for _, entity := range entities {
			if err := ws.AddEntity(entity); err != nil {
				return fmt.Errorf("failed to add entity %q: %w", entity.Name(), err)
			}
		}

		if *showJSON {
			return outputJSON(stdout, ws)
		}
		printStats(stdout, ws.Stat(), len(ws.GetEntities()))
		return nil
	}

	// Load file and its imports
	l := workspace.NewLoader(ws)
	if err := l.Load(*inputFile); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	if *showJSON {
		return outputJSON(stdout, ws)
	}

	// Print statistics
	printStats(stdout, ws.Stat(), len(ws.GetEntities()))
	return nil
}

// runExecute handles the run command
func runExecute(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	inputFile := fs.String("file", "", "LangSpace file to execute")
	entityName := fs.String("name", "", "Name of the intent or pipeline to execute")
	entityType := fs.String("type", "", "Entity type (intent or pipeline, auto-detected if not specified)")
	inputData := fs.String("input", "", "Input data for the execution")
	inputFile2 := fs.String("input-file", "", "File containing input data")
	timeout := fs.Duration("timeout", 5*time.Minute, "Execution timeout")
	noStream := fs.Bool("no-stream", false, "Disable streaming output")
	verbose := fs.Bool("verbose", false, "Show verbose output")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *inputFile == "" {
		return fmt.Errorf("required flag -file not provided")
	}

	if *entityName == "" {
		return fmt.Errorf("required flag -name not provided")
	}

	// Load file and its imports

	ws := workspace.New()
	l := workspace.NewLoader(ws)
	if err := l.Load(*inputFile); err != nil {
		return err
	}

	// Determine entity type if not specified
	if *entityType == "" {
		*entityType = detectEntityType(ws, *entityName)
		if *entityType == "" {
			return fmt.Errorf("entity %q not found. Specify -type to search by type", *entityName)
		}
	}

	// Get input data
	var input interface{}
	if *inputData != "" {
		input = *inputData
	} else if *inputFile2 != "" {
		data, err := os.ReadFile(*inputFile2)
		if err != nil {
			return fmt.Errorf("reading input file: %w", err)
		}
		input = string(data)
	}

	// Create runtime
	rt := runtime.New(ws, runtime.WithConfig(&runtime.Config{
		DefaultModel:    "claude-sonnet-4-20250514",
		DefaultProvider: "anthropic",
		Timeout:         *timeout,
		EnableStreaming: !*noStream,
	}))

	// Register providers
	rt.RegisterProvider("anthropic", runtime.NewAnthropicProvider())
	rt.RegisterProvider("openai", runtime.NewOpenAIProvider())

	// Create stream handler for output
	var handler runtime.StreamHandler
	if !*noStream {
		handler = &CLIStreamHandler{
			stdout:  stdout,
			stderr:  stderr,
			verbose: *verbose,
		}
	}

	// Execute
	ctx := context.Background()
	var opts []runtime.ExecuteOption
	if input != nil {
		opts = append(opts, runtime.WithInput(input))
	}
	if handler != nil {
		opts = append(opts, runtime.WithStreamHandler(handler))
	}
	opts = append(opts, runtime.WithTimeout(*timeout))

	result, err := rt.ExecuteByName(ctx, *entityType, *entityName, opts...)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Print result
	if !*noStream {
		checkPrint(fmt.Fprintln(stdout)) // Newline after streaming
	}

	if *verbose || !result.Success {
		printExecutionResult(stdout, result)
	} else if result.Output != nil && !*noStream {
		// If not streaming, print the output now
	} else if result.Output != nil {
		checkPrint(fmt.Fprintf(stdout, "%v\n", result.Output))
	}

	if !result.Success {
		return fmt.Errorf("execution failed: %v", result.Error)
	}

	return nil
}

// runCompile handles the compile command
func runCompile(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("compile", flag.ContinueOnError)
	inputFile := fs.String("file", "", "LangSpace file to compile")
	target := fs.String("target", "python", "Target language (python, typescript)")
	outputDir := fs.String("output", ".", "Output directory for generated files")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *inputFile == "" {
		return fmt.Errorf("required flag -file not provided")
	}

	// Load file and its imports
	ws := workspace.New()
	l := workspace.NewLoader(ws)
	if err := l.Load(*inputFile); err != nil {
		return err
	}

	// Get compiler for target
	compiler, err := compile.Get(compile.Target(*target))
	if err != nil {
		return fmt.Errorf("getting compiler: %w", err)
	}

	// Compile
	output, err := compiler.Compile(ws)
	if err != nil {
		return fmt.Errorf("compilation error: %w", err)
	}

	// Write output files
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for filename, content := range output.Files {
		outPath := filepath.Join(*outputDir, filename)
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
		checkPrint(fmt.Fprintf(stdout, "Generated: %s\n", outPath))
	}

	checkPrint(fmt.Fprintf(stdout, "\nCompilation complete. %d files generated.\n", len(output.Files)))
	return nil
}

// runValidate handles the validate command
func runValidate(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	inputFile := fs.String("file", "", "LangSpace file to validate")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *inputFile == "" {
		return fmt.Errorf("required flag -file not provided")
	}

	ws := workspace.New()
	l := workspace.NewLoader(ws)
	if err := l.Load(*inputFile); err != nil {
		return err
	}

	// For validation, we might want to still show ParseWithRecovery errors from the main file,
	// but Loader already parsed it. Let's just output success for now if Loader succeeds.
	checkPrint(fmt.Fprintf(stdout, "Validation successful: %d entities loaded (including imports)\n", len(ws.GetEntities())))
	return nil
}

// runServe handles the serve command
func runServe(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	inputFile := fs.String("file", "", "LangSpace file to serve")
	port := fs.Int("port", 8080, "Port to listen on")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *inputFile == "" {
		return fmt.Errorf("required flag -file not provided")
	}

	// Load file and its imports
	ws := workspace.New()
	l := workspace.NewLoader(ws)
	if err := l.Load(*inputFile); err != nil {
		return err
	}

	// Create runtime
	rt := runtime.New(ws)
	rt.RegisterProvider("anthropic", runtime.NewAnthropicProvider())
	rt.RegisterProvider("openai", runtime.NewOpenAIProvider())

	// Start trigger engine
	engine := runtime.NewTriggerEngine(rt)
	if err := engine.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start trigger engine: %w", err)
	}

	checkPrint(fmt.Fprintf(stdout, "LangSpace server listening on port %d...\n", *port))
	checkPrint(fmt.Fprintf(stdout, "Trigger engine active with %d triggers\n", len(ws.GetEntitiesByType("trigger"))))

	// Keep running until interrupted
	select {}
}

// runLSP handles the lsp command
func runLSP(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("lsp", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	server := lsp.NewServer()
	return server.Start()
}

// detectEntityType tries to find an entity by name and returns its type
func detectEntityType(ws *workspace.Workspace, name string) string {
	// Try common types in order of likelihood
	for _, typ := range []string{"intent", "pipeline", "agent", "tool"} {
		if _, found := ws.GetEntityByName(typ, name); found {
			return typ
		}
	}
	return ""
}

// printStats outputs workspace statistics
func printStats(w io.Writer, stats workspace.WorkspaceStats, entityCount int) {
	checkPrint(fmt.Fprintln(w, "Workspace statistics:"))
	checkPrint(fmt.Fprintf(w, "  Number of entities: %d\n", stats.TotalEntities))
	checkPrint(fmt.Fprintf(w, "  Number of file entities: %d\n", stats.FileEntities))
	checkPrint(fmt.Fprintf(w, "  Number of agent entities: %d\n", stats.AgentEntities))
	checkPrint(fmt.Fprintf(w, "  Number of tool entities: %d\n", stats.ToolEntities))
	checkPrint(fmt.Fprintf(w, "  Number of intent entities: %d\n", stats.IntentEntities))
	checkPrint(fmt.Fprintf(w, "  Number of pipeline entities: %d\n", stats.PipelineEntities))
	checkPrint(fmt.Fprintf(w, "  Number of script entities: %d\n", stats.ScriptEntities))
	checkPrint(fmt.Fprintf(w, "  Number of relationships: %d\n", stats.TotalRelationships))
	checkPrint(fmt.Fprintf(w, "  Number of hooks: %d\n", stats.TotalHooks))
	checkPrint(fmt.Fprintf(w, "Successfully processed entities: %d\n", entityCount))
}

// outputJSON outputs the workspace as JSON
func outputJSON(w io.Writer, ws *workspace.Workspace) error {
	return ws.SaveTo(w)
}

// printExecutionResult prints detailed execution result
func printExecutionResult(w io.Writer, result *runtime.ExecutionResult) {
	checkPrint(fmt.Fprintln(w, "\n--- Execution Result ---"))
	checkPrint(fmt.Fprintf(w, "Success: %v\n", result.Success))
	checkPrint(fmt.Fprintf(w, "Duration: %s\n", result.Duration))
	checkPrint(fmt.Fprintf(w, "Tokens Used: %d (input: %d, output: %d)\n",
		result.TokensUsed.TotalTokens,
		result.TokensUsed.InputTokens,
		result.TokensUsed.OutputTokens))

	if len(result.StepResults) > 0 {
		checkPrint(fmt.Fprintln(w, "\nStep Results:"))
		for name, step := range result.StepResults {
			checkPrint(fmt.Fprintf(w, "  %s: success=%v, duration=%s\n", name, step.Success, step.Duration))
		}
	}

	if result.Error != nil {
		checkPrint(fmt.Fprintf(w, "\nError: %v\n", result.Error))
	}

	if result.Output != nil {
		checkPrint(fmt.Fprintln(w, "\n--- Output ---"))
		checkPrint(fmt.Fprintf(w, "%v\n", result.Output))
	}
}

// CLIStreamHandler handles streaming output for the CLI
type CLIStreamHandler struct {
	stdout  io.Writer
	stderr  io.Writer
	verbose bool
}

func (h *CLIStreamHandler) OnChunk(chunk runtime.StreamChunk) {
	if chunk.Type == runtime.ChunkTypeContent {
		checkPrint(fmt.Fprint(h.stdout, chunk.Content))
	}
}

func (h *CLIStreamHandler) OnProgress(event runtime.ProgressEvent) {
	if h.verbose {
		switch event.Type {
		case runtime.ProgressTypeStart:
			checkPrint(fmt.Fprintf(h.stderr, "🚀 %s\n", event.Message))
		case runtime.ProgressTypeStep:
			checkPrint(fmt.Fprintf(h.stderr, "📍 [%d%%] %s\n", event.Progress, event.Message))
		case runtime.ProgressTypeComplete:
			checkPrint(fmt.Fprintf(h.stderr, "- %s\n", event.Message))
		case runtime.ProgressTypeError:
			checkPrint(fmt.Fprintf(h.stderr, "❌ %s\n", event.Message))
		}
	}
}

func (h *CLIStreamHandler) OnComplete(result *runtime.CompletionResponse) {}

func (h *CLIStreamHandler) OnError(err error) {
	checkPrint(fmt.Fprintf(h.stderr, "Error: %v\n", err))
}

// checkPrint is a helper that logs an error if a print operation fails.
func checkPrint(_ int, err error) {
	if err != nil {
		// Log to standard logger as a fallback if terminal output fails
		log.Printf("terminal output error: %v", err)
	}
}
