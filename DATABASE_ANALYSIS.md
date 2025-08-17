# Database Schema Analysis

## Table Relationships

Based on the provided database schema and sample data, here are the key relationships:

### 1. Devices Table
```sql
devices (
    device_id (UUID, PK),
    merchant_id (UUID, FK to merchants),
    deviceid (VARCHAR(255)), -- This is the actual device identifier
    msisdn (VARCHAR(15)),
    terminal_id (UUID, FK to terminals),
    -- other fields...
)
```

### 2. Terminals Table
```sql
terminals (
    terminal_id (UUID, PK),
    merchant_id (UUID, FK to merchants),
    bank_terminal_id (VARCHAR(255)),
    trx_type (VARCHAR(3)),
    -- other fields...
)
```

### 3. Merchants Table
```sql
merchants (
    merchant_id (UUID, PK),
    name (VARCHAR),
    provisioner_id (UUID),
    is_provisioner (BOOLEAN),
    -- other fields...
)
```

### 4. Payment Transaction Log Table
```sql
payment_tx_log (
    payment_tx_log_id (UUID, PK),
    device_id (VARCHAR), -- This joins with devices.deviceid
    merchant_id (UUID, FK to merchants),
    amount (INTEGER),
    result_code (VARCHAR),
    -- other fields...
)
```

## Key Join Relationships

### Correct Join Logic:
1. `payment_tx_log.device_id` → `devices.deviceid` (string to string)
2. `devices.terminal_id` → `terminals.terminal_id` (UUID to UUID)
3. `payment_tx_log.merchant_id` → `merchants.merchant_id` (UUID to UUID)

### Sample Data Flow:
```
payment_tx_log.device_id = "6C8632B4CC9364BCD8AF830131A0E0016EB5795B6A0F0B177F0D2670CB4639CC4B2368C6006839D264A38B7F58E5C8130447528BF4B7AEE1"
↓
devices.deviceid = "6C8632B4CC9364BCD8AF830131A0E0016EB5795B6A0F0B177F0D2670CB4639CC4B2368C6006839D264A38B7F58E5C8130447528BF4B7AEE1"
↓
devices.terminal_id = "9decde50-6ad5-11f0-9b79-c9f2d97c866a"
↓
terminals.terminal_id = "9decde50-6ad5-11f0-9b79-c9f2d97c866a"
↓
terminals.bank_terminal_id = "000001223435456"
```

## Issues Fixed

### 1. Column Name Mismatch
**Problem:** Code was using `d.deviceId` but actual column is `d.deviceid` (lowercase)

**Solution:** Updated all joins to use `d.deviceid`

### 2. Error Handling
**Problem:** Internal database errors were being exposed to API consumers

**Solution:** 
- Added error sanitization
- Removed nested error objects
- Added user-friendly error messages
- Prevented internal system details from leaking

## Sample Transaction Data

From the provided data, we can see:
- **Merchant:** "New WL App" (ID: 9dea9460-6ad5-11f0-9b79-c9f2d97c866a)
- **Device:** MSISDN 27721289793
- **Terminal:** Bank Terminal ID 000001223435456
- **Transactions:** 14 transactions with various amounts and result codes

### Transaction Statistics:
- **Total Transactions:** 14
- **Successful (00):** 9 transactions
- **Failed:** 5 transactions (codes: 57, 05, 91)
- **Total Amount:** 1,175,000 (in cents = R11,750.00)
- **Success Rate:** ~64%

## API Response Format

### Before (Problematic):
```json
{
    "error": {
        "code": "DATABASE_ERROR",
        "message": "Failed to get merchant summary: failed to get merchant summary: failed to get merchant summary: ERROR: column p.device_id does not exist (SQLSTATE 42703)",
        "request_id": "req_1755367656_26604257",
        "timestamp": "2025-08-16T18:07:36Z"
    }
}
```

### After (Fixed):
```json
{
    "code": "SERVICE_UNAVAILABLE",
    "message": "Service temporarily unavailable. Please try again later.",
    "timestamp": "2025-08-16T18:07:36Z",
    "request_id": "req_1755367656_26604257"
}
```

## Testing Recommendations

1. **Test with Real Data:** Use the provided sample data to test the API
2. **Verify Joins:** Ensure all table joins work correctly
3. **Test Error Scenarios:** Verify error handling works as expected
4. **Performance Testing:** Test with larger datasets
