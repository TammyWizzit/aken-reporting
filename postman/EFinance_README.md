# EFinance API Postman Collection

This directory contains Postman collection and environment files for testing the EFinance Transaction Reporting Service APIs.

## Files

- `EFinance_APIs.postman_collection.json` - Main Postman collection with all API endpoints
- `EFinance_Local_Environment.postman_environment.json` - Environment variables for local development

## Collection Features

### Authentication
- **Generate JWT Token**: Creates a JWT token using merchant credentials
  - Automatically saves the token to environment variables
  - Used for authenticating subsequent requests
- **Verify JWT Token**: Validates an existing JWT token

### EFinance Transaction APIs
- **Get Transaction Totals** (`POST /api/v1/efinance/transactions/totals`)
  - Retrieves transaction totals by description for a specific date and device
  - Groups transactions by `trx_descr` and sums amounts
- **Search Transaction Details** (`POST /api/v1/efinance/transactions/lookup`)  
  - Searches for detailed transaction information based on multiple criteria
  - Returns comprehensive transaction details including STAN, RRN, BIN, etc.

### Health & System
- **Service Health**: Check service and database health
- **API Info**: Get information about available endpoints and features

## Setup Instructions

### 1. Import into Postman

1. Open Postman
2. Click **Import** button
3. Select **File** tab
4. Choose both JSON files:
   - `EFinance_APIs.postman_collection.json`
   - `EFinance_Local_Environment.postman_environment.json`

### 2. Configure Environment

1. In Postman, select the **EFinance Local Environment** from the environment dropdown
2. Update the following variables as needed:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `base_url` | API base URL | `http://localhost:8080` |
| `merchant_id` | Merchant ID for authentication | `MERCHANT001` |
| `merchant_password` | Merchant password | `password123` |
| `transaction_date` | Date for transaction queries | `2024-01-15` |
| `device_id` | Device ID for filtering | `DEV001` |
| `panid` | PAN ID for transaction search | `1234` |
| `trx_rrn` | Transaction RRN for search | `000000123456` |
| `bank_group_id` | Bank group ID | `BANK001` |
| `amount` | Transaction amount (in cents) | `10000` |
| `trx_descr` | Transaction description | `Purchase` |

### 3. Authentication Workflow

1. **First, generate a JWT token:**
   - Run the **Generate JWT Token** request
   - The token will be automatically saved to the `jwt_token` environment variable
   - This token will be used for all subsequent authenticated requests

2. **Verify the token (optional):**
   - Run the **Verify JWT Token** request to confirm the token is valid

3. **Use the transaction endpoints:**
   - All transaction requests will automatically use the saved JWT token
   - Update the request body parameters as needed for your testing

## Request Examples

### Generate JWT Token
```json
POST /api/v2/auth/generate-token
{
    "merchant_id": "MERCHANT001",
    "password": "password123"
}
```

### Get Transaction Totals
```json
POST /api/v1/efinance/transactions/totals
Authorization: Bearer <jwt_token>
{
    "date": "2024-01-15",
    "device_id": "DEV001"
}
```

### Search Transaction Details
```json
POST /api/v1/efinance/transactions/lookup
Authorization: Bearer <jwt_token>
{
    "panid": "1234",
    "trx_rrn": "000000123456",
    "device_id": "DEV001", 
    "bank_group_id": "BANK001",
    "amount": 10000,
    "trx_descr": "Purchase",
    "date": "2024-01-15"
}
```

## Response Examples

### Transaction Totals Response
```json
{
    "date": "2024-01-15",
    "device_id": "DEV001",
    "totals": [
        {
            "trx_descr": "Purchase",
            "total_amount_egp": 1250.75
        },
        {
            "trx_descr": "Cash Withdrawal", 
            "total_amount_egp": 500.00
        }
    ]
}
```

### Transaction Search Response
```json
{
    "transactions": [
        {
            "datetime": "2024-01-15T14:30:45",
            "STAN": 123456,
            "RRN": "000000123456",
            "BIN": "123456",
            "PANID": "1234",
            "device_id": "DEV001",
            "group_id": "GROUP001",
            "trx_descr": "Purchase",
            "trx_type": "00",
            "bank_group_id": "BANK001",
            "transaction_code": "TXN001",
            "amount": 10000,
            "RC": "00",
            "trx_auth_code": "123456"
        }
    ]
}
```

## Built-in Tests

The collection includes automated tests that:
- Verify response status codes
- Check response structure and required fields
- Validate response times (< 5 seconds)
- Automatically save JWT tokens to environment variables

## Environment Variables

### Automatic Token Management
- The **Generate JWT Token** request automatically saves the received token
- All authenticated requests use the saved token via `{{jwt_token}}`
- Token expiration is tracked in `{{token_expires_in}}`

### Pre-request Scripts
- Automatically set default values for missing environment variables
- Generate current date if `transaction_date` is not set
- Provide fallback values for all required parameters

## Troubleshooting

### Authentication Issues
- Ensure the service is running on the correct port (default: 8080)
- Verify merchant credentials in the database
- Check that JWT secret is properly configured in the service

### Request Failures  
- Verify all required fields are provided in request bodies
- Check date format (YYYY-MM-DD)
- Ensure amounts are provided as integers (cents)
- Verify device_id exists in the database

### Database Issues
- Confirm the `iso_trx` table has data for the specified date
- Check that the device_id exists in transaction records
- Verify successful transactions exist (trx_rsp_code = '00')

## Notes

- All amounts are in cents (e.g., 10000 = 100.00 EGP)
- Dates should be in YYYY-MM-DD format
- The collection automatically handles request IDs with `{{$randomUUID}}`
- Response times are monitored and should be under 5 seconds
- JWT tokens expire after 24 hours by default