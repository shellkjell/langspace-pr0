// Package python provides Python/LangGraph code generation for LangSpace.
package python

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/shellkjell/langspace/pkg/ast"
	"github.com/shellkjell/langspace/pkg/compile"
	"github.com/shellkjell/langspace/pkg/workspace"
)

func init() {
	compile.Register(&Generator{})
}

// Generator generates Python/LangGraph code from LangSpace definitions.
type Generator struct{}

// Target returns the compilation target.
func (g *Generator) Target() compile.Target {
	return compile.TargetPython
}

// Compile generates Python code for the given workspace.
func (g *Generator) Compile(ws *workspace.Workspace) (*compile.Output, error) {
	output := &compile.Output{
		Files: make(map[string]string),
	}

	// Collect all entities by type
	agents := ws.GetEntitiesByType("agent")
	pipelines := ws.GetEntitiesByType("pipeline")
	intents := ws.GetEntitiesByType("intent")
	configs := ws.GetEntitiesByType("config")

	// Generate main workflow file
	mainCode, err := g.generateMain(agents, pipelines, intents, configs)
	if err != nil {
		return nil, fmt.Errorf("generating main: %w", err)
	}
	output.Files["workflow.py"] = mainCode

	// Generate requirements
	output.Files["requirements.txt"] = g.generateRequirements()

	return output, nil
}

// generateMain creates the main Python workflow file.
func (g *Generator) generateMain(agents, pipelines, intents, configs []ast.Entity) (string, error) {
	var buf bytes.Buffer

	// Write imports
	buf.WriteString(pythonImports)

	// Write config if present
	if len(configs) > 0 {
		if err := g.writeConfig(&buf, configs[0]); err != nil {
			return "", err
		}
	}

	// Write agent functions
	for _, agent := range agents {
		if err := g.writeAgent(&buf, agent); err != nil {
			return "", err
		}
	}

	// Write pipelines as StateGraphs
	for _, pipeline := range pipelines {
		if err := g.writePipeline(&buf, pipeline); err != nil {
			return "", err
		}
	}

	// Write intents as entry points
	for _, intent := range intents {
		if err := g.writeIntent(&buf, intent); err != nil {
			return "", err
		}
	}

	// Write main block
	buf.WriteString(pythonMain)

	return buf.String(), nil
}

func (g *Generator) writeConfig(buf *bytes.Buffer, config ast.Entity) error {
	model := getStringProp(config, "default_model", "claude-sonnet-4-20250514")
	fmt.Fprintf(buf, "\n# Configuration\nDEFAULT_MODEL = %q\n", model)
	return nil
}

func (g *Generator) writeAgent(buf *bytes.Buffer, agent ast.Entity) error {
	name := agent.Name()
	safeName := toSnakeCase(name)
	model := getStringProp(agent, "model", "DEFAULT_MODEL")
	temperature := getNumberProp(agent, "temperature", 0.7)
	instruction := getStringProp(agent, "instruction", "You are a helpful assistant.")

	tmpl := template.Must(template.New("agent").Funcs(funcMap).Parse(agentTemplate))
	return tmpl.Execute(buf, map[string]interface{}{
		"Name":        name,
		"SafeName":    safeName,
		"Model":       model,
		"Temperature": temperature,
		"Instruction": instruction,
	})
}

func (g *Generator) writePipeline(buf *bytes.Buffer, pipeline ast.Entity) error {
	name := pipeline.Name()
	safeName := toSnakeCase(name)

	// Collect steps
	var steps []map[string]string
	for _, val := range pipeline.Properties() {
		if nested, ok := val.(ast.NestedEntityValue); ok {
			if nested.Entity.Type() == "step" {
				stepName := nested.Entity.Name()
				usesAgent := ""
				if useVal, exists := nested.Entity.GetProperty("use"); exists {
					if ref, ok := useVal.(ast.ReferenceValue); ok && ref.Type == "agent" {
						usesAgent = ref.Name
					}
				}
				steps = append(steps, map[string]string{
					"name":      stepName,
					"safeName":  toSnakeCase(stepName),
					"usesAgent": usesAgent,
				})
			}
		}
	}

	tmpl := template.Must(template.New("pipeline").Funcs(funcMap).Parse(pipelineTemplate))
	return tmpl.Execute(buf, map[string]interface{}{
		"Name":     name,
		"SafeName": safeName,
		"Steps":    steps,
	})
}

func (g *Generator) writeIntent(buf *bytes.Buffer, intent ast.Entity) error {
	name := intent.Name()
	safeName := toSnakeCase(name)

	usesAgent := ""
	usesPipeline := ""
	if useVal, exists := intent.GetProperty("use"); exists {
		if ref, ok := useVal.(ast.ReferenceValue); ok {
			switch ref.Type {
			case "agent":
				usesAgent = ref.Name
			case "pipeline":
				usesPipeline = ref.Name
			default:
				return fmt.Errorf("invalid use type: %s", ref.Type)
			}
		}
	}

	tmpl := template.Must(template.New("intent").Funcs(funcMap).Parse(intentTemplate))
	return tmpl.Execute(buf, map[string]interface{}{
		"Name":         name,
		"SafeName":     safeName,
		"UsesAgent":    usesAgent,
		"UsesPipeline": usesPipeline,
	})
}

func (g *Generator) generateRequirements() string {
	return `# LangSpace Generated Requirements
langgraph>=0.2.60
langchain>=0.3.13
langchain-anthropic>=0.3.1
langchain-openai>=0.2.14
python-dotenv>=1.0.1
langsmith>=0.1.147
`
}

// Helper functions

func toSnakeCase(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return strings.ToLower(s)
}

func toTitle(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

var funcMap = template.FuncMap{
	"title":     toTitle,
	"snakecase": toSnakeCase,
	"add": func(a, b int) int {
		return a + b
	},
}

func getStringProp(entity ast.Entity, key, defaultVal string) string {
	if val, exists := entity.GetProperty(key); exists {
		if sv, ok := val.(ast.StringValue); ok {
			return sv.Value
		}
	}
	return defaultVal
}

func getNumberProp(entity ast.Entity, key string, defaultVal float64) float64 {
	if val, exists := entity.GetProperty(key); exists {
		if nv, ok := val.(ast.NumberValue); ok {
			return nv.Value
		}
	}
	return defaultVal
}

// Templates

const pythonImports = `"""
LangSpace Generated Workflow
Generated by: langspace compile --target python
"""

from typing import TypedDict, Annotated, List, Optional
from langgraph.graph import StateGraph, END
from langchain_anthropic import ChatAnthropic
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
import os
from dotenv import load_dotenv

load_dotenv()

# Observability (LangSmith)
if os.getenv("LANGCHAIN_TRACING_V2") == "true":
    if not os.getenv("LANGCHAIN_API_KEY"):
        print("Warning: LANGCHAIN_TRACING_V2 is enabled but LANGCHAIN_API_KEY is not set.")
    else:
        print("Observability enabled (LangSmith)")

`

const agentTemplate = `
# Agent: {{.Name}}
def {{.SafeName}}_agent(state: dict) -> dict:
    """{{.Instruction}}"""
    llm = ChatAnthropic(
        model={{if eq .Model "DEFAULT_MODEL"}}DEFAULT_MODEL{{else}}"{{.Model}}"{{end}},
        temperature={{.Temperature}},
    )
    
    messages = [
        SystemMessage(content="""{{.Instruction}}"""),
        HumanMessage(content=str(state.get("input", ""))),
    ]
    
    response = llm.invoke(messages)
    return {"{{.SafeName}}_output": response.content}

`

const pipelineTemplate = `
# Pipeline: {{.Name}}
class {{.SafeName | title}}State(TypedDict):
    input: str
{{- range .Steps}}
    {{.safeName}}_output: Optional[str]
{{- end}}

def create_{{.SafeName}}_pipeline():
    workflow = StateGraph({{.SafeName | title}}State)
    
{{- range .Steps}}
{{- if .usesAgent}}
    workflow.add_node("{{.name}}", {{.usesAgent | snakecase}}_agent)
{{- end}}
{{- end}}
    
{{- if .Steps}}
    workflow.set_entry_point("{{(index .Steps 0).name}}")
{{- range $i, $step := .Steps}}
{{- if lt (add $i 1) (len $.Steps)}}
    workflow.add_edge("{{$step.name}}", "{{(index $.Steps (add $i 1)).name}}")
{{- else}}
    workflow.add_edge("{{$step.name}}", END)
{{- end}}
{{- end}}
{{- end}}
    
    return workflow.compile()

{{.SafeName}}_app = create_{{.SafeName}}_pipeline()

`

const intentTemplate = `
# Intent: {{.Name}}
def run_{{.SafeName}}(input_data: str) -> str:
    """Execute the {{.Name}} intent."""
{{- if .UsesAgent}}
    result = {{.UsesAgent | snakecase}}_agent({"input": input_data})
    return result.get("{{.UsesAgent | snakecase}}_output", "")
{{- else if .UsesPipeline}}
    result = {{.UsesPipeline | snakecase}}_app.invoke({"input": input_data})
    return str(result)
{{- else}}
    return input_data
{{- end}}

`

const pythonMain = `
if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1:
        input_text = " ".join(sys.argv[1:])
        print(f"Input: {input_text}")
        # Add your entry point call here
    else:
        print("Usage: python workflow.py <input>")
`
