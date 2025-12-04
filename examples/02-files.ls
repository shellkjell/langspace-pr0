# LangSpace File Operations
# Demonstrating files with actual contents and file references

# Inline file contents
file "system-prompt.md" {
  contents: ```
    You are an expert Go developer. When reviewing code:
    1. Check for idiomatic Go patterns
    2. Look for potential race conditions
    3. Suggest performance improvements
    4. Ensure proper error handling
  ```
}

# Reference to external file (read at runtime)
file "project-config" {
  path: "./config/settings.json"
}

# File with metadata
file "output-report" {
  path: "./reports/analysis.md"
  mode: "write"  # This file will be created/overwritten
}

# Glob pattern for multiple files
file "source-files" {
  glob: "pkg/**/*.go"
  exclude: ["*_test.go"]
}

# Agent that uses files
agent "analyzer" {
  model: "claude-sonnet-4-20250514"
  instruction: file("system-prompt.md")  # Reference the file contents
}

# Intent that processes files
intent "analyze-project" {
  use: agent("analyzer")
  input: file("source-files")
  output: file("output-report")
}
