# eFinance v1 API Documentation

## Overview

The eFinance v1 API provides legacy transaction lookup and search functionality specifically designed for compatibility with existing eFinance systems. These APIs connect to the MySQL database and maintain backward compatibility with the original eFinance integration patterns.

**Base URL:** `{service_url}/api/v1/efinance`

**Authentication:** JWT Bearer Token (same as v2 APIs)

**Database:** MySQL (Atlas database)

---

## Quick Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/transactions/totals` | POST | Get transaction totals by date and device |
| `/transactions/lookup` | POST | Search for specific transactions |

---

## Authentication

All eFinance v1 APIs require JWT authentication. Include the Bearer token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

**Note:** Authentication is handled via PostgreSQL database, while transaction data is retrieved from MySQL.

---

## API Endpoints

### 1. Get Transaction Totals

**Endpoint:** `POST /api/v1/efinance/transactions/totals`

Retrieves transaction totals grouped by transaction type for a specific date and device.

#### Request

**Content-Type:** `application/json`

```json
{
  "date": "2024-01-15",
  "device_id": "DEVICE001"
}
```

**Request Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `date` | string | Yes | Transaction date in YYYY-MM-DD format |
| `device_id` | string | No | Unique identifier for the device/terminal. If not provided, totals will be returned for all devices |

#### Response

**Status Code:** `200 OK`

```json
{
  "date": "2024-01-15",
  "device_id": "DEVICE001",
  "totals": [
    {
      "trx_descr": "Cash Withdrawal",
      "total_amount_egp": 15000.50
    },
    {
      "trx_descr": "Balance Inquiry",
      "total_amount_egp": 0.00
    },
    {
      "trx_descr": "Transfer",
      "total_amount_egp": 8500.25
    }
  ]
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `date` | string | The requested date |
| `device_id` | string | The requested device ID |
| `totals` | array | Array of transaction type totals |
| `totals[].trx_descr` | string | Transaction description/type |
| `totals[].total_amount_egp` | number | Total amount in Egyptian Pounds |

#### Example Requests

**With device_id (specific device):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/totals" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15",
    "device_id": "DEVICE001"
  }'
```

**Without device_id (all devices):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/totals" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15"
  }'
```

---

### 2. Search Transaction Details

**Endpoint:** `POST /api/v1/efinance/transactions/lookup`

Searches for specific transaction details based on various criteria including PAN ID, RRN, device ID, and other transaction attributes. By default, returns all transactions (both successful and failed). Use the `response_code` filter to filter by transaction status.

#### Request

**Content-Type:** `application/json`

```json
{
  "panid": "123456****1234",
  "trx_rrn": "240115123456",
  "device_id": "DEVICE001",
  "bank_group_id": "BANK001",
  "amount": 500000,
  "trx_descr": "Cash Withdrawal",
  "date": "2024-01-15"
}
```

**Request Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `panid` | string | No | Masked PAN identifier (optional filter) |
| `trx_rrn` | string | No | Transaction Reference Number (optional filter) |
| `device_id` | string | No | Device/terminal identifier (optional filter) |
| `group_id` | string | No | Transaction group identifier from element 41 (optional filter) |
| `bank_group_id` | string | No | Bank group identifier from request_meta (optional filter) |
| `amount` | integer | No | Transaction amount in minor units (piasters) (optional filter) |
| `trx_descr` | string | No | Transaction description/type (optional filter) |
| `date` | string | No | Transaction date in YYYY-MM-DD format (optional filter) |
| `tx_id` | string | No | Transaction ID from request_meta (optional filter) |
| `response_code` | string | No | Response code (RC) filter (optional filter) |

#### Response

**Status Code:** `200 OK`

```json
{
  "transactions": [
    {
      "datetime": "2024-01-15T10:30:45Z",
      "STAN": 123456,
      "trx_rrn": "240115123456",
      "BIN": "123456",
      "PANID": "123456****1234",
      "device_id": "DEVICE001",
      "group_id": "GRP001",
      "trx_descr": "Cash Withdrawal",
      "trx_type": "WITHDRAWAL",
      "bank_group_id": "BANK001",
      "transaction_code": "01",
      "tx_id": "TXN-2024-001-123456",
      "amount": 500000,
      "RC": "00",
      "trx_auth_code": "123456"
    }
  ]
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `transactions` | array | Array of matching transactions |
| `transactions[].datetime` | string | Transaction timestamp (ISO 8601) |
| `transactions[].STAN` | integer | System Trace Audit Number |
| `transactions[].trx_rrn` | string | Retrieval Reference Number |
| `transactions[].BIN` | string | Bank Identification Number |
| `transactions[].PANID` | string | Masked PAN identifier |
| `transactions[].device_id` | string | Device/terminal identifier |
| `transactions[].group_id` | string | Transaction group identifier (from element 41) |
| `transactions[].trx_descr` | string | Transaction description |
| `transactions[].trx_type` | string | Transaction type code |
| `transactions[].bank_group_id` | string | Bank group identifier from request_meta (nullable) |
| `transactions[].transaction_code` | string | ISO transaction code (nullable) |
| `transactions[].tx_id` | string | Transaction ID from request_meta (nullable) |
| `transactions[].amount` | integer | Amount in minor units (piasters) |
| `transactions[].RC` | string | Response code |
| `transactions[].trx_auth_code` | string | Transaction authorization code (nullable) |

#### Example Requests

**All filters (most specific search):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "panid": "123456****1234",
    "trx_rrn": "240115123456",
    "device_id": "DEVICE001",
    "bank_group_id": "BANK001",
    "amount": 500000,
    "trx_descr": "Cash Withdrawal",
    "date": "2024-01-15"
  }'
```

**Date only (all transactions for a date):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15"
  }'
```

**Date + device_id (all transactions for a specific device):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15",
    "device_id": "DEVICE001"
  }'
```

**Date + amount + trx_descr (transactions by amount and type):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15",
    "amount": 500000,
    "trx_descr": "Cash Withdrawal"
  }'
```

**No filters (all successful transactions):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Date + tx_id (specific transaction by ID):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15",
    "tx_id": "TXN-2024-001-123456"
  }'
```

**Response code filter (successful transactions only):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15",
    "response_code": "00"
  }'
```

**Response code filter (failed transactions):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-01-15",
    "response_code": "05"
  }'
```

**Device only (all transactions for a specific device across all dates):**
```bash
curl -X POST "https://api.example.com/api/v1/efinance/transactions/lookup" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "DEVICE001"
  }'
```

---

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "timestamp": "2024-01-15T10:30:45Z",
  "request_id": "req_1705321845_123456"
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `BAD_REQUEST` | Invalid request format or missing required fields |
| 401 | `AUTH_FAILED` | Invalid or missing authentication credentials |
| 403 | `AUTHZ_FAILED` | Insufficient permissions |
| 404 | `TX_NOT_FOUND` | No transactions found matching criteria |
| 500 | `DATABASE_ERROR` | Internal database error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

### Example Error Response

```json
{
  "code": "BAD_REQUEST",
  "message": "Invalid request body: Key: 'TransactionLookupRequest.Date' Error:Field validation for 'Date' failed on the 'required' tag",
  "timestamp": "2024-01-15T10:30:45Z",
  "request_id": "req_1705321845_123456"
}
```

---

## Data Types and Formats

### Date Format
- **Format:** YYYY-MM-DD
- **Example:** "2024-01-15"
- **Timezone:** UTC

### Amount Format
- **Unit:** Minor currency units (piasters for EGP)
- **Example:** 500000 = 5000.00 EGP
- **Type:** Integer

### Device ID Format
- **Type:** String
- **Example:** "DEVICE001", "ATM_001", "POS_123"

### PAN ID Format
- **Type:** Masked string
- **Format:** First 6 digits + asterisks + last 4 digits
- **Example:** "123456****1234"

---

## Rate Limiting

The eFinance v1 APIs are subject to the same rate limiting as other service APIs:

- **Default:** 1000 requests per hour per client
- **Headers:** Rate limit information is included in response headers
- **Exceeded:** HTTP 429 status with retry-after information

---

## Database Architecture

### Database Selection
- **Authentication:** PostgreSQL database
- **Transaction Data:** MySQL database (Atlas)
- **Selection Logic:** Automatic based on API endpoint (/api/v1/efinance/* uses MySQL)

### Connection Details
The service maintains dual database connections:
- PostgreSQL for user authentication and v2 APIs
- MySQL for eFinance legacy transaction data

---

## Migration Notes

### From Legacy eFinance APIs

1. **Base URL Change:** Update from legacy endpoint to `/api/v1/efinance`
2. **Authentication:** Add JWT Bearer token authentication
3. **Request Format:** Ensure JSON content-type headers
4. **Response Format:** Response structure remains compatible
5. **Error Handling:** Updated error codes and format

### Backward Compatibility

The v1 eFinance APIs maintain backward compatibility with:
- Request/response data structures
- Transaction amount formats (minor units)
- Date formats and field names
- Core business logic and calculations

---

## Testing and Development

### Postman Collection

A comprehensive Postman collection is available including:
- Authentication examples
- Request templates for all endpoints
- Environment variables for different deployment stages
- Error scenario test cases

### Sample Test Data

```json
// Valid test request for totals
{
  "date": "2024-01-15",
  "device_id": "TEST_DEVICE_001"
}

// Valid test request for lookup
{
  "panid": "123456****1234",
  "trx_rrn": "240115123456",
  "device_id": "TEST_DEVICE_001",
  "bank_group_id": "TEST_BANK",
  "amount": 100000,
  "trx_descr": "Balance Inquiry",
  "date": "2024-01-15"
}
```

---

## Support and Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify JWT token is valid and not expired
   - Ensure Bearer token format in Authorization header

2. **No Data Returned**
   - Check date format (YYYY-MM-DD)
   - Verify device_id exists in the system
   - Confirm transaction data exists for the specified criteria

3. **Database Connection Issues**
   - Service handles MySQL connection failures gracefully
   - Check service health endpoint: `/api/v2/health`

### Logging and Monitoring

- All requests are logged with trace IDs for debugging
- Database queries include performance metrics
- Error rates and response times are monitored

### Health Checks

Monitor service health:
```bash
curl -X GET "https://api.example.com/api/v2/health"
```

The health endpoint reports status for both PostgreSQL and MySQL connections.

---

## Appendix

### Transaction Types (trx_descr)

Common transaction descriptions:
- "Cash Withdrawal"
- "Balance Inquiry" 
- "Transfer"
- "Mini Statement"
- "PIN Change"
- "Deposit"

### Response Codes (RC)

Common response codes:
- "00" - Successful
- "01" - Refer to card issuer
- "03" - Invalid merchant
- "05" - Do not honor
- "12" - Invalid transaction
- "13" - Invalid amount
- "14" - Invalid card number

### Amount Conversion

Egyptian Pound (EGP) conversion:
- API amounts are in piasters (1/100 EGP)
- 100000 piasters = 1000.00 EGP
- Always use integer values for amounts