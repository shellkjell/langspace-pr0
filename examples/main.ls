import "researcher.ls"

pipeline "main" {
    step "search" {
        use: agent("researcher")
        input: "Explain quantum computing."
    }
}
