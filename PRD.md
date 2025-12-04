# LangSpace Product Requirements Document

## Vision

**LangSpace is a declarative language for composing AI workflows.**

It sits in the gap between "write Python code" and "use a no-code builder" — providing a readable, versionable, shareable format for defining agents, their tools, and how they work together.

Think of it as:
- **Terraform for AI agents** — declare your agents and their relationships, then apply
- **Docker Compose for LLM workflows** — define multi-agent systems in a single file
- **SQL for agent orchestration** — a domain-specific query language that's simpler than general-purpose code

## Why LangSpace?

### The Problem

The AI agent ecosystem is fragmented:

| Approach | Problem |
|----------|---------|
| **LangGraph, CrewAI, AutoGen** | Powerful but require significant Python/TypeScript expertise |
| **No-code builders** | Easy but not versionable, shareable, or composable |
| **Raw API calls** | Maximum control but massive boilerplate |
| **Prompt templates** | Don't capture the full workflow, just the prompt |

Developers need to:
1. Share agent configurations with teammates
2. Version control their AI workflows
3. Compose simple agents into complex pipelines
4. Test and iterate on agent behavior without rewriting code
5. Generate code/configs for their preferred runtime

### The Solution

LangSpace provides a **human-readable DSL** that:
- Captures the full specification of an AI workflow
- Compiles to multiple targets (Python/LangGraph, TypeScript, API calls, etc.)
- Integrates with any LLM provider via a unified interface
- Supports direct execution with built-in LLM integration

## Core Concepts

### 1. Files — Data at Rest
Files represent static data: configuration, prompts, context documents, or generated outputs.

```langspace
file "system-prompt.md" {
  contents: ```
    You are a helpful coding assistant specializing in Go.
    Always explain your reasoning before providing code.
  ```
}

file "config.json" {
  path: "./config/app.json"  # Reference to external file
}
```

### 2. Agents — Intelligent Actors
Agents are LLM-powered entities with specific roles, instructions, and capabilities.

```langspace
agent "code-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You are a senior code reviewer. Analyze code for:
    - Security vulnerabilities
    - Performance issues
    - Best practice violations
    - Clarity and maintainability

    Provide specific, actionable feedback with line references.
  ```

  tools: [read_file, search_codebase, run_tests]
}
```

### 3. Tools — Agent Capabilities
Tools extend what agents can do, connecting them to the outside world.

```langspace
tool "search_codebase" {
  description: "Search the codebase for patterns or symbols"

  parameters: {
    query: string required "The search pattern"
    file_pattern: string optional "Glob pattern for files to search"
  }

  # Tools can be implemented inline or reference external handlers
  handler: mcp("filesystem-server")
}
```

### 4. Intentions — Expressing Desired Outcomes
Intentions are the heart of LangSpace. They express *what you want to happen* rather than *how to do it*.

```langspace
# Simple invocation
intent "review my code" {
  use: agent("code-reviewer")
  input: file("src/**/*.go")
  output: file("reviews/feedback.md")
}

# With explicit context
intent "improve documentation" {
  use: agent("doc-writer")

  input: file("pkg/parser/parser.go")

  context: [
    file("README.md"),
    file("CONTRIBUTING.md"),
    file("docs/style-guide.md")
  ]

  output: file("pkg/parser/README.md")
}
```

### 5. Pipelines — Multi-Step Workflows
Pipelines chain agents together, with data flowing between steps.

```langspace
pipeline "full-code-review" {
  # Step 1: Analyze the code
  step "analyze" {
    use: agent("code-analyzer")
    input: $input  # Pipeline input
  }

  # Step 2: Review based on analysis
  step "review" {
    use: agent("code-reviewer")
    input: step("analyze").output
    context: [$input]  # Include original code
  }

  # Step 3: Generate summary
  step "summarize" {
    use: agent("summarizer")
    input: step("review").output
  }

  output: step("summarize").output
}
```

### 6. Triggers — Event-Driven Execution
Triggers connect LangSpace workflows to real-world events.

```langspace
trigger "on-commit" {
  event: git.push
  filter: { branch: "main", files: "src/**/*.go" }

  run: pipeline("full-code-review") {
    input: changed_files
  }
}

trigger "daily-summary" {
  event: schedule("0 9 * * *")  # 9 AM daily

  run: intent("generate-standup") {
    input: git.commits(since: "yesterday")
  }
}
```

## Execution Model

### Direct Execution (Built-in Runtime)

LangSpace includes a built-in runtime that can execute workflows directly:

```bash
# Execute an intention
langspace run review-code.ls --input src/main.go

# Execute a pipeline
langspace run pipeline.ls --input ./project

# Start a trigger listener
langspace serve triggers.ls
```

The runtime:
- Manages LLM API connections (Anthropic, OpenAI, local models)
- Handles tool execution and MCP server integration
- Provides streaming output and progress reporting
- Supports checkpointing for long-running workflows

### Compilation Targets

For integration with existing systems, LangSpace can compile to:

| Target | Use Case |
|--------|----------|
| **Python/LangGraph** | Integration with existing Python ML pipelines |
| **TypeScript** | Node.js/Deno applications |
| **MCP Server** | Expose as tools for Claude, Cursor, etc. |
| **API Calls** | Direct HTTP/SDK calls for maximum control |
| **Prompt-only** | Just the prompts, for manual testing |

```bash
# Compile to Python
langspace compile --target python workflow.ls -o workflow.py

# Compile to MCP server
langspace compile --target mcp workflow.ls -o mcp-server/
```

## LLM Integration

### Provider Configuration

```langspace
# Global configuration
config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
    openai: {
      api_key: env("OPENAI_API_KEY")
    }
    ollama: {
      base_url: "http://localhost:11434"
    }
  }
}

# Per-agent override
agent "fast-classifier" {
  model: "gpt-4o-mini"  # Use OpenAI for this one
  # ...
}
```

### Tool Integration via MCP

LangSpace natively supports the Model Context Protocol:

```langspace
# Connect to MCP servers
mcp "filesystem" {
  command: "npx"
  args: ["-y", "@anthropic/mcp-filesystem", "/path/to/allowed"]
}

mcp "database" {
  transport: "sse"
  url: "http://localhost:3000/mcp"
}

# Use MCP tools in agents
agent "data-analyst" {
  tools: [
    mcp("filesystem").read_file,
    mcp("filesystem").write_file,
    mcp("database").query,
  ]
}
```

## Use Cases

### 1. Code Review Automation
```langspace
file "review-criteria.md" {
  contents: ```
    ## Code Review Standards
    - All functions must have doc comments
    - Error handling must be explicit
    - No magic numbers
  ```
}

agent "reviewer" {
  model: "claude-sonnet-4-20250514"
  instruction: file("review-criteria.md")
  tools: [read_file, list_directory]
}

intent "review-pr" {
  use: agent("reviewer")
  input: git.diff(base: "main")
  output: github.pr_comment()
}
```

### 2. Documentation Generation
```langspace
pipeline "generate-docs" {
  step "extract" {
    use: agent("api-extractor")
    input: file("pkg/**/*.go")
  }

  step "document" {
    use: agent("doc-writer")
    input: step("extract").output
    context: [file("docs/style-guide.md")]
  }

  step "format" {
    use: agent("markdown-formatter")
    input: step("document").output
  }

  output: file("docs/api-reference.md") <- step("format").output
}
```

### 3. Multi-Agent Debate
```langspace
agent "optimist" {
  instruction: "Always argue for the most optimistic interpretation"
}

agent "skeptic" {
  instruction: "Always challenge assumptions and find weaknesses"
}

agent "synthesizer" {
  instruction: "Find the balanced truth between opposing viewpoints"
}

pipeline "debate" {
  step "initial" {
    use: agent("optimist")
    input: $input
  }

  step "challenge" {
    use: agent("skeptic")
    input: step("initial").output
    context: [$input]
  }

  step "synthesize" {
    use: agent("synthesizer")
    input: [step("initial").output, step("challenge").output]
    context: [$input]
  }

  output: step("synthesize").output
}
```

## Roadmap

### Phase 1: Foundation (Current)
- [x] Tokenizer with block syntax support
- [x] Parser for declarations (file, agent, tool)
- [ ] AST representation for all entity types
- [ ] Basic validation and error reporting

### Phase 2: Execution
- [ ] Intent execution with LLM integration
- [ ] Pipeline orchestration
- [ ] Variable interpolation and data flow
- [ ] Progress reporting and streaming

### Phase 3: Integration
- [ ] MCP client integration
- [ ] Compilation to Python/LangGraph
- [ ] Compilation to TypeScript
- [ ] CLI tool (`langspace run`, `langspace compile`)

### Phase 4: Advanced
- [ ] Trigger system and event handling
- [ ] Checkpointing and resumption
- [ ] Multi-model routing
- [ ] Plugin system for custom tools

## Success Metrics

1. **Adoption**: 1000+ GitHub stars within 6 months
2. **Usability**: New users can create their first working agent in < 10 minutes
3. **Expressiveness**: 90% of common AI workflows can be expressed in LangSpace
4. **Performance**: Negligible overhead vs. direct API calls

## Non-Goals

- **Not a replacement for LangGraph/CrewAI**: LangSpace is a higher-level abstraction that can *target* these frameworks
- **Not a prompt engineering tool**: While prompts are part of agents, prompt optimization is out of scope
- **Not a model hosting platform**: LangSpace orchestrates, it doesn't serve models

## Technical Decisions

### Why Go?

1. **Single binary distribution**: Easy installation, no runtime dependencies
2. **Cross-platform**: Build for any OS/arch from one codebase
3. **Performance**: Fast parsing and execution
4. **Concurrency**: Built-in support for parallel agent execution

### Why a New Language?

Existing options were considered:

| Option | Why Not |
|--------|---------|
| YAML | Too rigid, no expressions, poor error messages |
| JSON | Not human-writable, no comments |
| HCL (Terraform) | Close, but too infrastructure-focused |
| Embedded DSL | Requires knowing the host language |

LangSpace syntax is designed specifically for AI workflows, with first-class support for multi-line strings, agent composition, and data flow.