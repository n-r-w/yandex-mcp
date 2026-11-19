# Technical Solution: Download a Wiki attachment from its link

## Problem Statement

An agent can find an attachment link in a Wiki page but cannot retrieve the file through MCP. See the [problem statement](problem.md) and [requirements and acceptance scenarios](requirements.md).

## Proposed Solution

### Retrieve by attachment path

Add `wiki_attachment_get` for a Wiki attachment path copied from the page text. The tool reads from Wiki and writes to the MCP server's filesystem only when the caller requests disk delivery. The caller passes the path copied from the link in `attachment_path`. The handler accepts `slug/.files/filename`, optionally preceded by one slash. It rejects malformed paths and absolute URLs. The path identifies the attachment on the configured Wiki API origin.

The Wiki adapter calls the documented `GET /v1/pages/attachments/download_by_url` operation on the configured Wiki API origin. The handler removes an optional leading slash and preserves the rest of the copied path. For example, `/team/project/.files/report.csv` produces the query value `team/project/.files/report.csv`. The adapter encodes that value once for the HTTP query and sets `download=true`. MCP leaves interpretation of the copied spelling to Wiki. Local percent-decoding or retrying under a different spelling could select a different filename. The adapter uses the current Wiki token and organization settings; Wiki makes the authorization decision. The adapter exposes the successful download body as a stream to its consumer, through the existing consumer-owned `IWikiAdapter` interface. The tool performs no resource-list or page-ID lookup. The download-by-ID operation would require extra input the page link does not supply. Fetching the link as an HTTP URL would give the caller control of the destination and is excluded.

The shared HTTP client executes the Wiki download through a download-specific request path. This path reuses request authentication and token refresh, disables HTTP redirects at execution, and makes the final HTTP status available to the Wiki adapter before any body read. Existing HTTP methods and their callers keep their behavior. The Wiki adapter exposes a body as attachment data only for HTTP 200. It maps a final 401 or 403 to a safe access error independently of the body or the inline size cap. Errors expose no upstream debugging bodies or credentials. An access denial yields no data or disk write.

An authorized API check confirmed the download flow for a CSV attachment. The Wiki API returned the page text. `download_by_url` accepted the attachment path from that text and returned HTTP 200 without a redirect. The response bytes matched the download by attachment ID.

### Deliver in the mode the agent requests

Reuse the Tracker delivery options: `get_content: true` requests text and `save_path` requests disk delivery. The caller selects exactly one. In disk mode, `override: true` permits overwriting. A distinct mode selector would add a second vocabulary for the same operation. Both modes use the same Wiki download. The adapter transfers bytes, and the tool boundary delivers them in the requested form.

- In the text mode, the Wiki tool applies the existing configured text extension policy to the attachment filename and bounds the bytes read by `YANDEX_MCP_ATTACH_INLINE_MAX_BYTES` (default: 10 MiB). The tool returns the complete text in `content` only when the bytes are valid UTF-8. A successful empty text file has `content: ""` and a size of zero bytes. If the file exceeds the cap or contains invalid UTF-8, the tool reports the reason and returns no partial contents. Disk writes require a separate, explicit call. For example, a 12 MiB CSV at the default limit yields a size error in text mode and can be saved by an explicit disk request.
- In disk mode, the caller provides an absolute destination on the machine running MCP. The tool checks the destination using Tracker's existing extension and resolved-directory rules before requesting the download. Use a destination that only trusted users can change. Keep its path components stable during the call, as the directory check takes place before the writer opens the file. A guarantee against concurrent path changes would require a different writer and is outside this change. After Wiki authorizes the download, the shared writer streams directly to the destination without applying the inline limit. The writer opens an existing file for replacement only when the caller explicitly permits overwrite. If `report.csv` is a symbolic link to an existing `archive.csv` and the path passes the directory checks, an authorized overwrite updates `archive.csv` and leaves the link in place. The writer uses the path the caller supplied for this operation. Replacing the link itself would break the alias, and rejecting an otherwise permitted link would differ from Tracker.

  MCP returns `saved_path` and the byte count after a successful write. If copying fails, MCP returns an error. Bytes already written can remain at the destination, as in Tracker. Staging a temporary file for later replacement would add a second write strategy and is outside this change. The agent must be able to read the destination to use the saved file.

Put shared attachment delivery in a focused package, `internal/tools/attachments`, at the tool boundary. Move the existing Tracker extension and destination rules and stream writer into that package. Both tools use that implementation. Tracker keeps its inputs, outputs, errors, configuration and write behavior. The Wiki tool owns Wiki link parsing and its MCP responses, and the Wiki adapter owns upstream status mapping. The shared package depends on neither tool package nor an upstream adapter. Sharing the implementation avoids different write rules for two attachment tools. The cost is a behavior-preserving change to Tracker call sites and verification against their existing tests.

At startup, `runMCP` assembles and injects the attachment settings from the existing configuration. The delivery component receives the allowed text extensions, destination extensions and directories. The Wiki tool receives the existing inline byte cap. The tools and the shared package do not load environment variables.

The explicit modes are the agreed choice. Disk-only delivery would require filesystem access for small text files. Text-only delivery would exclude files that exceed the limit or cannot be interpreted as text. Automatically switching to disk would add an unexpected write and still require a usable destination. An agent who requests text for an unsuitable file must make a second call to save it.

### Verification

For SCN.01, verify the MCP tool against a Wiki HTTP fixture that serves known bytes for the documented download operation under the expected Wiki authentication. Check complete UTF-8 text in text mode and byte-for-byte file contents in disk mode, including a file above the inline limit. In the target deployment, verify that MCP can write to an allowed destination and that the agent can read the saved bytes. Check that neither mode requires a resource-list call. For SCN.02, verify that a final access-denied status produces a safe denial in both modes, even when the error body exceeds the configured inline cap. Neither mode exposes file contents or creates a destination on denial.

Use boundary checks to verify that a redirect target receives no request and HTTP 206 exposes no attachment stream. Verify that the shared writer requires explicit permission to overwrite, preserves a permitted symbolic link when overwriting its target, and reports a failed copy as an error. A pre-download existence check alone would allow an unapproved overwrite if a file appears during the download. Preserve Tracker's existing no-overwrite rule when the writer opens the destination. Check the Wiki tool with non-default attachment settings and run existing Tracker delivery tests.

## Overengineering and Overspecification Considerations

Use the existing Wiki tool and adapter boundaries, authorization, attachment configuration and disk writer. A dedicated metadata call, cache, automatic fallback, temporary download staging, new configuration or new dependencies are outside this change. No Tracker feature, Wiki resource-list or Wiki grids behavior changes.

## Open Questions

None.

## References

- [Tracker tool reference](../../tracker-tools.md): attachment delivery modes and write scope.
- [Yandex Wiki download-by-path API](https://yandex.ru/support/wiki/ru/api-ref/attachments/pagesattachments__download_by_filename_slug_pair): documented request, read-access condition and HTTP 200 response.
