# Create UI Feedback Module

## Task Goal
Create ui/feedback.go with functions for spinner and progress bar management to centralize and standardize UI feedback components across the application. This will improve code organization and make UI feedback elements consistent, reusable, and easier to maintain.

## Implementation Approach
I'll create a new ui package with a feedback.go file that encapsulates the spinner and progress bar functionality. The implementation will include:

1. **Spinner Management**:
   - Create a `Spinner` struct that wraps the underlying briandowns/spinner library
   - Add methods for common operations (Start, Stop, Update)
   - Provide factory functions with standardized configurations
   - Include helper methods for common use cases with sensible defaults

2. **Progress Bar Management**:
   - Create a `ProgressBar` struct that wraps the schollz/progressbar/v3 library
   - Add methods for initialization, incrementation, and completion
   - Support customization of bar appearance through options
   - Standardize progress bar theme and style for consistent UI

3. **Factory Functions**:
   - `NewScanner()` - Create a spinner specifically for directory scanning
   - `NewProcessor()` - Create a progress bar for processing tasks
   - `NewCustomSpinner(options)` - Create a customizable spinner
   - `NewCustomProgressBar(total, options)` - Create a customizable progress bar

4. **Supporting Types**:
   - Define option types for configuring spinners and progress bars

## Key Reasoning
This approach is best because:

1. **Abstraction**: By wrapping the third-party libraries, we can provide a simpler, application-specific interface that hides implementation details.

2. **Consistency**: Centralizing UI feedback ensures consistent styling and behavior across the application, improving user experience.

3. **Maintainability**: If we need to change the underlying libraries or modify the appearance of UI elements, changes will be localized to this module.

4. **Reduced Duplication**: The current code duplicates spinner and progress bar configuration in multiple places. This module centralizes that configuration.

5. **Testability**: By abstracting UI feedback into a dedicated module, we can more easily mock these components in tests.

This approach follows the established pattern of module creation seen in the filesystem and llm packages, where functionality is grouped by purpose and abstracted behind clean interfaces.