# Project Specific Rules

Project: Yandex MCP Server

## Tech stack
1. go 1.27
2. `github.com/modelcontextprotocol/go-sdk` for MCP server implementation
3. `github.com/stretchr/testify` for tests
4. `go.uber.org/mock` (no custom mocks, `//go:generate go tool mockgen` directives in interface files)
5. `github.com/caarlos0/env/v11` for loading configuration from environment variables
6. `log/slog` for logging (must use structured logging with context)

## Instructions
1. Generate new documentation in English unless user specifically requests another language.
2. Maintain consistency of environment variables between `.env.example`, `.env`, Taskfile.yml, scripts, code, and documentation.
3. Use `task fmt`, `task lint`, `task build` to check code before completing changes.

## Golang rules
1. Use `go.uber.org/mock` for mocks
2. Use `github.com/stretchr/testify` for tests
3. All interfaces MUST be prefixed with uppercase `I` letter
4. For single package:
    1) All interfaces should be in file `interfaces.go`
    2) Main package struct and its constructor should be in `service.go` or `client.go`
    3) All internal structs (except main service struct and DTOs) should be in models.go
    4) All DTO should be in dto.go (structs with tag `json`, `yaml`, etc.)
    5) All configuration related code should be in config.go (using `github.com/caarlos0/env/v11`)
    6) All internal errors should be in errors.go file
    7) All internal constants should be in const.go file
    8) All mock generation commands should be in `interfaces.go`
5. Use following functions to log system errors: internal/adapters/token/errors.go:LogError, internal/domain/errors.go:LogError
6. Use `t.Context()` instead of `context.Background()` in tests
7. ALL DTOs MUST be not exported.

## Documentation
1. Write in English
2. Yandex Tracker Tools: `docs/tracker-tools.md`
3. Yandex Wiki Tools: `docs/wiki-tools.md`
4. Yandex API reference, golang MCP SDK: `docs/research/`
5. Each feature and each bug fix has a specification directory `docs/specs/{name}/`, where `{name}` names the feature or the bug in kebab case, for example `docs/specs/remote-authentication/`. A specification describes the system at the time of its change. A later change creates its own specification and leaves earlier specifications unchanged. The directory holds these files:
    1) `problem.md` describes what is wrong now, for whom, and what it costs, with the current behavior and the boundary of the problem. It holds no requirements and no solution.
    2) `requirements.md` lists the requirements under which the problem disappears, the scope and the non-scope, and the scenarios that show that the requirements hold. Each requirement is verifiable without knowledge of the implementation.
    3) `solution.md` describes how the requirements are met, each trade-off with its cost, and how the result is verified.
    4) `glossary.md` defines each term that the specification introduces, one item per term in the form `term: definition`. The directory has this file only when the specification introduces terms.