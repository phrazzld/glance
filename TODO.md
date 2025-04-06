# TODO

## Project Setup and Testing
- [x] **Setup Testing Framework**
  - **Action:** Add testify dependencies to go.mod, set up initial test structure with basic CLI execution tests
  - **Depends On:** None
  - **AC Ref:** RFT-01

- [x] **Create Main Test File**
  - **Action:** Create main_test.go with end-to-end test that verifies GLANCE.md creation with test directory
  - **Depends On:** Setup Testing Framework
  - **AC Ref:** RFT-01

## Configuration and Main Function Refactoring
- [x] **Create Config Package and Struct**
  - **Action:** Create config/config.go with Config struct for app settings (APIKey, TargetDir, Force, Verbose, PromptTemplate)
  - **Depends On:** None
  - **AC Ref:** RFT-02

- [x] **Implement LoadConfig Function**
  - **Action:** Move flag parsing, env var loading, and prompt template loading to LoadConfig function in config package
  - **Depends On:** Create Config Package and Struct
  - **AC Ref:** RFT-02

- [x] **Refactor Main Function**
  - **Action:** Shorten main.go by delegating setup, scanning, processing, and reporting to dedicated functions
  - **Depends On:** Create Config Package and Struct, Implement LoadConfig Function
  - **AC Ref:** RFT-03

## Filesystem Package
- [x] **Create Filesystem Scanner**
  - **Action:** Create filesystem/scanner.go with directory traversal functions (ListDirs, loadGitignore, etc.)
  - **Depends On:** None
  - **AC Ref:** RFT-04

- [x] **Create Filesystem Reader**
  - **Action:** Create filesystem/reader.go with file reading and text detection functions
  - **Depends On:** None
  - **AC Ref:** RFT-04

- [x] **Centralize Ignore Logic**
  - **Action:** Create filesystem/ignore.go with ShouldIgnore function for consistent file/directory skipping
  - **Depends On:** Create Filesystem Scanner
  - **AC Ref:** RFT-08

- [x] **Create Filesystem Utilities**
  - **Action:** Create filesystem/utils.go with helper functions like latestModTime and shouldRegenerate
  - **Depends On:** None
  - **AC Ref:** RFT-04

## LLM Package
- [x] **Define LLM Client Interface**
  - **Action:** Create llm/client.go with Client interface and GeminiClient implementation
  - **Depends On:** None
  - **AC Ref:** RFT-05

- [x] **Create Prompt Generation Module**
  - **Action:** Create llm/prompt.go to handle prompt template handling and generation
  - **Depends On:** None
  - **AC Ref:** RFT-05

- [x] **Refactor API Interaction**
  - **Action:** Move Gemini API interaction code to the llm package and implement interface methods
  - **Depends On:** Define LLM Client Interface
  - **AC Ref:** RFT-05

## UI Package
- [x] **Create UI Feedback Module**
  - **Action:** Create ui/feedback.go with functions for spinner and progress bar management
  - **Depends On:** None
  - **AC Ref:** RFT-06

## Error Handling
- [x] **Define Custom Error Types**
  - **Action:** Create custom error types (e.g., APIError) for more specific error handling
  - **Depends On:** None
  - **AC Ref:** RFT-07

- [x] **Implement Error Wrapping**
  - **Action:** Update error handling throughout codebase to use fmt.Errorf with %w for proper error wrapping
  - **Depends On:** Define Custom Error Types
  - **AC Ref:** RFT-07

## Testing Implementation
- [ ] **Create Filesystem Unit Tests**
  - **Action:** Add comprehensive tests for filesystem package functions using mocks and test directories
  - **Depends On:** Create Filesystem Scanner, Create Filesystem Reader, Create Filesystem Utilities
  - **AC Ref:** RFT-09

- [ ] **Create LLM Unit Tests**
  - **Action:** Add tests for LLM package with mocked API client to verify prompt generation and retry logic
  - **Depends On:** Define LLM Client Interface, Create Prompt Generation Module
  - **AC Ref:** RFT-09

- [ ] **Create Config Unit Tests**
  - **Action:** Add tests for config package to verify correct loading of settings from various sources
  - **Depends On:** Create Config Package and Struct, Implement LoadConfig Function
  - **AC Ref:** RFT-09

- [ ] **Create UI Unit Tests**
  - **Action:** Add tests for UI package functionality
  - **Depends On:** Create UI Feedback Module
  - **AC Ref:** RFT-09

- [ ] **Implement Integration Tests**
  - **Action:** Create tests in main_test.go that verify interaction between packages using test directories
  - **Depends On:** Create Filesystem Unit Tests, Create LLM Unit Tests, Create Config Unit Tests
  - **AC Ref:** RFT-10

## Finalization
- [ ] **Code Cleanup and Review**
  - **Action:** Review naming, comments, constants, and formatting with gofmt and golint
  - **Depends On:** All implementation tasks
  - **AC Ref:** RFT-11

- [ ] **Update Documentation**
  - **Action:** Update README.md and CLAUDE.md to reflect new structure and architecture
  - **Depends On:** All implementation tasks
  - **AC Ref:** RFT-12
