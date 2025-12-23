# Python >= 3.14.0
> Release Notes: https://docs.python.org/3/whatsnew/

## Structure

```
/src
├── package_name/
│   ├── __init__.py
│   ├── module.py
│   └── subpackage/
├── pyproject.toml
└── requirements.txt
/tests
├── test_module.py
└── conftest.py
```

## Style

- PEP 8 (mandatory)
- `ruff` for linting + formatting
- `mypy --strict` for type checking
- Maximum line length: 88 (Black default)

## Type Hints

- ALL functions must have type hints
- Use `typing` module for complex types
- `from __future__ import annotations`

## Naming

- Modules: snake_case
- Classes: PascalCase
- Functions/variables: snake_case
- Constants: UPPER_SNAKE_CASE
- Private: `_prefix`

## Testing

- `pytest` (mandatory)
- Tests in `/tests` directory
- `pytest-cov` minimum 80%
- Fixtures in `conftest.py`

## Dependencies

- `pyproject.toml` for config
- Virtual environments (venv/poetry/uv)
- Pin versions in requirements.txt

## Forbidden

- `from module import *`
- Mutable default arguments
- Bare `except:`
- `type: ignore` without reason
