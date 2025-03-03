# Glance Project Overview

This document provides a technical overview of the `glance` project, a command-line tool that generates directory summaries using Google's Gemini AI API.

## Purpose

Glance recursively scans a directory, generating a `GLANCE.md` file in each subdirectory.  These files provide concise, developer-oriented overviews of the directory's contents, leveraging the Gemini API for content generation.  This aids in project understanding and navigation, particularly in large or unfamiliar codebases.

## Structure and Architecture

The project's structure is generally well-organized. The main functionality resides in `glance.go`,  with supporting files for configuration (`prompt.txt`), licensing (LICENSE), and documentation (README.md, CLAUDE.md).  The architecture employs a breadth-first search (BFS) algorithm for directory traversal, efficiently handling nested structures.  `.gitignore` files are respected, preventing the processing of unwanted files and directories.


* **`glance.go`:** The core implementation, encompassing directory traversal, file processing, Gemini API interaction, and output generation.
* **`prompt.txt`:**  A template file for constructing prompts sent to the Gemini API.  This allows for customization of the generated summaries.  A fallback default prompt is included within the code.
* **`README.md`:**  Provides a user-oriented overview, installation instructions, and basic usage examples.
* **`CLAUDE.md`:** Supplements the README with build, run commands, and additional developer guidelines (code style and environment setup)
* **`LICENSE`:** Specifies the project's license (MIT).
* **`go.mod` and `go.sum`:** Manage project dependencies using Go modules.


## Subdirectory Contributions (N/A in this case)

The provided context does not reveal any subdirectories.  The functionality is self-contained within the main directory.

## Dependencies

The `go.mod` file clearly defines dependencies:

* **`github.com/google/generative-ai-go/genai`:** The Gemini API client library. Version `v0.19.0` is used.  This is a critical dependency and any updates should be carefully considered and tested.
* **`github.com/joho/godotenv`:** For loading environment variables from a `.env` file.
* **`github.com/sabhiram/go-gitignore`:**  For parsing `.gitignore` files.
* **`github.com/briandowns/spinner` and `github.com/schollz/progressbar/v3`:**  Provide user interface elements (spinner and progress bar) during the scan.
* **`github.com/sirupsen/logrus`:** A structured logging library.

The `go.sum` file ensures dependency integrity.

## Potential Pitfalls and Gotchas

* **API Key Management:** The reliance on the `GEMINI_API_KEY` environment variable introduces a security risk if not handled properly. Secure methods for managing API keys (e.g., dedicated secrets management solutions) should be implemented.
* **Gemini API Limits:** The Gemini API has rate limits and cost implications.  Error handling in `generateGlanceText` includes retry logic, but further strategies might be necessary for high-volume processing.  Consider implementing more sophisticated backoff strategies.
* **Large Files:** Files larger than 5MB are truncated. While this prevents excessively large prompts, it also limits the context available to Gemini, potentially impacting the accuracy of the generated summaries. A more sophisticated mechanism for handling large files (e.g., summarization before sending to the API) might improve results.
* **UTF-8 Handling:** The code sanitizes invalid UTF-8 characters. While helpful, this might lead to loss of information in severely corrupted files. Consider logging instances of such sanitization.
* **Error Handling:** Error handling is generally well-implemented, using explicit error returns and providing context in log messages. However,  more granular error classification and reporting (potentially using custom error types) could enhance debugging.

## Points to Consider

* **Custom Prompt Enhancements:** The `prompt.txt` file offers customization, but consider developing a more robust system for managing prompts, perhaps allowing users to specify different prompt templates for various file types or directory structures.
* **Testing:**  Adding comprehensive unit and integration tests would significantly improve the maintainability and reliability of the codebase.
* **Parallel Processing:** Consider using goroutines to parallelize the directory traversal and file processing to improve performance for very large directories.
* **Dependency Updates:** Regularly check for updates to the Gemini API client and other dependencies to ensure compatibility and leverage new features.


This overview aims to provide a comprehensive understanding of the `glance` project.  Addressing the points mentioned above will further enhance its robustness, performance, and usability.
