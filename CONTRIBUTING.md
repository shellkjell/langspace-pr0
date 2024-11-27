# Contributing to LangSpace

We love your input! We want to make contributing to LangSpace as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Code Quality Standards

### Go Guidelines

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Document all exported functions, types, and constants
- Write meaningful test cases
- Maintain test coverage above 80%

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Example:
```
Add token pooling for improved memory efficiency

- Implement global token pool
- Add token release mechanism
- Update parser to use pooled tokens
- Add benchmarks for token pooling

Fixes #123
```

## Testing

Before submitting a pull request, ensure all tests pass:

```bash
make test          # Run all tests
make coverage      # Generate coverage report
make benchmark     # Run benchmarks
make lint          # Run linter
```

## Performance Guidelines

1. **Memory Management**
   - Use token pooling for frequently allocated objects
   - Minimize allocations in hot paths
   - Profile memory usage for large inputs

2. **Parsing Performance**
   - Keep parsing O(n) where possible
   - Minimize string copies
   - Use efficient data structures

3. **Error Handling**
   - Provide detailed error messages
   - Include line/column information
   - Make errors actionable

## Documentation

- Update README.md with any new features
- Document all exported symbols
- Include examples in godoc
- Update PERFORMANCE.md with benchmark changes
- Keep ROADMAP.md current

## License

By contributing, you agree that your contributions will be licensed under the GNU GPL v2 License.

## References

* [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* [Effective Go](https://golang.org/doc/effective_go.html)
* [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)
* [Semantic Versioning](https://semver.org/)
