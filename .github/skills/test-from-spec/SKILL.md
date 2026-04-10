---
name: test-from-spec
description: "Generate unit tests from a spec's acceptance criteria. Use when: you have a written spec and need tests, generate tests from spec, TDD from spec, spec to tests."
---

# Test from Spec

Generate unit tests that cover every acceptance criterion in a spec.

## When to Use
- A spec is written and approved
- You're starting the TDD cycle (tests before code)
- You want to verify spec coverage with tests

## Procedure

1. Read the spec file indicated by the user (or search `specs/` directory)
2. Extract all acceptance criteria (the `Given...When...Then` items)
3. For each criterion, generate at least one test:
   - Happy path test for the main behavior
   - Edge case tests if the spec mentions edge cases
   - Error case tests for failure scenarios
4. Follow the **existing test patterns** in the project:
   - Same test framework
   - Same file naming convention
   - Same assertion library
   - Same fixture/helper patterns
5. If no existing tests are found, ask the user which framework to use
6. Save test files next to the source files they test (colocated)

## Rules

- One test per behavior — each test verifies exactly one acceptance criterion
- Test names must be descriptive enough to serve as documentation
- Tests must be self-contained — no shared mutable state between tests
- Do NOT implement the production code — only write the tests
- Tests should fail initially (red phase of TDD)

## Output

For each acceptance criterion, produce:
```
Criterion: "Given X, when Y, then Z"
Test: test_file.ext → test function name
```

Then write the actual test code.
