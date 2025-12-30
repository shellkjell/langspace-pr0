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
| **MCP/Tool-heavy patterns** | Context window bloat from round-trip data transfer |

Developers need to:
1. Share agent configurations with teammates
2. Version control their AI workflows
3. Compose simple agents into complex pipelines
4. Test and iterate on agent behavior without rewriting code
5. Generate code/configs for their preferred runtime
6. **Minimize context window usage for cost-effective, efficient agent execution**

### The Context Window Problem

Modern AI agents often interact with external systems through tools (MCP, function calling, etc.). Each tool call requires:
- Sending the full tool schema to the model
- Receiving structured input from the model
- Executing the tool and returning results
- Loading those results into the context window

For multi-step operations, this creates massive context bloat. Consider updating a database record:

```
Tool-heavy approach:
1. Call find_record tool → Load entire record into context (1KB+)
2. Model processes record → Context grows
3. Call update_record tool → Send modified record back
4. Result loaded into context → More bloat
Total: 4+ context round-trips, full data duplication
```

```
Script-first approach:
1. Agent writes a script that finds, modifies, and saves
2. Script executes outside the model context
3. Only the result ("Updated successfully") returns
Total: 1 context exchange, minimal data
```

LangSpace's script-first philosophy addresses this by letting agents write and execute code rather than making multiple tool calls.

### The Solution

LangSpace provides a **human-readable DSL** that:
- Captures the full specification of an AI workflow
- Compiles to multiple targets (Python/LangGraph, TypeScript, API calls, etc.)
- Integrates with any LLM provider via a unified interface
- Supports direct execution with built-in LLM integration

## Core Concepts

### 1. Files — Data at Rest
Files represent static data: configuration, prompts, context documents, or generated outputs.

````langspace
file "system-prompt.md" {
  contents: ```
    You are a helpful coding assistant specializing in Go.
    Always explain your reasoning before providing code.
  ```
}

file "config.json" {
  path: "./config/app.json"  # Reference to external file
}
````

### 2. Agents — Intelligent Actors
Agents are LLM-powered entities with specific roles, instructions, and capabilities.

````langspace
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
````

### 3. Tools — Agent Capabilities
Tools extend what agents can do, connecting them to the outside world.

````langspace
tool "search_codebase" {
  description: "Search the codebase for patterns or symbols"

  parameters: {
    query: string required "The search pattern"
    file_pattern: string optional "Glob pattern for files to search"
  }

  # Tools can be implemented inline or reference external handlers
  handler: mcp("filesystem-server")
}
````

### 4. Intentions — Expressing Desired Outcomes
Intentions are the heart of LangSpace. They express *what you want to happen* rather than *how to do it*.

````langspace
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
````

### 5. Pipelines — Multi-Step Workflows
Pipelines chain agents together, with data flowing between steps.

````langspace
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
````

### 6. Triggers — Event-Driven Execution
Triggers connect LangSpace workflows to real-world events.

````langspace
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
````

### 7. Scripts — Code-First Agent Actions

Scripts enable a fundamentally more efficient way for agents to interact with external systems. Instead of making multiple tool calls (each consuming context window space), agents write executable code that performs complex multi-step operations in a single execution.

**The Problem with Tool-Heavy Approaches:**

Traditional MCP/tool-based interactions suffer from context window bloat:
1. Agent requests a database record → full record loaded into context
2. Agent decides to modify one field → entire context grows
3. Agent saves record → more context consumed
4. Each round-trip adds to the context window burden

**The Script Solution:**

````langspace
script "update-record" {
  language: "python"
  runtime: "python3"

  # Define what the script needs access to
  capabilities: [database, filesystem]

  # Scripts can be written inline or referenced
  code: ```python
    import db

    # Find, modify, and save in one execution
    record = db.find("users", {"id": user_id})
    record["description"] = new_description
    db.save("users", record)

    # Return only what matters
    print(f"Updated user {user_id}")
  ```

  # Input parameters passed to the script
  parameters: {
    user_id: string required
    new_description: string required
  }

  # Timeout and resource limits
  timeout: "30s"
  max_memory: "256MB"
}
````

**Why Scripts are More Efficient:**

| Approach | Context Usage | Round Trips |
|----------|---------------|-------------|
| Tool calls | High (full data in/out) | Multiple |
| Scripts | Low (only results) | Single |

**Script Features:**

````langspace
# Template script for agents to customize
script "db-operation" {
  language: "python"
  runtime: "python3"
  capabilities: [database]

  # Agents can provide the code dynamically
  code: $agent_generated_code

  # Sandbox configuration
  sandbox: {
    network: false          # No network access
    filesystem: "readonly"  # Read-only fs access
    allowed_modules: ["db", "json", "datetime"]
  }
}

# Script that an agent can use directly
agent "data-manager" {
  model: "claude-sonnet-4-20250514"

  instruction: ```
    You manage database records efficiently. Instead of using
    individual tool calls, write Python scripts to perform
    multi-step operations in a single execution.
  ```

  # The agent can generate and execute scripts
  scripts: [
    script("db-operation"),
    script("file-batch")
  ]
}

# Intent using script execution
intent "batch-update" {
  use: agent("data-manager")

  # Agent will write a script to handle this efficiently
  input: file("updates.json")

  # Execute the agent's generated script
  execute: script("db-operation") {
    code: $agent.output
  }
}
````

**Supported Languages:**

Scripts can be written in any language with a compatible runtime:
- Python (recommended for data operations)
- JavaScript/TypeScript
- Shell/Bash
- SQL (for database-specific operations)
- Custom DSLs

**Security Model:**

Scripts run in sandboxed environments with explicit capability grants:

````langspace
script "safe-operation" {
  language: "python"

  # Explicit capability declarations
  capabilities: [
    database.read,      # Can read from database
    database.write,     # Can write to database
    filesystem.read,    # Can read files
    # filesystem.write  # NOT granted - can't write files
  ]

  # Resource limits
  limits: {
    timeout: "60s"
    memory: "512MB"
    cpu: "1 core"
  }

  code: file("scripts/operation.py")
}
````

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

````langspace
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
````

### Tool Integration via MCP

LangSpace natively supports the Model Context Protocol:

````langspace
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
````

## Use Cases

### 1. Code Review Automation
````langspace
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
````

### 2. Documentation Generation
````langspace
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
````

### 3. Multi-Agent Debate
````langspace
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
````

## Roadmap

### Phase 1: Foundation - Complete
- [x] Tokenizer with block syntax support
- [x] Parser for declarations (file, agent, tool, intent, pipeline, etc.)
- [x] AST representation for all entity types
- [x] Basic validation and error reporting
- [x] Method calls, comparison expressions, control flow

### Phase 2: Execution - Complete
- [x] Intent execution with LLM integration
- [x] Pipeline orchestration
- [x] Variable interpolation and data flow
- [x] Progress reporting and streaming
- [x] Sandboxed script execution (Python, Shell)

### Phase 3: Integration - In Progress
- [x] MCP client integration
- [x] Compilation to Python/LangGraph
- [x] Modular Import mechanism and recursive loader
- [ ] Compilation to TypeScript
- [x] CLI tool (`langspace run`, `langspace serve`, `langspace parse`, `langspace validate`, `langspace lsp`)

### Phase 4: Advanced
- [x] Trigger system and event handling
- [x] Intelligent IDE support (Go to Definition)
- [ ] Checkpointing and resumption
- [ ] Multi-model routing
- [ ] Plugin system for custom tools

## Success Metrics

1. **Usability**: New users can create their first working agent in < 10 minutes
2. **Expressiveness**: 90% of common AI workflows can be expressed in LangSpace
3. **Performance**: Negligible overhead vs. direct API calls

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