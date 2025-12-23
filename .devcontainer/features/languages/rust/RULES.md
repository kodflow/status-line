# Rust >= 1.92.0
> Release Notes: https://releases.rs/

## Structure

```
/src
├── main.rs (binary) or lib.rs (library)
├── modules/
│   └── mod.rs
├── Cargo.toml
└── Cargo.lock
/tests
└── integration_test.rs
```

## Style

- `rustfmt` (mandatory)
- `clippy` with `-D warnings`
- Rust 2021 edition minimum

## Naming

- Crates: snake_case
- Modules: snake_case
- Types/Traits: PascalCase
- Functions/variables: snake_case
- Constants: UPPER_SNAKE_CASE
- Lifetimes: lowercase (`'a`, `'static`)

## Error Handling

- `Result<T, E>` for recoverable errors
- `?` operator for propagation
- Custom error types with `thiserror`
- `anyhow` for applications

## Testing

- Unit tests in same file (`#[cfg(test)]`)
- Integration tests in `/tests`
- `cargo test` minimum 80% coverage
- Doc tests for public APIs

## Memory Safety

- Prefer borrowing over cloning
- Use `Arc`/`Rc` sparingly
- Avoid `unsafe` unless necessary
- Document all `unsafe` blocks

## Forbidden

- `unwrap()` in production code
- `expect()` without clear message
- `clone()` without justification
- Unused `Result` values
