# Donor Blood Request API Documentation

**Base URL:** `http://localhost:8080`

> **Authentication:** All endpoints require a `Bearer` token in the `Authorization` header.
> Donor endpoints require `RoleDonor`. Admin endpoints require `RoleBloodBankAdmin`.

---

## Business Rules Summary

| Rule | Detail |
|---|---|
| **Eligibility** | Donor must be in the **Top 10** of the leaderboard (ranked by CLEARED test results) |
| **Cooldown** | A donor can only make **1 request per 3 months** (90 days) |
| **Component Type** | Donor specifies the blood component type they need (e.g., `PRBC`, `WHOLE_BLOOD`) |
| **Units** | Requests are made by **number of units**, not by volume (ml) |
| **Partial Approval** | If only some units are available, request is `PARTIALLY APPROVED` |
| **Inventory Check** | Admin approval reserves matching units by blood type **AND** component type |

---

## DONOR ENDPOINTS

### 1. POST `/api/donor/blood-request` — Create a Blood Request

**Role:** `Donor`
**Description:** Creates a new blood request. The donor must be in the top 10 leaderboard and must not have made another request within the last 3 months.

**Request Body:**
```json
{
  "units": 2,
  "component_type": "PRBC",
  "reason": "Post-surgery recovery",
  "hospital_name": "St. Paul Hospital",
  "hospital_address": "Addis Ababa, Arada Sub-city",
  "hospital_phone": "0912345678"
}
```

**Fields:**
| Field | Type | Required | Description |
|---|---|---|---|
| `units` | `int` | ✅ | Number of blood units needed (e.g., `2`) |
| `component_type` | `string` | ✅ | Type of component: `WHOLE_BLOOD`, `PRBC`, `PLATELETS`, `PLASMA`, `CRYOPRECIPITATE`, `CRYO_POOR_PLASMA` |
| `reason` | `string` | ✅ | Medical reason for the request |
| `hospital_name` | `string` | ✅ | Name of the hospital where blood is needed |
| `hospital_address` | `string` | ✅ | Address of the hospital |
| `hospital_phone` | `string` | ✅ | Contact number for the hospital |

> **Note:** `blood_type` is automatically taken from the donor's registered profile. The donor does not need to specify it.

**Success Response `201 Created`:**
```json
{
  "message": "Blood request created successfully",
  "data": {
    "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "donor_id": "d5e6f789-...",
    "donor_name": "Abebe Kebede",
    "donor_email": "abebe@email.com",
    "donor_phone": "0911223344",
    "donor_address": "Addis Ababa",
    "blood_type": "O+",
    "component_type": "PRBC",
    "units": 2,
    "reason": "Post-surgery recovery",
    "hospital_name": "St. Paul Hospital",
    "hospital_address": "Addis Ababa, Arada Sub-city",
    "hospital_phone": "0912345678",
    "status": "PENDING",
    "created_at": "2026-05-16T05:23:00Z",
    "can_fulfill": false
  }
}
```

**Error Responses:**

`403 Forbidden` — Donor is NOT in the top 10 leaderboard:
```json
{
  "error": "Only top 10 leaderboard donors are eligible to request blood"
}
```

`403 Forbidden` — Donor requested within the last 3 months:
```json
{
  "error": "You can only request blood once every 3 months"
}
```

`500 Internal Server Error` — General error:
```json
{
  "error": "..."
}
```

---

### 2. GET `/api/donor/my-requests` — Get My Requests

**Role:** `Donor`
**Description:** Returns a list of all blood requests made by the authenticated donor, with optional filters.

**Query Parameters:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `status` | `string` | ❌ | Filter by status: `PENDING`, `APPROVED`, `PARTIALLY APPROVED`, `REJECTED`, `FULFILLED`, `PARTIALLY FULFILLED` |
| `start_date` | `string` | ❌ | Filter requests created on or after this date (`YYYY-MM-DD`) |
| `end_date` | `string` | ❌ | Filter requests created on or before this date (`YYYY-MM-DD`) |

**Example Request:**
```
GET /api/donor/my-requests?status=APPROVED
```

**Success Response `200 OK`:**
```json
{
  "message": "My requests fetched successfully",
  "data": [
    {
      "request_id": "a1b2c3d4-...",
      "donor_id": "d5e6f7...",
      "donor_name": "Abebe Kebede",
      "donor_email": "abebe@email.com",
      "donor_phone": "0911223344",
      "donor_address": "Addis Ababa",
      "blood_type": "O+",
      "component_type": "PRBC",
      "units": 2,
      "reason": "Post-surgery recovery",
      "hospital_name": "St. Paul Hospital",
      "hospital_address": "Addis Ababa, Arada Sub-city",
      "hospital_phone": "0912345678",
      "status": "APPROVED",
      "created_at": "2026-05-16T04:50:00Z",
      "can_fulfill": true
    }
  ]
}
```

**`can_fulfill` field:** `true` if status is `APPROVED` or `PARTIALLY APPROVED`. The frontend can use this to show a "Collect" or "Pickup" action to the donor.

**Empty response (no requests yet):**
```json
{
  "message": "My requests fetched successfully",
  "data": []
}
```

---

## BLOOD BANK ADMIN ENDPOINTS

### 3. GET `/api/admin/blood-requests` — Get All Donor Requests

**Role:** `BloodBankAdmin`
**Description:** Returns all donor blood requests, sorted by the donor's number of successful donations (descending). This allows the admin to prioritize loyal, high-donating donors. Supports filters.

**Query Parameters:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `blood_type` | `string` | ❌ | Filter by blood type: `A+`, `A-`, `B+`, `B-`, `AB+`, `AB-`, `O+`, `O-` |
| `status` | `string` | ❌ | Filter by status: `PENDING`, `APPROVED`, `PARTIALLY APPROVED`, `REJECTED`, `FULFILLED`, `PARTIALLY FULFILLED` |
| `start_date` | `string` | ❌ | Filter by `created_at` >= date (`YYYY-MM-DD`) |
| `end_date` | `string` | ❌ | Filter by `created_at` <= date (`YYYY-MM-DD`) |

**Example Request:**
```
GET /api/admin/blood-requests?status=PENDING&blood_type=O+
```

**Success Response `200 OK`:**
```json
{
  "message": "All requests fetched successfully",
  "data": [
    {
      "request_id": "a1b2c3d4-...",
      "donor_id": "d5e6f7...",
      "donor_name": "Abebe Kebede",
      "donor_email": "abebe@email.com",
      "donor_phone": "0911223344",
      "donor_address": "Addis Ababa",
      "blood_type": "O+",
      "component_type": "PRBC",
      "units": 2,
      "reason": "Post-surgery recovery",
      "hospital_name": "St. Paul Hospital",
      "hospital_address": "Addis Ababa, Arada Sub-city",
      "hospital_phone": "0912345678",
      "status": "PENDING",
      "created_at": "2026-05-16T04:50:00Z",
      "can_fulfill": false,
      "successful_donations": 8
    }
  ]
}
```

> **`successful_donations`** is the count of CLEARED test results for that donor. The list is sorted highest first, meaning the most loyal donors appear at the top.

---

### 4. POST `/api/admin/blood-requests/:id/approve` — Approve a Request

**Role:** `BloodBankAdmin`
**Description:** Attempts to reserve blood units from inventory for the given request. The system searches for `AVAILABLE` units that match both the donor's `blood_type` AND `component_type`, ordered by soonest expiry first. The request status is automatically set based on availability.

**URL Parameter:** `:id` = `request_id`

**Request Body:** None (no body required)

**Status Outcomes:**

| Scenario | Status Set | Message Returned |
|---|---|---|
| 0 units available | `REJECTED` | `"no enough blood in the inventory"` |
| Some units available (< requested) | `PARTIALLY APPROVED` | `"partially approved"` |
| All requested units available | `APPROVED` | `"fully approved"` |

**Success Response `200 OK` (all units available):**
```json
{
  "message": "fully approved",
  "data": {
    "request_id": "a1b2c3d4-...",
    "donor_name": "Abebe Kebede",
    "blood_type": "O+",
    "component_type": "PRBC",
    "units": 2,
    "status": "APPROVED",
    "can_fulfill": true,
    ...
  }
}
```

**Success Response `200 OK` (partial units available):**
```json
{
  "message": "partially approved",
  "data": {
    "request_id": "a1b2c3d4-...",
    "donor_name": "Abebe Kebede",
    "blood_type": "O+",
    "component_type": "PRBC",
    "units": 3,
    "reserved_units": 2,
    "status": "PARTIALLY APPROVED",
    "can_fulfill": true,
    ...
  }
}
```

> **`reserved_units`** indicates how many units were actually found and reserved in inventory. In this example, the donor requested 3 units but only 2 were available, so `reserved_units: 2`. This field is only present when units were reserved (non-zero).

**Success Response `200 OK` (no units available):**
```json
{
  "message": "no enough blood in the inventory",
  "data": {
    "request_id": "a1b2c3d4-...",
    "status": "REJECTED",
    "can_fulfill": false,
    ...
  }
}
```

**Error Response `400 Bad Request`** — Already processed:
```json
{
  "error": "request already processed"
}
```

---

### 5. POST `/api/admin/blood-requests/:id/reject` — Manually Reject a Request

**Role:** `BloodBankAdmin`
**Description:** Manually rejects a pending request. Cannot reject a request that is already fulfilled.

**URL Parameter:** `:id` = `request_id`

**Request Body:** None

**Success Response `200 OK`:**
```json
{
  "message": "Request rejected successfully",
  "data": {
    "request_id": "a1b2c3d4-...",
    "status": "REJECTED",
    "can_fulfill": false,
    ...
  }
}
```

**Error Responses:**

`400 Bad Request` — Already fulfilled:
```json
{
  "error": "cannot reject a fulfilled request"
}
```

`400 Bad Request` — Already rejected:
```json
{
  "error": "request already rejected"
}
```

---

## Request Status Flow

```
                    ┌──────────────────────────────────────┐
                    │           DONOR SUBMITS               │
                    └─────────────┬────────────────────────┘
                                  │
                                  ▼
                             [ PENDING ]
                                  │
                  ┌───────────────┼──────────────────┐
                  ▼               ▼                   ▼
            (admin rejects)  (0 units)     (some/all units found)
                  │               │                   │
             [REJECTED]      [REJECTED]    ┌──────────┴───────────┐
                                           ▼                       ▼
                                    [PARTIALLY             [APPROVED]
                                     APPROVED]                     │
                                           │                       │
                                           └──────────┬────────────┘
                                                      │
                                              (admin clicks fulfill)
                                                      │
                                         ┌────────────┴────────────┐
                                         ▼                         ▼
                                [PARTIALLY FULFILLED]         [FULFILLED]
```

---

## Component Types Reference

| Value | Description |
|---|---|
| `WHOLE_BLOOD` | Unprocessed whole blood |
| `PRBC` | Packed Red Blood Cells |
| `PLATELETS` | Platelet concentrate |
| `PLASMA` | Fresh Frozen Plasma |
| `CRYOPRECIPITATE` | Clotting factor concentrate |
| `CRYO_POOR_PLASMA` | Cryo-poor plasma (post-cryo extraction) |
