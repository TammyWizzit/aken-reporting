# AKEN Reporting Service v2.0 - Postman Collection

This directory contains the complete Postman collection for testing and exploring the AKEN Reporting Service v2.0 API.

## 📦 Collection Contents

### Main Collection
- **File**: `AKEN_Reporting_Service_v2.postman_collection.json`
- **Requests**: 70+ comprehensive API test requests
- **Categories**: System Health, Transactions, Merchants, Future Features
- **Features**: Pre/post-request scripts, automated testing, variable support

### Environment Files
- **Development**: `environments/Development.postman_environment.json`
- **Staging**: `environments/Staging.postman_environment.json` 
- **Production**: `environments/Production.postman_environment.json`

## 🚀 Quick Setup

### 1. Import Collection
1. Open Postman
2. Click **Import** 
3. Select `AKEN_Reporting_Service_v2.postman_collection.json`
4. Collection will appear as "AKEN Reporting Service v2.0"

### 2. Import Environment
1. Click **Import** again
2. Select your environment file:
   - `environments/Development.postman_environment.json` for local testing
   - `environments/Staging.postman_environment.json` for staging
   - `environments/Production.postman_environment.json` for production
3. Select the imported environment from the dropdown

### 3. Configure Variables
Update these key variables in your selected environment:
- `base_url`: Your service URL (default: http://localhost:8090)
- `merchant_id`: Your merchant UUID
- `merchant_password`: Your merchant password
- `transaction_id`: Sample transaction ID for testing

## 📋 Collection Structure

### 🏥 System Health (3 requests)
- **Health Check**: Basic service health verification
- **API Info**: Detailed API information and capabilities  
- **Debug Info**: Development debugging information

### 📊 Transactions (8 requests)
- **List All Transactions**: Basic transaction retrieval
- **Transactions with Field Selection**: Specific field filtering
- **Filtered Transactions - Success Only**: Response code filtering
- **Date Range Filter**: Time-based filtering with timezone support
- **Complex Filter - Amount & Status**: Boolean logic filtering
- **Paginated Results**: Pagination example
- **Get Single Transaction**: Individual transaction lookup
- **Advanced Search Query**: Elasticsearch-style complex search

### 🏢 Merchants (3 requests)
- **Merchant Summary**: Comprehensive merchant statistics
- **Merchant Summary with Date Filter**: Time-filtered statistics
- **Merchant Transactions**: Merchant-specific transaction list

### 🚧 Future Features (3 requests)
- **Export Transactions**: Bulk export functionality (501 Not Implemented)
- **Batch Operations**: Multiple request batching (501 Not Implemented)
- **Analytics Summary**: Advanced analytics (501 Not Implemented)

## 🔧 Pre-configured Features

### Authentication
- **Basic Auth** pre-configured using environment variables
- **Dynamic credentials** using `{{merchant_id}}` and `{{merchant_password}}`
- **Development mode** support with DISABLE_AUTH

### Request Scripts
- **Pre-request**: Automatic timestamp and request ID generation
- **Global variables**: Shared across all requests
- **Dynamic values**: Request correlation and tracking

### Response Testing
- **Global tests**: Applied to all requests automatically
- **Response time validation**: Ensures performance standards
- **Structure validation**: Verifies response format
- **Security header checks**: API version and service headers
- **Error format validation**: Structured error response verification

## 🎯 Example Usage

### Basic Testing Flow

1. **Start with Health Check**:
   - Run "Health Check" to verify service is running
   - Expected: 200 OK with healthy status

2. **Test Authentication**:
   - Run "List All Transactions" 
   - Expected: 200 OK with transaction data (or 401 if auth failed)

3. **Explore Filtering**:
   - Run "Filtered Transactions - Success Only"
   - Modify filter parameter to test different conditions
   - Expected: Filtered results matching criteria

4. **Test Pagination**:
   - Run "Paginated Results - Page 2"
   - Check pagination metadata in response
   - Expected: Page 2 data with correct pagination links

5. **Advanced Features**:
   - Run "Advanced Search Query"
   - Explore complex filtering with request body
   - Expected: Aggregated results with metadata

### Field Selection Examples

```bash
# Modify the fields parameter in requests:

# Minimal fields for performance
fields=tx_log_id,amount,merchant_name

# Extended transaction details  
fields=tx_log_id,amount,merchant_name,tx_date_time,response_code,auth_code,rrn

# Complete transaction record
fields=tx_log_id,tx_log_type,tx_date_time,amount,merchant_id,merchant_name,device_id,response_code,auth_code,rrn,pan,reversed,settlement_status,stan,user_ref,meta,settlement_date,card_type
```

### Filter Examples

```bash
# Simple filters
filter=response_code:eq:00
filter=amount:gte:1000  
filter=merchant_id:eq:123

# Date ranges
filter=tx_date_time:between:2024-01-01,2024-12-31
filter=tx_date_time:gte:2024-01-01

# Complex boolean logic
filter=(response_code:eq:00 OR response_code:eq:10) AND amount:gte:1000
filter=merchant_id:eq:123 AND tx_date_time:between:2024-01-01,2024-12-31 AND NOT reversed:eq:true
```

## 🔍 Testing Scenarios

### Performance Testing
1. **Load Testing**: Use "List All Transactions" with varying limit values
2. **Field Selection**: Test different field combinations for payload optimization
3. **Complex Filters**: Test advanced filtering for query performance

### Data Validation
1. **Pagination**: Verify page navigation works correctly
2. **Sorting**: Test different sort combinations
3. **Field Filtering**: Ensure only requested fields are returned
4. **Error Handling**: Test invalid parameters for proper error responses

### Authentication Testing
1. **Valid Credentials**: Normal operation testing
2. **Invalid Credentials**: 401 error testing  
3. **Development Mode**: DISABLE_AUTH=true testing
4. **Production Mode**: Full authentication flow

## 🌍 Environment Configuration

### Development Environment
```json
{
  "base_url": "http://localhost:8090",
  "merchant_id": "9cda37a0-4813-11ef-95d7-c5ac867bb9fc",
  "merchant_password": "password",
  "transaction_id": "f3b24260-903f-11ef-b4a1-4982815eb203"
}
```

### Staging Environment
```json
{
  "base_url": "https://aken-reporting-staging.wizzitdigital.com",
  "merchant_id": "staging-merchant-uuid",
  "merchant_password": "staging-password",
  "transaction_id": "staging-transaction-uuid"
}
```

### Production Environment
```json
{
  "base_url": "https://aken-reporting.wizzitdigital.com", 
  "merchant_id": "production-merchant-uuid",
  "merchant_password": "production-password",
  "transaction_id": "production-transaction-uuid"
}
```

## 🔄 Collection Maintenance

### Adding New Requests
When adding new API endpoints:

1. **Create new request** in appropriate folder
2. **Set authentication** to inherit from parent
3. **Add example response** for documentation
4. **Include description** explaining the endpoint
5. **Add tests** for response validation

### Environment Updates
When API changes occur:

1. **Update collection** with new endpoints
2. **Test all requests** against new API version
3. **Update environment variables** if needed
4. **Document breaking changes** in descriptions

### Best Practices
1. **Use descriptive names** for requests and folders
2. **Include examples** in request descriptions
3. **Add response examples** for documentation
4. **Use variables** instead of hardcoded values
5. **Add tests** to validate responses
6. **Document complex scenarios** in descriptions

## 📞 Support & Troubleshooting

### Common Issues

1. **401 Unauthorized**:
   - Check merchant_id and merchant_password in environment
   - Verify service is running and accessible
   - Confirm DISABLE_AUTH setting for development

2. **Connection Refused**:
   - Verify base_url in environment
   - Check if service is running on correct port
   - Confirm network connectivity

3. **Empty Responses**:
   - Check if merchant has transaction data
   - Verify filter parameters aren't too restrictive
   - Confirm database connectivity

4. **Slow Responses**:
   - Reduce page size (limit parameter)
   - Use field selection to reduce payload
   - Check database performance and indexes

### Getting Help
- **API Documentation**: See `/DOCUMENTATION.md`
- **Developer Guide**: See `/README.md` 
- **Service Health**: Check `GET /api/v2/health`
- **Debug Info**: Check `GET /debug` (development only)

---

**Collection Version**: 2.0.0  
**Last Updated**: January 28, 2025  
**Compatible With**: AKEN Reporting Service v2.0+