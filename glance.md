# Glance: Directory Overview

This directory contains the source code for `glance`, a command-line tool that recursively generates descriptive Markdown summaries of directories.  These summaries are produced using Google's Gemini large language model via the `generative-ai-go` library.

## Purpose

Glance aims to provide quick, high-level overviews of directory structures and their contents, assisting developers in understanding unfamiliar codebases or quickly reviewing project organization.

## Structure and Architecture

The project uses a breadth-first search (BFS) algorithm (`listAllDirsWithIgnores` function in `glance.go`) to traverse the directory tree.  It respects `.gitignore` files and skips hidden files and directories.  The core logic resides in `glance.go`, which handles command-line argument parsing, directory traversal, Gemini API interaction, and output generation. The `prompt.txt` file provides the template for the prompt sent to the Gemini API.

The tool employs several external libraries for functionalities such as progress display (`progressbar/v3`), spinner animation (`briandowns/spinner`), environment variable loading (`joho/godotenv`), `.gitignore` parsing (`sabhiram/go-gitignore`), and structured logging (`sirupsen/logrus`).  The `go.mod` file specifies these dependencies.  `go.sum` contains the checksums for these dependencies, ensuring reproducibility and security. The `LICENSE` file indicates the project's licensing terms (MIT License).  The `README.md` file serves as the primary documentation for the tool.

## Notable Dependencies

*   **`github.com/google/generative-ai-go`**:  The core dependency for interacting with the Google Gemini API.  The API key is expected to be provided via the `GEMINI_API_KEY` environment variable or a `.env` file.
*   **`github.com/sabhiram/go-gitignore`**: Used for parsing and respecting `.gitignore` files during directory traversal.
*   **`github.com/sirupsen/logrus`**: Provides structured logging capabilities.  Debug-level logging is enabled with the `--verbose` flag.
*   **`github.com/schollz/progressbar/v3`**: Displays a progress bar during the directory scanning and summary generation phases.


## Quirks and Gotchas

*   **API Key Requirement:** The tool requires a valid Gemini API key to function.  Failure to provide one will result in a fatal error.
*   **File Size Limits:** Files larger than 5MB are truncated to avoid exceeding the Gemini API's input size limitations.
*   **UTF-8 Handling:** The tool attempts to sanitize invalid UTF-8 sequences, replacing them with a replacement character ('�').
*   **Regeneration Control:** The `--force` flag is needed to regenerate existing `glance.md` files.  Otherwise, existing files are skipped.
*   **Error Handling:** While the code attempts to handle various errors, some edge cases (e.g., network issues during API calls) might not be fully covered. The `printDebrief` function provides a post-run summary of successes and failures.
*   **Prompt Template:** The prompt sent to the Gemini API is customizable using the `--prompt-file` flag.  A default prompt is available in `prompt.txt`.  The `defaultPrompt` variable in `glance.go` provides a fallback.

## Code Patterns

The codebase exhibits a relatively straightforward structure, with clear separation of concerns.  The use of channels or goroutines for concurrent processing is absent, leading to potentially suboptimal performance for very large directory trees. The BFS approach ensures that subdirectory summaries are available before processing parent directories.  Retry logic is implemented in `processDirWithRetry` to handle transient issues during Gemini API interactions.


## Summary

Glance provides a functional, though potentially improvable, solution for generating directory summaries using a large language model.  Its reliance on external dependencies is well-managed, and the code is reasonably well-documented.  However, performance optimization and more robust error handling might be areas for future improvement.
