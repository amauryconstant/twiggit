## 1. Type Definitions

- [x] 1.1 Define `matcherFunc func(error) bool` and `formatterFunc func(error) string` types in cmd/error_formatter.go

## 2. Matcher Functions

- [x] 2.1 Create `isValidationError(err error) bool` using errors.As()
- [x] 2.2 Create `isWorktreeError(err error) bool` using errors.As()
- [x] 2.3 Create `isProjectError(err error) bool` using errors.As()
- [x] 2.4 Create `isServiceError(err error) bool` using errors.As()

## 3. ErrorFormatter Refactoring

- [x] 3.1 Replace reflection map with matcher-formatter slice in ErrorFormatter struct
- [x] 3.2 Refactor `register` method to accept matcher+formatter pair
- [x] 3.3 Refactor `Format` method to iterate through matchers instead of reflection lookup
- [x] 3.4 Update formatter functions to accept error directly (remove ErrorFormatter receiver)

## 4. Testing

- [x] 4.1 Update tests in cmd/error_formatter_test.go to verify behavior unchanged
- [x] 4.2 Run `mise run test:full` to verify no regressions
