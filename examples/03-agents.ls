# LangSpace Agent Configuration
# Showing the full range of agent configuration options

# Minimal agent - just an instruction
agent "simple" {
  instruction: "You are helpful."
}

# Full-featured agent
agent "code-reviewer" {
  # Model configuration
  model: "claude-sonnet-4-20250514"
  temperature: 0.3
  max_tokens: 4096

  # The agent's core instruction
  instruction: ```
    You are a senior code reviewer with 15 years of experience.

    ## Your Review Process
    1. First, understand the overall purpose of the code
    2. Check for correctness and edge cases
    3. Evaluate code style and readability
    4. Look for security vulnerabilities
    5. Suggest performance optimizations

    ## Output Format
    Structure your review as:
    - **Summary**: One paragraph overview
    - **Issues**: Numbered list of problems found
    - **Suggestions**: Improvements that aren't bugs
    - **Verdict**: APPROVE, REQUEST_CHANGES, or NEEDS_DISCUSSION
  ```

  # Tools this agent can use
  tools: [
    read_file,
    search_codebase,
    run_tests,
    get_git_diff
  ]

  # System-level configuration
  system: {
    # Retry failed requests
    retry: 3

    # Timeout for LLM calls
    timeout: "120s"

    # Enable streaming output
    stream: true
  }
}

# Agent with tool-calling disabled (pure text generation)
agent "poet" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.9
  tools: []  # Explicitly no tools

  instruction: ```
    You are a whimsical poet. Write short poems about whatever
    topic is given to you. Use vivid imagery and unexpected metaphors.
  ```
}

# Agent that uses another agent's output format
agent "reviewer-v2" {
  extends: agent("code-reviewer")  # Inherit configuration

  # Override just the model
  model: "gpt-4o"

  # Add to the instruction
  instruction_append: ```

    Additionally, check for:
    - Proper logging practices
    - API documentation completeness
  ```
}
