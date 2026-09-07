# Project Specific Rules and Information

Project: Yandex MCP Server

NO BACKWARDS COMPATIBILITY! NO FALLBACK! NO DEPRECATIONS! JUST REMOVE OLD CODE/DOCUMENTATION AS NEEDED.
THIS IS A NEW PROJECT, NOT IN PRODUCTION YET. NO DATA IN DATABASE YET. FEEL FREE TO MAKE BREAKING CHANGES AS NEEDED.

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