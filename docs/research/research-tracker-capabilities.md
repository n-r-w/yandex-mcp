# Yandex Tracker API Capabilities (Raw Extraction)

**Related memories:** 01-research-segmentation-overview, 02-research-sources-map-official-docs

**Purpose:** Complete extraction of Yandex Tracker API capabilities from official documentation, focusing on actionable endpoints and operations for MCP tool implementation.

---

## Base Configuration

### API Base URL
- **URL:** `https://api.tracker.yandex.net`
- **Version:** v3 (recommended), v2 (legacy)
- **Source:** https://yandex.ru/support/tracker/en/common-format

### Authentication Requirements

**OAuth 2.0 (Yandex 360 and Yandex Cloud):**
- Header: `Authorization: OAuth <OAuth_token>`
- Header: `X-Org-ID: <organization_ID>` (for Yandex 360)
- Header: `X-Cloud-Org-ID: <organization_ID>` (for Yandex Cloud)
- **Source:** https://yandex.ru/support/tracker/en/concepts/access#about_OAuth

**IAM Token (Yandex Cloud only):**
- Header: `Authorization: Bearer <IAM_token>`
- Header: `X-Cloud-Org-ID: <organization_ID>`
- Token lifetime: Maximum 12 hours
- **Source:** https://yandex.ru/support/tracker/en/concepts/access#iam-token

**Organization ID:** Obtain from Administration → Organizations in Tracker interface

**Permissions:**
- `tracker:read`: Read-only access
- `tracker:write`: Full access (create, delete, edit)
- **Source:** https://yandex.ru/support/tracker/en/concepts/access

---

## HTTP Methods

- **GET:** Retrieve information about objects
- **POST:** Create new objects
- **PATCH:** Edit existing objects (partial updates)
- **DELETE:** Delete objects
- **Source:** https://yandex.ru/support/tracker/en/common-format#methods

---

## Pagination Model

**Standard Pagination (for < 10,000 results):**
- Query parameters: `perPage` (default: 50), `page` (default: 1)
- Response headers: `X-Total-Pages`, `X-Total-Count`
- **Source:** https://yandex.ru/support/tracker/en/common-format#displaying-results

**Scrollable Results (for > 10,000 results):**
- Query parameters:
  - `scrollType`: Scrolling type.
    - Acceptable values:
      - `sorted`: Use sorting specified in the request.
      - `unsorted`: No sorting.
    - Notes:
      - Used only in the **first** request of the scrollable sequence.
      - Not used together with `keys` or `queue` search parameters (use `filter`/`query` instead).
  - `perScroll`: Max issues per response (max: 1,000, default: 100). Used only in the first request.
  - `scrollTTLMillis`: Scroll context lifetime in milliseconds (default: 60,000).
  - `scrollId`: Page ID for 2nd+ requests.
- Response headers: `Link` (pagination links), `X-Scroll-Id`, `X-Scroll-Token`, `X-Total-Count`
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/search-issues#scroll

---

## Error Response Codes

**Success Codes:**
- 200: Request executed successfully
- 201: Object created successfully
- 204: Object deleted successfully

**Client Errors:**
- 400: Invalid parameter values
- 401: User not authorized
- 403: Action not authorized (insufficient permissions)
- 404: Object not found
- 409: Conflict during edit (version mismatch)
- 412: Conflict during edit (invalid version)
- 422: JSON validation error
- 423: Object edits disabled (version limit exceeded: max 10100 for robots, 11100 for users)
- 428: Access denied (missing required conditions)
- 429: Rate limit exceeded

**Source:** https://yandex.ru/support/tracker/en/error-codes

---

## Rate Limits

- **Status:** Rate limit exists (code 429) but specific limits are NOT documented in retrieved pages
- **Recommendation:** Implement retry logic with exponential backoff
- **Source:** https://yandex.ru/support/tracker/en/error-codes (code 429 mentioned)

---

## Capabilities by Area

### 1. Issues (Tasks)

#### 1.1 Get Issue
- **Endpoint:** `GET /v3/issues/<issue_ID>`
- **Required IDs:** Issue ID or key (e.g., "TEST-1")
- **Operation:** Read
- **Parameters:**
  - `expand`: Additional fields to include in the response.
    - Documented values:
      - `attachments`: Attached files.
- **Response:** Full issue object with fields:
  - Basic: `self`, `id`, `key`, `version`, `summary`, `description`
  - Status: `status`, `previousStatus`
  - People: `assignee`, `createdBy`, `updatedBy`, `followers`
  - Classification: `type`, `priority`, `queue`
  - Relations: `parent` (parent issue), `aliases`
  - Projects: `project` (primary and secondary)
  - Sprints: `sprint` array
  - Metadata: `createdAt`, `updatedAt`, `lastCommentUpdatedAt`, `votes`, `favorite`
- **Limitations:** None documented
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-issue

#### 1.2 Create Issue
- **Endpoint:** `POST /v3/issues/`
- **Required Fields:**
  - `summary`: Issue name
  - `queue`: Queue object with `id` or `key`
- **Optional Fields:**
  - `description`: Issue description
  - `parent`: Parent issue ID or key
  - `type`: Issue type (id, key, or name)
  - `priority`: Priority (id, key, or name)
  - `assignee`: User login or ID
  - `attachmentIds`: Array of temporary attachment IDs
  - `tags`: Array of tags
  - `sprint`: Array of sprint IDs
- **Operation:** Write
- **Response:** 201 Created with full issue object
- **Limitations:** None documented
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/create-issue

#### 1.3 Edit Issue
- **Endpoint:** `PATCH /v3/issues/<issue_ID>`
- **Required IDs:** Issue ID or key
- **Operation:** Write
- **Special Features:**
  - Partial updates (only specified fields are changed)
  - Array manipulation with `add`, `remove`, `set`, `replace` commands
  - Field reference by ID, key, or display name
- **Key Operations:**
  - Update fields: `summary`, `description`, `type`, `priority`, etc.
  - Modify arrays: `followers`, `tags` (add/remove items)
  - Update relations: `parent` (change parent issue)
  - Manage projects: `project.primary` sets the main project (request uses a project ID number); `project.secondary.add` adds additional projects (array of project ID numbers). In issue responses, `project.primary` is a project object (`self`, `id`, `display`) and `project.secondary` is an array of project objects.
  - Add to sprints: `sprint` array with IDs
- **Response:** 200 OK with updated issue object
- **Limitations:**
  - Status changes MUST use transition endpoint (not direct PATCH)
  - Version field may be required for concurrent edit protection
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/patch-issue

#### 1.4 Delete Issue
- **Status:** Not supported in Yandex Tracker (issues cannot be deleted).
- **Workarounds:**
  - Close the issue with an appropriate resolution (example: Canceled, Duplicate).
  - If you must remove an issue, move it to a separate queue and delete the queue (this deletes the queue and its issues). The deleted queue (and its issues) can be restored via API.
- **Related API:** `POST /v3/queues/<queue_ID>/_restore`
- **Source:**
  - https://yandex.ru/support/tracker/en/user/ticket-cancel
  - https://yandex.ru/support/tracker/en/concepts/queues/restore-queue

#### 1.5 Search Issues
- **Endpoint:** `POST /v3/issues/_search`
- **Operation:** Read
- **Query Methods:**
  1. **Filter object:** JSON with field-value pairs
     - Example: `{"filter": {"queue": "TREK", "assignee": "empty()"}}`
  2. **Query language:** String-based query syntax
     - Example: `{"query": "epic: notEmpty() Queue: TREK \"Sort by\": Updated DESC"}`
- **Parameters:**
  - `filter`: Field-based filtering
  - `query`: Query language filter
  - `order`: Issue sorting direction and field.
    - Format: `[+/-]<field_key>`.
    - Note: This parameter is only used together with the `filter` parameter. If `query` is used, sorting is configured via the query language.
  - `expand`: Additional fields to include in the response.
    - Acceptable values:
      - `transitions`: Workflow transitions between statuses.
      - `attachments`: Attached files.
  - `scrollType`, `perScroll`, `scrollTTLMillis`: For scrollable results (>10K issues)
- **Response:** Array of issue objects
- **Limitations:**
  - Pagination required for large result sets
  - Scrollable results recommended for >10,000 issues
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/search-issues

#### 1.6 Count Issues
- **Endpoint:** `POST /v3/issues/_count`
- **Operation:** Read
- **Parameters:**
  - `filter`: Field-based filtering (same as search)
  - `query`: Query language filter (same as search)
- **Response:** Plain number (count of matching issues)
- **Example:**
  ```json
  {"filter": {"queue": "JUNE", "assignee": "empty()}}
  ```
  Returns: `5221186`
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/count-issues

#### 1.7 Get Issue Transitions
- **Endpoint:** `GET /v3/issues/<issue_ID>/transitions`
- **Required IDs:** Issue ID or key
- **Operation:** Read
- **Response:** Array of available transitions with:
  - `id`: Transition ID
  - `display`: Transition name (matches UI button)
  - `to`: Target status object (id, key, display)
- **Use Case:** Determine available status transitions before executing
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-transitions

#### 1.8 Execute Status Transition
- **Endpoint:** `POST /v3/issues/<issue_ID>/transitions/<transition_ID>/_execute`
- **Required IDs:** Issue ID/key, Transition ID
- **Operation:** Write
- **Request Body (Optional):**
  - Issue fields that can be edited during transition
  - `comment`: Comment to add with transition
- **Response:** 200 OK with array of available transitions in new status
- **Limitations:** Must use valid transition ID from get-transitions
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/new-transition

#### 1.9 Get Issue Changelog
- **Endpoint:** `GET /v3/issues/<issue_ID>/changelog`
- **Required IDs:** Issue ID or key
- **Operation:** Read
- **Query Parameters:**
  - `perPage`: Number of changelog records per response (default: 50). Use when the issue history has more than 50 records.
- **Response:** Array of changelog entries. Each entry can include:
  - `id`, `self`
  - `issue` (issue reference)
  - `updatedAt`, `updatedBy`
  - `type`: `IssueCreated`, `IssueUpdated`, `IssueWorkflow`
  - `transport`
  - `fields`: array of changed fields (`field`, `from`, `to`)
  - Optional: `comments.added`, `executedTriggers`
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-changelog

#### 1.10 Move Issue to Another Queue
- **Endpoint:** `POST /v3/issues/<issue_ID>/_move?queue=<queue_key>`
- **Required IDs:** Issue ID or key; target queue key (via `queue` query parameter)
- **Operation:** Write
- **Query Parameters:**
  - `queue`: Target queue key.
  - `InitialStatus`: Reset issue status to the initial value when moving.
  - `MoveAllFields`: Attempt to keep components, versions, and projects if the target queue supports the same values.
- **Response:** 200 OK with moved issue object. Includes:
  - `key`: New issue key in the target queue
  - `aliases`: Contains the old key
  - `previousQueue`: Previous queue reference
- **Limitations:**
  - Caller must be allowed to edit the issue and create issues in the target queue.
  - If the issue type and status are missing in the target queue, the move is not performed.
  - Local field values are reset when moving an issue to a different queue.
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/move-issue

#### 1.11 Get Priorities
- **Endpoint:** Not found in the accessible English support docs at time of extraction (the previous page URL returns 404).
- **Observed in issue responses:** `priority` is an object with fields like `self` (example pattern: `https://api.tracker.yandex.net/v3/priorities/<priority_ID>`), `id`, `key`, `display`.
- **Source (404 page):** https://yandex.ru/support/tracker/en/concepts/issues/get-priorities
- **Source (priority object in issue examples):**
  - https://yandex.ru/support/tracker/en/concepts/issues/get-issue
  - https://yandex.ru/support/tracker/en/concepts/issues/patch-issue

---

### 2. Comments

#### 2.1 Add Comment to Issue
- **Endpoint:** `POST /v3/issues/<issue_ID>/comments`
- **Required IDs:** Issue ID or key
- **Required Fields:** `text`: Comment text
- **Optional Fields:**
  - `attachmentIds`: Array of temporary file IDs
  - `summonees`: Array of user IDs/usernames to summon
  - `maillistSummonees`: Array of mailing list addresses
  - `markupType`: Type of text markup.
    - Documented value:
      - `md`: YFM (Yandex Flavored Markdown).
- **Query Parameters:**
  - `isAddToFollowers`: Add commenter to followers (default: true)
- **Operation:** Write
- **Response:** 201 Created with comment object:
  - `id`, `longId`: Comment ID (numeric and string)
  - `text`: Comment text
  - `createdBy`, `updatedBy`: User info
  - `createdAt`, `updatedAt`: Timestamps
  - `version`: Comment version
  - `type`: Comment type:
    - `standard`: Comment sent via the Yandex Tracker interface.
    - `incoming`: Comment created from an incoming message.
    - `outcoming`: Comment created from an outgoing message.
  - `transport`: Method of adding a comment:
    - `internal`: Via the Yandex Tracker interface.
    - `email`: Via email.
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/add-comment

#### 2.2 Edit Comment
- **Endpoint:** `PATCH /v3/issues/<issue_ID>/comments/<comment_ID>`
- **Required IDs:** Issue ID/key, Comment ID (numeric `id` or string `longId`)
- **Required Fields:** `text`: New comment text
- **Optional Fields:** Same as add comment
- **Operation:** Write
- **Response:** 200 OK with updated comment object
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/edit-comment

#### 2.3 Delete Comment
- **Endpoint:** `DELETE /v3/issues/<issue_ID>/comments/<comment_ID>`
- **Required IDs:** Issue ID/key, Comment ID (numeric `id` or string `longId`)
- **Operation:** Write
- **Response:** 204 No Content
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/delete-comment

#### 2.4 Get Issue Comments
- **Endpoint:** `GET /v3/issues/<issue_ID>/comments`
- **Required IDs:** Issue ID or key
- **Operation:** Read
- **Query Parameters:**
  - `expand`: Additional fields to include in the response.
    - Acceptable values:
      - `attachments`: Attached files.
      - `html`: Comment HTML markup.
      - `all`: All additional fields.
  - Pagination:
    - `perPage`: Number of comments per page (default: 50).
    - `id`: Comment numeric `id` value after which the requested page will begin.
- **Pagination headers:** `Link` (first/next)
- **Response:** Array of comment objects
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-comments

---

### 3. Attachments

#### 3.1 Upload Temporary Attachment
- **Endpoint:** `POST /v3/attachments/`
- **Required:** File in multipart/form-data format
- **Operation:** Write
- **Content-Type:** `multipart/form-data`
- **Response:** 201 Created with attachment object:
  - `id`: Unique ID (can be used only once)
  - `name`: Filename
  - `content`: Download URL
  - `thumbnail`: Thumbnail URL (images only)
  - `mimetype`: File type
  - `size`: File size in bytes
  - `metadata`: Additional metadata
- **Important:** The temporary file ID received in response can be used to add an attachment only once.
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/temp-attachment

#### 3.2 Attach a File to an Issue
- **Endpoint:** `POST /v3/issues/<issue_ID>/attachments/`
- **Required IDs:** Issue ID or key
- **Operation:** Write
- **Content-Type:** `multipart/form-data`
- **Request Body:** File contents (multipart)
- **Response:** 201 Created with attachment object
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/post-attachment

#### 3.3 Get Issue Attachments List
- **Endpoint:** `GET /v3/issues/<issue_ID>/attachments`
- **Required IDs:** Issue ID or key
- **Operation:** Read
- **Notes:** The list includes files attached to the issue and to comments below it.
- **Response:** JSON array of attachment objects
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-attachments-list

#### 3.4 Download Attachment
- **Endpoint:** `GET /v3/issues/<issue_ID>/attachments/<file_ID>/<file_name>`
- **Required IDs:** Issue ID or key, File ID, File name
- **Operation:** Read
- **Response:** 200 OK (file contents)
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-attachment

#### 3.5 Download Thumbnail (Images)
- **Endpoint:** `GET /v3/issues/<issue_ID>/thumbnails/<file_ID>`
- **Required IDs:** Issue ID or key, File ID
- **Operation:** Read
- **Response:** 200 OK (thumbnail image)
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-attachment-preview

#### 3.6 Delete Attachment
- **Endpoint:** `DELETE /v3/issues/<issue_ID>/attachments/<file_ID>/`
- **Required IDs:** Issue ID or key, File ID
- **Operation:** Write
- **Response:** 204 No Content
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/delete-attachment

---

### 4. Queues

#### 4.1 List Queues
- **Endpoint:** `GET /v3/queues/`
- **Operation:** Read
- **Query Parameters:**
  - `expand`: Additional fields (projects, components, versions, types, team, workflows)
  - `perPage`: Number of queues per page (default: 50)
- **Response:** Array of queue objects with:
  - Basic: `self`, `id`, `key`, `version`, `name`, `description`
  - Owner: `lead` (user object)
  - Settings: `assignAuto`, `defaultType`, `defaultPriority`, `denyVoting`
  - Team: `teamUsers` array
  - Configuration: `issueTypes`, `versions`, `workflows`, `issueTypesConfig`
- **Pagination:** Standard pagination for >50 queues
- **Source:** https://yandex.ru/support/tracker/en/concepts/queues/get-queues

#### 4.2 Get Queue Details
- **Endpoint:** `GET /v3/queues/<queue_ID>`
- **Required IDs:** Queue ID (queue key in examples)
- **Operation:** Read
- **Query Parameters:**
  - `expand`: Additional fields to include. The documentation example uses `expand=all`.
- **Response:** Queue object (same base fields as queue list, plus additional fields when expanded).
- **Source:** https://yandex.ru/support/tracker/en/concepts/queues/get-queue

#### 4.3 Create/Delete/Restore Queue
- **Create queue endpoint:** `POST /v3/queues/`
  - **Required fields (request body):** `key`, `name`, `lead`, `defaultType`, `defaultPriority`.
  - **Optional/advanced:** `issueTypesConfig` (workflow + resolutions per issue type).
  - **Response:** 201 Created with queue object.
- **Delete queue endpoint:** `DELETE /v3/queues/<queue_ID>`
  - **Response:** 204 No Content.
- **Restore queue endpoint:** `POST /v3/queues/<queue_ID>/_restore`
  - **Response:** 200 OK with queue object.
  - **Limitations:** Can only be performed on behalf of an administrator.
- **Operation:** Write
- **Source:**
  - https://yandex.ru/support/tracker/en/concepts/queues/create-queue
  - https://yandex.ru/support/tracker/en/concepts/queues/delete-queue
  - https://yandex.ru/support/tracker/en/concepts/queues/restore-queue

---

### 5. Users

#### 5.1 Get Current User Info
- **Endpoint:** `GET /v3/myself`
- **Operation:** Read
- **Response:** User object with:
  - IDs: `uid` (default since Oct 2023), `trackerUid`, `passportUid`, `cloudUid`
  - Basic info: `login`, `firstName`, `lastName`, `display`, `email`
  - Status: `hasLicense`, `dismissed`, `external`
  - Settings: `useNewFilters`, `disableNotifications`
  - Dates: `firstLoginDate`, `lastLoginDate`
- **Important:** Default user ID type changed from `passportUid` to `uid` in Oct 2023
- **Source:** https://yandex.ru/support/tracker/en/get-user-info

#### 5.2 List Users / Get User

- **List users endpoint:** `GET /v3/users`
  - **Notes:** The response is paginated (see the common pagination format in `common-format`).
  - **Response:** JSON array of user objects.
  - **Source:** https://yandex.ru/support/tracker/en/get-users
- **Get user endpoint:** `GET /v3/users/<user_login_or_ID>`
  - **Notes:** The path parameter accepts either the user's login or numeric ID.
  - **Response:** User object.
  - **Source:** https://yandex.ru/support/tracker/en/get-user

---

### 6. Issue Relations

#### 6.1 Parent-Child Relations
- **Supported:** Yes - via `parent` field in issue object
- **Get:** Included in issue GET response
- **Set:** Via issue PATCH (set `parent` field)
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-issue, https://yandex.ru/support/tracker/en/concepts/issues/patch-issue

#### 6.2 Issue Links/Dependencies
- **List links:** `GET /v3/issues/<issue_ID>/links`
- **Create link:** `POST /v3/issues/<issue_ID>/links`
- **Delete link:** `DELETE /v3/issues/<issue_ID>/links/<link_ID>`
- **Operation:** Read/Write
- **Request Body (create):**
  - `relationship`: Link type ID.
  - `issue`: Issue ID or key to link to.
- **Example (create):**
  ```json
  {
    "relationship": "relates",
    "issue": "TREK-2"
  }
  ```
- **Response (list/create):** Link object(s) include:
  - `id`, `self`
  - `type`: link type object (`id`, `inward`, `outward`)
  - `direction`: `inward` or `outward`
  - `object`: linked issue reference (`id`, `key`, `display`)
  - `createdBy`, `updatedBy`, `createdAt`, `updatedAt`
- **Source:**
  - https://yandex.ru/support/tracker/en/concepts/issues/link-issue
  - https://yandex.ru/support/tracker/en/concepts/issues/get-links
  - https://yandex.ru/support/tracker/en/concepts/issues/delete-link-issue

---

### 7. Projects

#### 7.1 Get Issue Projects
- **Method:** Included in issue GET response (`GET /v3/issues/<issue_ID>`)
- **Field:** `project` object:
  - `project.primary`: Main project (project object with `self`, `id`, `display`).
  - `project.secondary`: Additional projects (array of project objects with `self`, `id`, `display`).
- **Operation:** Read
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-issue

#### 7.2 Update Issue Projects
- **Method:** Via issue PATCH (`PATCH /v3/issues/<issue_ID>`)
- **Fields (request body):**
  - `project.primary`: Set the main project (project ID number).
  - `project.secondary.add`: Add projects to the additional projects list (array of project ID numbers).
- **Operation:** Write
- **Example:**
  ```json
  {
    "project": {
      "primary": 1234,
      "secondary": {
        "add": [5678]
      }
    }
  }
  ```
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/patch-issue

#### 7.3 Project Entity Comments
- **Endpoint:** `GET /v3/entities/project/<project_ID>/comments`
- **Required IDs:** Project ID (use `id` or `shortId`)
- **Operation:** Read
- **Parameters:**
  - `expand`: Include additional data (all, html, attachments, reactions)
  - Paginated version: `/v3/entities/project/<project_ID>/comments/_relative`
- **Response:** Array of comment objects with reactions, attachments, summonses
- **Source:** https://yandex.ru/support/tracker/en/concepts/entities/comments/get-all-comments

---

### 8. Sprints

#### 8.1 Get Issue Sprints
- **Method:** Included in issue GET response
- **Field:** `sprint` array
- **Operation:** Read
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/get-issue

#### 8.2 Add Issue to Sprint
- **Method:** Via issue PATCH
- **Field:** `sprint` array with sprint IDs
- **Limitations:** Sprints must be on different boards
- **Example:**
  ```json
  {"sprint": [{"id": "3"}, {"id": "2"}]}
  ```
- **Source:** https://yandex.ru/support/tracker/en/concepts/issues/patch-issue

#### 8.3 Boards (Issue boards) and Columns

**Boards:**
- **List boards:** `GET /v3/boards`
- **Get board:** `GET /v3/boards/<board_ID>`
- **Create board:** `POST /v3/boards/`
- **Edit board:** `PATCH /v3/boards/<board_ID>`
  - Requires `If-Match: "<version_number>"` header.
- **Delete board:** `DELETE /v3/boards/<board_ID>`

**Columns (per board):**
- **List columns:** `GET /v3/boards/<board_ID>/columns`

- **Source (boards):**
  - https://yandex.ru/support/tracker/en/get-boards
  - https://yandex.ru/support/tracker/en/get-board
  - https://yandex.ru/support/tracker/en/post-board
  - https://yandex.ru/support/tracker/en/patch-board
  - https://yandex.ru/support/tracker/en/delete-board
- **Source (columns):** https://yandex.ru/support/tracker/en/get-columns


#### 8.4 Sprints (Agile boards)

- **List board sprints:** `GET /v3/boards/<board_ID>/sprints`
- **Get sprint:** `GET /v3/sprints/<sprint_ID>`
- **Create sprint:** `POST /v3/sprints`
  - **Required fields (request body):** `name`, `board.id`, `startDate`, `endDate`
- **Operation:** Read/Write
- **Notes:** The official English docs describe create + get + list operations for sprints. Endpoints for editing or deleting sprints are not present in these API reference pages.
- **Source:**
  - https://yandex.ru/support/tracker/en/get-sprints
  - https://yandex.ru/support/tracker/en/get-sprint
  - https://yandex.ru/support/tracker/en/post-sprint

---

### 9. Other Entities

#### 10.1 Portfolio and Goal Comments
- **Endpoint:** `GET /v3/entities/<entity_type>/<entity_ID>/comments`
- **Entity Types:** project, portfolio, goal
- **Operation:** Read
- **Source:** https://yandex.ru/support/tracker/en/concepts/entities/comments/get-all-comments

#### 10.2 Entity Search

- **Endpoint:** `POST /v3/entities/<entity_type>/_search?fields=<comma_separated_field_keys>`
- **Entity Types:** `project`, `portfolio`, `goal`
- **Operation:** Read
- **Query Parameters:**
  - `fields`: Additional entity parameters to include in the response body (comma-separated field keys).
- **Request Body:**
  - `filter`: Search conditions.
  - Optional sorting:
    - `orderBy`: Field key to sort by.
    - `orderAsc`: Sort direction (`true` for ascending, `false` for descending).
- **Response:** JSON object with:
  - `hits`: Number of entities found
  - `pages`: Number of result pages
  - `values`: Array of entity objects
  - Optional: `orderBy`
- **Source:** https://yandex.ru/support/tracker/en/concepts/entities/search-entities

---

## Request/Response Formats

### Date/Time Format
- All dates in UTC±00:00 timezone
- Format: `YYYY-MM-DDThh:mm:ss.sss±hhmm`
- Example: `2020-11-03T13:24:52.575+0000`
- **Source:** https://yandex.ru/support/tracker/en/common-format

### Text Formatting
- **Markup:** Yandex Flavored Markdown (YFM) for descriptions and comments
- **Line breaks:** Use `\n`
- **Variables:** `{{issue.<field_key>}}`, `{{currentUser}}`, `{{currentDateTime}}`
- **Special characters:** Escape `"`, `\`, `/` with backslash
- **Source:** https://yandex.ru/support/tracker/en/common-format#text-format

### Array Operations
- **Add:** `{"field": {"add": ["value1", "value2"]}}`
- **Remove:** `{"field": {"remove": ["value1"]}}`
- **Set (overwrite):** `{"field": {"set": ["newvalues"]}}`
- **Replace:** `{"field": {"replace": [{"target": "old", "replacement": "new"}]}}`
- **Reset:** `{"field": null}` or `{"field": []}`
- **Source:** https://yandex.ru/support/tracker/en/common-format#edit-fields

### Object References
- Can reference objects by: ID, key, or display name
- Examples for type field:
  - By ID: `{"type": 1}`
  - By key: `{"type": "bug"}`
  - By name: `{"type": {"name": "Error"}}`
  - By ID object: `{"type": {"id": "1"}}`
  - Set command: `{"type": {"set": "bug"}}`
- **Source:** https://yandex.ru/support/tracker/en/common-format#edit-fields

---

## Unclear/Missing Items

The following capabilities/definitions are still missing from this extraction (need authoritative docs pages or confirmed absence):

1. **Link type discovery**: endpoint(s) to list available `relationship` values (link types) not extracted.
2. **Get Priorities**: listing endpoint documentation is missing in accessible English docs; `https://yandex.ru/support/tracker/en/concepts/issues/get-priorities` returns 404.
3. **Rate limits**: HTTP 429 is documented, but specific numeric limits are not present in extracted pages.
