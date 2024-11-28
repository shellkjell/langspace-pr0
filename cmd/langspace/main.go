package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shellkjell/langspace/pkg/parser"
	"github.com/shellkjell/langspace/pkg/workspace"
)

func main() {
	inputFile := flag.String("file", "", "Input file to parse")
	flag.Parse()

	var input string
	var err error

	if *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
		input = string(data)
	} else {
		reader := bufio.NewReader(os.Stdin)
		data, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		input = string(data)
	}

	// Create a new workspace
	ws := workspace.New()

	// Parse input
	p := parser.New(input)
	entities, err := p.Parse()
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Add entities to workspace
	for _, entity := range entities {
		if err := ws.AddEntity(entity); err != nil {
			log.Fatalf("Error adding entity: %v", err)
		}
	}

	data := ws.Stat()
	fmt.Println("Workspace statistics:")
	fmt.Println("  Number of entities: ", data.TotalEntities)
	fmt.Println("  Number of file entities: ", data.FileEntities)
	fmt.Println("  Number of agent entities: ", data.AgentEntities)

	fmt.Println("Successfully processed amount of entities: ", len(entities))
}
