# Dart >= 3.10.0 | Flutter >= 3.38.0
> Release Notes: https://dart.dev/get-dart | https://docs.flutter.dev/release/release-notes

## Structure

```
/src
├── lib/
│   ├── main.dart
│   ├── src/
│   │   ├── features/
│   │   ├── widgets/
│   │   └── services/
│   └── app.dart
├── pubspec.yaml
└── pubspec.lock
/tests
├── unit/
└── widget/
```

## Style

- Effective Dart (mandatory)
- `dart format` for formatting
- `dart analyze` with no warnings
- Maximum line length: 80

## Naming

- Files: snake_case.dart
- Classes: PascalCase
- Functions/variables: camelCase
- Constants: camelCase or UPPER_SNAKE_CASE
- Private: `_prefix`

## Modern Dart

- Null safety (mandatory)
- Pattern matching
- Records
- Class modifiers (sealed, final, interface)
- Extension types

## Flutter Specific

- Prefer `const` constructors
- Extract widgets when >50 lines
- Use `BuildContext` extensions
- Riverpod or Bloc for state

## Testing

- `flutter_test` package
- Tests in `/tests` directory
- Widget tests for UI
- Minimum 80% coverage

## Forbidden

- `dynamic` without reason
- Nullable types without handling
- `print()` in production
- Deeply nested widgets (>4 levels)
- `setState` in complex widgets
