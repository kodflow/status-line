# Java >= 25
> Release Notes: https://adoptium.net/temurin/release-notes/

## Structure

```
/src
├── main/
│   └── java/
│       └── com/company/project/
│           ├── Application.java
│           ├── controller/
│           ├── service/
│           └── model/
├── pom.xml (Maven) or build.gradle (Gradle)
/tests
└── java/
    └── com/company/project/
```

## Style

- Google Java Style Guide
- `google-java-format` for formatting
- Checkstyle for linting
- Maximum line length: 100

## Naming

- Packages: lowercase, reverse domain
- Classes: PascalCase
- Methods/variables: camelCase
- Constants: UPPER_SNAKE_CASE

## Modern Java

- Records for DTOs
- Pattern matching (`instanceof`)
- Sealed classes when appropriate
- `var` for local variables
- Text blocks for multiline strings

## Testing

- JUnit 5 (mandatory)
- Tests in `/tests` directory
- Mockito for mocking
- Minimum 80% coverage

## Dependencies

- Maven or Gradle
- Spring Boot for web apps
- Lombok sparingly
- Keep dependencies minimal

## Forbidden

- Raw types (use generics)
- Checked exceptions for control flow
- `null` returns (use `Optional`)
- Mutable public fields
- `System.out.println` in production
