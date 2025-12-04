# LangSpace Pipelines
# Multi-step workflows with data flowing between agents

# Simple sequential pipeline
pipeline "basic-review" {
  step "analyze" {
    use: agent("code-analyzer")
    input: $input
  }

  step "review" {
    use: agent("code-reviewer")
    input: step("analyze").output
  }

  output: step("review").output
}

# Pipeline with parallel steps
pipeline "comprehensive-review" {
  # First step: analyze the code
  step "analyze" {
    use: agent("code-analyzer")
    input: $input
  }

  # These run in parallel after analyze completes
  parallel {
    step "security" {
      use: agent("security-auditor")
      input: step("analyze").output
    }

    step "performance" {
      use: agent("performance-analyzer")
      input: step("analyze").output
    }

    step "style" {
      use: agent("style-checker")
      input: step("analyze").output
    }
  }

  # Final step combines all results
  step "summarize" {
    use: agent("summarizer")
    input: [
      step("security").output,
      step("performance").output,
      step("style").output
    ]
    context: [$input]  # Include original code for reference
  }

  output: step("summarize").output
}

# Pipeline with conditional branching
pipeline "smart-process" {
  step "classify" {
    use: agent("classifier")
    input: $input

    # This step outputs a classification
    output_schema: {
      type: enum ["bug", "feature", "refactor", "docs"]
    }
  }

  # Branch based on classification
  branch step("classify").output.type {
    "bug" => step "fix-bug" {
      use: agent("bug-fixer")
      input: $input
    }

    "feature" => step "implement" {
      use: agent("feature-builder")
      input: $input
    }

    "refactor" => step "refactor" {
      use: agent("refactorer")
      input: $input
    }

    "docs" => step "document" {
      use: agent("doc-writer")
      input: $input
    }
  }

  output: $branch.output
}

# Pipeline with loops (for iterative refinement)
pipeline "iterative-improvement" {
  step "initial" {
    use: agent("writer")
    input: $input
  }

  # Loop until quality threshold is met
  loop max: 3 {
    step "evaluate" {
      use: agent("critic")
      input: $current  # Current iteration's output

      output_schema: {
        score: number
        feedback: string
      }
    }

    # Exit loop if score is high enough
    break_if: step("evaluate").output.score >= 8

    step "improve" {
      use: agent("improver")
      input: $current
      context: [step("evaluate").output.feedback]
    }

    # Update current for next iteration
    set $current: step("improve").output
  }

  output: $current
}

# Real-world example: Documentation generation
pipeline "generate-docs" {
  # Extract API information from source
  step "extract" {
    use: agent("api-extractor")
    input: file("pkg/**/*.go")

    instruction: ```
      Extract all public functions, types, and methods.
      Include their signatures, doc comments, and example usage.
    ```
  }

  # Generate documentation for each component
  step "document" {
    use: agent("doc-writer")
    input: step("extract").output

    context: [
      file("docs/style-guide.md"),
      file("docs/templates/api-doc.md")
    ]
  }

  # Format and validate
  step "format" {
    use: agent("markdown-formatter")
    input: step("document").output
  }

  # Write output files
  step "write" {
    input: step("format").output
    output: file("docs/api/{{component}}.md")  # One file per component
  }

  output: step("write").files
}
