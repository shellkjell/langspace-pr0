# LangSpace Complete Example: Code Review System
# A production-ready code review automation workflow

config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
  }
}

# ============================================
# FILES
# ============================================

file "review-guidelines" {
  contents: ```
    # Code Review Guidelines

    ## Security
    - Check for SQL injection vulnerabilities
    - Validate all user inputs
    - Ensure secrets are not hardcoded

    ## Performance
    - Look for N+1 query patterns
    - Check for unnecessary memory allocations
    - Identify blocking operations in hot paths

    ## Style
    - Follow Go idioms and conventions
    - Use meaningful variable names
    - Keep functions focused and small

    ## Testing
    - Verify adequate test coverage
    - Check for edge case handling
    - Ensure tests are deterministic
  ```
}

file "output-template" {
  contents: ```
    # Code Review Report

    **Reviewed**: {{date}}
    **Files**: {{file_count}}
    **Author**: {{author}}

    ## Summary
    {{summary}}

    ## Issues Found
    {{issues}}

    ## Recommendations
    {{recommendations}}

    ## Verdict
    {{verdict}}
  ```
}

# ============================================
# TOOLS
# ============================================

mcp "git" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@anthropic/mcp-git"]
}

tool "run_linter" {
  description: "Run golangci-lint on the codebase"

  parameters: {
    paths: array optional ["."] "Paths to lint"
  }

  handler: shell {
    command: "golangci-lint run --out-format json {{params.paths | join(' ')}}"
    timeout: "2m"
  }
}

tool "run_tests" {
  description: "Run Go tests with coverage"

  parameters: {
    package: string optional "./..." "Package pattern"
    race: bool optional true "Enable race detector"
  }

  handler: shell {
    command: "go test {{if params.race}}-race{{end}} -coverprofile=coverage.out {{params.package}}"
    timeout: "5m"
  }
}

# ============================================
# AGENTS
# ============================================

agent "code-analyzer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You are a code analysis specialist. Your job is to understand
    code structure and identify patterns.

    When given code, extract:
    1. Main purpose and functionality
    2. Public API surface
    3. Dependencies and imports
    4. Complexity hotspots

    Output structured JSON.
  ```

  tools: [
    mcp("git").get_diff,
    mcp("git").get_file,
    tool("run_linter"),
  ]
}

agent "security-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.1  # Low temperature for precise analysis

  instruction: ```
    You are a security specialist focused on identifying vulnerabilities.

    Check for:
    - Injection vulnerabilities (SQL, command, etc.)
    - Authentication/authorization issues
    - Secrets exposure
    - Input validation gaps
    - Cryptographic weaknesses

    Rate each finding by severity: CRITICAL, HIGH, MEDIUM, LOW
  ```

  tools: [
    mcp("git").get_file,
    mcp("git").search,
  ]
}

agent "style-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("review-guidelines")

  tools: [
    mcp("git").get_file,
    tool("run_linter"),
  ]
}

agent "summarizer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.5

  instruction: ```
    You synthesize multiple review perspectives into a cohesive report.

    Your output should:
    1. Prioritize issues by severity and impact
    2. Remove duplicate findings
    3. Provide a clear overall verdict
    4. Be actionable and specific

    Use the provided template for formatting.
  ```
}

# ============================================
# PIPELINE
# ============================================

pipeline "full-review" {
  # Step 1: Analyze the code structure
  step "analyze" {
    use: agent("code-analyzer")
    input: $input

    instruction: "Analyze this code change and identify areas of concern."
  }

  # Step 2: Parallel reviews
  parallel {
    step "security" {
      use: agent("security-reviewer")
      input: $input
      context: [step("analyze").output]
    }

    step "style" {
      use: agent("style-reviewer")
      input: $input
      context: [step("analyze").output]
    }

    step "tests" {
      # Run tests and capture results
      tools: [tool("run_tests")]
      input: $input
    }
  }

  # Step 3: Synthesize results
  step "summarize" {
    use: agent("summarizer")

    input: [
      step("security").output,
      step("style").output,
      step("tests").output,
    ]

    context: [
      $input,
      file("output-template"),
    ]
  }

  output: step("summarize").output
}

# ============================================
# TRIGGERS
# ============================================

trigger "pr-opened" {
  event: github.pull_request {
    actions: ["opened", "synchronize"]
  }

  run: pipeline("full-review") {
    input: github.pr.diff
  }

  on_complete: {
    github.pr.comment(output)
  }
}

trigger "manual-review" {
  event: cli.command("review")

  run: pipeline("full-review") {
    input: git.staged_files()
  }

  on_complete: {
    write_file("review-output.md", output)
    print("Review complete! See review-output.md")
  }
}

# ============================================
# CLI ENTRYPOINTS
# ============================================

# Run with: langspace run code-review.ls review <files>
intent "review" {
  params: {
    files: array required "Files to review"
  }

  run: pipeline("full-review") {
    input: params.files
  }

  output: stdout
}

# Run with: langspace run code-review.ls quick-check
intent "quick-check" {
  use: agent("style-reviewer")
  input: git.staged_files()
  output: stdout
}
