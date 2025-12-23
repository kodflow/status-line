# Elixir >= 1.19.0 | OTP >= 27
> Release Notes: https://github.com/elixir-lang/elixir/releases

## Structure

```
/src
├── lib/
│   └── project_name/
│       ├── application.ex
│       └── module.ex
├── mix.exs
└── mix.lock
/tests
├── test_helper.exs
└── module_test.exs
```

## Style

- `mix format` (mandatory)
- Credo for linting
- Dialyzer for type checking
- Maximum line length: 98

## Naming

- Files: snake_case.ex/.exs
- Modules: PascalCase
- Functions/variables: snake_case
- Module attributes: @snake_case

## Modern Elixir

- Pattern matching everywhere
- With clauses for complex flows
- Behaviours for polymorphism
- GenServer for state
- Supervisors for fault tolerance

## OTP Patterns

- Let it crash philosophy
- Supervision trees
- GenServer for stateful processes
- Task for async work

## Testing

- ExUnit (mandatory)
- Tests in `/tests` directory
- Doctests for examples
- Property-based testing

## Dependencies

- Hex packages via mix.exs
- Keep dependencies minimal
- Phoenix for web apps

## Forbidden

- `try/catch` for control flow
- Mutable state outside GenServer
- Deep nesting (>3 levels)
- Long functions (>20 lines)
