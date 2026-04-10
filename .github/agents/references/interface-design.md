# Interface Design for Testability

Designing interfaces with testability in mind leads to more maintainable and robust code. A well-designed interface allows you to write tests that are easy to set up, understand, and maintain. Here are some principles and heuristics to guide you in creating testable interfaces.

## Heuristics

### 1. Accept dependencies, don't create them

```typescript
// Testable — dependency injected
function processOrder(order: Order, gateway: PaymentGateway) {}

// Hard to test — dependency created internally
function processOrder(order: Order) {
  const gateway = new StripeGateway(); // locked to Stripe in tests
}
```

### 2. Return results, don't produce side effects

```typescript
// Testable — pure function, assert on return value
function calculateDiscount(cart: Cart): Discount {}

// Hard to test — mutates input, must inspect cart after call
function applyDiscount(cart: Cart): void {
  cart.total -= discount;
}
```

### 3. Small surface area

- Fewer methods = fewer tests needed
- Fewer params = simpler test setup
- If a function takes more than 3 parameters, consider a config/options object

### 4. Hide information

The interface should not reveal implementation details:

```go
// Good — caller doesn't know about SQL
func (s *BookService) FindByAuthor(ctx context.Context, authorID string) ([]Book, error)

// Bad — SQL leaks into the interface
func (s *BookService) FindByAuthor(ctx context.Context, authorID string, orderByColumn string, joinTables []string) ([]Book, error)
```

### 5. Sensible defaults

The common case should be trivial; the complex case should be possible:

```typescript
// Good — common case requires no config
const client = createClient("https://api.example.com");

// Also good — complex case is supported
const client = createClient("https://api.example.com", {
  timeout: 5000,
  retries: 3,
  headers: { "X-Custom": "value" },
});
```

## Quick Test

For any interface, ask:

- Can I write a test without setting up the universe?
- Can I swap the real dependency for a fake in one line?
- Does the test read like documentation of the behavior?
