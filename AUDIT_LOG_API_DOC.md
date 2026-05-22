# Administrative Audit Log API Documentation

This document outlines the API endpoints available for the Blood Bank Admin to manage and view audit logs. All endpoints require the user to be authenticated and have the `Blood Bank Admin` role.

### Base Path
`GET /api/bloodbankadmin/audit-logs`

---

## 1. Get All Audit Logs (with Filtering & Pagination)

Retrieves a paginated list of audit logs. Supports filtering by various parameters to help the admin track specific actions or timeframes.

**Endpoint:** `GET /api/bloodbankadmin/audit-logs`

**Query Parameters:**
| Parameter   | Type     | Description                                                                 |
|-------------|----------|-----------------------------------------------------------------------------|
| `page`      | Integer  | The page number to retrieve (default: 1).                                   |
| `limit`     | Integer  | The number of items per page (default: 20).                                 |
| `action`    | String   | Filter by specific action (e.g., `CREATE_CAMPAIGN`, `REGISTER_LAB_TECH`).   |
| `target_type`| String  | Filter by the type of target affected (e.g., `campaigns`, `users`).         |
| `start_date`| String   | Filter logs from this date (format: YYYY-MM-DD).                            |
| `end_date`  | String   | Filter logs up to this date (format: YYYY-MM-DD).                           |

**Response (200 OK):**
```json
{
  "total": 45,
  "page": 1,
  "limit": 20,
  "logs": [
    {
      "log_id": "c1234567-89ab-cdef-0123-456789abcdef",
      "admin_id": "a1234567-89ab-cdef-0123-456789abcdef",
      "action": "CREATE_CAMPAIGN",
      "target_type": "campaigns",
      "target_id": "t1234567-89ab-cdef-0123-456789abcdef",
      "details": "Created campaign: Summer Blood Drive",
      "created_at": "2026-05-22T10:00:00Z"
    }
  ]
}
```

---

## 2. Get Audit Log by ID

Retrieves the full details of a specific audit log entry by its unique identifier.

**Endpoint:** `GET /api/bloodbankadmin/audit-logs/:id`

**Path Parameters:**
| Parameter | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `id`      | String | The unique ID of the audit log.   |

**Response (200 OK):**
```json
{
  "log_id": "c1234567-89ab-cdef-0123-456789abcdef",
  "admin_id": "a1234567-89ab-cdef-0123-456789abcdef",
  "action": "CREATE_CAMPAIGN",
  "target_type": "campaigns",
  "target_id": "t1234567-89ab-cdef-0123-456789abcdef",
  "details": "Created campaign: Summer Blood Drive",
  "created_at": "2026-05-22T10:00:00Z"
}
```

**Response (404 Not Found):**
```json
{
  "error": "audit log not found"
}
```

---

## 3. Export Audit Logs

Exports the audit logs as a downloadable CSV file. This endpoint respects filtering parameters but ignores pagination limits to allow downloading the full dataset of matching logs. It is highly useful for administrators who want to open the logs in spreadsheet software like Excel.

**Endpoint:** `GET /api/bloodbankadmin/audit-logs/export`

**Query Parameters:**
| Parameter   | Type     | Description                                                                 |
|-------------|----------|-----------------------------------------------------------------------------|
| `action`    | String   | Filter by specific action (e.g., `CREATE_CAMPAIGN`).                        |
| `target_type`| String  | Filter by the type of target affected (e.g., `campaigns`).                  |
| `start_date`| String   | Filter logs from this date (format: YYYY-MM-DD).                            |
| `end_date`  | String   | Filter logs up to this date (format: YYYY-MM-DD).                           |

**Response (200 OK):**
- **Headers:** 
  - `Content-Type: text/csv`
  - `Content-Disposition: attachment; filename=audit_logs.csv`
- **Body:** CSV file contents.

```csv
Log ID,Admin ID,Action,Target Type,Target ID,Details,Created At
c1234567-89ab-cdef-0123-456789abcdef,a1234567-89ab-cdef-0123-456789abcdef,CREATE_CAMPAIGN,campaigns,t1234567-89ab-cdef-0123-456789abcdef,Created campaign: Summer Blood Drive,2026-05-22 10:00:00
```

---

## 4. Delete Audit Log

Deletes a specific audit log entry from the system. This allows the admin to prune old or expired log records manually.

**Endpoint:** `DELETE /api/bloodbankadmin/audit-logs/:id`

**Path Parameters:**
| Parameter | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `id`      | String | The unique ID of the audit log.   |

**Response (200 OK):**
```json
{
  "message": "Audit log deleted successfully"
}
```

**Response (404 Not Found):**
```json
{
  "error": "audit log not found"
}
```

---

## Common Tracked Actions & Targets
The system automatically logs the following common administrative actions:

| Action Category                 | Target Type (`target_type`) | Common Actions (`action`)                                     |
|---------------------------------|-----------------------------|---------------------------------------------------------------|
| **Campaigns**                   | `campaigns`                 | `CREATE_CAMPAIGN`, `UPDATE_CAMPAIGN`, `DELETE_CAMPAIGN`       |
| **Inventory**                   | `blood_units`               | `MARK_UNIT_USED`, `DELETE_UNIT`                               |
| **Staff Registration**          | `users`                     | `REGISTER_LAB_TECH`, `REGISTER_BLOOD_COLLECTOR`               |
| **Hospital Requests (Reg.)**    | `hospital_requests`         | `APPROVE_HOSPITAL_REQUEST`, `REJECT_HOSPITAL_REQUEST`         |
| **Hospital Contracts**          | `hospital_contracts`        | `SIGN_CONTRACT`, `REJECT_CONTRACT`                            |
| **Contract Templates**          | `contract_templates`        | `CREATE_CONTRACT_TEMPLATE`, `UPDATE_CONTRACT_TEMPLATE`, `DELETE_CONTRACT_TEMPLATE` |
| **Emergency Blood Requests**    | `emergency_requests`        | `CREATE_MANUAL_EMERGENCY`, `PUBLISH_EMERGENCY`, `REJECT_EMERGENCY` |
| **Donor Blood Requests**        | `donor_blood_requests`      | `APPROVE_DONOR_BLOOD_REQUEST`, `REJECT_DONOR_BLOOD_REQUEST`, `FULFILL_DONOR_BLOOD_REQUEST` |
| **Hospital Blood Requests**     | `blood_requests`            | `APPROVE_BLOOD_REQUEST`, `REJECT_BLOOD_REQUEST`               |
