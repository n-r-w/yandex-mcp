# Wiki attachment download problem

## Key definitions and abbreviations

MCP and agent are defined in the [Yandex 360 OAuth glossary](../yandex-360-oauth/glossary.md).

## Context

An agent can read a Yandex Wiki page through MCP and receive references to files attached to the page.

## Observed problem

- PRM.01: An agent can read a Wiki page and find an attachment link, but it cannot retrieve the file through MCP. For example, the agent can find a link to a CSV report and still be unable to read its rows through MCP.

## Affected audience

Users whose agents need to work with the contents of files linked from Yandex Wiki pages.

## Evidence

- `Registrator.Register` in `internal/tools/wiki/service.go` registers the Wiki tools. Its registrations contain no operation that retrieves the file behind an attachment link.
- The [Yandex Wiki download API reference](https://yandex.ru/support/wiki/ru/api-ref/attachments/pagesattachments__download_by_filename_slug_pair) documents retrieval by attachment path and requires read access to the page that owns the file.

## Impact

HIGH: when the task depends on an attachment's contents, the agent cannot complete that step through MCP even after it has found the link.

## Current state

`wiki_page_get` returns the Wiki markup in `content` when the upstream response includes it. For example, the markup can contain a `:file[…]` reference with a path of the form `team/project/.files/report.csv`. MCP passes the reference through as part of the page text.

## Desired state

An agent with a Wiki attachment link can obtain the file through MCP when it has read access to the page that owns the file.

## Problem boundary

The problem covers Wiki attachments whose links the agent already has from page content. Finding attachments through the page resource list and handling Tracker attachments are outside this boundary.

## Assumptions

No assumptions.

## Open questions

None about the problem.
