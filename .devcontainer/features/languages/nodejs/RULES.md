# Node.js >= 25.0.0
> Release Notes: https://nodejs.org/en/blog/release

## Structure

```
/src
├── index.ts
├── components/
├── services/
├── utils/
├── types/
├── package.json
└── tsconfig.json
/tests
├── unit/
└── integration/
```

## Style

- TypeScript (mandatory for new code)
- ESLint + Prettier
- ES Modules (`"type": "module"`)
- Strict mode in tsconfig.json

## TypeScript

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true
  }
}
```

## Naming

- Files: kebab-case
- Classes: PascalCase
- Functions/variables: camelCase
- Constants: UPPER_SNAKE_CASE
- Types/Interfaces: PascalCase

## Testing

- Vitest or Jest
- Tests in `/tests` directory
- Minimum 80% coverage
- Mock external dependencies

## Dependencies

- `package.json` with exact versions
- `pnpm` preferred (faster, stricter)
- No `any` type
- Audit dependencies regularly

## Forbidden

- `var` keyword
- `==` (use `===`)
- `console.log` in production
- Synchronous file operations
- `require()` in ES modules
