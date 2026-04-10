# Complexity Signals

When reviewing code, look for signals of complexity that indicate the code is hard to change safely. These signals include:

## 1. Change Amplification

A simple change requires editing many files.

**Heuristic**: "If adding a new field touches more than 3 files, there's amplification."

**Examples**:

- Adding a column requires changing: model, DAO, mapper, service, handler, DTO, test fixtures
- A new API endpoint requires modifying 5+ files in different directories

**Fix**: Consolidate related logic into deeper modules that hide the change.

## 2. Cognitive Load

You need to know a lot of things to make a change safely.

**Heuristic**: "If a new developer needs to read more than 5 files to understand one flow, the load is high."

**Examples**:

- Understanding user creation requires reading auth, validation, database, events, and notifications
- A function signature requires understanding the caller's caller to use correctly

**Fix**: Deep modules with obvious interfaces that don't require reading the implementation.

## 3. Unknown Unknowns

It's not obvious what needs to change, or what will break.

**Heuristic**: "If a change breaks something unexpected, there are unknown unknowns."

**Examples**:

- Changing a utility function breaks an unrelated feature because of a hidden dependency
- Renaming a field causes a runtime error in a template that's not type-checked

**Fix**: Make dependencies explicit, enforce invariants at compile time or with tests.

## During Exploration

When navigating code, rate each module:

| Signal | Score |
|--------|-------|
| I can make a change in one place | Low complexity |
| I need to touch 2-3 related files | Normal |
| I need to touch 4+ files across directories | Change amplification |
| I can understand the flow by reading one file | Low cognitive load |
| I need to read 5+ files to understand | High cognitive load |
| Changes break in predictable places | No unknown unknowns |
| Changes break in surprising places | Unknown unknowns |
