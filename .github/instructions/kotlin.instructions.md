---
applyTo: "**/*.kt"
---

# Kotlin / Android Conventions

## Architecture: MVVM + Repository
- **ViewModel** exposes `StateFlow<UiState>` — never LiveData
- **UiState** is an immutable `data class` with sensible defaults
- **Repository** is an interface; implementation injected via Hilt
- **Screen** collects state with `collectAsStateWithLifecycle()`

## Hilt DI
- ViewModels: `@HiltViewModel` with `@Inject constructor`
- Repositories: `@Singleton` scope, bound via `@Binds` in `@Module`
- Use constructor injection — never field injection

## Room
- DAOs return `Flow<List<Entity>>` for observable queries
- Use `suspend` for write operations
- Soft delete with `deletedAt` timestamp column
- Filter with `WHERE deletedAt IS NULL` by default

## Compose
- Screen composables take `viewModel` as parameter (default `hiltViewModel()`)
- Extract private helper composables for complex UI sections
- Use `Modifier` as first optional parameter in reusable composables
- Follow Material 3 design patterns

## Testing
- Backtick-delimited test names: `` fun `returns error when name is blank`() ``
- Use test builders: `createTestCat()`, `createWeightEntry()`
- Fake repositories for ViewModel tests
- `Truth` for assertions, `Turbine` for Flow testing

## Formatting
- Spotless with ktfmt (Google style)
- Run `./gradlew spotlessCheck` before commit
