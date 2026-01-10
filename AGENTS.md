# Project Specific Rules and Information

Project: Yandex MCP Server

NO BACKWARDS COMPATIBILITY! NO FALLBACK! NO DEPRECATIONS! JUST REMOVE OLD CODE/DOCUMENTATION AS NEEDED.
THIS IS A NEW PROJECT, NOT IN PRODUCTION YET. NO DATA IN DATABASE YET. FEEL FREE TO MAKE BREAKING CHANGES AS NEEDED.

## Yandex Services Supported
- Yandex Tracker
- Yandex Wiki

## Command Line Options
- `--wiki-write` enable write operations for Yandex Wiki tools (default: false)
- `--tracker-write` enable write operations for Yandex Tracker tools (default: false)

## Tech stack
- go 1.25.5
- `github.com/modelcontextprotocol/go-sdk` for MCP server implementation
- `github.com/stretchr/testify` for tests
- `go.uber.org/mock` (no custom mocks, `//go:generate` directives in interface files)
- `github.com/caarlos0/env/v11` for loading configuration from environment variables
- `log/slog` for logging (must use structured logging with context)

## Instructions
- MUST generate new documentation in English unless user specifically requests another language.
- When updating documents, the original language of the document must be used.
- MUST ALWAYS use English version of official sources.
- Use `Taskfile.yml` to run tasks.
- MUST maintain consistency of environment variables between `.env.example`, `.env`, Taskfile.yml, scripts, code, and documentation.
- Use `task lint` and `task build` to check code before completing changes.
- MUST NOT use tables in user-facing markdown docs, use lists or sections instead.
- ALL DTOs MUST be not exported.

## Golang rules
- MUST use `go.uber.org/mock` for mocks
- MUST use `github.com/stretchr/testify` for tests
- All interfaces MUST be prefixed with uppercase `I` letter
- For single package:
    * All interfaces should be in file `interfaces.go`
    * Main package struct and its constructor should be in `service.go` or `client.go`
    * All internal structs (except main service struct and DTOs) should be in models.go
    * All DTO should be in dto.go (structs with tag `json`, `yaml`, etc.)
    * All configuration related code should be in config.go (using `github.com/caarlos0/env/v11`)
    * All internal errors should be in errors.go file
    * All internal constants should be in const.go file
    * All mock generation commands should be in `interfaces.go`
- MUST use `golangci-lint-v2` for linting

## Documentation
- Yandex Tracker Tools: `docs/tracker-tools.md`
- Yandex Wiki Tools: `docs/wiki-tools.md`
- Yandex API reference, golang MCP SDK: `docs/research/`

## Environment Variables (.env)
- `YANDEX_WIKI_BASE_URL`: Base URL for Yandex Wiki API
- `YANDEX_TRACKER_BASE_URL`: Base URL for Yandex Tracker API (default https://api.tracker.yandex.net)
- `YANDEX_CLOUD_ORG_ID`: Yandex Cloud Organization ID
- `YANDEX_IAM_TOKEN_REFRESH_PERIOD`: Token refresh period in hours (default 10)

## Folder structure
```
├── docs # Documentation files
├── cmd
│   └── yandex-mcp # Main application entry point
└── internal
    ├── adapters
    │   ├── token # Yandex Token API adapter
    │   ├── tracker # Yandex Tracker API adapter
    │   └── wiki # Yandex Wiki API adapter
    ├── config # Loading configuration from env
    ├── itest # Integration tests
    ├── domain # Domain models and errors
    ├── tools
    │   ├── helpers # Common tools helpers
    │   ├── tracker # Yandex Tracker related tools
    │   └── wiki # Yandex Wiki related tools
    └── server # MCP server initialization       
```

## Architecture

### General

**cmd/yandex-mcp**
Application entry point and dependencies initialization

**internal/config/**
Configuration loading from environment variables.
Passes configuration struct via pointer to all components that need it.

**internal/domain/**
Domain models and errors definitions.
Contains ALL models used between different layers, including requests and responses models.
Not contains any DTOs.
Example: `internal/adapters/token` requests and responses models are defined here, not in adapter package.

### Adapters

**internal/adapters/token/**
Yandex Token API adapter.
Responsible for obtaining and refreshing IAM_token.
Use `yc` cli command internally to obtain and parse token from output via regex `t1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{86}[=]{0,2}`.
Implement interfaces: `wiki.ITokenProvider`, `tracker.ITokenProvider`

**internal/adapters/wiki/**
Yandex Wiki API adapter.
Responsible for making API calls to Yandex Wiki service.
Uses Token adapter via interface `wiki.ITokenProvider` for retrieving IAM_token.

**internal/adapters/tracker/**
Yandex Tracker API adapter.
Responsible for making API calls to Yandex Tracker service.
Uses Token adapter via interface `tracker.ITokenProvider` for retrieving IAM_token.

### MCP Server

**internal/server/**
MCP server implementation.
Responsible for initializing and starting MCP server.
Register Wiki Tools and Tracker Tools via interfaces `server.IToolsRegistrator`.
Know nothing about which tools to register. Receives slices of `server.IToolsRegistrator` via dependency injection.

### Tools

Tool names defined in `internal/domain/tracker_tools.go` and `internal/domain/wiki_tools.go`.

Tools Description Rules: 
- Tools descriptions MUST be short, clear, and self-contained.
- If field has predefined set of values, MUST document all possible values.
- MUST NOT contain text like "etc.", "and so on", "and more", "such as", etc.
- MUST use consistent terminology with Yandex official documentation.

MCP Tools Implementation Rules:
- MCP tools should be primarily convenient for LLM usage and do not necessarily match the parameters of the official Yandex API.
- For example:
    * If the API has a parameter that passes a list as a delimited string, then in the MCP tool this parameter should be a list of strings
    * If the API has a parameter that is a string, but contains integer value, then in the MCP tool this parameter should be an integer    

**internal/tools/wiki/**
Yandex Wiki related tools implementation.
Responsible for registering Wiki related tools to MCP server.
Uses Wiki adapter via interface `wiki.IWikiAdapter` for making API calls to Yandex Wiki.

**internal/tools/tracker/**
Yandex Tracker related tools implementation.
Responsible for registering Tracker related tools to MCP server.
Uses Tracker adapter via interface `tracker.ITrackerAdapter` for making API calls to Yandex Tracker

### Authentication
- Use `IAM_token` for authentication to Yandex services
- `IAM_token` has limited lifetime and MUST be refreshed if api calls start failing with authentication errors.