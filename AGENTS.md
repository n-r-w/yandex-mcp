# Project Specific Rules and Information

Project: Yandex MCP Server

NO BACKWARDS COMPATIBILITY! NO FALLBACK! NO DEPRECATIONS! JUST REMOVE OLD CODE/DOCUMENTATION AS NEEDED.
THIS IS A NEW PROJECT, NOT IN PRODUCTION YET. NO DATA IN DATABASE YET. FEEL FREE TO MAKE BREAKING CHANGES AS NEEDED.

## Yandex Services Supported
1. Yandex Tracker
2. Yandex Wiki

## Tech stack
1. go 1.25.5
2. `github.com/modelcontextprotocol/go-sdk` for MCP server implementation
3. `github.com/stretchr/testify` for tests
4. `go.uber.org/mock` (no custom mocks, `//go:generate` directives in interface files)
5. `github.com/caarlos0/env/v11` for loading configuration from environment variables
6. `log/slog` for logging (must use structured logging with context)

## Instructions
1. DON'T edit AGENTS.md without DIRECT user request.
2. Generate new documentation in English unless user specifically requests another language.
3. When updating documents, the original language of the document must be used.
4. ALWAYS use English version of official sources.
5. Use `Taskfile.yml` to run tasks.
6. Maintain consistency of environment variables between `.env.example`, `.env`, Taskfile.yml, scripts, code, and documentation.
7. Use `task lint` and `task build` to check code before completing changes.
8. DON'T use tables in user-facing markdown docs, use lists or sections instead.
9. ALL DTOs MUST be not exported.

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
5. Use `golangci-lint-v2` for linting
6. Use following functions to log system errors: internal/adapters/token/errors.go:LogError, internal/domain/errors.go:LogError
7. Use `t.Context()` instead of `context.Background()` in tests

## Documentation
1. Yandex Tracker Tools: `docs/tracker-tools.md`
2. Yandex Wiki Tools: `docs/wiki-tools.md`
3. Yandex API reference, golang MCP SDK: `docs/research/`

## Environment Variables (.env)
1. `YANDEX_WIKI_BASE_URL`: Base URL for Yandex Wiki API
2. `YANDEX_TRACKER_BASE_URL`: Base URL for Yandex Tracker API (default https://api.tracker.yandex.net)
3. `YANDEX_CLOUD_ORG_ID`: Yandex Cloud Organization ID
4. `YANDEX_IAM_TOKEN_REFRESH_PERIOD`: Token refresh period in hours (default 10)
5. `YANDEX_HTTP_TIMEOUT`: HTTP timeout for Yandex API requests in seconds (default 30)
6. `YANDEX_MCP_ATTACH_EXT`: Comma-separated list of allowed attachment extensions without dots (e.g. "jpg,png,txt"). Optional, if not set, default to predefined list of common extensions. 
7. `YANDEX_MCP_ATTACH_VIEW_EXT`: Comma-separated list of allowed attachment extensions without dots for inline viewing (e.g. "txt,json,md"). Optional, if not set, default to predefined list of text formats.
8. `YANDEX_MCP_ATTACH_DIR`: Directory path for saving attachments (must be absolute path, e.g. "/home/user/attachments"). Optional, if not set, default to rules that allow saving attachments in user home directory except hidden top-level subdirectories (e.g. "~/.ssh").

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
1. Configuration loading from environment variables.
2. Passes configuration struct via pointer to all components that need it.

**internal/domain/**
1. Domain models and errors definitions.
2. Contains ALL models used between different layers, including requests and responses models.
3. Not contains any DTOs.
4. Example: `internal/adapters/token` requests and responses models are defined here, not in adapter package.

### Adapters

**internal/adapters/token/**
1. Yandex Token API adapter.
2 .Responsible for obtaining and refreshing IAM_token.
3. Use `yc` cli command internally to obtain and parse token from output via regex `t1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{86}[=]{0,2}`.
4. Implement interfaces: `wiki.ITokenProvider`, `tracker.ITokenProvider`

**internal/adapters/wiki/**
1. Yandex Wiki API adapter.
2. Responsible for making API calls to Yandex Wiki service.
3. Uses Token adapter via interface `wiki.ITokenProvider` for retrieving IAM_token.

**internal/adapters/tracker/**
1. Yandex Tracker API adapter.
2. Responsible for making API calls to Yandex Tracker service.
3. Uses Token adapter via interface `tracker.ITokenProvider` for retrieving IAM_token.

### MCP Server

**internal/server/**
1. MCP server implementation.
2. Responsible for initializing and starting MCP server.
3. Register Wiki Tools and Tracker Tools via interfaces `server.IToolsRegistrator`.
4. Know nothing about which tools to register. Receives slices of `server.IToolsRegistrator` via dependency injection.

### Tools

Tool names defined in `internal/domain/tracker_tools.go` and `internal/domain/wiki_tools.go`.

Tools Description Rules: 
1. Tools descriptions MUST be short, clear, and self-contained.
2. If field has predefined set of values, MUST document all possible values.
3. MUST NOT contain text like "etc.", "and so on", "and more", "such as", etc.
4. MUST use consistent terminology with Yandex official documentation.

MCP Tools Implementation Rules:
1. MCP tools should be primarily convenient for LLM usage and do not necessarily match the parameters of the official Yandex API.
2. For example:
    1) If the API has a parameter that passes a list as a delimited string, then in the MCP tool this parameter should be a list of strings
    2) If the API has a parameter that is a string, but contains integer value, then in the MCP tool this parameter should be an integer    

**internal/tools/wiki/**
1. Yandex Wiki related tools implementation.
2. Responsible for registering Wiki related tools to MCP server.
3. Uses Wiki adapter via interface `wiki.IWikiAdapter` for making API calls to Yandex Wiki.

**internal/tools/tracker/**
1. Yandex Tracker related tools implementation.
2. Responsible for registering Tracker related tools to MCP server.
3. Uses Tracker adapter via interface `tracker.ITrackerAdapter` for making API calls to Yandex Tracker

### Authentication
1. Use `IAM_token` for authentication to Yandex services
2. `IAM_token` has limited lifetime and MUST be refreshed if api calls start failing with authentication errors.