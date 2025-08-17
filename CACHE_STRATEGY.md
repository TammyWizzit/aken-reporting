# 🚀 AKEN Reporting Service - Smart Caching Strategy

## 📋 **Overview**

The AKEN Reporting Service implements a **smart caching strategy** that balances performance with data freshness. We cache data that doesn't change frequently while ensuring transaction data is always fresh.

## 🎯 **What We Cache (30 minutes TTL)**

### ✅ **Cached Data Types**

#### **1. Merchant Summaries** 
```go
// Aggregated statistics - safe to cache
GET /api/v2/merchants/:id/summary
{
  "total_transactions": 1000,
  "successful_transactions": 850,
  "success_rate": 85.0,
  "total_amount": 50000
}
```
**Why Cache**: Aggregated data, expensive to calculate, doesn't change frequently

#### **2. Merchant Data**
```go
// Static merchant information
{
  "id": "merchant123",
  "name": "Test Merchant", 
  "merchant_code": "TEST123",
  "active": true
}
```
**Why Cache**: Rarely changes, frequently accessed

#### **3. Configuration Data**
```go
// System configuration and settings
{
  "field_mappings": {...},
  "filter_operators": {...},
  "default_values": {...}
}
```
**Why Cache**: Static data, expensive to load

## ❌ **What We DON'T Cache**

### **Transaction Data**
```go
// Fresh transaction data - never cached
GET /api/v2/transactions?page=1&limit=100
{
  "data": [...latest_transactions...],
  "total_count": 1500,
  "page": 1
}
```
**Why Not Cache**: 
- Users need to see latest transactions
- Data changes frequently
- Critical for real-time reporting
- Fresh data is more important than performance

## ⏱️ **Cache Duration: 30 Minutes**

### **Configuration**
```yaml
# docker-compose.yml
environment:
  - REDIS_TTL_SECONDS=1800  # 30 minutes
```

### **Benefits of 30-Minute TTL**
- **High Cache Hit Rate**: 80-95% for cached data types
- **Fresh Transaction Data**: Always up-to-date
- **Performance**: Fast responses for summaries and config
- **Database Load**: Significant reduction for expensive queries

## 🔄 **Cache Invalidation Strategy**

### **1. Automatic Expiration**
```go
// Redis automatically expires cached data after 30 minutes
// No manual cleanup needed
```

### **2. Manual Invalidation**
```go
// When merchant data changes
func (c *gin.Context) {
    if shouldInvalidateCache(c.Request.Method) {
        merchantID := getMerchantID(c)
        // Invalidate merchant cache only
        cacheService.InvalidateMerchantCache(merchantID)
        // Don't invalidate transaction cache (not cached anyway)
    }
}
```

## 📊 **Performance Impact**

### **Before Smart Caching**
```
All Requests → Database Query (50-200ms)
Cache Hit Rate: 0%
Database Load: 100%
```

### **After Smart Caching**
```
Transaction Requests → Database Query (50-200ms) - Always Fresh
Summary Requests → Cache Hit (1-5ms) - 80-95% hit rate
Config Requests → Cache Hit (1-5ms) - 95%+ hit rate
Overall Cache Hit Rate: 40-60%
Database Load: 40-60% reduction
```

## 🎯 **Cache Key Strategy**

### **1. Merchant Summary Keys**
```go
// Format: aken:reporting:summary:merchant123:filter_hash
"aken:reporting:a1b2c3d4" // merchant123, no filters
"aken:reporting:e5f6g7h8" // merchant123, date range filter
"aken:reporting:i9j0k1l2" // merchant123, amount filter
```

### **2. Merchant Data Keys**
```go
// Format: aken:reporting:merchant:merchant123
"aken:reporting:m3n4o5p6" // merchant123 data
```

## 🛡️ **Fail-Safe Design**

### **1. Graceful Degradation**
```go
// If Redis is down, service continues without caching
func NewCacheService() (CacheService, error) {
    if !config.IsRedisEnabled() {
        return &noOpCacheService{}, nil
    }
    // ... Redis connection logic
}
```

### **2. No-Op Cache Service**
```go
// When Redis unavailable, these methods do nothing
func (n *noOpCacheService) GetCachedMerchantSummary(key string) (*MerchantSummary, error) { 
    return nil, nil // Always cache miss
}
```

## 📈 **Expected Results**

### **Performance Improvements**
- **Transaction Data**: Always fresh (no caching)
- **Summary Data**: 80-95% cache hit rate
- **Config Data**: 95%+ cache hit rate
- **Overall**: 40-60% reduction in database load

### **User Experience**
- **Transaction Queries**: Fresh data, 50-200ms response
- **Summary Queries**: Fast cached data, 1-5ms response
- **Config Queries**: Instant cached data, 1-5ms response

## 🔧 **Monitoring & Debugging**

### **1. Cache Headers**
```go
// Response headers indicate cache status
X-Cache-Status: HIT/MISS
X-Cache-TTL: 1800
```

### **2. Cache Metrics**
```go
// Track cache performance
type CacheMetrics struct {
    SummaryHitRate: float64
    ConfigHitRate:  float64
    TransactionCacheHitRate: float64 // Should be 0%
    AverageResponseTime: time.Duration
}
```

## 🎯 **Summary**

This smart caching strategy provides:
- ✅ **Fresh transaction data** (never cached)
- ✅ **Fast summary responses** (cached for 30 minutes)
- ✅ **Instant config access** (cached for 30 minutes)
- ✅ **High performance** (40-60% database load reduction)
- ✅ **Fail-safe operation** (works without Redis)

The perfect balance of performance and data freshness! 🚀
