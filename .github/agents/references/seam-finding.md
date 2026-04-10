# Seam Finding

A **seam** is a place in code where you can change behavior without editing the code itself. Seams are essential for making legacy code testable.

## Types of Seams

### Object Seam

Substitute behavior via inheritance or polymorphism (override a method).

```typescript
// Production
class OrderProcessor {
  protected sendEmail(to: string, body: string) {
    emailService.send(to, body);
  }
}

// Test — override the seam
class TestableOrderProcessor extends OrderProcessor {
  emails: string[] = [];
  protected sendEmail(to: string, body: string) {
    this.emails.push(body); // capture instead of sending
  }
}
```

### Link Seam

Substitute a dependency at the import/require level (module mocking).

```typescript
// Production imports a real module
import { fetchUser } from "./api";

// Test replaces the module
jest.mock("./api", () => ({
  fetchUser: jest.fn().mockResolvedValue({ id: 1, name: "Test" }),
}));
```

### Preprocessing Seam

Change behavior via environment variables, build flags, or configuration.

```python
# Production
STORAGE_BACKEND = os.environ.get("STORAGE_BACKEND", "s3")

# Test — set env before import
os.environ["STORAGE_BACKEND"] = "local"
```

## How to Find Seams

| Look for... | Seam type |
|-------------|-----------|
| `new` / direct instantiation | Object seam candidate |
| `import` / `require` of external modules | Link seam candidate |
| API calls, HTTP requests | Natural boundary for seam |
| Config-based conditionals | Preprocessing seam already exists |

## Patterns to Open Seams

### Extract and Override

Extract a call into a protected method, then override in tests.

- **When**: You can't change the constructor or inject dependencies
- **Risk**: Creates a test-only subclass; keep it minimal

### Wrap Method

Add behavior before/after without modifying the original.

- **When**: You need to add logging, validation, or instrumentation
- **Risk**: Can obscure the original flow if overused

### Sprout Method / Sprout Class

New code goes in a separate method or class, testable in isolation.

- **When**: Adding new behavior to an untested function
- **Risk**: Can fragment logic if the sprout grows too large; consolidate later

## Red Flags (No Seams Available)

| Signal | Problem |
|--------|---------|
| Constructor does real work | Can't instantiate in tests without side effects |
| Static methods with side effects | Can't override or mock without link-level hacks |
| Global singletons | Hidden dependencies, shared state between tests |
| Business logic mixed with I/O | Can't test logic without triggering I/O |
