package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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
		data, err := ioutil.ReadFile(*inputFile)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
		input = string(data)
	} else {
		data, err := ioutil.ReadAll(os.Stdin)
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

	fmt.Printf("Successfully processed %d entities\n", len(entities))
}
