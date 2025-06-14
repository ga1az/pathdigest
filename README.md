# PathDigest

PathDigest is a command-line tool written in Go that analyzes Git repositories, local directories, or individual files and generates a structured, LLM-friendly text digest of their content. This digest can be easily used as context for Large Language Models (LLMs) in tools like Cursor, VSCode, Gemini, and others.

Inspired by web-based tools like [gitingest](https://github.com/cyclotruc/gitingest) by @cyclotruc, PathDigest brings similar powerful code-to-context capabilities directly to your terminal as a native, efficient binary.

> I made this tool for my own use, but I thought it might be useful for others.

## Installation

### Via `go install` (Recommended if you have Go installed)

Make sure you have Go (1.20 or higher recommended) installed and `$HOME/go/bin` (or `$GOPATH/bin`) is in your `PATH`.

```bash
go install github.com/ga1az/pathdigest@latest
```

### Via `install.sh` (Recommended if you don't have Go installed)

```bash
curl -sSfL https://raw.githubusercontent.com/ga1az/pathdigest/main/install.sh | sh -s

# Install to a specific directory
curl -sSfL https://raw.githubusercontent.com/ga1az/pathdigest/main/install.sh | sh -s -- -b /usr/local/bin
```

## Usage

```bash
pathdigest <source>

//Git URL
pathdigest https://github.com/ga1az/pathdigest

//Local Directory
pathdigest ./projects/pathdigest -o pathdigest_digest.txt
```

## Features

*   **Versatile Source Input**: Process Git repository URLs (cloning specific branches/commits if needed), local directories, or single files.
*   **Structured Output**: Generates a clear text output including:
    *   A summary of the processed source.
    *   A tree-like representation of the directory structure.
    *   Concatenated content of all processed text files.
*   **Customizable Filtering**:
    *   Use glob patterns to exclude specific files or directories (e.g., `*.log`, `node_modules/`).
    *   Use glob patterns to explicitly include files or directories, overriding exclusion rules.
*   **File Size Control**: Set a maximum file size to skip very large files, keeping the context manageable.
*   **Git Integration**:
    *   Specify a branch or commit hash when providing a Git URL.
    *   Handles different Git URL formats (HTTPS, SSH, slugs).
*   **Go-Powered**: Built with Go for fast performance and easy cross-platform binary creation.


Output will be a json file (In progress) or txt file that can be used as a context for cursor, vscode, gemini, etc.

### Flags

*   `-o, --output <file_path>`: Specifies the output file path for the digest. Defaults to `pathdigest_digest.txt`. Use `-` to print to standard output.
*   `-s, --max-size <bytes>`: Maximum file size in bytes to process (e.g., `10485760` for 10MB). Defaults to 10MB.
*   `-e, --exclude-pattern <pattern>`: Glob pattern to exclude files/directories. Can be used multiple times. Adds to a list of default exclude patterns (e.g., `.git/`, `node_modules/`).
*   `-i, --include-pattern <pattern>`: Glob pattern to include files/directories. Can be used multiple times. If include patterns are provided, only files matching these patterns will be processed (overriding excludes if a path matches both).
*   `-b, --branch <branch_name>`: Specifies the branch to clone and ingest if the source is a Git URL. Can also be a commit hash.

To see all available flags and default values:
```bash
pathdigest --help
```

### Build Binary

```bash
go build ./cmd/pathdigest
```

