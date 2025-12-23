# Scala >= 3.7.0
> Release Notes: https://github.com/scala/scala3/releases

## Structure

```
/src
├── main/
│   └── scala/
│       └── com/company/project/
│           ├── Main.scala
│           └── domain/
├── build.sbt
└── project/
/tests
└── scala/
    └── com/company/project/
```

## Style

- Scalafmt (mandatory)
- Scalafix for linting
- Wartremover for additional checks
- Maximum line length: 100

## Naming

- Files: PascalCase.scala
- Classes/Traits/Objects: PascalCase
- Methods/values: camelCase
- Constants: PascalCase or UPPER_SNAKE_CASE
- Type parameters: single uppercase

## Modern Scala 3

- Indentation-based syntax (optional)
- Given/using for implicits
- Extension methods
- Opaque types
- Enums
- Union/intersection types

## Functional Style

- Immutability by default
- Case classes for data
- Pattern matching
- For comprehensions
- Higher-order functions

## Testing

- ScalaTest or MUnit
- Tests in `/tests` directory
- ScalaCheck for properties
- Minimum 80% coverage

## Dependencies

- sbt (standard)
- Coursier for resolution
- Cats/ZIO for FP

## Forbidden

- `var` without justification
- `null` (use Option)
- `return` statements
- Side effects in constructors
- Throwing exceptions for control flow
