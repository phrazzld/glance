# Glance - Claude Guide

## Build & Run Commands
- Build: `go build -o glance`
- Run: `./glance [--force] [--verbose] /path/to/directory`
- Direct run: `go run glance.go [--force] [--verbose] /path/to/directory`
- Flags: `--force` (regenerate existing GLANCE.md files), `--verbose` (detailed logging)

## Environment Setup
- Required: Set `GEMINI_API_KEY` in environment or `.env` file
- Optional: Create `.env` file in project root

## Code Style Guidelines
- Language: Go 1.23+ with proper error handling via returns
- Logging: Use logrus with appropriate levels (debug/info/warn)
- File Structure: Keep code organized by functionality with clear comments
- Naming: Use camelCase for vars/functions, PascalCase for exported items
- Error Handling: Always check errors and provide context in error messages
- Template Strings: Use text/template for prompt generation and configuration

## Architecture Notes
- Directory traversal: BFS approach with .gitignore awareness
- File processing: Skip binary/non-text files, limit large files to 5MB
- Retry mechanism: Multiple attempts for API calls with clear logging