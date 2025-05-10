# Go Coding Standards - Ultima SDK Port

This document outlines the coding standards and best practices to follow while porting the C# Ultima SDK to Go.

## 1. Project Goal & Philosophy

- **Public API:** Strive for an idiomatic Go public API. It should feel natural for Go developers to use, not like a direct C# translation.
- **Internal Implementation:** The internal logic (especially within `internal/` packages) should closely mimic the structure and algorithms of the C# reference implementation (`csharp/`) to ensure correctness and simplify verification during the porting process.
- **Correctness First:** The primary goal is a correct and verified implementation against the original C# SDK's behavior.

## 2. Go Version

- Target **Go 1.22+** to leverage features like generics and iterators where appropriate.

## 3. Package Structure

- **Root Package:** Contains the primary public API types and functions for interacting with the SDK.
- **`internal/` Packages:** Contain implementation details not meant for direct external use (e.g., `internal/binary` for low-level reading, `internal/uop` for UOP file handling). This enforces encapsulation.

## 4. Instantiation and State Management

- **No Global State/Singletons:** Avoid global variables for SDK state or file handles. Do not rely on implicit `init()` functions for setup.
- **Explicit Instantiation:** Use an explicit function, like `Open(directory string) (*SDK, error)`, to initialize and return an SDK instance. This instance will hold necessary state (like file paths or handles) and provide methods for accessing data (`sdk.GetArt()`, `sdk.GetMap()`, etc.). This improves testability and makes dependencies clear.

## 5. Dependency Management

- **Go Modules:** Use Go modules for managing dependencies. Initialize with `go mod init` and keep dependencies tidy with `go mod tidy`.

## 6. Error Handling

- **Idiomatic Go Errors:** Use Go's standard error handling mechanism. Functions that can fail must return an `error` as their last return value.
- **No Panics:** Do not use `panic` for recoverable errors. Panics should only be used for unrecoverable situations.
- **Error Wrapping:** Use `fmt.Errorf` with the `%w` verb or the `errors` package (Go 1.13+) to wrap errors when adding context, allowing callers to inspect the underlying error cause if needed.
- **Clear Error Messages:** Provide informative error messages.

## 7. API Design

- **Idiomatic Naming:** Follow Go naming conventions (CamelCase for exported identifiers, camelCase for unexported). Keep names short and descriptive.
- **Interfaces:** Use interfaces thoughtfully to define behavior contracts, especially where multiple implementations might exist or for decoupling. Prefer smaller, focused interfaces.
- **Return Values:** Return values directly, avoiding pointer parameters for returning data unless necessary (e.g., large structs, optional results).
- **Zero Values:** Ensure the zero value for structs is meaningful and usable whenever possible.
- **Simplicity:** Favor simple, clear designs over overly complex abstractions.

## 8. Image Handling

- **Standard `image` Package:** Use Go's standard `image` package as the foundation for image manipulation.
- **16-bit Pixel Format:** Carefully implement handling for the 16-bit ARGB1555 pixel format common in UO files. `image.RGBA64` might be adaptable, or direct byte manipulation within a custom `image.Image` implementation might be required.
- **Efficiency:** Be mindful of performance when reading and manipulating pixel data.

## 9. `unsafe` Code

- **Avoid If Possible:** Prioritize safe Go code.
- **Last Resort:** Use the `unsafe` package only when absolutely necessary for performance-critical sections (e.g., tight loops in image processing) or for interoperability that cannot be achieved otherwise.
- **Justification & Documentation:** If `unsafe` is used, clearly document _why_ it's necessary, what assumptions are being made, and ensure its usage is correct and localized.

## 10. Testing

- **`testify/assert`:** Use the `github.com/stretchr/testify/assert` package for assertions in unit tests.
- **Test Data:** Use the official test data located in `d:\Workspace\Go\src\github.com\kelindar\ultima-sdk-testdata\` for all tests requiring UO data files.
- **Coverage:** Aim for good test coverage, especially for core logic (file parsing, data structures, image handling).
- **Table-Driven Tests:** Use table-driven tests where appropriate to cover multiple input/output scenarios concisely.
- **Verification:** Compare test output (e.g., image dimensions, tile IDs, data values) against known values from the original C# SDK or verified sources.

## 11. Performance

- **Correctness First:** Focus on writing correct, clear, and safe code initially.
- **Optimize Judiciously:** Apply optimizations only where profiling indicates a significant benefit is needed.

## 12. Documentation

- **Go Doc Comments:** Write clear Go documentation comments (`//`) for all exported types, functions, constants, and variables. Explain _what_ they do and _how_ to use them.
- **Package Comments:** Provide package comments (`// package mypackage ...`) explaining the purpose of each package.
- **README.md:** Maintain a clear `README.md` for the module explaining its purpose, installation, and basic usage.

## 13. Concurrency

- **Assume Single-Threaded Use Initially:** Design the core library logic assuming single-threaded access unless explicitly stated otherwise.
- **Concurrency Safety:** If concurrency features are added later, ensure data structures accessed by multiple goroutines are properly protected using mutexes (`sync.Mutex`, `sync.RWMutex`) or other synchronization primitives. Document the concurrency safety guarantees of public APIs.

## 14. Logging

- **Avoid `fmt.Print*`:** Do not use `fmt.Println` or similar functions for logging within the library code.
- **Standard `log` (Optional):** If logging is deemed necessary within the library (use sparingly), consider using the standard `log` package behind an optional configuration flag or interface.
- **Return Errors:** Prefer returning errors to indicate problems rather than logging them directly. Let the calling application decide on the logging strategy.
