# Dependency Categories

When analyzing module dependencies, look for these four types of dependencies between modules. Use these to classify coupling during architecture reviews and decide how to consolidate or isolate.

## 1. Shared State

Modules that read/write the same data.

**How to identify**: Multiple modules access the same database table, global variable, or shared file.

**Example**: `OrderService` and `InvoiceService` both read/write the `orders` table directly.

**Deepening strategy**: Create a single module that owns the data and exposes an interface. Others go through it.

**Test impact**: Shared state makes tests order-dependent. Deepening lets you test each consumer in isolation with a fake data owner.

## 2. Shared Interface

Modules that implement the same interface or contract.

**How to identify**: Multiple classes implement the same interface, or multiple functions have the same signature pattern.

**Example**: `StripeGateway`, `PayPalGateway`, and `MockGateway` all implement `PaymentGateway`.

**Deepening strategy**: This is usually healthy coupling. Ensure the interface is small and stable. If the interface keeps changing, it's too shallow.

**Test impact**: Shared interfaces enable easy test doubles. Keep the interface minimal.

## 3. Call Chain

A calls B, which calls C (transitive coupling).

**How to identify**: Following a function call leads through 3+ layers before reaching the actual logic.

**Example**: `Handler` → `Service` → `Repository` → `Database`, where each layer does minimal work.

**Deepening strategy**: If intermediate layers just pass through, collapse them. The handler can call the repository directly if the "service" adds no logic.

**Test impact**: Long call chains require mocking at every level. Shorter chains = simpler tests.

## 4. Temporal

A must run before B (order matters).

**How to identify**: Functions that must be called in sequence, initialization that must happen before use, or data that must be prepared before processing.

**Example**: `migrate()` must run before `startServer()`. `buildIndex()` must complete before `search()`.

**Deepening strategy**: Make temporal dependencies explicit in the interface (constructor takes ready-to-use dependencies) or enforce order with types (builder pattern, state machines).

**Test impact**: Temporal coupling causes flaky tests when order isn't guaranteed. Make setup explicit in each test.
