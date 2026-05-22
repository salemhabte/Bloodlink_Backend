# BloodLink — Slot-Based Blood Storage API Documentation

**Version:** 2.0  
**Base URL:** `http://<your-server>/api`  
**Authentication:** All protected endpoints require a `Bearer` JWT token in the `Authorization` header.

---

## Table of Contents

1. [Core Concepts & Terminology](#core-concepts)
2. [Slot Lifecycle Rules](#slot-lifecycle-rules)
3. [Cell Capacity Rules](#cell-capacity-rules)
4. [Endpoint: Submit Test Result (Create)](#1-submit-test-result)
5. [Endpoint: Update Test Result](#2-update-test-result)
6. [Endpoint: Convert Plasma to Cryoprecipitate](#3-convert-plasma-to-cryoprecipitate)
7. [Endpoint: Get All Blood Units](#4-get-all-blood-units)
8. [Endpoint: Get Blood Unit By ID](#5-get-blood-unit-by-id)
9. [Endpoint: Mark Blood Unit as Used](#6-mark-blood-unit-as-used)
10. [Endpoint: Delete Blood Unit](#7-delete-blood-unit)
11. [Error Reference](#error-reference)

---

## Core Concepts

| Term | Definition |
|---|---|
| **Cell** | A unique physical storage compartment identified by the combination of `storage_location` + `rack_number` + `shelf_number`. Example: `Freezer-A / R1 / S2`. |
| **Slot** | A single physical position inside a cell, identified by a `position_number` (e.g., `"P1"`, `"P2"`). Each slot holds exactly **one** blood unit at a time. |
| **Active Unit** | A blood unit with status `AVAILABLE`, `RESERVED`, or `EXPIRED`. Active units **occupy** their slot. |
| **Freed Slot** | A slot is freed when its blood unit is marked `USED` or is deleted. |
| **Cell Capacity** | Each cell supports a maximum of **12 active slots**. |

---

## Slot Lifecycle Rules

```
Blood Unit Created (AVAILABLE)
    └── Slot is OCCUPIED, counts toward 12-slot cell capacity

Blood Unit Status → RESERVED
    └── Slot remains OCCUPIED (reserved for a hospital)

Blood Unit Status → USED
    └── Slot is immediately FREED and available for a new unit

Blood Unit Status → EXPIRED
    └── Slot remains OCCUPIED. Admin must explicitly DELETE the unit to free the slot.

Blood Unit Deleted (by Admin)
    └── Slot is FREED regardless of previous status

Plasma Unit Converted to Cryo
    └── Plasma slot FREED (-1)
    └── Two new units created with their own position numbers (+2)
    └── Net change: +1 unit in the cell
```

---

## Cell Capacity Rules

- A cell holds a maximum of **12 active blood units**.
- `USED` and deleted units do **not** count toward the limit.
- `EXPIRED` units **do** count and hold their slot until explicitly deleted.
- When converting Plasma to Cryo: the plasma frees 1 slot, then 2 new units consume 2 slots. The cell must have at least **1 free slot** before the conversion starts.
- Validation checks happen **before** any data is saved.

---

## 1. Submit Test Result

Records the lab test results for a blood donation. When the result is `CLEARED`, blood components are created and stored in the inventory with precise slot assignments.

### Endpoint

```
POST /api/lab/tests
Authorization: Bearer <LabTechnician JWT>
```

### Request Body

```json
{
  "donation_id":        "string (required)",
  "hiv_result":         "NEGATIVE | POSITIVE (required)",
  "hepatitis_b_result": "NEGATIVE | POSITIVE (required)",
  "hepatitis_c_result": "NEGATIVE | POSITIVE (required)",
  "syphilis_result":    "NEGATIVE | POSITIVE (required)",
  "blood_type":         "A+ | A- | B+ | B- | AB+ | AB- | O+ | O- (required)",
  "overall_status":     "CLEARED | PERMANENTLY_DEFERRED | TEMPORARILY_DEFERRED (required)",

  "storage_location":   "string (required if CLEARED)",
  "rack_number":        "string (required if CLEARED)",
  "shelf_number":       "string (required if CLEARED)",

  "components": [
    {
      "component_type":   "PRBC | PLATELETS | PLASMA | CRYOPRECIPITATE (required if CLEARED)",
      "quantity":         450,
      "position_number":  "string (required if CLEARED, e.g. 'P1')"
    }
  ]
}
```

#### Field Details

| Field | Type | Required | Notes |
|---|---|---|---|
| `donation_id` | string | ✅ Yes | UUID of the donation to be tested |
| `hiv_result` | string | ✅ Yes | `NEGATIVE` or `POSITIVE` |
| `hepatitis_b_result` | string | ✅ Yes | `NEGATIVE` or `POSITIVE` |
| `hepatitis_c_result` | string | ✅ Yes | `NEGATIVE` or `POSITIVE` |
| `syphilis_result` | string | ✅ Yes | `NEGATIVE` or `POSITIVE` |
| `blood_type` | string | ✅ Yes | One of the 8 standard blood types |
| `overall_status` | string | ✅ Yes | `CLEARED`, `PERMANENTLY_DEFERRED`, or `TEMPORARILY_DEFERRED` |
| `storage_location` | string | ⚠️ If CLEARED | Name/ID of the storage unit (e.g., `"Freezer-A"`) |
| `rack_number` | string | ⚠️ If CLEARED | Rack identifier (e.g., `"R1"`) |
| `shelf_number` | string | ⚠️ If CLEARED | Shelf identifier (e.g., `"S2"`) |
| `components` | array | ⚠️ If CLEARED | List of blood components produced |
| `components[].component_type` | string | ⚠️ If CLEARED | Component type |
| `components[].quantity` | int | ⚠️ If CLEARED | Volume in mL |
| `components[].position_number` | string | ⚠️ If CLEARED | Physical slot in the cell (e.g., `"P1"`) |

#### Example Request — CLEARED (3 components)

```json
{
  "donation_id": "a3f5-...",
  "hiv_result": "NEGATIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "A+",
  "overall_status": "CLEARED",
  "storage_location": "Freezer-A",
  "rack_number": "R1",
  "shelf_number": "S2",
  "components": [
    { "component_type": "PRBC",      "quantity": 250, "position_number": "P1" },
    { "component_type": "PLATELETS", "quantity": 60,  "position_number": "P2" },
    { "component_type": "PLASMA",    "quantity": 140, "position_number": "P3" }
  ]
}
```

#### Example Request — PERMANENTLY_DEFERRED

```json
{
  "donation_id": "b7c2-...",
  "hiv_result": "POSITIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "B+",
  "overall_status": "PERMANENTLY_DEFERRED"
}
```

### Success Response — `201 Created`

```json
{
  "message": "Test result submitted successfully"
}
```

### Error Responses

| HTTP | Error Message | Cause |
|---|---|---|
| `400` | `"position_number is required for all components when overall_status is CLEARED"` | A component is missing its `position_number` |
| `400` | `"duplicate position_number 'P1' found in request"` | Two components in the same request share a position |
| `400` | `"Slot [Rack R1, Shelf S2, Pos P1] in Freezer-A is already occupied"` | That position already contains an active blood unit |
| `400` | `"Only 1 positions available in this cell. You are trying to store 3 components."` | Cell does not have enough free slots |
| `400` | `"storage_location is required when overall_status is CLEARED"` | Storage location field missing |
| `400` | `"total component quantity (500 mL) exceeds donation quantity (450 mL)"` | Component volumes exceed the donation volume |
| `400` | `"components must be empty when overall_status is PERMANENTLY_DEFERRED"` | Components sent for a deferred test |
| `400` | `"storage fields must be empty when overall_status is PERMANENTLY_DEFERRED"` | Storage location sent for a deferred test |

---

## 2. Update Test Result

Updates an existing lab test result for a donation. The same slot and capacity validation applies as for Submit. When re-submitting for `CLEARED`, old blood units from the same donation are **automatically deleted** first, so the lab tech may safely reuse the same position numbers.

### Endpoint

```
PUT /api/lab/tests/:donation_id
Authorization: Bearer <LabTechnician JWT>
```

### URL Parameter

| Param | Type | Description |
|---|---|---|
| `:donation_id` | string (UUID) | The donation whose test result is being updated |

### Request Body

Identical structure to [Submit Test Result](#1-submit-test-result).

#### Example Request — Changing from DEFERRED to CLEARED

```json
{
  "donation_id": "a3f5-...",
  "hiv_result": "NEGATIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "O+",
  "overall_status": "CLEARED",
  "storage_location": "Freezer-B",
  "rack_number": "R3",
  "shelf_number": "S1",
  "components": [
    { "component_type": "PRBC",   "quantity": 300, "position_number": "P5" },
    { "component_type": "PLASMA", "quantity": 150, "position_number": "P6" }
  ]
}
```

### Success Response — `200 OK`

```json
{
  "message": "Test result updated successfully"
}
```

### Error Responses

Same as [Submit Test Result](#1-submit-test-result), plus:

| HTTP | Error Message | Cause |
|---|---|---|
| `403` | `"you are not allowed to update another lab tech's test"` | The authenticated lab tech did not create this test |
| `404` | `"test result not found"` | No test exists for the given `donation_id` |

---

## 3. Convert Plasma to Cryoprecipitate

Converts an existing `AVAILABLE` Plasma unit into two new components:
- **Cryoprecipitate** — extracted portion
- **Cryo-poor Plasma** — the remainder after extraction

The original Plasma unit is **soft-deleted** (freeing its slot). Two new blood units are created, each assigned to their own position number in the same cell.

### Endpoint

```
POST /api/inventory/:id/convert-cryo
Authorization: Bearer <LabTechnician JWT>
```

### URL Parameter

| Param | Type | Description |
|---|---|---|
| `:id` | string (UUID) | The `blood_unit_id` of the PLASMA unit to convert |

### Request Body

```json
{
  "cryoprecipitate_quantity":  80,
  "cryo_poor_plasma_quantity": 60,
  "cryo_position_number":      "P4",
  "cryo_poor_position_number": "P5"
}
```

#### Field Details

| Field | Type | Required | Notes |
|---|---|---|---|
| `cryoprecipitate_quantity` | int | ✅ Yes | Volume (mL) of Cryoprecipitate. Must be > 0 and < total plasma volume. |
| `cryo_poor_plasma_quantity` | int | ❌ No | Volume (mL) of Cryo-poor Plasma. If not provided, defaults to `plasma_volume - cryo_quantity`. Must not exceed remaining volume. |
| `cryo_position_number` | string | ✅ Yes | Slot for the new Cryoprecipitate unit (e.g., `"P4"`). |
| `cryo_poor_position_number` | string | ✅ Yes | Slot for the new Cryo-poor Plasma unit (e.g., `"P5"`). Must differ from `cryo_position_number`. |

#### Example — Full Request

```json
{
  "cryoprecipitate_quantity":  80,
  "cryo_poor_plasma_quantity": 60,
  "cryo_position_number":      "P4",
  "cryo_poor_position_number": "P5"
}
```

#### Example — Let system calculate Cryo-poor Plasma quantity

```json
{
  "cryoprecipitate_quantity": 80,
  "cryo_position_number":     "P4",
  "cryo_poor_position_number":"P5"
}
```

### Success Response — `200 OK`

```json
{
  "message": "Plasma converted to Cryoprecipitate successfully"
}
```

### Error Responses

| HTTP | Error Message | Cause |
|---|---|---|
| `400` | `"only PLASMA units can be converted to Cryoprecipitate"` | The unit is not of type PLASMA |
| `400` | `"only AVAILABLE units can be converted"` | The plasma is RESERVED, USED, or EXPIRED |
| `400` | `"cryoprecipitate quantity must be greater than 0 and less than total plasma quantity"` | Invalid cryo quantity |
| `400` | `"cryo_position_number and cryo_poor_position_number are required"` | One or both position fields are missing |
| `400` | `"cryo_position_number and cryo_poor_position_number cannot be the same"` | Both positions are identical |
| `400` | `"Slot [Rack R1, Shelf S2, Pos P4] is already occupied"` | A chosen slot already contains an active unit |
| `400` | `"Only 0 positions available in this cell. You are trying to store 2 components."` | Not enough room in the cell for the 2 new units |
| `400` | `"cryo-poor plasma quantity cannot exceed the remaining plasma quantity"` | Cryo-poor volume is too large |
| `400` | `"cryo-poor plasma quantity cannot be negative"` | Negative value provided |
| `404` | `"blood unit not found"` | No unit exists with the given `:id` |

---

## 4. Get All Blood Units

Returns all blood units in the inventory with optional filters. Each unit now includes its `position_number` in the response.

### Endpoint

```
GET /api/inventory/
Authorization: Bearer <BloodBankAdmin or LabTechnician JWT>
```

### Query Parameters (all optional)

| Param | Type | Description | Example |
|---|---|---|---|
| `blood_type` | string | Filter by blood type | `A+` |
| `component_type` | string | Filter by component | `PRBC` |
| `status` | string | Filter by status | `AVAILABLE` |
| `quantity` | int | Minimum volume (mL) | `200` |
| `start_date` | string | Collection date from (YYYY-MM-DD) | `2026-01-01` |
| `end_date` | string | Collection date to (YYYY-MM-DD) | `2026-12-31` |
| `near_expired` | bool | Only units expiring within 7 days | `true` |

### Success Response — `200 OK`

```json
{
  "total_blood_units": 23,
  "available_blood": 20,
  "reserved_blood": 0,
  "used_blood": 1,
  "expired_blood": 2,
  "near_expired_blood": 3,
  "by_blood_type": {
    "A+": 9,
    "AB+": 1,
    "B+": 1,
    "O+": 12
  },
  "by_component_type": {
    "CRYOPRECIPITATE": 2,
    "CRYO_POOR_PLASMA": 2,
    "PLASMA": 5,
    "PLATELETS": 5,
    "PRBC": 7,
    "WHOLE_BLOOD": 2
  },
  "by_blood_and_component": {
    "A+_PLASMA": 3,
    "A+_PLATELETS": 3,
    "A+_PRBC": 3,
    "AB+_PRBC": 1,
    "B+_WHOLE_BLOOD": 1,
    "O+_CRYOPRECIPITATE": 2,
    "O+_CRYO_POOR_PLASMA": 2,
    "O+_PLASMA": 2,
    "O+_PLATELETS": 2,
    "O+_PRBC": 3,
    "O+_WHOLE_BLOOD": 1
  },
  "units": [
    {
      "blood_unit_id": "237ec831-7fce-4c99-892a-bedf39bd7d25",
      "blood_type": "O+",
      "component_type": "PLATELETS",
      "quantity_ml": 250,
      "collection_date": "2026-05-14T00:00:00Z",
      "expiration_date": "2026-05-19T00:00:00Z",
      "status": "EXPIRED",
      "is_deleted": false,
      "storage_location": "Main Fridge",
      "rack_number": "A1",
      "shelf_number": "2",
      "position_number": "P1",
      "created_at": "2026-05-16T00:20:14.138898Z"
    },
    {
      "blood_unit_id": "1e9d880f-12ba-478b-8d66-6a7232b3997d",
      "blood_type": "A+",
      "component_type": "PLATELETS",
      "quantity_ml": 60,
      "collection_date": "2026-05-18T00:00:00Z",
      "expiration_date": "2026-05-23T00:00:00Z",
      "status": "AVAILABLE",
      "is_deleted": false,
      "storage_location": "Freezer-A",
      "rack_number": "R1",
      "shelf_number": "S2",
      "position_number": "P2",
      "created_at": "2026-05-22T01:56:03.607228Z"
    }
  ]
}
```

---

## 5. Get Blood Unit By ID

Returns a single blood unit by its ID, including its slot information.

### Endpoint

```
GET /api/inventory/:id
Authorization: Bearer <BloodBankAdmin or LabTechnician JWT>
```

### URL Parameter

| Param | Type | Description |
|---|---|---|
| `:id` | string (UUID) | The `blood_unit_id` of the unit to retrieve |

### Success Response — `200 OK`

```json
{
  "blood_unit_id": "237ec831-7fce-4c99-892a-bedf39bd7d25",
  "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
  "blood_type": "O+",
  "component_type": "PLATELETS",
  "quantity_ml": 250,
  "collection_date": "2026-05-14T00:00:00Z",
  "expiration_date": "2026-05-19T00:00:00Z",
  "status": "EXPIRED",
  "is_deleted": false,
  "storage_location": "Main Fridge",
  "rack_number": "A1",
  "shelf_number": "2",
  "position_number": "P1",
  "created_at": "2026-05-16T00:20:14.138898Z"
}
```

### Error Responses

| HTTP | Error Message | Cause |
|---|---|---|
| `404` | `"blood unit not found"` | No active unit found with the given ID |

---

## 6. Mark Blood Unit as Used

Marks a `RESERVED` blood unit as `USED`. This **immediately frees the slot** so the position becomes available for a new blood unit.

### Endpoint

```
PUT /api/inventory/:id/used
Authorization: Bearer <BloodBankAdmin JWT>
```

### URL Parameter

| Param | Type | Description |
|---|---|---|
| `:id` | string (UUID) | The `blood_unit_id` to mark as used |

### Request Body

None required.

### Success Response — `200 OK`

```json
{
  "message": "Blood unit marked as used"
}
```

### Error Responses

| HTTP | Error Message | Cause |
|---|---|---|
| `400` | `"unit not found or not in RESERVED status"` | The unit does not exist or is not currently RESERVED |
| `500` | `"internal server error"` | Unexpected database error |

---

## 7. Delete Blood Unit

Soft-deletes a blood unit from the inventory. This **frees the slot** regardless of the unit's current status. This is the mechanism for releasing slots held by `EXPIRED` units.

### Endpoint

```
DELETE /api/inventory/:id
Authorization: Bearer <BloodBankAdmin or LabTechnician JWT>
```

### URL Parameter

| Param | Type | Description |
|---|---|---|
| `:id` | string (UUID) | The `blood_unit_id` to delete |

### Request Body

None required.

### Success Response — `200 OK`

```json
{
  "message": "Blood unit deleted successfully"
}
```

### Error Responses

| HTTP | Error Message | Cause |
|---|---|---|
| `404` | `"blood unit not found"` | No active unit found with the given ID |
| `500` | `"internal server error"` | Unexpected database error |

---

## Error Reference

### Common Slot Validation Errors

| Error | Description | How to Fix |
|---|---|---|
| `"position_number is required for all components when overall_status is CLEARED"` | A component in the array is missing `position_number` | Add `position_number` to every component object |
| `"duplicate position_number 'P1' found in request"` | Two components share the same position | Give each component a unique `position_number` |
| `"Slot [Rack R1, Shelf S2, Pos P1] in Freezer-A is already occupied"` | That slot has an active blood unit | Choose a different slot, or delete/use the existing unit first |
| `"Only X positions available in this cell. You are trying to store Y components."` | Not enough room in the cell | Choose a different cell with more free slots, or delete expired units to free space |
| `"cryo_position_number and cryo_poor_position_number cannot be the same"` | Two new cryo products can't share a slot | Use two different position numbers |

### HTTP Status Code Reference

| Code | Meaning |
|---|---|
| `200` | Request succeeded |
| `201` | Resource created successfully |
| `400` | Bad request — validation error (check error message) |
| `401` | Unauthorized — missing or invalid JWT token |
| `403` | Forbidden — valid token but insufficient role permissions |
| `404` | Resource not found |
| `500` | Internal server error |
