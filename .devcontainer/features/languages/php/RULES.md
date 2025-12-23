# PHP >= 8.5.0
> Release Notes: https://www.php.net/releases/

## Structure

```
/src
├── index.php
├── App/
│   ├── Controllers/
│   ├── Models/
│   └── Services/
├── composer.json
└── composer.lock
/tests
├── Unit/
└── Feature/
```

## Style

- PSR-12 (mandatory)
- PHP-CS-Fixer for formatting
- PHPStan level 9 for static analysis
- Maximum line length: 120

## Naming

- Files: PascalCase.php (match class name)
- Classes: PascalCase
- Methods: camelCase
- Variables: camelCase
- Constants: UPPER_SNAKE_CASE

## Modern PHP

- Typed properties
- Constructor promotion
- Named arguments
- Match expressions
- Enums
- Readonly classes
- Attributes

## Testing

- PHPUnit 11+
- Tests in `/tests` directory
- Pest for BDD style (optional)
- Minimum 80% coverage

## Dependencies

- Composer (mandatory)
- PSR-4 autoloading
- Laravel/Symfony for frameworks
- Keep dependencies minimal

## Forbidden

- Short array syntax `array()`
- `@` error suppression
- `eval()`
- `global` keyword
- Mixed types without reason
