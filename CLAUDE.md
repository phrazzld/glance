# Glance - Claude Guide

## Build & Run Commands
- Build: `go build -o glance`
- Run: `./glance [--force] [--verbose] /path/to/directory`
- Direct run: `go run . [--force] [--verbose] /path/to/directory`
- Flags: 
  - `--force` (regenerate existing GLANCE.md files)
  - `--verbose` (detailed logging)
  - `--prompt-file` (specify custom prompt template file)
- Testing: `go test ./...` (run all tests)

## Environment Setup
- Required: Set `GEMINI_API_KEY` in environment or `.env` file
- Optional: Create `.env` file in project root

## Code Style Guidelines
- Language: Go 1.23+ with proper error handling via returns
- Logging: Use logrus with appropriate levels (debug/info/warn)
- File Structure: Keep code organized by functionality with clear comments
- Naming: Use camelCase for vars/functions, PascalCase for exported items
- Error Handling: Use custom error types and wrap errors with context
- Template Strings: Use text/template for prompt generation and configuration

## Package Structure
- `config`: Configuration handling (flags, env vars, defaults)
- `errors`: Custom error types and error handling utilities
- `filesystem`: Directory scanning and file operations
- `llm`: LLM client interface and Gemini implementation
- `ui`: User interface components (spinners, progress bars)

## Architecture Notes
- Directory traversal: BFS approach with .gitignore awareness
- File processing: Skip binary/non-text files, limit large files to 5MB
- Retry mechanism: Multiple attempts for API calls with exponential backoff
- Error handling: Structured error types with context and suggestions