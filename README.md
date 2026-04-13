# PathDigest

PathDigest is a command-line tool written in Go that analyzes Git repositories, local directories, or individual files and generates a structured, LLM-friendly text digest of their content. The digest can be used as context for Large Language Models (LLMs) in tools like Cursor, VSCode, Gemini, and others.

Inspired by web-based tools like [gitingest](https://github.com/cyclotruc/gitingest) by @cyclotruc, PathDigest brings similar powerful code-to-context capabilities directly to your terminal as a native, efficient binary.

> I made this tool for my own use, but I thought it might be useful to others.

## Installation

### Via `go install` (requires Go 1.23+)

Make sure `$HOME/go/bin` (or `$GOPATH/bin`) is in your `PATH`.

```bash
go install github.com/ga1az/pathdigest@latest
```

### Via install script (macOS & Linux)

```bash
# Install to default directory ($GOPATH/bin or ./bin)
curl -sSfL https://raw.githubusercontent.com/ga1az/pathdigest/main/install.sh | sh -s

# Install to a specific directory
curl -sSfL https://raw.githubusercontent.com/ga1az/pathdigest/main/install.sh | sh -s -- -b /usr/local/bin

# Install a specific version
curl -sSfL https://raw.githubusercontent.com/ga1az/pathdigest/main/install.sh | sh -s -- v0.2.0
```

### Via GitHub Releases

Download the binary for your platform from [github.com/ga1az/pathdigest/releases](https://github.com/ga1az/pathdigest/releases).

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/ga1az/pathdigest/releases/latest/download/pathdigest_0.0.0_darwin_arm64.tar.gz | tar xz
chmod +x pathdigest
mv pathdigest /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/ga1az/pathdigest/releases/latest/download/pathdigest_0.0.0_darwin_amd64.tar.gz | tar xz
chmod +x pathdigest
mv pathdigest /usr/local/bin/

# Linux (amd64)
curl -sL https://github.com/ga1az/pathdigest/releases/latest/download/pathdigest_0.0.0_linux_amd64.tar.gz | tar xz
chmod +x pathdigest
mv pathdigest /usr/local/bin/

# Linux (arm64)
curl -sL https://github.com/ga1az/pathdigest/releases/latest/download/pathdigest_0.0.0_linux_arm64.tar.gz | tar xz
chmod +x pathdigest
mv pathdigest /usr/local/bin/
```

## Usage

### Basic

```bash
# Digest a local directory (outputs to pathdigest_digest.txt by default)
pathdigest ./my-project

# Digest a Git repository URL
pathdigest https://github.com/ga1az/pathdigest

# Output to stdout (useful for piping)
pathdigest ./my-project -o -

# Output to a specific file
pathdigest ./my-project -o digest.txt
```

### JSON Output

Use `--format json` (or `-f json`) for structured output that tools and scripts can consume:

```bash
# JSON to file
pathdigest ./my-project -f json -o digest.json

# JSON to stdout
pathdigest ./my-project -f json -o -

# Default text format (unchanged behavior)
pathdigest ./my-project
```

The JSON output includes:
- `summary` — source path, file count, total size, patterns used
- `tree` — nested directory structure with names, paths, types, sizes
- `files` — flat array of all processed files with content
- `git_info` — repository metadata when processing a Git URL

### Filtering

```bash
# Exclude additional patterns (adds to built-in defaults)
pathdigest ./my-project -e "*.test.ts" -e "coverage/"

# Include only specific patterns (overrides excludes)
pathdigest ./my-project -i "*.go" -i "go.mod"

# Limit max file size (default: 10MB)
pathdigest ./my-project -s 1048576  # 1MB limit
```

### Git Integration

```bash
# Clone and digest a specific branch
pathdigest https://github.com/user/repo -b develop

# Digest a specific path within a repo
pathdigest https://github.com/user/repo/tree/main/internal/pkg
```

### All Flags

```
Flags:
  -b, --branch string             Branch to clone and ingest (if source is a Git URL)
  -e, --exclude-pattern strings   Glob patterns to exclude (adds to defaults)
  -f, --format string             Output format: text or json (default "text")
  -h, --help                      Help for pathdigest
  -i, --include-pattern strings   Glob patterns to include (overrides excludes)
  -s, --max-size int              Maximum file size in bytes (default 10485760)
  -o, --output string             Output file path (default "pathdigest_digest.txt")
```

## Shell Completions

PathDigest supports autocompletion for **bash**, **zsh**, and **fish**.

### Bash

```bash
# Generate and load completions
pathdigest completion bash > ~/.config/pathdigest/completion.bash
echo 'source ~/.config/pathdigest/completion.bash' >> ~/.bashrc

# Or if you use bash-completion (recommended)
pathdigest completion bash > $(brew --prefix)/etc/bash_completion.d/pathdigest  # macOS
pathdigest completion bash > /etc/bash_completion.d/pathdigest                  # Linux
```

### Zsh

```bash
# Generate completions to your fpath
pathdigest completion zsh > "${fpath[1]}/_pathdigest"

# Or manually:
mkdir -p ~/.zfunc
pathdigest completion zsh > ~/.zfunc/_pathdigest

# Make sure ~/.zfunc is in your fpath in ~/.zshrc:
# fpath+=~/.zfunc
# autoload -U compinit && compinit
```

### Fish

```bash
pathdigest completion fish > ~/.config/fish/completions/pathdigest.fish
```

## Features

- **Versatile Source Input** — Process Git repository URLs (cloning specific branches/commits), local directories, or single files.
- **Text & JSON Output** — Default text format for human consumption, JSON format (`-f json`) for tools and scripts.
- **Smart Filtering** — Built-in exclude patterns for common noise (`.git/`, `node_modules/`, `build/`, etc.) plus custom glob patterns.
- **Git Integration** — Specify branches, commits, and sub-paths when providing a Git URL.
- **File Size Control** — Set a maximum file size to skip very large files.
- **Cross-Platform** — Pre-built binaries for macOS, Linux, and Windows (amd64 and arm64).
- **Shell Completions** — Autocompletion for bash, zsh, and fish.

## Build from Source

```bash
git clone https://github.com/ga1az/pathdigest.git
cd pathdigest
go build ./cmd/pathdigest
```
