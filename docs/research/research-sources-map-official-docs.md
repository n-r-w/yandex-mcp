# Official Documentation Sources Map

**Related memory:** 01-research-segmentation-overview

**Purpose:** Canonical entrypoints and key subpages for Yandex Wiki API, Yandex Tracker API, and IAM authentication documentation (English).

## Yandex Wiki API

### Entry Points
- **Main API Overview** - https://yandex.ru/support/wiki/en/api-ref/about
  - Purpose: Describes the Wiki API purpose, design for web services/apps, and user permission model
  - Why relevant: Primary documentation root explaining what the API does and how permissions work

### Key Subpages
- **API Access** - https://yandex.ru/support/wiki/en/api-ref/access
  - Purpose: Complete authentication documentation for OAuth 2.0 and IAM tokens
  - Why relevant: Critical for understanding auth methods (OAuth vs IAM), required headers (Authorization, X-Org-Id/X-Cloud-Org-Id), and base URL (api.wiki.yandex.net)

- **API Usage Examples** - https://yandex.ru/support/wiki/en/api-ref/examples
  - Purpose: Python script examples for page creation, content editing, and dynamic table operations
  - Why relevant: Shows practical usage patterns for both OAuth and IAM token authentication

- **API Reference Index** - https://yandex.ru/support/wiki/en/api-ref/
  - Purpose: Top-level index of all Wiki API reference sections
  - Why relevant: Navigation hub for finding specific endpoints and resources
  - **Note:** Some section titles are in Russian, but the documentation structure is visible

- **Dynamic Tables (Grids)** - https://yandex.ru/support/wiki/en/api-ref/grids/
  - Purpose: Complete endpoint documentation for dynamic table operations
  - Why relevant: Documents 12 grid operations (create, get, update, delete, add/remove rows and columns, update cells, move rows/columns, clone)

- **Pages Section** - https://yandex.ru/support/wiki/en/api-ref/pages/
  - Purpose: Wiki page-related endpoints
  - Why relevant: Covers page-specific operations like getting dynamic tables associated with a page

- **Page Resources** - https://yandex.ru/support/wiki/en/api-ref/pagesresources/
  - Purpose: Documentation for page resource endpoints
  - Why relevant: Covers resource retrieval operations for Wiki pages

## Yandex Tracker API

### Entry Points
- **API Overview** - https://yandex.ru/support/tracker/en/about-api
  - Purpose: Explains Tracker API purpose, capabilities (search/create/edit issues, boards, queue settings), and permission model
  - Why relevant: Primary documentation root; links to common format and access documentation

### Key Subpages
- **API Access** - https://yandex.ru/support/tracker/en/concepts/access
  - Purpose: Authentication documentation for OAuth 2.0 and IAM tokens
  - Why relevant: Details auth methods, required headers (Authorization: OAuth/Bearer, X-Org-ID/X-Cloud-Org-ID), and Python client setup

- **Common Request Format** - https://yandex.ru/support/tracker/en/common-format
  - Purpose: Comprehensive guide to API request structure (methods, resources, headers, body format, pagination)
  - Why relevant: Documents HTTP methods (GET/POST/PATCH/DELETE), API versions (v3 vs v2), pagination (perPage, page), and request body patterns

- **Error Response Codes** - https://yandex.ru/support/tracker/en/error-codes
  - Purpose: Lists all HTTP response codes with explanations
  - Why relevant: Essential for error handling (200/201/204 success, 400/401/403/404/409/412/422/423/428/429 errors)

- **Get Issue** - https://yandex.ru/support/tracker/en/concepts/issues/get-issue
  - Purpose: Documents retrieving issue information via API
  - Why relevant: Shows issue object structure with all fields (self, id, key, version, summary, description, queue, status, assignee, etc.)

- **Create Issue** - https://yandex.ru/support/tracker/en/concepts/issues/create-issue
  - Purpose: Documents creating issues via API
  - Why relevant: Shows required fields and response format (201 Created on success)

- **Developer Tools** - https://yandex.ru/support/tracker/en/user/API
  - Purpose: Hub page linking to API reference and Python client information
  - Why relevant: Provides overview of capabilities and links to yandex_tracker_client Python package

## IAM Token / Authentication

### Entry Points
- **IAM Operations Index** - https://yandex.cloud/en/docs/iam/operations/
  - Purpose: Central hub for all IAM token operations
  - Why relevant: Navigation page for all token-related operations (create, reissue, refresh tokens, API keys)

### Key Subpages
- **Create IAM Token (CLI)** - https://yandex.cloud/en/docs/iam/operations/iam-token/create
  - Purpose: How to create IAM tokens using `yc iam create-token` for Yandex accounts
  - Why relevant: Primary method for obtaining tokens; includes examples and usage patterns

- **IAM Token Concepts** - https://yandex.cloud/en/docs/iam/concepts/authorization/iam-token
  - Purpose: Explains IAM token nature, lifetime (12 hours maximum), and usage patterns
  - Why relevant: Critical for understanding token expiration and refresh requirements

- **Create IAM Token (Service Account)** - https://yandex.cloud/en/docs/iam/operations/iam-token/create-for-sa
  - Purpose: Creating IAM tokens for service accounts using authorized keys
  - Why relevant: Alternative auth method for non-user accounts; covers JWT exchange

- **Create IAM Token (Local Account)** - https://yandex.cloud/en/docs/iam/operations/iam-token/create-for-local
  - Purpose: Simple IAM token creation for local accounts
  - Why relevant: Simplified token creation method for certain account types

- **Authorization with REST APIs** - https://yandex.cloud/en/docs/iam/concepts/authorization/api
  - Purpose: Using IAM tokens with REST APIs (header format, folder/org IDs)
  - Why relevant: Documents `Authorization: Bearer <IAM_token>` header format and context headers (x-folder-id, x-org-id)

## Access Verification

**Access Status (January 2026):**
- All Wiki API pages: Accessible without authorization ✓
- All Tracker API pages: Accessible without authorization ✓
- Yandex Cloud IAM pages: Some pages may redirect to SSO authentication, but content is available via search snippets and alternative sources ✓

**Note:** Yandex Cloud documentation (yandex.cloud) may trigger authentication redirects in some regions/browsers. If access issues occur, the information is consistently documented across multiple Yandex properties and can be retrieved via official search or GitHub mirrors.

## API Base URLs

From the documentation:
- **Yandex Wiki API:** `https://api.wiki.yandex.net` (v1)
- **Yandex Tracker API:** `https://api.tracker.yandex.net` (v3 recommended)
- **Yandex OAuth:** `https://oauth.yandex.com` or `https://oauth.yandex.ru`

## Authentication Headers Summary

**Wiki API:**
- OAuth: `Authorization: OAuth <OAuth_token>` + `X-Org-Id` or `X-Cloud-Org-Id`
- IAM: `Authorization: Bearer <IAM_token>` + `X-Org-Id` or `X-Cloud-Org-Id`

**Tracker API:**
- OAuth: `Authorization: OAuth <OAuth_token>` + `X-Org-ID` or `X-Cloud-Org-ID`
- IAM: `Authorization: Bearer <IAM_token>` + `X-Org-ID` or `X-Cloud-Org-ID`
