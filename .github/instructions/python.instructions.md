---
applyTo: "**/*.py"
---

# Python Conventions

## Type Hints
- Type hints on all function signatures
- Use Python 3.12+ union syntax: `str | None` (not `Optional[str]`)
- Use `from __future__ import annotations` if needed for older Python

## Async
- Prefer `async def` for I/O-bound operations
- Use `asyncio` patterns consistent with the project's framework

## Dependencies
- Use `uv` for dependency management (not pip directly)
- Use `pydantic` for settings and validation
- Cache settings with `@lru_cache`: one settings instance per process

## Structure
- snake_case for functions, modules, variables
- PascalCase for classes and type names
- One class per file when the class is substantial
- Group imports: stdlib → third-party → local (enforced by ruff)

## Testing
- Use `pytest` with fixtures
- Mirror `src/` structure in `tests/`
- Mark integration tests with `@pytest.mark.integration`
- Async tests: use `pytest-asyncio` with auto mode

## Logging
- Use `structlog` with JSON rendering for structured logs
- Create logger per module: `structlog.get_logger(__name__)`
