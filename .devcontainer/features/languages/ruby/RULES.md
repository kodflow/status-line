# Ruby >= 3.4.0
> Release Notes: https://www.ruby-lang.org/en/news/

## Structure

```
/src
├── lib/
│   └── project_name/
│       ├── version.rb
│       └── module.rb
├── Gemfile
└── Gemfile.lock
/tests
├── test_helper.rb
└── test_module.rb
```

## Style

- RuboCop (mandatory)
- Ruby Style Guide
- Maximum line length: 80
- 2 spaces indentation

## Naming

- Files: snake_case.rb
- Classes/Modules: PascalCase
- Methods/variables: snake_case
- Constants: UPPER_SNAKE_CASE
- Predicates: end with `?`
- Dangerous methods: end with `!`

## Modern Ruby

- Pattern matching
- Endless methods
- Data classes
- `...` argument forwarding
- Numbered block params (`_1`, `_2`)

## Testing

- Minitest or RSpec
- Tests in `/tests` or `/spec`
- SimpleCov for coverage
- Minimum 80% coverage

## Dependencies

- Bundler (mandatory)
- Gemfile with versions
- `bundle install --frozen`

## Forbidden

- `eval` unless absolutely necessary
- Global variables (`$var`)
- `for` loops (use iterators)
- Monkey patching core classes
- Implicit returns in long methods
