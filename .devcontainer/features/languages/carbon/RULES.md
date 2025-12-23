# Carbon >= 0.1.0 (Experimental)
> Release Notes: https://github.com/carbon-language/carbon-lang/releases

**WARNING:** Carbon is experimental. Rules may change.

## Structure

```
/src
├── main.carbon
├── lib/
│   └── module.carbon
└── BUILD
/tests
└── test_module.carbon
```

## Style

- Official Carbon formatter (when available)
- Google C++ style influence
- Maximum line length: 80

## Naming

- Files: snake_case.carbon
- Types: PascalCase
- Functions: PascalCase
- Variables: snake_case
- Constants: UPPER_SNAKE_CASE

## Modern Carbon

- Memory safety by default
- Generics
- Interfaces
- Pattern matching
- Interop with C++

## Testing

- Carbon test framework (when available)
- Tests in `/tests` directory
- Bazel for build system

## C++ Interop

- Seamless C++ integration
- Use Carbon for new code
- Migrate C++ incrementally

## Forbidden

- Raw pointers for ownership
- Manual memory management
- C-style casts
- Undefined behavior

## Note

Carbon is a successor to C++ by Google.
Follow official documentation as it evolves.
