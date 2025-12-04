# LangSpace Intentions
# The primary way to invoke agents and express desired outcomes

# Simple intention - just use an agent
intent "quick-review" {
  use: agent("code-reviewer")
  input: file("src/main.go")
}

# Intention with explicit output destination
intent "documented-review" {
  use: agent("code-reviewer")
  input: file("src/**/*.go")
  output: file("reviews/{{date}}-review.md")  # Variable interpolation
}

# Intention with context - additional information for the agent
intent "contextual-review" {
  use: agent("code-reviewer")

  # The primary input to review
  input: file("pkg/parser/parser.go")

  # Additional context the agent should consider
  context: [
    file("CONTRIBUTING.md"),
    file("docs/style-guide.md"),
    file("pkg/parser/README.md")
  ]

  output: file("reviews/parser-review.md")
}

# Intention with parameters - for reusable templates
intent "review-module" {
  params: {
    module: string required "The module path to review"
    depth: string optional "shallow" "How deep to analyze"
  }

  use: agent("code-reviewer")
  input: file("{{params.module}}/**/*.go")

  context: [
    file("{{params.module}}/README.md")
  ]
}

# Intention with conditional output
intent "smart-review" {
  use: agent("code-reviewer")
  input: git.staged_files()

  # Different outputs based on result
  on_success: {
    output: file("reviews/latest.md")
    notify: slack("#code-reviews")
  }

  on_failure: {
    output: file("reviews/failed.md")
    notify: slack("#dev-alerts")
  }
}

# Intention with human-in-the-loop
intent "careful-refactor" {
  use: agent("refactorer")
  input: file("src/legacy.go")

  # Pause for human approval before applying changes
  require_approval: true
  approval_prompt: "Review the proposed refactoring changes"

  output: file("src/legacy.go")  # In-place modification
}
