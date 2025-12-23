# C++ >= C++23
> Release Notes: https://en.cppreference.com/w/cpp/compiler_support

## Structure

```
/src
├── main.cpp
├── include/
│   └── project/
│       └── header.hpp
├── lib/
│   └── module.cpp
├── CMakeLists.txt
/tests
└── test_module.cpp
```

## Style

- `clang-format` (mandatory)
- `clang-tidy` for linting
- Google C++ Style Guide
- Maximum line length: 80

## Naming

- Files: snake_case.cpp/.hpp
- Classes: PascalCase
- Functions: snake_case or camelCase
- Variables: snake_case
- Constants: kConstantName or UPPER_SNAKE_CASE
- Namespaces: lowercase

## Modern C++

- Smart pointers (`unique_ptr`, `shared_ptr`)
- Range-based for loops
- `auto` for complex types
- `constexpr` where possible
- `std::optional`, `std::variant`
- Modules (C++20+)

## Testing

- Google Test (gtest)
- Tests in `/tests` directory
- CTest for integration
- Sanitizers (ASan, UBSan)

## Memory Management

- RAII pattern
- No raw `new`/`delete`
- `std::span` for array views
- `std::string_view` for string params

## Forbidden

- Raw pointers for ownership
- C-style casts
- `#define` for constants
- `using namespace std;` in headers
- Manual memory management
