config {
  default_model: "claude-3-5-sonnet-20240620"
}

agent "writer" {
  instruction: "Write a short poem about space."
}

agent "critic" {
  instruction: "Provide a briefly critique of the poem."
}

pipeline "poetry-gen" {
  step "generate" {
    use: agent("writer")
  }
  step "critique" {
    use: agent("critic")
  }
}

intent "get-poetry" {
  use: pipeline("poetry-gen")
}
