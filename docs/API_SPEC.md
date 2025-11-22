# API Specification - Agnos Demo Service

**Version:** 1.0  
**Base URL:** `http://localhost:80`

---

## 1. Authentication

### 1.1 Staff Login
Authenticates a staff member and returns a JWT token for accessing protected endpoints.

- **Endpoint:** `POST /staff/login`
- **Content-Type:** `application/json`

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | Staff username |
| `password` | string | Yes | Staff password |
| `hospital` | string | Yes | Hospital code (e.g., "hn-001") |

**Example Request:**
```json
{
  "username": "admin",
  "password": "password",
  "hospital": "hn-001"
}
```

**Success Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**
- `400 Bad Request`: Invalid input format.
- `401 Unauthorized`: Invalid credentials.

---

### 1.2 Create Staff
Registers a new staff member.

- **Endpoint:** `POST /staff/create`
- **Content-Type:** `application/json`

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | Unique username |
| `password` | string | Yes | Password |
| `hospital` | string | Yes | Hospital code |

**Example Request:**
```json
{
  "username": "new_staff",
  "password": "secure_password",
  "hospital": "hn-001"
}
```

**Success Response (201 Created):**
```json
{
  "message": "Staff created successfully",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 2. Patient Management

**Authentication Required:** All patient endpoints require a valid JWT token in the header.
`Authorization: Bearer <token>`

### 2.1 Search Patients
Search for patients within the staff's hospital. Results are automatically filtered to match the staff's hospital code.

- **Endpoint:** `GET /patient/search`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `patient_hn` | string | Exact match for Hospital Number |
| `national_id` | string | Exact match for National ID |
| `passport_id` | string | Exact match for Passport ID |
| `first_name` | string | Partial match (Thai or English) |
| `middle_name` | string | Partial match (Thai or English) |
| `last_name` | string | Partial match (Thai or English) |
| `date_of_birth` | string | Exact match (YYYY-MM-DD) |

**Example Request:**
`GET /patient/search?first_name=John&date_of_birth=1980-01-01`

**Success Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "patient_hn": "hn-001",
    "first_name_th": "จอห์น",
    "last_name_th": "โด",
    "first_name_en": "John",
    "last_name_en": "Doe",
    "date_of_birth": "1980-01-01",
    "gender": "M",
    "national_id": "1234567890123",
    "passport_id": "A1234567",
    "phone_number": "0812345678",
    "email": "john@example.com"
  }
]
```

---

### 2.2 Get Patient by Identifier
Retrieve a specific patient using their National ID or Passport ID.
**Security Note:** You can only retrieve patients belonging to your own hospital.

- **Endpoint:** `GET /patient/search/:id`
- **Path Parameter:** `:id` can be a National ID or Passport ID.

**Example Request:**
`GET /patient/search/1234567890123`

**Success Response (200 OK):**
```json
{
  "id": "uuid-string",
  "patient_hn": "hn-001",
  "first_name_th": "จอห์น",
  "last_name_th": "โด",
  "first_name_en": "John",
  "last_name_en": "Doe",
  "date_of_birth": "1980-01-01",
  "gender": "M",
  "national_id": "1234567890123",
  "passport_id": "A1234567"
}
```

**Error Responses:**
- `404 Not Found`: Patient does not exist.
- `403 Forbidden`: Patient belongs to a different hospital.

---

## 3. Data Models

### Patient Object
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique system identifier |
| `patient_hn` | string | Hospital Code / Number |
| `first_name_th` | string | First Name (Thai) |
| `middle_name_th` | string | Middle Name (Thai) |
| `last_name_th` | string | Last Name (Thai) |
| `first_name_en` | string | First Name (English) |
| `middle_name_en` | string | Middle Name (English) |
| `last_name_en` | string | Last Name (English) |
| `date_of_birth` | date | YYYY-MM-DD |
| `gender` | enum | 'M' or 'F' |
| `national_id` | string | 13-digit Thai ID |
| `passport_id` | string | Passport Number |
| `phone_number` | string | Contact Number |
| `email` | string | Email Address |
