# Idea: Download a Wiki attachment from its link

## Definitions

See [key definitions and abbreviations](problem.md#key-definitions-and-abbreviations) in the problem statement.

## Context and Problem

See the [problem statement](problem.md).

## Goal

An agent can work with the complete contents of an attachment linked from a Yandex Wiki page.

## Affected Scenarios

An agent has an attachment link from the text of a Wiki page and requests the file through MCP.

## Scope and Non-Scope

The scope covers retrieving Wiki attachments by a known link and reporting denied access. Finding attachments through the page resource list and handling Tracker attachments are outside the scope.

## Requirements

### Functional Requirements

- [X] FRQ.01: When an agent passes a link to an attachment from Wiki page text and Wiki permits the download, MCP makes the complete file contents available to the agent. The link is sufficient to request the file without listing page resources first. If Wiki denies access, MCP reports the denial and exposes none of the file contents.
  - Origin: PRM.01.
  - Goal: The agent can work with an attachment after reading the page that links to it.
  - Feasibility: Confirmed. The [Yandex Wiki download API](https://yandex.ru/support/wiki/ru/api-ref/attachments/pagesattachments__download_by_filename_slug_pair) supports retrieval by attachment path and requires read access to the page that owns the file.

### Non-Functional Requirements

No separate non-functional requirements.

## Justification of Requirements

| What does not work now | Requirements that fix it |
|---|---|
| PRM.01: The agent finds the attachment link in the Wiki page but cannot retrieve the file through MCP. | FRQ.01: The agent obtains the complete file from the known link when Wiki permits the download. |

Sufficiency: FRQ.01 covers the blocked retrieval step.

Necessity: Without FRQ.01, obtaining the link does not let the agent retrieve the file through MCP.

## Critical Scenarios

- SCN.01: Wiki permits the download.
  - Requirements: FRQ.01, permitted download by known link.
  - Level: external.
  - Area: agent, MCP, Yandex Wiki.
  - Flow: The agent finds the attachment link in the page text and passes it to MCP. The agent obtains the complete file without requesting the page resource list.
- SCN.02: Wiki denies access.
  - Requirements: FRQ.01, denial without file contents.
  - Level: external.
  - Area: agent, MCP, Yandex Wiki.
  - Flow: The agent passes a known attachment link to MCP. Wiki denies access. MCP reports the denial to the agent and exposes none of the file contents.

## Open Questions

No open requirements questions. The technical solution will define how MCP delivers the file to the agent.

## Technical Supplement

No technical supplement.

## References

- [Problem statement](problem.md): current behavior and the problem boundary.
