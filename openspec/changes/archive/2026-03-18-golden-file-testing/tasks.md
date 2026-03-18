## 1. Infrastructure Setup

- [x] 1.1 Create test/helpers/golden.go with CompareGolden function
- [x] 1.2 Create test/golden/ directory structure (list/, errors/)

## 2. Core Implementation

- [x] 2.1 Add UPDATE_GOLDEN environment variable support in CompareGolden
- [x] 2.2 Add mise tasks test:golden and test:golden:update to mise/config.toml

## 3. Golden File Tests

> Note: Golden file tests are E2E tests using the Ginkgo framework with //go:build e2e build tag.

- [x] 3.1 Create golden file tests for list command (text output)
- [x] 3.2 Create golden file tests for list command (JSON output)
- [x] 3.3a Create golden file tests for validation errors
- [x] 3.3b Create golden file tests for service errors
- [x] 3.3c Create golden file tests for not-found errors

## 4. Verification

- [x] 4.1 Run mise run test:golden to verify infrastructure

## 5. Documentation

- [x] 5.1 Update test/AGENTS.md with golden file documentation
- [x] 5.2 Update test/helpers/AGENTS.md with golden file documentation
