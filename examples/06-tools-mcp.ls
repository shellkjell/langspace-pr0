# LangSpace Tools and MCP Integration
# Defining tools and connecting to external capabilities

# Simple inline tool definition
tool "get_weather" {
  description: "Get current weather for a location"

  parameters: {
    location: string required "City name or coordinates"
    units: string optional "metric" "Temperature units (metric/imperial)"
  }

  # Inline implementation using HTTP
  handler: http {
    method: "GET"
    url: "https://api.weather.com/v1/current"
    query: {
      q: params.location
      units: params.units
    }
    headers: {
      "X-API-Key": env("WEATHER_API_KEY")
    }
  }
}

# Tool implemented by a shell command
tool "run_tests" {
  description: "Run the project's test suite"

  parameters: {
    package: string optional "./..." "Package pattern to test"
    verbose: bool optional false "Enable verbose output"
  }

  handler: shell {
    command: "go test {{if params.verbose}}-v{{end}} {{params.package}}"
    timeout: "5m"
    working_dir: $project_root
  }
}

# Tool that reads from the filesystem
tool "read_file" {
  description: "Read the contents of a file"

  parameters: {
    path: string required "Path to the file"
  }

  handler: builtin("fs.read")
}

# MCP Server connection
mcp "filesystem" {
  # Stdio transport (spawns a process)
  transport: "stdio"
  command: "npx"
  args: ["-y", "@modelcontextprotocol/server-filesystem", "/allowed/path"]

  # Environment variables for the server
  env: {
    "DEBUG": "mcp:*"
  }
}

# MCP Server with SSE transport
mcp "database" {
  transport: "sse"
  url: "http://localhost:3000/mcp"

  # Authentication
  headers: {
    "Authorization": "Bearer {{env('DB_MCP_TOKEN')}}"
  }
}

# Using MCP tools in agents
agent "data-analyst" {
  model: "claude-sonnet-4-20250514"

  instruction: ```
    You are a data analyst. Use the available tools to query
    databases and analyze results.
  ```

  # Reference specific tools from MCP servers
  tools: [
    mcp("filesystem").read_file,
    mcp("filesystem").write_file,
    mcp("filesystem").list_directory,
    mcp("database").query,
    mcp("database").list_tables,
  ]
}

# Agent with mixed tools (built-in + MCP + custom)
agent "full-stack" {
  model: "claude-sonnet-4-20250514"

  tools: [
    # Built-in tools
    read_file,
    write_file,

    # Custom tools defined above
    tool("run_tests"),
    tool("get_weather"),

    # MCP tools
    mcp("database").query,
  ]
}

# Tool with structured output schema
tool "analyze_code" {
  description: "Perform static analysis on code"

  parameters: {
    file: string required "Path to source file"
    checks: array optional ["all"] "List of checks to run"
  }

  # Define the expected output shape
  output_schema: {
    issues: array {
      line: number
      column: number
      severity: enum ["error", "warning", "info"]
      message: string
      rule: string
    }
    summary: {
      errors: number
      warnings: number
      info: number
    }
  }

  handler: shell {
    command: "golangci-lint run --out-format json {{params.file}}"
  }
}
