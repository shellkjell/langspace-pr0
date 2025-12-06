# LangSpace Advanced Example: Multi-Agent Software Factory
# A complete software development lifecycle with specialized agents
# handling requirements, design, implementation, testing, and deployment.
#
# This example demonstrates:
# - Specialized agent roles (architect, developer, tester, etc.)
# - Complex pipeline with parallel and sequential steps
# - Test-driven development workflow
# - Code quality gates
# - Git integration for version control
# - Documentation generation

# ============================================================================
# CONFIGURATION
# ============================================================================

config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
    openai: {
      api_key: env("OPENAI_API_KEY")
    }
  }

  # Development settings
  development: {
    language: "go"
    test_coverage_threshold: 80
    lint_severity_threshold: "warning"
    require_documentation: true
  }
}

# ============================================================================
# STANDARDS AND GUIDELINES
# ============================================================================

file "coding-standards" {
  contents: ```
    # Coding Standards for Go

    ## Naming Conventions
    - Use MixedCaps or mixedCaps for multi-word names
    - Acronyms should be all caps: HTTP, URL, ID
    - Interface names should describe behavior: Reader, Writer
    - Single-method interfaces should use method name + 'er'

    ## Package Design
    - Package names should be short, lowercase, no underscores
    - Avoid generic names: util, common, misc
    - Package should provide a focused set of related features

    ## Error Handling
    - Always check errors
    - Wrap errors with context: fmt.Errorf("operation failed: %w", err)
    - Use sentinel errors for expected conditions
    - Implement custom error types for complex error handling

    ## Documentation
    - Every exported symbol must have a doc comment
    - Doc comments should be complete sentences
    - Start with the name of the element being documented
    - Examples should be runnable

    ## Testing
    - Table-driven tests for multiple cases
    - Test file naming: xxx_test.go
    - Use testify for assertions where appropriate
    - Mock external dependencies
  ```
}

file "architecture-principles" {
  contents: ```
    # Architecture Principles

    ## Clean Architecture
    1. Independence of frameworks
    2. Testability
    3. Independence of UI
    4. Independence of database
    5. Independence of any external agency

    ## Dependency Rule
    - Dependencies point inward
    - Inner layers don't know about outer layers
    - Use interfaces at boundaries

    ## Layer Structure
    - Domain: Business entities and rules
    - Use Cases: Application-specific business rules
    - Interface Adapters: Controllers, presenters, gateways
    - Frameworks & Drivers: External frameworks, DB, UI

    ## Error Handling Strategy
    - Domain errors: Semantic, business-meaningful
    - Application errors: Wrap with context
    - Infrastructure errors: Convert to application errors
  ```
}

file "pr-template" {
  contents: ```
    ## Description
    {{description}}

    ## Changes
    {{changes}}

    ## Testing
    - [ ] Unit tests added/updated
    - [ ] Integration tests verified
    - [ ] Manual testing completed

    ## Documentation
    - [ ] README updated (if needed)
    - [ ] API documentation updated
    - [ ] Code comments added

    ## Checklist
    - [ ] Code follows project style guidelines
    - [ ] No new warnings or errors
    - [ ] All tests pass
    - [ ] Coverage threshold met
  ```
}

# ============================================================================
# TOOLS
# ============================================================================

mcp "git" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@anthropic/mcp-git"]
}

mcp "filesystem" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
}

tool "run_tests" {
  description: "Run Go tests with coverage"

  parameters: {
    package: string optional "./..."
    verbose: bool optional false
    race: bool optional true
    cover: bool optional true
  }

  handler: shell {
    command: ```
      go test {{if params.verbose}}-v{{end}} \
        {{if params.race}}-race{{end}} \
        {{if params.cover}}-coverprofile=coverage.out{{end}} \
        {{params.package}}
    ```
    timeout: "10m"
  }
}

tool "run_linter" {
  description: "Run golangci-lint"

  parameters: {
    paths: array optional ["."]
    fix: bool optional false
  }

  handler: shell {
    command: ```
      golangci-lint run \
        {{if params.fix}}--fix{{end}} \
        --out-format json \
        {{params.paths | join(' ')}}
    ```
    timeout: "5m"
  }
}

tool "go_build" {
  description: "Build the Go project"

  parameters: {
    output: string optional "./bin/app"
    target: string optional "./cmd/app"
  }

  handler: shell {
    command: "go build -o {{params.output}} {{params.target}}"
    timeout: "5m"
  }
}

tool "check_coverage" {
  description: "Check test coverage percentage"

  handler: shell {
    command: "go tool cover -func=coverage.out | tail -1 | awk '{print $3}'"
  }
}

# Efficient script for multi-file code generation
script "generate-code" {
  language: "python"
  runtime: "python3"

  capabilities: [filesystem]

  parameters: {
    files: array required "Array of {path, content} objects"
    base_dir: string optional "."
  }

  code: ```python
    import json
    import os

    files = json.loads(files)
    created = []
    errors = []

    for file_spec in files:
        path = os.path.join(base_dir, file_spec["path"])
        content = file_spec["content"]

        try:
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, 'w') as f:
                f.write(content)
            created.append(path)
            print(f"✓ Created: {path}")
        except Exception as e:
            errors.append(f"{path}: {e}")
            print(f"✗ Failed: {path} - {e}")

    print(f"\nSummary: {len(created)} files created, {len(errors)} errors")
  ```
}

# Script for code analysis
script "analyze-codebase" {
  language: "python"
  runtime: "python3"

  capabilities: [filesystem.read]

  parameters: {
    root: string optional "."
    extensions: array optional [".go"]
  }

  code: ```python
    import os
    import json
    from collections import defaultdict

    stats = {
        "total_files": 0,
        "total_lines": 0,
        "by_package": defaultdict(lambda: {"files": 0, "lines": 0}),
        "largest_files": []
    }

    extensions = json.loads(extensions) if isinstance(extensions, str) else extensions

    for root_dir, dirs, files in os.walk(root):
        # Skip vendor and hidden directories
        dirs[:] = [d for d in dirs if not d.startswith('.') and d != 'vendor']

        for filename in files:
            if not any(filename.endswith(ext) for ext in extensions):
                continue

            filepath = os.path.join(root_dir, filename)
            rel_path = os.path.relpath(filepath, root)
            package = os.path.dirname(rel_path) or "root"

            try:
                with open(filepath) as f:
                    lines = len(f.readlines())

                stats["total_files"] += 1
                stats["total_lines"] += lines
                stats["by_package"][package]["files"] += 1
                stats["by_package"][package]["lines"] += lines
                stats["largest_files"].append((rel_path, lines))
            except Exception:
                pass

    # Sort largest files
    stats["largest_files"].sort(key=lambda x: x[1], reverse=True)
    stats["largest_files"] = stats["largest_files"][:10]

    print(f"Codebase Analysis")
    print(f"=" * 40)
    print(f"Total Files: {stats['total_files']}")
    print(f"Total Lines: {stats['total_lines']}")
    print(f"\nBy Package:")
    for pkg, data in sorted(stats["by_package"].items()):
        print(f"  {pkg}: {data['files']} files, {data['lines']} lines")
    print(f"\nLargest Files:")
    for path, lines in stats["largest_files"]:
        print(f"  {path}: {lines} lines")
  ```
}

# ============================================================================
# SPECIALIZED AGENTS
# ============================================================================

agent "product-owner" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.4

  instruction: ```
    You are a product owner who translates high-level requirements into
    clear, actionable user stories and acceptance criteria.

    Your responsibilities:
    1. Clarify ambiguous requirements
    2. Break down features into implementable user stories
    3. Define clear acceptance criteria
    4. Prioritize based on value and complexity
    5. Identify dependencies between stories

    Output Format (JSON):
    {
      "feature_summary": "...",
      "user_stories": [
        {
          "id": "US-001",
          "title": "...",
          "as_a": "...",
          "i_want": "...",
          "so_that": "...",
          "acceptance_criteria": ["..."],
          "priority": "high|medium|low",
          "story_points": 3,
          "dependencies": ["US-002"]
        }
      ],
      "non_functional_requirements": ["..."],
      "out_of_scope": ["..."]
    }
  ```
}

agent "architect" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("architecture-principles")

  tools: [
    mcp("filesystem").read_file,
    mcp("filesystem").list_directory,
  ]
}

agent "senior-developer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: file("coding-standards")

  tools: [
    mcp("git").get_diff,
    mcp("filesystem").read_file,
    mcp("filesystem").write_file,
    tool("go_build"),
  ]
}

agent "test-engineer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You are a test engineer specializing in Go. You write comprehensive
    tests that verify functionality and prevent regressions.

    Testing philosophy:
    1. Test behavior, not implementation
    2. Use table-driven tests for multiple cases
    3. Mock external dependencies
    4. Test edge cases and error conditions
    5. Keep tests readable and maintainable

    Test types to generate:
    - Unit tests for individual functions
    - Integration tests for component interactions
    - Benchmark tests for performance-critical code
    - Example tests for documentation

    Output runnable, idiomatic Go test code.
  ```

  tools: [
    mcp("filesystem").read_file,
    mcp("filesystem").write_file,
    tool("run_tests"),
  ]
}

agent "code-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You are a senior code reviewer focused on code quality, security,
    and maintainability. You provide constructive, actionable feedback.

    Review checklist:
    1. Code correctness and logic errors
    2. Security vulnerabilities
    3. Performance issues
    4. Error handling completeness
    5. Code style and consistency
    6. Documentation adequacy
    7. Test coverage
    8. Design patterns and architecture

    For each issue found:
    - Explain why it's a problem
    - Suggest a specific fix
    - Rate severity: critical, major, minor, suggestion

    Be thorough but constructive. Acknowledge good patterns too.
  ```

  tools: [
    mcp("git").get_diff,
    mcp("filesystem").read_file,
    tool("run_linter"),
    tool("check_coverage"),
  ]
}

agent "security-auditor" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.1  # Very precise for security

  instruction: ```
    You are a security auditor specializing in Go applications.
    Your job is to identify security vulnerabilities and risks.

    Vulnerability categories to check:
    1. Injection (SQL, command, template)
    2. Broken authentication/authorization
    3. Sensitive data exposure
    4. Security misconfiguration
    5. Insecure deserialization
    6. Using components with known vulnerabilities
    7. Insufficient logging/monitoring

    Go-specific checks:
    - Unsafe package usage
    - Race conditions
    - Improper error handling that leaks info
    - Hardcoded secrets
    - Weak cryptography

    Rate each finding:
    - CRITICAL: Immediate exploitation possible
    - HIGH: Significant risk, fix before release
    - MEDIUM: Should be fixed, but not blocking
    - LOW: Minor risk, fix when convenient
  ```

  tools: [
    mcp("filesystem").read_file,
    tool("run_linter"),
  ]
}

agent "documentation-writer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.4

  instruction: ```
    You write clear, comprehensive documentation for Go code.

    Documentation types:
    1. Package-level documentation
    2. Type and function doc comments
    3. README files
    4. API documentation
    5. Examples and tutorials

    Style guidelines:
    - Start with a one-sentence summary
    - Use complete sentences
    - Include usage examples
    - Document edge cases and errors
    - Keep it concise but complete
  ```

  tools: [
    mcp("filesystem").read_file,
    mcp("filesystem").write_file,
  ]
}

# ============================================================================
# DEVELOPMENT PIPELINE
# ============================================================================

pipeline "develop-feature" {
  # Step 1: Analyze and plan
  step "plan" {
    use: agent("product-owner")
    input: $input

    instruction: "Break down this feature request into user stories."
  }

  # Step 2: Design architecture
  step "design" {
    use: agent("architect")

    input: step("plan").output

    context: [
      file("architecture-principles")
    ]

    instruction: ```
      Design the architecture for implementing these user stories.
      Consider the existing codebase structure.

      Output:
      1. Component diagram
      2. File structure
      3. Interface definitions
      4. Data models
      5. API contracts
    ```
  }

  # Step 3: Implement in TDD fashion
  step "implement" {
    # First, generate tests
    step "write-tests" {
      use: agent("test-engineer")
      input: step("design").output

      instruction: "Write failing tests for this design (TDD red phase)."
    }

    # Then implement to make tests pass
    step "write-code" {
      use: agent("senior-developer")

      input: step("write-tests").output
      context: [step("design").output]

      instruction: "Implement code to make all tests pass (TDD green phase)."
    }

    # Refactor if needed
    step "refactor" {
      use: agent("senior-developer")
      input: step("write-code").output

      instruction: "Refactor for clarity and efficiency (TDD refactor phase)."
    }
  }

  # Step 4: Quality gates (parallel)
  parallel {
    step "review" {
      use: agent("code-reviewer")
      input: step("implement").output
    }

    step "security" {
      use: agent("security-auditor")
      input: step("implement").output
    }

    step "test" {
      tools: [tool("run_tests"), tool("check_coverage")]
      input: step("implement").output
    }

    step "lint" {
      tools: [tool("run_linter")]
      input: step("implement").output
    }
  }

  # Step 5: Evaluate quality gate results
  step "evaluate-quality" {
    input: [
      step("review").output,
      step("security").output,
      step("test").output,
      step("lint").output
    ]

    instruction: ```
      Evaluate all quality gate results.
      Determine if the code is ready for merge.

      Output:
      {
        "ready_to_merge": true/false,
        "blocking_issues": [...],
        "warnings": [...],
        "test_coverage": "XX%",
        "security_findings": {...}
      }
    ```
  }

  # Step 6: Fix issues if not ready
  branch step("evaluate-quality").output.ready_to_merge {
    false => step "fix-issues" {
      use: agent("senior-developer")

      input: step("evaluate-quality").output.blocking_issues
      context: [step("implement").output]

      instruction: "Fix the blocking issues identified in the review."
    }
  }

  # Step 7: Generate documentation
  step "document" {
    use: agent("documentation-writer")

    input: step("implement").output

    context: [
      step("plan").output,
      step("design").output
    ]

    instruction: "Generate comprehensive documentation for the new code."
  }

  # Step 8: Prepare PR
  step "prepare-pr" {
    input: [
      step("plan").output,
      step("implement").output,
      step("document").output
    ]

    context: [file("pr-template")]

    instruction: "Generate a pull request description using the template."
  }

  output: {
    code: step("implement").output,
    documentation: step("document").output,
    pr_description: step("prepare-pr").output,
    quality_report: step("evaluate-quality").output
  }
}

# Hotfix pipeline for urgent fixes
pipeline "hotfix" {
  step "analyze" {
    use: agent("senior-developer")
    input: $input

    instruction: "Analyze this bug report and identify the root cause."
  }

  step "fix" {
    use: agent("senior-developer")
    input: step("analyze").output

    instruction: "Implement the minimal fix for this bug."
  }

  step "test" {
    use: agent("test-engineer")
    input: step("fix").output

    instruction: "Add a regression test for this bug fix."
  }

  step "verify" {
    tools: [tool("run_tests"), tool("run_linter")]
    input: [step("fix").output, step("test").output]
  }

  output: {
    fix: step("fix").output,
    test: step("test").output,
    verification: step("verify").output
  }
}

# ============================================================================
# TRIGGERS
# ============================================================================

trigger "feature-request" {
  event: github.issue {
    labels: ["feature", "enhancement"]
  }

  run: pipeline("develop-feature") {
    input: github.issue.body
  }

  on_complete: {
    github.create_pr(
      title: "Implement: " + github.issue.title,
      body: output.pr_description,
      branch: "feature/" + github.issue.number
    )
    github.issue.comment("Implementation ready for review: " + output.pr_url)
  }

  on_error: {
    github.issue.comment("Implementation failed: " + error.message)
    github.issue.add_label("needs-attention")
  }
}

trigger "bug-report" {
  event: github.issue {
    labels: ["bug", "urgent"]
  }

  run: pipeline("hotfix") {
    input: github.issue.body
  }

  on_complete: {
    github.create_pr(
      title: "Fix: " + github.issue.title,
      body: "Fixes #" + github.issue.number,
      branch: "hotfix/" + github.issue.number
    )
  }
}

trigger "pr-review" {
  event: github.pull_request {
    actions: ["opened", "synchronize"]
  }

  run: parallel {
    step "code-review" {
      use: agent("code-reviewer")
      input: github.pr.diff
    }

    step "security-scan" {
      use: agent("security-auditor")
      input: github.pr.diff
    }
  }

  on_complete: {
    github.pr.review(
      body: output,
      event: "COMMENT"
    )
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

# Develop a new feature
intent "develop" {
  params: {
    spec: string required "Feature specification or path to spec file"
  }

  run: pipeline("develop-feature") {
    input: params.spec
  }

  on_complete: {
    print("Development complete!")
    print("Files modified: " + output.code.files.length)
    print("Quality score: " + output.quality_report.score)
  }
}

# Fix a bug
intent "fix" {
  params: {
    bug: string required "Bug description or issue number"
  }

  run: pipeline("hotfix") {
    input: params.bug
  }
}

# Review code
intent "review" {
  params: {
    target: string optional "HEAD" "Commit, branch, or file to review"
  }

  use: agent("code-reviewer")
  input: git.diff(params.target)

  output: stdout
}

# Generate documentation
intent "document" {
  params: {
    path: string required "Path to file or directory"
  }

  use: agent("documentation-writer")
  input: file(params.path)

  output: stdout
}

# Run full quality check
intent "check" {
  run: parallel {
    step "test" {
      tools: [tool("run_tests")]
    }
    step "lint" {
      tools: [tool("run_linter")]
    }
    step "coverage" {
      tools: [tool("check_coverage")]
    }
  }

  on_complete: {
    print("Tests: " + output.test.summary)
    print("Lint: " + output.lint.summary)
    print("Coverage: " + output.coverage)
  }
}
