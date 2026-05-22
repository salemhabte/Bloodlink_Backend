# Friendly ID API Documentation

## Overview

This document covers all endpoints affected by the **User-Friendly Identifier** feature.

### ID Format
| Type | Format | Example |
|---|---|---|
| Donation Number | `DON-YYYY-NNNNNN` | `DON-2026-000001` |
| Blood Unit Number | `UNIT-YYYY-NNNNNN` | `UNIT-2026-000003` |

- The numeric part is **zero-padded to 6 digits**, auto-expanding to 7 when > 999,999.
- **UUIDs** are still used internally for all relations and `GET by ID` routes.
- Each blood **component** from the same donation gets its **own unique `unit_number`**, but they all share the **same `donation_number`**.

---

## 1. Create Donation

**Generates a `DON-YYYY-NNNNNN` automatically upon creation.**

```
POST /api/donations/
Authorization: Bearer <BloodCollector JWT>
```

### Request Body
```json
{
  "donor_id": "uuid",
  "campaign_id": "uuid | null",
  "collection_date": "2026-05-22",
  "weight": 70.5,
  "blood_pressure": "120/80",
  "hemoglobin": 14.2,
  "temperature": 36.8,
  "pulse": 72,
  "quantity_ml": 450,
  "status": "APPROVED",
  "rejection_reason": ""
}
```

### Request Fields
| Field | Type | Required | Notes |
|---|---|---|---|
| `donor_id` | string (UUID) | ✅ | Must be a registered donor |
| `campaign_id` | string (UUID) | ❌ | Optional campaign link |
| `collection_date` | string (date) | ✅ | ISO date format |
| `weight` | float | ✅ | Donor weight in kg |
| `blood_pressure` | string | ✅ | e.g. `"120/80"` |
| `hemoglobin` | float | ✅ | |
| `temperature` | float | ✅ | Celsius |
| `pulse` | int | ✅ | Beats per minute |
| `quantity_ml` | int | ✅ | Must be 350 or 450 |
| `status` | string | ✅ | `APPROVED` or `REJECTED_TEMPORARY` |
| `rejection_reason` | string | ⚠️ | Required if status is `REJECTED_TEMPORARY` |

### Success Response — `201 Created`
```json
{
  "message": "Donation recorded successfully",
  "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
  "donation_number": "DON-2026-000001"
}
```

### Error Responses
| Status | Error Message | Cause |
|---|---|---|
| `400` | `"status must be APPROVED or REJECTED_TEMPORARY"` | Invalid status |
| `400` | `"rejection_reason is required when status is REJECTED_TEMPORARY"` | Missing rejection reason |
| `400` | `"quantity_ml must be 350 or 450"` | Invalid donation volume |
| `500` | `"failed to generate donation number: ..."` | Database sequence failure |

---

## 2. Get All Donations (blood Collector)

**Now includes `donation_number` in every record.**

```
GET /api/donations/
Authorization: Bearer <BloodCollector or BloodBankAdmin JWT>
```

### Query Parameters
| Param | Type | Description |
|---|---|---|
| `collector_id` | string | Filter by collector UUID |
| `donor_id` | string | Filter by donor UUID |
| `status` | string | Filter by status |
| `donation_number` | string | Filter by friendly donation ID |
| `start_date` | string | Collection date from (YYYY-MM-DD) |
| `end_date` | string | Collection date to (YYYY-MM-DD) |

### Success Response — `200 OK`
```json
{
  "total": 5,
  "approved": 4,
  "temporarily_rejected": 1,
  "donations": [
    {
      "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
      "donation_number": "DON-2026-000001",
      "donor_id": "uuid",
      "donor_name": "Abebe Kebede",
      "collector_name": "Dr. Sara",
      "campaign_title": "World Blood Day Drive",
      "campaign_address": "Addis Ababa",
      "collection_date": "2026-05-22T00:00:00Z",
      "weight": 70.5,
      "blood_pressure": "120/80",
      "hemoglobin": 14.2,
      "temperature": 36.8,
      "pulse": 72,
      "quantity_ml": 450,
      "status": "APPROVED",
      "rejection_reason": "",
      "overall_status": "CLEARED",
      "created_at": "2026-05-22T06:00:00Z"
    }
  ]
}
```

---

## 3. Get My Donations (Blood Collector)

**Now includes `donation_number` in every record.**

```
GET /api/donations/my
Authorization: Bearer <BloodCollector JWT>
```

### Query Parameters
| Param | Type | Description |
|---|---|---|
| `status` | string | Filter by status |
| `donation_number` | string | Filter by friendly donation ID |
| `start_date` | string | Collection date from (YYYY-MM-DD) |
| `end_date` | string | Collection date to (YYYY-MM-DD) |

### Success Response — `200 OK`
```json
[
  {
    "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
    "donation_number": "DON-2026-000001",
    "donor_id": "uuid",
    "collection_date": "2026-05-22T00:00:00Z",
    "quantity_ml": 450,
    "status": "APPROVED",
    "rejection_reason": "",
    "created_at": "2026-05-22T06:00:00Z"
  }
]
```

---

## 4. Get Donation By ID

**Now includes `donation_number`.**

```
GET /api/donations/:id
Authorization: Bearer <BloodCollector>
```

### URL Parameter
| Param | Type | Description |
|---|---|---|
| `:id` | string (UUID) | The internal `donation_id` UUID |

### Success Response — `200 OK`
```json
{
  "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
  "donation_number": "DON-2026-000001",
  "donor_id": "uuid",
  "donor_name": "Abebe Kebede",
  "collector_name": "Dr. Sara",
  "campaign_title": "World Blood Day Drive",
  "campaign_address": "Addis Ababa",
  "collection_date": "2026-05-22T00:00:00Z",
  "weight": 70.5,
  "blood_pressure": "120/80",
  "hemoglobin": 14.2,
  "temperature": 36.8,
  "pulse": 72,
  "quantity_ml": 450,
  "status": "APPROVED",
  "rejection_reason": "",
  "created_at": "2026-05-22T06:00:00Z"
}
```

### Error Responses
| Status | Error | Cause |
|---|---|---|
| `404` | `"donation not found"` | Invalid UUID |

---

## 5. Process Test Result (Lab Technician)

**Generates unique `UNIT-YYYY-NNNNNN` for every component when status is `CLEARED`. Each component gets its own number, all sharing the same `donation_number`.**

```
POST /api/lab/test-results/
Authorization: Bearer <LabTechnician JWT>
```

### Request Body — CLEARED
```json
{
  "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
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

### Request Body — PERMANENTLY_DEFERRED
```json
{
  "donation_id": "uuid",
  "hiv_result": "POSITIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "B+",
  "overall_status": "PERMANENTLY_DEFERRED",
  "storage_location": "",
  "rack_number": "",
  "shelf_number": "",
  "components": []
}
```

### Request Body — TEMPORARILY_DEFERRED
```json
{
  "donation_id": "uuid",
  "hiv_result": "NEGATIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "POSITIVE",
  "blood_type": "O+",
  "overall_status": "TEMPORARILY_DEFERRED",
  "storage_location": "",
  "rack_number": "",
  "shelf_number": "",
  "components": []
}
```

### Success Response — `200 OK` (CLEARED)
```json
{
  "message": "Test result recorded and blood units created",
  "donation_number": "DON-2026-000001",
  "units_created": [
    {
      "blood_unit_id": "uuid-1",
      "unit_number": "UNIT-2026-000001",
      "component_type": "PRBC",
      "quantity_ml": 250,
      "position_number": "P1"
    },
    {
      "blood_unit_id": "uuid-2",
      "unit_number": "UNIT-2026-000002",
      "component_type": "PLATELETS",
      "quantity_ml": 60,
      "position_number": "P2"
    },
    {
      "blood_unit_id": "uuid-3",
      "unit_number": "UNIT-2026-000003",
      "component_type": "PLASMA",
      "quantity_ml": 140,
      "position_number": "P3"
    }
  ]
}
```

### Error Responses
| Status | Error Message | Cause |
|---|---|---|
| `400` | `"a test for this donation already exists"` | Duplicate test |
| `400` | `"components must be empty when overall_status is PERMANENTLY_DEFERRED"` | Invalid payload |
| `400` | `"components must be empty when overall_status is TEMPORARILY_DEFERRED"` | Invalid payload |
| `400` | `"storage fields must be empty when overall_status is PERMANENTLY_DEFERRED"` | Storage sent for deferred |
| `400` | `"storage fields must be empty when overall_status is TEMPORARILY_DEFERRED"` | Storage sent for deferred |
| `400` | `"duplicate position_number 'P1' found in request"` | Two components share same slot |
| `400` | `"Slot [Rack R1, Shelf S2, Pos P1] in Freezer-A is already occupied"` | Slot taken by another unit |
| `400` | `"Only X positions available in this cell. You are trying to store Y components."` | Cell at capacity |
| `400` | `"⚠ Suggestion: Based on test results, overall status should be 'CLEARED'"` | Status conflicts with results |
| `500` | `"failed to generate unit number: ..."` | DB sequence failure |

---

## 6. Get Pending Donations (Lab Technician)

**Now includes `donation_number` in the response payload for correlation.**

```
GET /api/lab/pending-donations
Authorization: Bearer <LabTechnician JWT>
```

### Success Response — `200 OK`
```json
[
  {
    "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
    "donation_number": "DON-2026-000001",
    "donor_id": "uuid",
    "donor_name": "Abebe Kebede",
    "collected_by": "uuid",
    "collector_name": "Dr. Sara",
    "collection_date": "2026-05-22T00:00:00Z",
    "weight": 70.5,
    "blood_pressure": "120/80",
    "hemoglobin": 14.2,
    "temperature": 36.8,
    "pulse": 72,
    "quantity_ml": 450,
    "status": "APPROVED",
    "created_at": "2026-05-22T06:00:00Z"
  }
]
```

---

## 7. Get All Test Results

**Now includes `donation_number` and `unit_number` per blood unit.**

```
GET /api/lab/test-results/
Authorization: Bearer <LabTechnician JWT>
```

### Query Parameters
| Param | Type | Description |
|---|---|---|
| `donation_number` | string | Filter by friendly donation ID |
| `overall_status` | string | Filter by CLEARED, TEMPORARILY_DEFERRED, etc. |
| `blood_type` | string | Filter by blood type |
| `component_type` | string | Filter by component |
| `start_date` | string | Test date from (YYYY-MM-DD) |
| `end_date` | string | Test date to (YYYY-MM-DD) |

### Success Response — `200 OK`
```json
[
  {
    "test_id": "uuid",
    "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
    "donation_number": "DON-2026-000001",
    "donor_id": "uuid",
    "tested_by": "uuid",
    "hiv_result": "NEGATIVE",
    "hepatitis_b_result": "NEGATIVE",
    "hepatitis_c_result": "NEGATIVE",
    "syphilis_result": "NEGATIVE",
    "blood_type": "A+",
    "overall_status": "CLEARED",
    "created_at": "2026-05-22T06:00:00Z",
    "blood_units": [
      {
        "blood_unit_id": "uuid-1",
        "unit_number": "UNIT-2026-000001",
        "component_type": "PRBC",
        "quantity_ml": 250,
        "position_number": "P1"
      }
    ]
  }
]
```

---

## 8. Get My Test Results (Lab Technician)

**Same structure as Get All, filtered to the authenticated lab tech.**

```
GET /api/lab/test-results/my
Authorization: Bearer <LabTechnician JWT>
```

Response structure and **Query Parameters** are identical to **Section 7** above.

---

## 9. Get Test Results By Donation ID

```
GET /api/lab/test-results/:donation_id
Authorization: Bearer <LabTechnician or BloodBankAdmin JWT>
```

### URL Parameter
| Param | Type | Description |
|---|---|---|
| `:donation_id` | string (UUID) | Internal donation UUID |

### Success Response — `200 OK`
```json
{
  "test_id": "uuid",
  "donation_id": "09e17a56-19f9-4d78-b40a-38514b089a01",
  "donation_number": "DON-2026-000001",
  "donor_id": "uuid",
  "tested_by": "uuid",
  "hiv_result": "NEGATIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "A+",
  "overall_status": "CLEARED",
  "created_at": "2026-05-22T06:00:00Z",
  "blood_units": [
    {
      "blood_unit_id": "uuid-1",
      "unit_number": "UNIT-2026-000001",
      "component_type": "PRBC",
      "quantity_ml": 250,
      "position_number": "P1"
    }
  ]
}
```

### Error Responses
| Status | Error | Cause |
|---|---|---|
| `404` | `"test result not found"` | No test for this donation |

---

## 9. Get All Blood Units (Inventory)

**Now includes `unit_number` per blood unit.**

```
GET /api/inventory/
Authorization: Bearer <BloodBankAdmin or LabTechnician JWT>
```

### Query Parameters
| Param | Type | Description |
|---|---|---|
| `blood_type` | string | Filter by blood type |
| `component_type` | string | Filter by component |
| `status` | string | Filter by status |
| `unit_number` | string | Filter by friendly unit ID |
| `quantity` | int | Filter by minimum quantity |
| `near_expired` | boolean | True for near-expired units |
| `start_date` | string | Collection date from (YYYY-MM-DD) |
| `end_date` | string | Collection date to (YYYY-MM-DD) |

### Success Response — `200 OK`
```json
{
  "total_blood_units": 23,
  "available_blood": 20,
  "reserved_blood": 0,
  "used_blood": 1,
  "expired_blood": 2,
  "near_expired_blood": 3,
  "by_blood_type": { "A+": 9, "O+": 12 },
  "by_component_type": { "PRBC": 7, "PLASMA": 5 },
  "by_blood_and_component": { "A+_PRBC": 3 },
  "units": [
    {
      "blood_unit_id": "uuid-1",
      "unit_number": "UNIT-2026-000001",
      "donation_id": "uuid",
      "blood_type": "A+",
      "component_type": "PRBC",
      "quantity_ml": 250,
      "collection_date": "2026-05-22T00:00:00Z",
      "expiration_date": "2026-07-03T00:00:00Z",
      "status": "AVAILABLE",
      "is_deleted": false,
      "storage_location": "Freezer-A",
      "rack_number": "R1",
      "shelf_number": "S2",
      "position_number": "P1",
      "created_at": "2026-05-22T06:00:00Z"
    }
  ]
}
```

---

## 10. Get Blood Unit By ID

**Now includes `unit_number`.**

```
GET /api/inventory/:id
Authorization: Bearer <BloodBankAdmin or LabTechnician JWT>
```

### Success Response — `200 OK`
```json
{
  "blood_unit_id": "uuid-1",
  "unit_number": "UNIT-2026-000001",
  "donation_id": "uuid",
  "blood_type": "A+",
  "component_type": "PRBC",
  "quantity_ml": 250,
  "collection_date": "2026-05-22T00:00:00Z",
  "expiration_date": "2026-07-03T00:00:00Z",
  "status": "AVAILABLE",
  "is_deleted": false,
  "storage_location": "Freezer-A",
  "rack_number": "R1",
  "shelf_number": "S2",
  "position_number": "P1",
  "created_at": "2026-05-22T06:00:00Z"
}
```

### Error Responses
| Status | Error | Cause |
|---|---|---|
| `404` | `"blood unit not found"` | Invalid UUID |

---

## 11. Convert Plasma to Cryoprecipitate

**Generates two unique `UNIT-YYYY-NNNNNN` numbers for the new units.**

```
POST /api/inventory/convert-plasma/:id
Authorization: Bearer <LabTechnician JWT>
```

### Request Body
```json
{
  "cryo_quantity_ml": 15,
  "cryo_poor_quantity_ml": 185,
  "cryo_position_number": "P4",
  "cryo_poor_position_number": "P5"
}
```

### Success Response — `200 OK`
```json
{
  "message": "Plasma successfully converted to Cryoprecipitate and Cryo-poor Plasma",
  "cryoprecipitate": {
    "blood_unit_id": "uuid-new-1",
    "unit_number": "UNIT-2026-000004",
    "component_type": "CRYOPRECIPITATE",
    "quantity_ml": 15,
    "position_number": "P4"
  },
  "cryo_poor_plasma": {
    "blood_unit_id": "uuid-new-2",
    "unit_number": "UNIT-2026-000005",
    "component_type": "CRYO_POOR_PLASMA",
    "quantity_ml": 185,
    "position_number": "P5"
  }
}
```

### Error Responses
| Status | Error Message | Cause |
|---|---|---|
| `400` | `"only PLASMA units can be converted to Cryoprecipitate"` | Wrong component type |
| `400` | `"only AVAILABLE units can be converted"` | Unit is reserved/used/expired |
| `400` | `"cryo_position_number and cryo_poor_position_number are required"` | Missing positions |
| `400` | `"cryo_position_number and cryo_poor_position_number cannot be the same"` | Same slot |
| `400` | `"Slot [Rack R1, Shelf S2, Pos P4] is already occupied"` | Slot taken |
| `400` | `"Only X positions available in this cell. You are trying to store 2 components."` | Cell full |
| `500` | `"failed to generate unit number: ..."` | DB sequence failure |

---

## 12. CSV / PDF Export

**Both exports now replace `blood_unit_id` with `unit_number`.**

```
GET /api/inventory/export/csv
GET /api/inventory/export/pdf
Authorization: Bearer <BloodBankAdmin JWT>
```

### CSV Header Row
```
unit_number,blood_type,component_type,quantity_ml,collection_date,expiration_date,status,storage_location,rack_number,shelf_number,position_number
```

### PDF Table Columns
```
Unit Number | Blood Type | Component | Qty (ml) | Expiry | Status | Location | Rack | Shelf | Pos
```

---

## ID Rules Summary

| Rule | Detail |
|---|---|
| Auto-generated | Never sent by client — server always generates |
| Year reset | Counter resets automatically each calendar year |
| Scalable | 6-digit → 7-digit when counter > 999,999 |
| Unique per component | Each component (PRBC, PLASMA, etc.) from one donation gets its own `unit_number` |
| Shared donation number | All components from one donation share the same `donation_number` |
| UUID preserved | Internal `donation_id` and `blood_unit_id` UUIDs still used for all DB relations and GET-by-ID routes |

---

## HTTP Status Code Reference

| Code | Meaning |
|---|---|
| `200` | Request succeeded |
| `201` | Resource created successfully |
| `400` | Validation error — check error message |
| `401` | Unauthorized — missing or invalid JWT |
| `403` | Forbidden — valid token, insufficient role |
| `404` | Resource not found |
| `500` | Internal server error |
