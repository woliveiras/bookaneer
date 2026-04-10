# Pragmatic Heuristics

Actionable heuristics for evaluating code. Each has a definition, a red flag to detect violations, and a common trap to avoid.

## 1. DRY (Don't Repeat Yourself)

**Definition**: Every piece of knowledge should have a single, unambiguous representation.

**Red flag**: Changing one behavior requires editing the same logic in multiple places.

**Trap**: Accidental duplication is NOT a DRY violation. Two functions with identical code that change for different reasons should stay separate. Premature DRY creates coupling.

## 2. Orthogonality

**Definition**: Changing A should not affect B. Modules should be independent.

**Red flag**: "If I change X, how many things break?" If the answer is more than 1, they're not orthogonal.

**Trap**: Sharing code between modules feels DRY but creates coupling. Prefer duplication over the wrong abstraction.

## 3. Reversibility

**Definition**: Decisions should be easy to reverse. Hard-to-reverse decisions need more care.

**Red flag**: You chose a database/framework/service and it's deeply wired into every layer.

**Trap**: Overengineering "just in case" — not every decision needs an abstraction layer. Focus reversibility on decisions that are genuinely risky.

## 4. Tracer Bullets

**Definition**: Build thin vertical slices end-to-end before filling in horizontal layers.

**Red flag**: Building the entire database layer before touching the UI, or designing all APIs before implementing one.

**Trap**: Confusing tracer bullets with prototypes. Tracer bullets are production-quality code in a narrow slice. Prototypes are throwaway.

## 5. ETC (Easy to Change)

**Definition**: When in doubt between two designs, choose the one that's easier to change later.

**Red flag**: You're debating two approaches and neither has a clear advantage. Pick the one with fewer commitments.

**Trap**: "Easy to change" doesn't mean "infinitely flexible." It means fewer hardcoded assumptions, not more abstraction layers.

## 6. Tell, Don't Ask

**Definition**: Tell objects what to do; don't ask for their state and decide for them.

**Red flag**: `if (order.getStatus() === "paid") { order.setStatus("shipped"); }` — you're making decisions that belong to the order.

**Trap**: Taking this too far and hiding all state. Sometimes you need to read state (for display, logging, serialization). The heuristic applies to behavior, not reads.

## 7. Law of Demeter

**Definition**: Only talk to your direct dependencies. Don't reach through objects.

**Red flag**: `user.getAddress().getCity().getZipCode()` — chaining through multiple objects.

**Trap**: Creating bloated wrapper methods just to avoid dots. If the chain represents a genuine navigation path in the domain, a dedicated method on the top-level object is better than mechanical wrapping.
