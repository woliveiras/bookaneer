---
applyTo: "**/*.test.*,**/*.spec.*,**/tests/**,**/test/**,**/*_test.*"
---

# Testing Conventions

- One test per behavior — each test verifies exactly one thing
- Descriptive test names that read as documentation
- Follow the Arrange-Act-Assert pattern (or Given-When-Then)
- Tests are self-contained — no shared mutable state between tests
- Test the public interface, not implementation details
- Each acceptance criterion from a spec should have at least one test
- Do NOT modify existing tests to make code pass — fix the code instead
- Use the test framework and helpers already present in the project
- Prefer real objects over mocks; mock only at system boundaries (network, disk, time)
