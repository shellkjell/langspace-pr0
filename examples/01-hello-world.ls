# LangSpace Hello World
# The simplest possible LangSpace file: a single agent that responds to input

agent "greeter" {
  model: "claude-sonnet-4-20250514"

  instruction: ```
    You are a friendly greeter. When someone introduces themselves,
    welcome them warmly and ask how you can help them today.
  ```
}

# Run this with: langspace run 01-hello-world.ls --input "Hi, I'm Alice"
