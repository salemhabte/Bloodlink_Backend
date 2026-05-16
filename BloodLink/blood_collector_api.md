# Blood Collector API Documentation

Base Path: `/api/bloodcollector`

Authentication: All endpoints require a valid Bearer token in the `Authorization` header with the `bloodcollector` role.

---

## 1. Donors

### 1.1 Get Eligible Donors
Fetches a list of donors who are currently eligible to donate blood. This includes "New Eligible Donors" (never donated) and "Returning Eligible Donors" (last approved donation > 90 days ago).

**Endpoint:** `GET /eligible-donors`
**Query Parameters:**
- `q` (string, optional): Search query matching full name, email, or phone.

**Response (200 OK):**
```json
{
  "total_eligible": 15,
  "returning_eligible": 5,
  "new_eligible_donors": 10,
  "donors": [
    {
      "donor_id": "uuid",
      "user_id": "uuid",
      "full_name": "John Doe",
      "email": "johndoe@example.com",
      "phone": "+1234567890",
      "address": "123 Main St",
      "blood_type": "O+",
      "status": "Pending",
      "overall_status": "CLEARED",
      "registration_date": "2023-01-01T00:00:00Z",
      "last_donation_date": "2023-05-01T00:00:00Z"
    }
  ]
}
```

### 1.2 Get All Donors
Fetches the complete directory of registered donors along with summary statistics.

**Endpoint:** `GET /all-donors`
**Query Parameters:**
- `blood_type` (string, optional): Filter by blood type (e.g., "O+", "A-").
- `overall_status` (string, optional): Filter by medical status (e.g., "CLEARED", "TEMPORARILY_DEFERRED").

**Response (200 OK):**
```json
{
  "total_donors": 150,
  "cleared": 120,
  "temporarily_deferred": 25,
  "permanently_deferred": 5,
  "donors": [
    {
      "donor_id": "uuid",
      "full_name": "Jane Smith",
      "email": "janesmith@example.com",
      "overall_status": "CLEARED",
      "registration_date": "2023-01-01T00:00:00Z"
    }
  ]
}
```

### 1.3 Get Eligible Donor by ID
Fetches details of a specific eligible donor, including their last donation info.

**Endpoint:** `GET /eligible-donor/:id`

**Response (200 OK):**
```json
{
  "donor_id": "uuid",
  "full_name": "John Doe",
  "blood_type": "O+",
  "status": "Pending",
  "overall_status": "CLEARED",
  "registration_date": "2023-01-01T00:00:00Z",
  "last_donation_date": "2023-05-01T00:00:00Z"
}
```

---

## 2. Donations

### 2.1 Create Donation Record
Records a new blood donation. Enforces eligibility rules and medical constraints.

**Endpoint:** `POST /donation`

**Request Body:**
```json
{
  "donor_id": "uuid",
  "campaign_id": "uuid", // Optional
  "weight": 70.5,
  "hemoglobin": 13.5,
  "blood_pressure": "120/80",
  "pulse": 75,
  "temperature": 36.6,
  "quantity_ml": 450, // Must be 350 or 450 if APPROVED, 0 if REJECTED_TEMPORARY
  "status": "APPROVED", // "APPROVED" or "REJECTED_TEMPORARY"
  "rejection_reason": "", // Required if status is REJECTED_TEMPORARY
  "collection_date": "2023-10-10T10:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "message": "Donation created successfully"
}
```

### 2.2 Get All Donations
Fetches all donation records with optional filtering.

**Endpoint:** `GET /donation`
**Query Parameters:**
- `status` (string, optional): E.g., "APPROVED", "REJECTED_TEMPORARY".
- `start_date` (string, optional): ISO8601 Date.
- `end_date` (string, optional): ISO8601 Date.

**Response (200 OK):**
```json
[
  {
    "donation_id": "uuid",
    "donor_id": "uuid",
    "status": "APPROVED",
    "quantity_ml": 450,
    "collection_date": "2023-10-10T10:00:00Z"
  }
]
```

### 2.3 Get Donation By ID
Fetches details of a specific donation record.

**Endpoint:** `GET /donation/:id`

**Response (200 OK):**
```json
{
  "donation_id": "uuid",
  "donor_id": "uuid",
  "status": "APPROVED",
  "quantity_ml": 450,
  "weight": 70.5,
  "hemoglobin": 13.5,
  "blood_pressure": "120/80",
  "collection_date": "2023-10-10T10:00:00Z"
}
```

### 2.4 Update Donation Record
Updates medical data or status for a specific donation. Only the collector who created the donation can update it.

**Endpoint:** `PUT /donation/:id`

**Request Body:**
```json
{
  "weight": 72.0,
  "hemoglobin": 14.0,
  "blood_pressure": "115/75",
  "pulse": 72,
  "temperature": 36.5,
  "quantity_ml": 450,
  "status": "APPROVED",
  "rejection_reason": "",
  "collection_date": "2023-10-10T10:00:00Z"
}
```

**Response (200 OK):**
```json
{
  "message": "Donation updated successfully"
}
```

### 2.5 Get My Donations
Fetches all donation records created by the currently authenticated blood collector.

**Endpoint:** `GET /donation/my`

**Response (200 OK):**
```json
[
  {
    "donation_id": "uuid",
    "status": "APPROVED",
    "quantity_ml": 450,
    "collection_date": "2023-10-10T10:00:00Z"
  }
]
```
