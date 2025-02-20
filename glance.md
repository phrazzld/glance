The `/Users/phaedrus/Development/glance` directory contains a Go command-line tool named `Glance`.  Glance recursively scans a directory tree and generates a `glance.md` file in each directory summarizing its contents using Google's Gemini AI API.

**Architecture:** Glance uses a breadth-first search to traverse directories.  It processes directories from deepest to shallowest, ensuring that subdirectory summaries are available before parent directory summaries are generated.  It handles `.gitignore` files and skips hidden files and directories.  Large files are truncated.  Invalid UTF-8 is sanitized.

**Key File Roles:**

- `LICENSE`: Specifies the MIT open-source license.
- `README.md`: Documents the tool's purpose, usage, environment variables, dependencies, and limitations.
- `glance.go`: Contains the main Go source code for the Glance tool.
- `go.mod` and `go.sum`: Manage Go module dependencies and versioning.


**Dependencies:** The tool depends on several external packages: `github.com/briandowns/spinner`, `github.com/google/generative-ai-go/genai`, `github.com/joho/godotenv`, `github.com/sabhiram/go-gitignore`, `github.com/schollz/progressbar/v3`, `github.com/sirupsen/logrus`, and `google.golang.org/api`.

**Gotchas:**  A valid `GEMINI_API_KEY` environment variable or a `.env` file is required for operation.  The tool's performance is dependent on the Gemini API's response time and may be impacted by large directory trees or complex file structures.  Error handling is implemented with retries up to a maximum of 3 attempts.  Files larger than approximately 5MB are truncated before being sent to the Gemini API.
