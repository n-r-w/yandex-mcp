# Research: MCP Go SDK (github.com/modelcontextprotocol/go-sdk)

This repository is intended to ship an MCP server that exposes read-only Yandex Wiki and Yandex Tracker capabilities as MCP tools. This document summarizes the practical parts of the official MCP Go SDK that matter for implementation planning in this repo: server creation, transport selection, tool registration, request/response shapes, error surfacing, and testing hooks.

The focus here is on how to wire the SDK into the existing skeleton:
- `cmd/yandex-mcp` (entrypoint)
- `internal/server` (MCP server initialization)
- `internal/tools/*` (tool registration + tool handlers)
- `internal/adapters/*` (HTTP calls to Yandex APIs)
- `internal/domain` (domain errors / models used across layers)

No code in this repo currently implements the wiring (most Go files are stubs), so the examples below are intentionally “shape only” and may require minor adjustments when implemented.

## What the Go SDK provides (core concepts)

### Server

The SDK’s `mcp` package includes the server implementation and protocol types.

A server is created via `mcp.NewServer(implementation, options)` and started with `server.Run(ctx, transport)`.

Key points relevant to this repo:
- The server is a long-running JSON-RPC endpoint implementing MCP.
- Each connection creates a server-side session; tools can be invoked concurrently.
- Tools can be registered on the server; the server advertises them to the client.

### Tool

A tool is a named capability exposed to the MCP client (the LLM host). At the protocol level, a tool has:
- `name` (stable identifier)
- `description` (for the model; keep it short and precise)
- `inputSchema` (JSON Schema; must be an object)
- `outputSchema` (optional JSON Schema; must be an object when present)

The SDK’s `mcp.Tool` struct represents this metadata.

### Handler

A tool handler is the Go function that runs when the client calls the tool.

There are two registration styles:
- A low-level handler that receives the raw request (`*mcp.CallToolRequest`) and must validate/unmarshal/marshal manually.
- A typed (generic) handler that receives a typed input value (`In`) and returns a typed output value (`Out`), with the SDK handling schema inference, validation, and JSON marshaling.

### Transport

The transport carries JSON-RPC messages between client and server.

The SDK provides:
- Stdio transport for server-side subprocess use (`mcp.StdioTransport`)
- Command transport for client-side spawning of a subprocess (`mcp.CommandTransport`)
- In-memory transport for tests (`mcp.InMemoryTransport`)
- Logging wrapper transport to capture traffic (`mcp.LoggingTransport`)
- HTTP-based transports (streamable HTTP) as part of newer MCP specs

For this project, stdio is the intended production transport.

## How this repo should structure MCP server wiring

This section maps the SDK concepts onto the current repository structure.

### Recommended ownership boundaries

- `cmd/yandex-mcp/main.go`
  - Load env config (eventually via `internal/config`).
  - Build adapters (wiki/tracker clients).
  - Construct the MCP server (delegate to `internal/server`).
  - Register tool sets (delegate to `internal/tools/wiki` and `internal/tools/tracker`).
  - Run the server on stdio.

- `internal/server` (server bootstrap)
  - Provide a constructor like `New(...) *mcp.Server` (exact signature TBD).
  - Server options (logger, initialized hooks) belong here.
  - Keep this package focused on MCP concerns, not Yandex API details.

- `internal/tools/wiki` and `internal/tools/tracker`
  - Export a `Register(server *mcp.Server, deps ...)` function.
  - Define tool input/output DTOs for MCP (with `json` + `jsonschema` tags) in these packages.
  - Tool handlers call adapter interfaces; handlers should not do HTTP themselves.

- `internal/adapters/wiki`, `internal/adapters/tracker`, `internal/adapters/token`
  - Own HTTP request/response mapping and upstream error parsing.
  - Return domain-level models/errors to the tool handlers.

- `internal/domain`
  - Own cross-layer errors (for example, an “upstream HTTP error” type that includes status code and sanitized response body).

This keeps MCP concerns (“how to speak MCP”) separate from upstream concerns (“how to speak Yandex APIs”).

## Tool registration approaches (typed vs raw)

The SDK supports two approaches.

### Preferred: typed tool registration (`mcp.AddTool[In, Out]`)

Use typed registration for almost all tools in this project.

Why:
- Schema inference from Go structs means less boilerplate.
- Arguments are validated against the inferred (or provided) JSON Schema.
- The SDK unmarshals `params.arguments` into `In` for you.
- The SDK marshals `Out` back into a structured output and also populates `content` automatically when not set.
- Regular errors returned by the handler are turned into a tool error result (`isError = true`).

When it’s a good fit:
- Inputs and outputs are naturally JSON objects.
- You want consistent validation and consistent output formatting.
- You want to minimize manual JSON-RPC glue.

### Use with care: raw tool registration (`(*mcp.Server).AddTool`)

Use the low-level API only when you have a specific reason.

Typical reasons:
- You need full control over schema validation (custom schemas beyond what inference can express).
- You need to accept “untyped” or dynamic inputs that do not map well to a Go struct.
- You need to return non-standard content types (for example, images or embedded resources) and want full control over the `CallToolResult`.

Tradeoffs:
- You must supply a non-nil `InputSchema` (and it must be a JSON Schema object).
- You must unmarshal/validate request arguments yourself.
- You must build `CallToolResult` (including `Content`, `StructuredOutput`, and `IsError`) yourself.

Planning guidance for this repo:
- Start with typed tools everywhere.
- Reach for raw tools only if a specific tool cannot be expressed cleanly with typed DTOs.

## Tool input/output design guidance (schemas and JSON shapes)

### Keep inputs and outputs as JSON objects

The SDK expects tool input schemas to be JSON Schema objects (top-level `type: "object"`). For typed handlers, that naturally means:
- `In` should be a struct or a map-like type representing an object.
- `Out` should be a struct or a map-like type representing an object.

Avoid top-level arrays or “primitive-only” tools; wrap everything in an object, even if it’s a single field.

### Use `json` tags for stable wire names

Define `json:"..."` tags on input/output structs so you control the public API shape.

This is especially important for:
- pagination tokens (`cursor`, `page_size`, `scroll_id`, etc.)
- query parameters (`query`, `filter`)
- identifiers (`issue_id`, `page_id`, `slug`)

### Use `jsonschema` tags for model-facing field descriptions

The SDK uses JSON Schema inference; field descriptions can be provided via struct tags. A short description helps models call tools correctly.

Keep these descriptions:
- short
- specific
- aligned with upstream semantics

### Design outputs for “LLM consumption” first

Outputs should be:
- stable (don’t rename fields casually)
- minimal but sufficient (return key identifiers, titles, and links/URLs when meaningful)
- safe (do not echo tokens, secrets, or raw Authorization headers)

For larger upstream payloads, consider:
- returning a summarized subset plus an optional “raw” field only when necessary
- truncating long text fields (and documenting that truncation)

## Request/response flow (what the handler sees)

At runtime:
- The MCP client lists tools and calls one by name.
- Tool calls include `params.arguments` (JSON object).
- With typed tools, the SDK validates and unmarshals arguments into `In`.
- The handler returns:
  - an optional `*mcp.CallToolResult` (can be nil)
  - a typed output value `Out`
  - an error

The SDK then:
- populates structured output based on `Out`
- ensures `content` is non-null (empty list instead of JSON null)
- if the handler returns a regular error, marks the result as `isError = true` and packs the error for the client

Concurrency note:
- The server can handle multiple sessions.
- Tool handlers may be called concurrently.
- Adapters and shared clients should be safe for concurrent use (or protected with appropriate synchronization).

## Error handling (including upstream HTTP error surfacing)

The SDK distinguishes between:
- protocol-level JSON-RPC errors (structured, e.g., `*jsonrpc.Error`)
- tool execution errors (regular Go errors returned by the handler)

### Recommended pattern for upstream HTTP failures

Goal: WHEN a Yandex API request fails, tool callers should receive a useful MCP tool error result that includes the upstream HTTP status and a sanitized error body.

A practical approach for this repo:
- In adapters, convert non-2xx responses into a domain error type (for example: `domain.UpstreamHTTPError`).
- Include in that error:
  - upstream service identifier (wiki/tracker)
  - HTTP status code
  - request correlation id if available (headers)
  - a sanitized and size-limited response body (or parsed error fields)
- In the tool handler, catch that error and return a tool error with enough context.

Two options for returning the error through MCP:

1) Return a normal Go error and rely on the SDK’s automatic tool error packaging.
   - Pros: simplest; consistently sets `isError = true`.
   - Cons: clients may only see a human-oriented message unless you embed structure in the message.

2) Return an explicit `CallToolResult` with `IsError = true` and put a structured error object into `content` (as JSON text), while keeping `error` nil.
   - Pros: you control the error payload precisely.
   - Cons: slightly more boilerplate; you must ensure consistent formatting.

Planning recommendation:
- Start with option (1), but format the error message predictably (include status code and a short upstream summary).
- If you later find that clients need machine-readable errors, evolve toward option (2) or a structured “error output” object.

Security and privacy:
- Never include the IAM token / OAuth token in tool outputs.
- Be careful about echoing request URLs if they may contain sensitive query parameters.

### When to use protocol errors (`*jsonrpc.Error`)

Use protocol errors only for “MCP protocol / request contract” problems, such as:
- invalid params (client sent malformed arguments)
- invalid request

For upstream HTTP failures, prefer tool errors (not protocol errors), because the request was valid but the operation failed.

## Transport choice for this project (stdio)

### Why stdio fits this repo

Stdio transport is the default deployment mode for MCP servers that run as subprocesses spawned by an LLM host (for example, a desktop app). It avoids managing ports, HTTP servers, and session lifecycles.

Server-side transport:
- `mcp.StdioTransport` reads newline-delimited JSON-RPC from stdin and writes to stdout.

Client-side transport (for tests and tooling):
- `mcp.CommandTransport` can spawn the server process and communicate over stdin/stdout.

### How to run locally (once main wiring exists)

Once `cmd/yandex-mcp/main.go` is implemented to call `server.Run(context.Background(), &mcp.StdioTransport{})`, you should be able to:
- build: `task build` (produces `bin/yandex-mcp`)
- run as a subprocess under an MCP host that supports stdio servers
- run directly in a terminal (useful mainly for debugging when wrapped with logging transport)

Note: since the repo’s Go code is currently a stub, these commands will not yet produce a working server until implementation is added.

## Testing guidance (in-memory transport and debugging hooks)

### In-memory transport for unit-style tests

The SDK includes an in-memory transport intended for tests. The typical pattern is:
- create a server
- register tools
- connect a client session to the server via in-memory transport
- call tools and assert on results

This avoids:
- OS-level pipes
- spawning subprocesses
- flaky timing around stdio

Planning recommendation for this repo:
- Use in-memory transport for tool handler tests (validate inputs, outputs, and error shaping).
- Keep adapter tests separate (they may use HTTP mocks).

### Logging transport for debugging

Wrap a transport with `mcp.LoggingTransport` and write logs to a chosen `io.Writer`.

This is useful to:
- inspect JSON-RPC traffic
- debug schema issues
- debug unexpected client behavior

### Conformance / “everything” examples

The upstream repository includes example servers, including a “conformance” server. These are useful reference points when validating that:
- tool registration is correct
- schemas are accepted
- tool errors are surfaced as expected

## References

Primary SDK documentation and source entry points (raw URLs so they’re agent-friendly):

- README (overview, quickstart):
  https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/README.md

- MCP package API docs:
  https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp

- Server documentation (server options, tool registration behavior):
  https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/docs/server.md

- Protocol + transport documentation:
  https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/docs/protocol.md

Useful source pointers inside the repo (browse via GitHub):
- `mcp/server.go` (server implementation and tool dispatch)
- `mcp/tool.go` (typed tool helpers and error packaging behavior)
- `mcp/protocol.go` (protocol structs like `Tool`, `CallToolRequest`, `CallToolResult`)
- `mcp/transport.go` (stdio, in-memory, logging transports)
- `examples/server/*` (hello, basic/in-memory, everything, conformance)
