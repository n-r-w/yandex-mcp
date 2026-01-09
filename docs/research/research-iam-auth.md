# IAM Token Authentication - Yandex Wiki & Tracker APIs

**Related memories:**
- 02-research-sources-map-official-docs
- 03-research-wiki-capabilities-raw
- 04-research-tracker-capabilities-raw

**Purpose:** Extracted IAM token authentication requirements for Yandex Wiki and Tracker APIs from official documentation.

---

## Summary

* Both Yandex Wiki and Tracker APIs support IAM token authentication for Yandex Cloud organizations only
* IAM tokens are valid for maximum 12 hours
* Both services require organization ID headers alongside the Authorization header
* OAuth 2.0 is an alternative authentication method for both services (works in both Yandex 360 and Yandex Cloud organizations)
* Service accounts are NOT supported for Wiki API but ARE supported for Tracker API (with prior approval from Yandex support)

---

## Yandex Wiki API Authentication

### IAM Token Support

* **Organization type:** Yandex Cloud organizations only
* **Cannot use service accounts** - must send requests from user accounts only
* **Source:** https://yandex.ru/support/wiki/en/api-ref/access#iam-token

### Authorization Header Format

**For IAM tokens:**
```
Authorization: Bearer <IAM_token>
```

**For OAuth tokens:**
```
Authorization: OAuth <OAuth_token>
```

**Source:** https://yandex.ru/support/wiki/en/api-ref/access

### Required Headers

All Wiki API requests must include:

1. **Host header:**
   ```
   Host: api.wiki.yandex.net
   ```

2. **Authorization header:**
   * IAM: `Authorization: Bearer <IAM_token>`
   * OAuth: `Authorization: OAuth <OAuth_token>`

3. **Organization ID header** (one of the following):
   * `X-Org-Id: <organization_ID>` - for Yandex 360 for Business organizations
   * `X-Cloud-Org-Id: <organization_ID>` - for Yandex Cloud Organization organizations

**Source:** https://yandex.ru/support/wiki/en/api-ref/access

### Example Request

```
Host: api.wiki.yandex.net
Authorization: Bearer t1.ab123cd45*****************
X-Org-Id: bpfv7***************
```

**Source:** https://yandex.ru/support/wiki/en/api-ref/access

### OAuth Token Requirements

When using OAuth 2.0:
* Permissions needed:
  * `wiki:write` - all operations (create, delete, edit)
  * `wiki:read` - read-only access

**Source:** https://yandex.ru/support/wiki/en/api-ref/access#about_OAuth

---

## Yandex Tracker API Authentication

### IAM Token Support

* **Organization type:** Yandex Cloud organizations only
* **Service accounts:** Supported but require prior approval from Yandex support (provide organization ID and service account ID)
* **Without approval:** API requests return 401 Unauthorized
* **Source:** https://yandex.ru/support/tracker/en/concepts/access#iam-token

### Authorization Header Format

**For IAM tokens:**
```
Authorization: Bearer <IAM_token>
```

**For OAuth tokens:**
```
Authorization: OAuth <OAuth_token>
```

**Source:** https://yandex.ru/support/tracker/en/concepts/access, https://yandex.ru/support/tracker/en/common-format#headings

### Required Headers

All Tracker API requests must include:

1. **Host header:**
   ```
   Host: api.tracker.yandex.net
   ```

2. **Authorization header:**
   * IAM: `Authorization: Bearer <IAM_token>`
   * OAuth: `Authorization: OAuth <OAuth_token>`

3. **Organization ID header** (one of the following):
   * `X-Org-ID: <organization_ID>` - for Yandex 360 for Business organizations
   * `X-Cloud-Org-ID: <organization_ID>` - for Yandex Cloud Organization organizations

**Source:** https://yandex.ru/support/tracker/en/common-format#headings

### Example Request

**Using OAuth:**
```
curl -X GET 'api.tracker.yandex.net/v3/myself' \
     -H 'Authorization: OAuth ABC-def12GH_******' \
     -H 'X-Cloud-Org-Id: abcd12******'
```

**Using IAM:**
```
Authorization: Bearer t1.ab123cd45*****************
X-Org-ID: 1234***
```

**Source:** https://yandex.ru/support/tracker/en/concepts/access, https://yandex.ru/support/tracker/en/common-format#headings

### OAuth Token Requirements

When using OAuth 2.0:
* Permissions needed:
  * `tracker:write` - create, delete, and edit data
  * `tracker:read` - read-only access

**Source:** https://yandex.ru/support/tracker/en/concepts/access#about_OAuth

---

## IAM Token Lifecycle

### Token Lifetime

* **Maximum lifetime:** 12 hours
* **Limited by:** Federation cookie lifetime (for federated accounts)
* **Expiration behavior:** When token expires, API returns 401 Unauthorized error
* **Sources:**
  * Wiki: https://yandex.ru/support/wiki/en/api-ref/access#iam-token
  * Tracker: https://yandex.ru/support/tracker/en/concepts/access#iam-token
  * IAM concepts: https://yandex.cloud/en/docs/iam/concepts/authorization/iam-token (via Perplexity)

### Token Creation

**Using Yandex Cloud CLI:**
```bash
yc iam create-token
```

* Can be used for Yandex accounts or service accounts
* For service accounts, must set appropriate profile first
* **Recommended:** Request tokens frequently (e.g., every hour) to avoid expiration issues
* Previous tokens remain valid until their end time or manual revocation

**Source:** https://yandex.cloud/en/docs/iam/operations/iam-token/create (via Perplexity)

### Token Format

IAM tokens returned by CLI are typically in the format:
```
t1.<random_characters>
```

Example: `t1.ab123cd45*****************`

**Sources:** Wiki and Tracker documentation examples

---

## Authentication Failure Modes

### HTTP Error Codes

**401 Unauthorized:**
* **Wiki:** User is not authorized - token missing, invalid, or expired
* **Tracker:** User is not authorized - token missing, invalid, or expired
* **Wiki source:** https://yandex.ru/support/wiki/en/api-ref/access#about_OAuth
* **Tracker source:** https://yandex.ru/support/tracker/en/error-codes

**403 Forbidden:**
* **Tracker:** Action not authorized - user lacks permission for this specific action
* **Troubleshooting:** Check user rights in Tracker interface (same rights required for API and UI)
* **Source:** https://yandex.ru/support/tracker/en/error-codes

**428 Access Denied:**
* **Tracker:** Access to resource denied - missing required conditions in request
* **Source:** https://yandex.ru/support/tracker/en/error-codes

### Common Causes

1. **Missing or invalid Authorization header**
   * Symptom: 401 Unauthorized
   * Fix: Verify Authorization header format and token validity

2. **Missing organization ID header**
   * Symptom: 401 Unauthorized or 428 Access Denied
   * Fix: Include X-Org-Id/X-Cloud-Org-Id (Wiki) or X-Org-ID/X-Cloud-Org-ID (Tracker)

3. **Expired IAM token**
   * Symptom: 401 Unauthorized
   * Fix: Create new IAM token using `yc iam create-token`

4. **Insufficient permissions**
   * Symptom: 403 Forbidden
   * Fix: Verify user has required permissions in Wiki/Tracker interface

5. **Service account without approval** (Tracker only)
   * Symptom: 401 Unauthorized
   * Fix: Contact Yandex support with organization ID and service account ID

---

## Cross-Service Comparison

### Header Name Differences

**Wiki API:**
* Org headers: `X-Org-Id` or `X-Cloud-Org-Id`

**Tracker API:**
* Org headers: `X-Org-ID` or `X-Cloud-Org-ID` (note: capital "ID")

**Authorization header:** Same format for both services

### Organization Support

| Feature | Yandex Wiki | Yandex Tracker |
|---------|-------------|----------------|
| OAuth 2.0 (Yandex 360) | Yes | Yes |
| OAuth 2.0 (Yandex Cloud) | Yes | Yes |
| IAM Token | Yes (Cloud only) | Yes (Cloud only) |
| Service Accounts | No | Yes (with support approval) |

### Token Lifetime

Both services:
* IAM token: Maximum 12 hours
* OAuth token: Not specified in retrieved documentation (typically longer-lived)
