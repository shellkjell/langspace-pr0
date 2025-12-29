# LangSpace VSCode Extension

Syntax highlighting and language support for [LangSpace](https://github.com/shellkjell/langspace) files.

## Features

- Syntax highlighting for `.ls` files
- Bracket matching and auto-closing
- Comment toggling (`Cmd+/` or `Ctrl+/`)
- Code folding

## Installation

### From Source

1. Clone the repository
2. Open the `vscode-langspace` folder in VSCode
3. Press `F5` to launch Extension Development Host
4. Open a `.ls` file to see syntax highlighting

### Manual Installation

```bash
cd vscode-langspace
npm install -g @vscode/vsce
vsce package
code --install-extension langspace-0.1.0.vsix
```

## Supported Syntax

- **Entity types**: `agent`, `file`, `tool`, `intent`, `pipeline`, `step`, `trigger`, `config`, `mcp`, `script`
- **Control flow**: `branch`, `loop`, `break_if`, `parallel`
- **Types**: `string`, `number`, `bool`, `array`, `object`, `enum`
- **References**: `agent("name")`, `step("x").output`, `$input`
- **Multi-line strings**: Triple backticks with optional language tag
- **Comments**: `# single line comments`

## Example

```langspace
agent "reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3
  
  instruction: ```
    Review the code for best practices.
  ```
  
  tools: [read_file, search_codebase]
}
```

## License

GNU GPL v2
