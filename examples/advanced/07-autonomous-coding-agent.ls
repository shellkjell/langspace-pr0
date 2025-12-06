# LangSpace Advanced Example: Autonomous Coding Agent
# Self-improving coding agent with TDD, debugging, and git integration.
#
# This example demonstrates:
# - Test-driven development workflow
# - Self-debugging capabilities
# - Git-aware code changes
# - Multi-file refactoring
# - Code generation with context
# - Autonomous problem solving

# ============================================================================
# CONFIGURATION
# ============================================================================

config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
  }

  # Development settings
  development: {
    language: "typescript"
    framework: "node"
    test_runner: "jest"
    linter: "eslint"
    formatter: "prettier"
  }

  # Safety settings
  safety: {
    require_tests: true
    max_files_per_change: 10
    protected_paths: [".env", "secrets/", "node_modules/"]
    require_review_for: ["package.json", "tsconfig.json", "*.config.*"]
  }

  # Git settings
  git: {
    auto_commit: false
    branch_prefix: "ai/"
    commit_style: "conventional"
  }
}

# ============================================================================
# CODING STANDARDS
# ============================================================================

file "coding-standards" {
  contents: ```
    # Coding Standards

    ## TypeScript Guidelines

    ### Types
    - Prefer interfaces over type aliases for object shapes
    - Use strict mode always
    - Avoid `any`, use `unknown` if type is truly unknown
    - Export types from the module that owns them

    ### Functions
    - Pure functions when possible
    - Maximum 3 parameters, use objects for more
    - Single responsibility
    - Early returns for guard clauses

    ### Error Handling
    - Use custom error classes
    - Never swallow errors silently
    - Provide context in error messages
    - Use Result<T, E> pattern for recoverable errors

    ### Testing
    - Test behavior, not implementation
    - One assertion per test when possible
    - Descriptive test names: "should [action] when [condition]"
    - AAA pattern: Arrange, Act, Assert

    ### Documentation
    - JSDoc for public APIs
    - Inline comments for complex logic only
    - README for each module

    ## Project Structure
    ```
    src/
      domain/       # Business logic, no framework deps
      application/  # Use cases, orchestration
      infrastructure/  # External integrations
      presentation/ # API/UI layer
    tests/
      unit/
      integration/
      e2e/
    ```

    ## Git Commit Format
    ```
    <type>(<scope>): <subject>

    [optional body]

    [optional footer]
    ```

    Types: feat, fix, docs, style, refactor, test, chore
  ```
}

file "test-template" {
  contents: ```
    import { describe, it, expect, beforeEach, afterEach } from 'vitest';

    describe('{{ClassName}}', () => {
      // Setup
      let sut: {{ClassName}};

      beforeEach(() => {
        sut = new {{ClassName}}();
      });

      afterEach(() => {
        // Cleanup
      });

      describe('{{methodName}}', () => {
        it('should {{expectedBehavior}} when {{condition}}', () => {
          // Arrange
          const input = {{testInput}};

          // Act
          const result = sut.{{methodName}}(input);

          // Assert
          expect(result).toEqual({{expectedOutput}});
        });

        it('should throw {{ErrorType}} when {{errorCondition}}', () => {
          // Arrange
          const invalidInput = {{invalidInput}};

          // Act & Assert
          expect(() => sut.{{methodName}}(invalidInput)).toThrow({{ErrorType}});
        });
      });
    });
  ```
}

# ============================================================================
# TOOLS
# ============================================================================

mcp "filesystem" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@anthropic/mcp-filesystem", env("PROJECT_ROOT")]
}

mcp "git" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@anthropic/mcp-git", env("PROJECT_ROOT")]
}

tool "read_file" {
  description: "Read a file from the project"
  handler: mcp("filesystem").read_file
}

tool "write_file" {
  description: "Write content to a file"
  handler: mcp("filesystem").write_file
}

tool "list_dir" {
  description: "List directory contents"
  handler: mcp("filesystem").list_directory
}

tool "search_files" {
  description: "Search for files matching a pattern"
  handler: mcp("filesystem").search_files
}

tool "git_status" {
  description: "Get git status"
  handler: mcp("git").status
}

tool "git_diff" {
  description: "Get diff of changes"
  handler: mcp("git").diff
}

tool "git_commit" {
  description: "Create a commit"
  handler: mcp("git").commit
}

tool "git_branch" {
  description: "Create or switch branches"
  handler: mcp("git").branch
}

tool "run_tests" {
  description: "Run tests"

  parameters: {
    pattern: string optional "Test file pattern"
    watch: bool optional false
    coverage: bool optional false
  }

  handler: shell {
    command: "npm test -- {{params.pattern}} {{params.coverage ? '--coverage' : ''}}"
    timeout: "120s"
  }
}

tool "run_linter" {
  description: "Run linter"

  parameters: {
    fix: bool optional false
    path: string optional "src/"
  }

  handler: shell {
    command: "npm run lint {{params.path}} {{params.fix ? '-- --fix' : ''}}"
    timeout: "60s"
  }
}

tool "run_typecheck" {
  description: "Run TypeScript type checker"

  handler: shell {
    command: "npx tsc --noEmit"
    timeout: "60s"
  }
}

tool "run_build" {
  description: "Build the project"

  handler: shell {
    command: "npm run build"
    timeout: "120s"
  }
}

tool "install_package" {
  description: "Install an npm package"

  parameters: {
    package: string required "Package name"
    dev: bool optional false "Install as dev dependency"
  }

  handler: shell {
    command: "npm install {{params.dev ? '-D' : ''}} {{params.package}}"
    timeout: "60s"
  }
}

# ============================================================================
# CODE ANALYSIS SCRIPTS
# ============================================================================

# Parse and understand code structure
script "analyze-code-structure" {
  language: "javascript"
  runtime: "node"

  capabilities: [filesystem.read]

  parameters: {
    file_path: string required
  }

  code: ```javascript
    const fs = require('fs');
    const path = require('path');

    const content = fs.readFileSync(file_path, 'utf-8');
    const lines = content.split('\n');

    const analysis = {
      imports: [],
      exports: [],
      classes: [],
      functions: [],
      interfaces: [],
      types: [],
      dependencies: []
    };

    // Parse imports
    const importRegex = /import\s+(?:{([^}]+)}|(\w+))\s+from\s+['"]([^'"]+)['"]/g;
    let match;
    while ((match = importRegex.exec(content)) !== null) {
      analysis.imports.push({
        named: match[1]?.split(',').map(s => s.trim()) || [],
        default: match[2] || null,
        from: match[3]
      });
      if (!match[3].startsWith('.')) {
        analysis.dependencies.push(match[3].split('/')[0]);
      }
    }

    // Parse exports
    const exportRegex = /export\s+(?:default\s+)?(?:async\s+)?(?:function|class|const|interface|type)\s+(\w+)/g;
    while ((match = exportRegex.exec(content)) !== null) {
      analysis.exports.push(match[1]);
    }

    // Parse classes
    const classRegex = /class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+([^{]+))?/g;
    while ((match = classRegex.exec(content)) !== null) {
      analysis.classes.push({
        name: match[1],
        extends: match[2] || null,
        implements: match[3]?.split(',').map(s => s.trim()) || []
      });
    }

    // Parse functions
    const funcRegex = /(?:async\s+)?function\s+(\w+)|const\s+(\w+)\s*=\s*(?:async\s*)?\(/g;
    while ((match = funcRegex.exec(content)) !== null) {
      analysis.functions.push(match[1] || match[2]);
    }

    // Parse interfaces
    const interfaceRegex = /interface\s+(\w+)(?:\s+extends\s+([^{]+))?/g;
    while ((match = interfaceRegex.exec(content)) !== null) {
      analysis.interfaces.push({
        name: match[1],
        extends: match[2]?.split(',').map(s => s.trim()) || []
      });
    }

    // Parse type aliases
    const typeRegex = /type\s+(\w+)\s*=/g;
    while ((match = typeRegex.exec(content)) !== null) {
      analysis.types.push(match[1]);
    }

    analysis.lineCount = lines.length;
    analysis.dependencies = [...new Set(analysis.dependencies)];

    console.log(JSON.stringify(analysis, null, 2));
  ```
}

# Find related files
script "find-related-files" {
  language: "javascript"
  runtime: "node"

  capabilities: [filesystem.read]

  parameters: {
    file_path: string required
    project_root: string required
  }

  code: ```javascript
    const fs = require('fs');
    const path = require('path');

    const content = fs.readFileSync(file_path, 'utf-8');
    const dir = path.dirname(file_path);
    const baseName = path.basename(file_path, path.extname(file_path));

    const related = {
      tests: [],
      imports: [],
      importedBy: [],
      sameModule: []
    };

    // Find test files
    const testPatterns = [
      path.join(dir, `${baseName}.test.ts`),
      path.join(dir, `${baseName}.spec.ts`),
      path.join(project_root, 'tests', 'unit', `${baseName}.test.ts`),
      path.join(project_root, '__tests__', `${baseName}.test.ts`)
    ];

    for (const testPath of testPatterns) {
      if (fs.existsSync(testPath)) {
        related.tests.push(testPath);
      }
    }

    // Find imports in this file
    const importRegex = /from\s+['"](\.[^'"]+)['"]/g;
    let match;
    while ((match = importRegex.exec(content)) !== null) {
      const importPath = path.resolve(dir, match[1]);
      const extensions = ['.ts', '.tsx', '.js', '.jsx', '/index.ts'];
      for (const ext of extensions) {
        const fullPath = importPath + ext;
        if (fs.existsSync(fullPath)) {
          related.imports.push(fullPath);
          break;
        }
      }
    }

    // Find files in same directory
    const dirFiles = fs.readdirSync(dir);
    for (const f of dirFiles) {
      if (f !== path.basename(file_path) && f.endsWith('.ts')) {
        related.sameModule.push(path.join(dir, f));
      }
    }

    console.log(JSON.stringify(related, null, 2));
  ```
}

# Extract test failures
script "parse-test-output" {
  language: "python"
  runtime: "python3"

  parameters: {
    output: string required "Test runner output"
  }

  code: ```python
    import json
    import re

    output = output

    failures = []
    current_failure = None

    # Jest-style parsing
    failure_pattern = r'● (.+)'
    expect_pattern = r'expect\((.+)\)\.(.+)\((.+)\)'
    received_pattern = r'Received: (.+)'
    expected_pattern = r'Expected: (.+)'
    at_pattern = r'at .+ \((.+):(\d+):(\d+)\)'

    lines = output.split('\n')
    i = 0

    while i < len(lines):
        line = lines[i]

        # New failure
        match = re.match(failure_pattern, line)
        if match:
            if current_failure:
                failures.append(current_failure)
            current_failure = {
                "test_name": match.group(1),
                "expected": None,
                "received": None,
                "file": None,
                "line": None,
                "assertion": None
            }

        # Extract assertion
        if current_failure and 'expect' in line:
            match = re.search(expect_pattern, line)
            if match:
                current_failure["assertion"] = line.strip()

        # Extract expected/received
        if current_failure:
            exp_match = re.search(expected_pattern, line)
            if exp_match:
                current_failure["expected"] = exp_match.group(1)

            rec_match = re.search(received_pattern, line)
            if rec_match:
                current_failure["received"] = rec_match.group(1)

            at_match = re.search(at_pattern, line)
            if at_match:
                current_failure["file"] = at_match.group(1)
                current_failure["line"] = int(at_match.group(2))

        i += 1

    if current_failure:
        failures.append(current_failure)

    # Summary
    passed = len(re.findall(r'✓', output))
    failed = len(failures)

    result = {
        "passed": passed,
        "failed": failed,
        "failures": failures,
        "all_passing": failed == 0
    }

    print(json.dumps(result, indent=2))
  ```
}

# ============================================================================
# CODING AGENTS
# ============================================================================

agent "architect" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.4

  instruction: ```
    You are a software architect who designs solutions before implementation.

    Your approach:
    1. Understand requirements fully before designing
    2. Consider existing code patterns and conventions
    3. Plan for testability from the start
    4. Design for extensibility but not over-engineering
    5. Document your design decisions

    When designing:
    - Identify affected files and dependencies
    - Consider edge cases and error scenarios
    - Plan the test strategy
    - Estimate complexity and risks

    Output a structured design document:
    ```json
    {
      "summary": "Brief description of the solution",
      "affected_files": ["list of files to modify or create"],
      "new_dependencies": ["any packages to install"],
      "design": {
        "components": [...],
        "interactions": [...],
        "data_flow": [...]
      },
      "test_strategy": {
        "unit_tests": [...],
        "integration_tests": [...],
        "edge_cases": [...]
      },
      "implementation_order": ["ordered list of tasks"],
      "risks": ["potential issues and mitigations"]
    }
    ```
  ```

  tools: [
    tool("read_file"),
    tool("list_dir"),
    tool("search_files"),
  ]

  scripts: [
    script("analyze-code-structure"),
    script("find-related-files")
  ]
}

agent "test-writer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("coding-standards") + ```

    You write tests BEFORE implementation (TDD).

    Test writing process:
    1. Understand the requirement
    2. Write the simplest failing test
    3. Only test public behavior, not implementation details
    4. Use descriptive test names
    5. Follow the AAA pattern

    Test categories:
    - Happy path: Normal successful execution
    - Edge cases: Boundary conditions
    - Error cases: Invalid inputs, failure scenarios
    - Integration: Component interactions

    Use the test template when creating new test files.
    Make tests deterministic - no flaky tests.
    Mock external dependencies appropriately.
  ```

  tools: [
    tool("read_file"),
    tool("write_file"),
    tool("run_tests"),
  ]
}

agent "implementer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("coding-standards") + ```

    You implement code to make tests pass.

    Implementation approach:
    1. Write the MINIMUM code to make the test pass
    2. Follow existing patterns in the codebase
    3. Keep functions small and focused
    4. Handle errors appropriately
    5. Add necessary type annotations

    After writing code:
    - Run the tests to verify
    - Check types with TypeScript compiler
    - Ensure linter passes

    Do NOT:
    - Add features not covered by tests
    - Refactor while implementing (that's a separate step)
    - Leave TODO comments for later
    - Ignore compiler or linter warnings
  ```

  tools: [
    tool("read_file"),
    tool("write_file"),
    tool("run_tests"),
    tool("run_typecheck"),
    tool("run_linter"),
  ]
}

agent "debugger" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You debug failing tests and code issues.

    Debugging approach:
    1. Understand the failure message completely
    2. Identify what was expected vs what happened
    3. Trace the data flow
    4. Form a hypothesis
    5. Verify with targeted logging or assertions
    6. Fix the root cause, not symptoms

    When analyzing failures:
    - Read the full stack trace
    - Check the test setup
    - Verify mocks are configured correctly
    - Look for off-by-one errors
    - Check for async timing issues

    Output your diagnosis:
    ```json
    {
      "failure_summary": "What's failing",
      "root_cause": "Why it's failing",
      "fix": "What needs to change",
      "files_to_modify": ["list of files"],
      "confidence": "high/medium/low"
    }
    ```
  ```

  tools: [
    tool("read_file"),
    tool("run_tests"),
  ]

  scripts: [
    script("parse-test-output"),
    script("analyze-code-structure")
  ]
}

agent "refactorer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("coding-standards") + ```

    You refactor code while maintaining behavior.

    Refactoring principles:
    1. NEVER change behavior while refactoring
    2. Run tests after EVERY change
    3. Make small, incremental changes
    4. One refactoring at a time

    Common refactorings:
    - Extract function/method
    - Extract class
    - Rename for clarity
    - Simplify conditionals
    - Remove duplication
    - Improve type safety

    Before refactoring:
    - Ensure tests pass
    - Understand the code fully

    After each change:
    - Run tests
    - Check types
    - Run linter
  ```

  tools: [
    tool("read_file"),
    tool("write_file"),
    tool("run_tests"),
    tool("run_typecheck"),
    tool("run_linter"),
  ]
}

agent "code-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.4

  instruction: file("coding-standards") + ```

    You review code for quality, correctness, and maintainability.

    Review checklist:
    1. **Correctness**: Does it work? Edge cases handled?
    2. **Tests**: Adequate coverage? Testing behavior?
    3. **Readability**: Clear names? Good structure?
    4. **Performance**: Obvious issues? N+1 queries?
    5. **Security**: Input validation? Injection risks?
    6. **Types**: Proper typing? Any 'any' usage?
    7. **Error handling**: Appropriate? Informative messages?

    Provide actionable feedback:
    - Specific line references
    - Suggested alternatives
    - Severity: blocker/major/minor/suggestion

    Output:
    ```json
    {
      "approval": "approve/request-changes/comment",
      "summary": "Overall assessment",
      "issues": [
        {
          "file": "path",
          "line": 42,
          "severity": "major",
          "issue": "Description",
          "suggestion": "How to fix"
        }
      ],
      "positives": ["Good things about the code"]
    }
    ```
  ```

  tools: [
    tool("read_file"),
    tool("git_diff"),
  ]
}

# ============================================================================
# TDD PIPELINE
# ============================================================================

pipeline "implement-feature" {
  # Step 1: Design the solution
  step "design" {
    use: agent("architect")

    input: {
      requirement: $input.requirement,
      context: $input.context
    }
  }

  # Step 2: Create a feature branch
  step "branch" {
    tools: [tool("git_branch")]

    input: {
      name: config.git.branch_prefix + $input.feature_name
    }
  }

  # Step 3: Write tests first (Red phase)
  step "write-tests" {
    use: agent("test-writer")

    input: {
      design: step("design").output,
      template: file("test-template")
    }
  }

  # Step 4: Verify tests fail
  step "verify-red" {
    tools: [tool("run_tests")]
    input: { pattern: step("write-tests").output.test_file }
  }

  # Step 5: Implementation loop (Green phase)
  loop max: 5 {
    step "implement" {
      use: agent("implementer")

      input: {
        design: step("design").output,
        tests: step("write-tests").output,
        previous_attempt: $previous_implementation
      }
    }

    step "test" {
      tools: [tool("run_tests")]
      input: { pattern: step("write-tests").output.test_file }
    }

    break_if: step("test").output.all_passing == true

    # Debug if tests fail
    step "debug" {
      use: agent("debugger")

      input: {
        test_output: step("test").output,
        implementation: step("implement").output
      }
    }

    set $previous_implementation: step("debug").output.fix
  }

  # Step 6: Refactor (Refactor phase)
  step "refactor" {
    use: agent("refactorer")

    input: {
      files: step("implement").output.files_modified,
      design: step("design").output
    }
  }

  # Step 7: Final verification
  step "verify" {
    parallel {
      step "all-tests" {
        tools: [tool("run_tests")]
        input: { coverage: true }
      }

      step "typecheck" {
        tools: [tool("run_typecheck")]
      }

      step "lint" {
        tools: [tool("run_linter")]
      }
    }
  }

  # Step 8: Code review
  step "review" {
    use: agent("code-reviewer")

    input: {
      diff: tool("git_diff").output
    }
  }

  # Step 9: Commit if approved
  branch step("review").output.approval {
    "approve" => step "commit" {
      tools: [tool("git_commit")]

      input: {
        message: "feat(" + $input.feature_name + "): " + $input.requirement,
        files: step("implement").output.files_modified
      }
    }

    "request-changes" => step "report-issues" {
      output: step("review").output.issues
    }
  }

  output: {
    success: step("review").output.approval == "approve",
    design: step("design").output,
    tests: step("write-tests").output,
    implementation: step("implement").output,
    review: step("review").output,
    commit: $branch.output
  }
}

# Fix bug pipeline
pipeline "fix-bug" {
  # Step 1: Understand the bug
  step "analyze" {
    use: agent("debugger")

    input: {
      description: $input.bug_description,
      reproduction: $input.reproduction_steps
    }
  }

  # Step 2: Write a failing test that reproduces the bug
  step "write-repro-test" {
    use: agent("test-writer")

    input: {
      bug: step("analyze").output,
      test_type: "regression"
    }
  }

  # Step 3: Verify the test fails (confirms bug exists)
  step "verify-bug" {
    tools: [tool("run_tests")]
    input: { pattern: step("write-repro-test").output.test_file }
  }

  # Step 4: Fix the bug
  step "fix" {
    use: agent("implementer")

    input: {
      diagnosis: step("analyze").output,
      test: step("write-repro-test").output
    }
  }

  # Step 5: Verify fix
  step "verify-fix" {
    tools: [tool("run_tests")]
  }

  # Step 6: Ensure no regressions
  step "full-test" {
    tools: [tool("run_tests")]
    input: { coverage: true }
  }

  # Step 7: Commit
  step "commit" {
    tools: [tool("git_commit")]

    input: {
      message: "fix: " + $input.bug_description
    }
  }

  output: {
    diagnosis: step("analyze").output,
    fix: step("fix").output,
    test_results: step("full-test").output
  }
}

# Refactor pipeline
pipeline "refactor-code" {
  # Step 1: Analyze current code
  step "analyze" {
    execute: script("analyze-code-structure") {
      file_path: $input.file
    }
  }

  # Step 2: Find related files and tests
  step "find-related" {
    execute: script("find-related-files") {
      file_path: $input.file
      project_root: env("PROJECT_ROOT")
    }
  }

  # Step 3: Ensure tests pass before refactoring
  step "baseline-tests" {
    tools: [tool("run_tests")]
  }

  # Step 4: Plan refactoring
  step "plan" {
    use: agent("architect")

    input: {
      file: $input.file,
      analysis: step("analyze").output,
      related: step("find-related").output,
      goal: $input.refactoring_goal
    }
  }

  # Step 5: Incremental refactoring loop
  loop max: 10 {
    step "refactor-step" {
      use: agent("refactorer")

      input: {
        plan: step("plan").output,
        current_step: $step_number,
        previous_changes: $changes
      }
    }

    step "test-step" {
      tools: [tool("run_tests")]
    }

    break_if: step("refactor-step").output.complete == true

    step "update-state" {
      set $changes: step("refactor-step").output.changes
      set $step_number: $step_number + 1
    }
  }

  # Step 6: Final verification
  step "final-verify" {
    parallel {
      step "tests" {
        tools: [tool("run_tests")]
        input: { coverage: true }
      }
      step "types" {
        tools: [tool("run_typecheck")]
      }
      step "lint" {
        tools: [tool("run_linter")]
      }
    }
  }

  # Step 7: Review changes
  step "review" {
    use: agent("code-reviewer")
    input: { diff: tool("git_diff").output }
  }

  output: {
    changes: step("refactor-step").output.all_changes,
    verification: step("final-verify").output,
    review: step("review").output
  }
}

# ============================================================================
# TRIGGERS
# ============================================================================

# GitHub issue trigger
trigger "new-issue" {
  event: github.issue.opened

  filter: github.issue.labels contains "ai-assist"

  run: {
    # Analyze the issue
    analysis: agent("architect") {
      input: {
        title: github.issue.title,
        body: github.issue.body,
        labels: github.issue.labels
      }
    }

    # Determine if it's a bug or feature
    branch analysis.output.type {
      "bug" => pipeline("fix-bug") {
        input: {
          bug_description: github.issue.title,
          reproduction_steps: analysis.output.reproduction
        }
      }

      "feature" => pipeline("implement-feature") {
        input: {
          requirement: github.issue.body,
          feature_name: analysis.output.feature_name
        }
      }
    }
  }

  on_complete: {
    github.issue.comment(
      body: "I've analyzed this issue and created a solution.\n\n" +
            "**Summary:** " + output.design.summary + "\n\n" +
            "**Changes:** " + output.implementation.files_modified.join(", ") + "\n\n" +
            "A PR will be created shortly."
    )
  }
}

# PR review trigger
trigger "pr-opened" {
  event: github.pull_request.opened

  run: agent("code-reviewer") {
    input: {
      diff: github.pull_request.diff,
      title: github.pull_request.title,
      description: github.pull_request.body
    }
  }

  on_complete: {
    github.pull_request.review(
      event: output.approval == "approve" ? "APPROVE" : "REQUEST_CHANGES",
      body: output.summary,
      comments: output.issues.map(i => ({
        path: i.file,
        line: i.line,
        body: "**" + i.severity + "**: " + i.issue + "\n\n" + i.suggestion
      }))
    )
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

intent "implement" {
  params: {
    requirement: string required "What to implement"
    feature_name: string required "Feature branch name"
  }

  run: pipeline("implement-feature") {
    input: params
  }

  output: stdout
}

intent "fix" {
  params: {
    bug: string required "Bug description"
    steps: string optional "Reproduction steps"
  }

  run: pipeline("fix-bug") {
    input: {
      bug_description: params.bug,
      reproduction_steps: params.steps
    }
  }

  output: stdout
}

intent "refactor" {
  params: {
    file: string required "File to refactor"
    goal: string required "Refactoring goal"
  }

  run: pipeline("refactor-code") {
    input: params
  }

  output: stdout
}

intent "review" {
  params: {
    diff: string optional "Git diff or 'staged'"
  }

  use: agent("code-reviewer")

  input: {
    diff: params.diff == "staged" ? tool("git_diff").output : params.diff
  }

  output: stdout
}

intent "test" {
  params: {
    pattern: string optional "Test pattern"
    coverage: bool optional false
  }

  tools: [tool("run_tests")]
  input: params

  output: stdout
}

intent "analyze" {
  params: {
    file: string required "File to analyze"
  }

  run: script("analyze-code-structure") {
    file_path: params.file
  }

  output: stdout
}
