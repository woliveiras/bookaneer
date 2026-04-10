# Deep Modules

**Deep module** = small interface + lots of hidden implementation.

┌─────────────────────┐
│   Small Interface   │  ← Few methods, simple params
├─────────────────────┤
│                     │
│  Deep Implementation│  ← Complex logic hidden inside
│                     │
└─────────────────────┘


**Shallow module** = large interface + little implementation (avoid).

┌─────────────────────────────────┐
│       Large Interface           │  ← Many methods, complex params
├─────────────────────────────────┤
│  Thin Implementation            │  ← Just passes through
└─────────────────────────────────┘


## Evaluation Questions

When reviewing or designing a module, ask:

1. Can I reduce the number of methods?
2. Can I simplify the parameters?
3. Can I hide more complexity inside?
4. Does the caller need to know implementation details to use this?

## Red Flags (Shallow Modules)

| Signal | Example |
|--------|---------|
| **Pass-through method** | Method that just calls another method with the same args |
| **Wrapper class** | Class that wraps another class without adding meaningful logic |
| **Single-method delegator** | Class exists only to forward a call to another class |
| **Interface-heavy** | Interface has as many methods as the implementation has lines |
| **Configuration bloat** | Caller must set many options because the module doesn't pick good defaults |

## Deep Module Test

A module is deep when:

- You can describe what it does in one sentence
- Its tests only need to touch the public interface
- Changing the implementation doesn't change the callers
- New team members can use it without reading the source
