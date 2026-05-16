# Lab Technician API Documentation

Base Paths: `/api/lab`, `/api/inventory`, `/api/analytics/lab`

Authentication: All endpoints require a valid Bearer token in the `Authorization` header with the `labtech` role.

---

## 1. Pending Donations

### 1.1 Get All Pending Donations
Fetches a list of all donations that have been collected but not yet tested.

**Endpoint:** `GET /api/lab/pending-donations`

**Response (200 OK):**
```json
[
  {
    "donation_id": "uuid",
    "donor_id": "uuid",
    "status": "APPROVED",
    "quantity_ml": 450,
    "collection_date": "2023-10-10T10:00:00Z",
    "donor_name": "John Doe"
  }
]
```

### 1.2 Get Pending Donation By ID
Fetches a specific pending donation before processing.

**Endpoint:** `GET /api/lab/pending-donations/:donation_id`

**Response (200 OK):**
```json
{
  "donation_id": "uuid",
  "donor_id": "uuid",
  "status": "APPROVED",
  "quantity_ml": 450,
  "collection_date": "2023-10-10T10:00:00Z"
}
```

---

## 2. Test Processing

### 2.1 Submit Test Result (Process Blood)
Processes a pending donation, records the 4 disease markers, calculates the overall status, and if CLEARED, splits it into inventory components.

**Endpoint:** `POST /api/lab/tests`

**Rules:**
- `overall_status` must match the 4 disease markers (e.g., if HIV is POSITIVE, it must be PERMANENTLY_DEFERRED).
- If `overall_status` is `CLEARED`:
  - `storage_location` is required.
  - If donation was 350ml: exactly 1 component of type `WHOLE_BLOOD` is required.
  - If donation was 450ml: 1 to 3 components of types `PRBC`, `PLATELETS`, or `PLASMA` are required.
  - The sum of component quantities cannot exceed the donation `quantity_ml`.

**Request Body:**
```json
{
  "donation_id": "uuid",
  "hiv_result": "NEGATIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "O+",
  "overall_status": "CLEARED",
  "storage_location": "Main Fridge",
  "rack_number": "A1",
  "shelf_number": "2",
  "components": [
    {
      "component_type": "PRBC",
      "quantity": 250
    },
    {
      "component_type": "PLASMA",
      "quantity": 200
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "message": "test result processed successfully",
  "test_result": { ... }
}
```

### 2.2 Update Test Result
Allows the Lab Technician who originally processed the test to update it (e.g., fix a typo in the storage location or component split). 

**Endpoint:** `PUT /api/lab/tests/:donation_id`

**Request Body:** Same as Submit Test Result.

**Response (200 OK):**
```json
{
  "message": "updated"
}
```

### 2.3 Quick Reject Blood
Instantly marks a pending donation as rejected without filling out components. Status will be marked as `PERMANENTLY_DEFERRED`.

**Endpoint:** `PATCH /api/lab/tests/:donation_id/reject`

**Response (200 OK):**
```json
{
  "message": "blood rejected"
}
```

---

## 3. Test History & Summaries

### 3.1 Get All Test History
Fetches a list of all historical test results processed by the lab, along with summary cards.

**Endpoint:** `GET /api/lab/all-tests`

**Query Parameters (Filters):**
- `overall_status` (string, optional): Filter by medical status (e.g., "CLEARED", "PERMANENTLY_DEFERRED").
- `blood_type` (string, optional): Filter by blood type (e.g., "O+", "A-").
- `component_type` (string, optional): Filter by component type (e.g., "PRBC").
- `storage_location` (string, optional): Filter by storage location string.
- `start_date` (string, optional): ISO8601 Date.
- `end_date` (string, optional): ISO8601 Date.

**Response (200 OK):**
```json
{
  "total_tests": 150,
  "cleared": 120,
  "temporarily_deferred": 25,
  "permanently_deferred": 5,
  "tests": [
    {
      "test_id": "uuid",
      "donation_id": "uuid",
      "donor_id": "uuid",
      "tested_by": "uuid",
      "hiv_result": "NEGATIVE",
      "hepatitis_b_result": "NEGATIVE",
      "hepatitis_c_result": "NEGATIVE",
      "syphilis_result": "NEGATIVE",
      "blood_type": "O+",
      "overall_status": "CLEARED",
      "storage_location": "Main Fridge",
      "rack_number": "A1",
      "shelf_number": "2",
      "components": [
        {
          "component_type": "PRBC",
          "quantity": 250
        }
      ],
      "created_at": "2023-10-10T12:00:00Z"
    }
  ]
}
```

### 3.2 Get My Tests
Fetches a list of test results processed specifically by the currently authenticated Lab Technician. Returns the exact same response structure (with cards) and supports the exact same filters as `/all-tests`.

**Endpoint:** `GET /api/lab/tests/my`

### 3.3 Get Single Test Result
Fetches the details of a specific test.

**Endpoint:** `GET /api/lab/tests/:donation_id`

**Response (200 OK):**
```json
{
  "test_id": "uuid",
  "donation_id": "uuid",
  "donor_id": "uuid",
  "tested_by": "uuid",
  "hiv_result": "NEGATIVE",
  "hepatitis_b_result": "NEGATIVE",
  "hepatitis_c_result": "NEGATIVE",
  "syphilis_result": "NEGATIVE",
  "blood_type": "O+",
  "overall_status": "CLEARED",
  "created_at": "2023-10-10T12:00:00Z"
}
```

---

## 4. Shared Inventory Management

Lab Techs have access to the core inventory system alongside Blood Bank Admins.

### 4.1 Get All Blood Units (Inventory)
Fetches the current blood inventory with comprehensive summary cards.

**Endpoint:** `GET /api/inventory/`

**Query Parameters:**
- `blood_type` (string, optional)
- `component_type` (string, optional)
- `status` (string, optional) - e.g., AVAILABLE, RESERVED, USED, EXPIRED
- `quantity` (int, optional)
- `near_expired` (boolean, optional)
- `start_date`, `end_date` (string, optional)

**Response (200 OK):**
```json
{
  "total_blood_units": 500,
  "available_blood": 450,
  "reserved_blood": 20,
  "used_blood": 20,
  "expired_blood": 10,
  "near_expired_blood": 5,
  "by_blood_type": {
    "O+": 200,
    "A-": 50
  },
  "by_component_type": {
    "PRBC": 300,
    "PLASMA": 200
  },
  "by_blood_and_component": {
    "O+_PRBC": 150
  },
  "units": [
    {
      "blood_unit_id": "uuid",
      "blood_type": "O+",
      "component_type": "PRBC",
      "quantity_ml": 250,
      "collection_date": "2023-10-10T10:00:00Z",
      "expiration_date": "2023-11-21T10:00:00Z",
      "status": "AVAILABLE",
      "storage_location": "Fridge A",
      "rack_number": "1",
      "shelf_number": "2",
      "created_at": "2023-10-10T12:00:00Z"
    }
  ]
}
```

### 4.2 Convert Plasma to Cryoprecipitate (Lab Only)
Allows a Lab Technician to convert an existing Plasma unit into Cryoprecipitate and Cryo-poor Plasma. The original Plasma unit is marked as soft-deleted.

**Endpoint:** `POST /api/inventory/:id/convert-cryo`

**Request Body:**
```json
{
  "cryoprecipitate_quantity": 15,
  "cryo_poor_plasma_quantity": 185
}
```
*(Note: `cryo_poor_plasma_quantity` is optional. If omitted, the system calculates it as a perfect split: Original Plasma Quantity - Cryoprecipitate Quantity. If the lab tech specifies a quantity, it handles standard processing/tubing losses).*

**Response (200 OK):**
```json
{
  "message": "Plasma converted to Cryoprecipitate successfully"
}
```

### 4.3 Export Inventory
Exports the filtered inventory list. The endpoint forces a file download.

**Endpoints:**
- `GET /api/inventory/export/csv`
- `GET /api/inventory/export/pdf`

*(Accepts the exact same query parameters as the Get All Blood Units endpoint).*

---

## 5. Lab Analytics Dashboard

Fetches high-level metrics for the Lab Technician dashboard.

**Endpoint:** `GET /api/analytics/lab/dashboard`

**Response (200 OK):**
```json
{
  "total_donations_processed": 1500,
  "cleared_donations": 1400,
  "rejected_donations": 100,
  "processing_accuracy": 98.5,
  "recent_tests": [ ... ]
}
```
