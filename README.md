# Glance README

## What is it?

Glance is a command-line tool that recursively scans a directory tree and generates a `glance.md` file in each directory. This file provides a high-level summary of the directory's contents, generated using Google's Generative AI API (Gemini).

## Quick Usage

1. **Build or Run:**
   Clone the repository and run the tool with:

       go build -o glance && ./glance [--force] [--verbose] /path/to/directory

   or run directly:

       go run glance.go [--force] [--verbose] /path/to/directory

2. **Set Up Environment:**
   Ensure you have a valid `GEMINI_API_KEY` set in your environment or in a `.env` file.

3. **Flags:**
   - `--force` will regenerate `glance.md` even if it already exists.
   - `--verbose` enables detailed logging output.

## Environment Variables

- **GEMINI_API_KEY:**
  Your Google Generative AI API key. This must be valid for the Gemini calls to succeed.

## .env File

Optionally, create a `.env` file in the same directory as the tool to automatically load your environment variables. For example:

       GEMINI_API_KEY=your_api_key_here

If the `.env` file is absent, Glance will fall back to your system's environment variables.

## What Does It Skip?

- **Hidden Files and Directories:**
  Glance ignores hidden directories (e.g., `.git`) and dotfiles.

- **.gitignore Matches:**
  Files or directories that are listed in a local `.gitignore` are not processed.

- **Existing `glance.md` Files:**
  It won’t overwrite an existing `glance.md` unless you use the `--force` flag.

- **Large Files:**
  Files larger than approximately 5MB are truncated to keep the prompt size manageable.

- **Invalid UTF-8:**
  Any invalid UTF-8 in file contents is sanitized before sending data to the API.

## Logging

Glance uses [logrus](https://github.com/sirupsen/logrus) for logging:
- **Info-level Logging:** Default logging of key actions.
- **Debug-level Logging:** Enabled with the `--verbose` flag for more detailed output.
- Additionally, it features a spinner and a progress bar during scanning and generation.

## Dependencies

- [github.com/briandowns/spinner](https://github.com/briandowns/spinner) – Spinner animation.
- [github.com/google/generative-ai-go/genai](https://github.com/google/generative-ai-go) – Gemini API client.
- [github.com/joho/godotenv](https://github.com/joho/godotenv) – Loads environment variables from a `.env` file.
- [github.com/sabhiram/go-gitignore](https://github.com/sabhiram/go-gitignore) – Parses `.gitignore` files.
- [github.com/schollz/progressbar/v3](https://github.com/schollz/progressbar) – Displays a progress bar.
- [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) – Provides structured logging.

## License

This repository is provided under an open license (e.g., Apache-2.0). Please refer to the LICENSE file for details.

## Enjoy!

Use Glance to quickly generate summaries of your projects and make your directories easier to understand. Happy coding!
