package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/models"
	"aken_reporting_service/internal/repositories"

	"github.com/go-redis/redis/v8"
)

// CacheService provides Redis caching functionality
type CacheService interface {
	// Transaction caching
	GetCachedTransactions(key string) (*repositories.TransactionListResult, error)
	SetCachedTransactions(key string, result *repositories.TransactionListResult, ttl time.Duration) error
	InvalidateTransactionCache(merchantID string) error
	
	// Merchant caching
	GetCachedMerchant(merchantID string) (*models.Merchant, error)
	SetCachedMerchant(merchantID string, merchant *models.Merchant, ttl time.Duration) error
	InvalidateMerchantCache(merchantID string) error
	
	// Summary caching
	GetCachedMerchantSummary(key string) (*models.MerchantSummary, error)
	SetCachedMerchantSummary(key string, summary *models.MerchantSummary, ttl time.Duration) error
	
	// Generic caching
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	DeletePattern(pattern string) error
	
	// Health check
	Ping() error
	Close() error
}

type cacheService struct {
	client *redis.Client
	prefix string
	ctx    context.Context
}

// NewCacheService creates a new Redis cache service
func NewCacheService() (CacheService, error) {
	if !config.IsRedisEnabled() {
		return &noOpCacheService{}, nil
	}

	redisConfig := config.GetRedisConfig()
	
	// Configure Redis options
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		DialTimeout:  redisConfig.Timeout,
		ReadTimeout:  redisConfig.Timeout,
		WriteTimeout: redisConfig.Timeout,
	}
	
	// Only set password if it's not empty
	if redisConfig.Password != "" {
		options.Password = redisConfig.Password
	}
	
	client := redis.NewClient(options)

	ctx := context.Background()
	
	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &cacheService{
		client: client,
		prefix: config.GetRedisKeyPrefix(),
		ctx:    ctx,
	}, nil
}

// generateCacheKey creates a cache key with prefix and hash
func (c *cacheService) generateCacheKey(parts ...string) string {
	key := c.prefix + fmt.Sprintf("%v", parts)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%s:%x", c.prefix, hash[:8])
}

// GetCachedTransactions retrieves cached transaction results
func (c *cacheService) GetCachedTransactions(key string) (*repositories.TransactionListResult, error) {
	cacheKey := c.generateCacheKey("transactions", key)
	
	data, err := c.client.Get(c.ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached transactions: %v", err)
	}

	var result repositories.TransactionListResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached transactions: %v", err)
	}

	return &result, nil
}

// SetCachedTransactions stores transaction results in cache
func (c *cacheService) SetCachedTransactions(key string, result *repositories.TransactionListResult, ttl time.Duration) error {
	cacheKey := c.generateCacheKey("transactions", key)
	
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal transactions for cache: %v", err)
	}

	if err := c.client.Set(c.ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cached transactions: %v", err)
	}

	return nil
}

// InvalidateTransactionCache removes cached transactions for a merchant
func (c *cacheService) InvalidateTransactionCache(merchantID string) error {
	pattern := c.generateCacheKey("transactions", merchantID) + "*"
	return c.DeletePattern(pattern)
}

// GetCachedMerchant retrieves cached merchant data
func (c *cacheService) GetCachedMerchant(merchantID string) (*models.Merchant, error) {
	cacheKey := c.generateCacheKey("merchant", merchantID)
	
	data, err := c.client.Get(c.ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached merchant: %v", err)
	}

	var merchant models.Merchant
	if err := json.Unmarshal(data, &merchant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached merchant: %v", err)
	}

	return &merchant, nil
}

// SetCachedMerchant stores merchant data in cache
func (c *cacheService) SetCachedMerchant(merchantID string, merchant *models.Merchant, ttl time.Duration) error {
	cacheKey := c.generateCacheKey("merchant", merchantID)
	
	data, err := json.Marshal(merchant)
	if err != nil {
		return fmt.Errorf("failed to marshal merchant for cache: %v", err)
	}

	if err := c.client.Set(c.ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cached merchant: %v", err)
	}

	return nil
}

// InvalidateMerchantCache removes cached merchant data
func (c *cacheService) InvalidateMerchantCache(merchantID string) error {
	cacheKey := c.generateCacheKey("merchant", merchantID)
	return c.Delete(cacheKey)
}

// GetCachedMerchantSummary retrieves cached merchant summary
func (c *cacheService) GetCachedMerchantSummary(key string) (*models.MerchantSummary, error) {
	cacheKey := c.generateCacheKey("summary", key)
	
	data, err := c.client.Get(c.ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached summary: %v", err)
	}

	var summary models.MerchantSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached summary: %v", err)
	}

	return &summary, nil
}

// SetCachedMerchantSummary stores merchant summary in cache
func (c *cacheService) SetCachedMerchantSummary(key string, summary *models.MerchantSummary, ttl time.Duration) error {
	cacheKey := c.generateCacheKey("summary", key)
	
	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary for cache: %v", err)
	}

	if err := c.client.Set(c.ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cached summary: %v", err)
	}

	return nil
}

// Get retrieves a value from cache
func (c *cacheService) Get(key string, dest interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil // Cache miss
		}
		return fmt.Errorf("failed to get from cache: %v", err)
	}

	return json.Unmarshal(data, dest)
}

// Set stores a value in cache
func (c *cacheService) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for cache: %v", err)
	}

	return c.client.Set(c.ctx, key, data, ttl).Err()
}

// Delete removes a key from cache
func (c *cacheService) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// DeletePattern removes keys matching a pattern
func (c *cacheService) DeletePattern(pattern string) error {
	iter := c.client.Scan(c.ctx, 0, pattern, 0).Iterator()
	for iter.Next(c.ctx) {
		if err := c.client.Del(c.ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

// Ping tests Redis connection
func (c *cacheService) Ping() error {
	return c.client.Ping(c.ctx).Err()
}

// Close closes Redis connection
func (c *cacheService) Close() error {
	return c.client.Close()
}

// noOpCacheService provides a no-operation cache service when Redis is disabled
type noOpCacheService struct{}

func (n *noOpCacheService) GetCachedTransactions(key string) (*repositories.TransactionListResult, error) { return nil, nil }
func (n *noOpCacheService) SetCachedTransactions(key string, result *repositories.TransactionListResult, ttl time.Duration) error { return nil }
func (n *noOpCacheService) InvalidateTransactionCache(merchantID string) error { return nil }
func (n *noOpCacheService) GetCachedMerchant(merchantID string) (*models.Merchant, error) { return nil, nil }
func (n *noOpCacheService) SetCachedMerchant(merchantID string, merchant *models.Merchant, ttl time.Duration) error { return nil }
func (n *noOpCacheService) InvalidateMerchantCache(merchantID string) error { return nil }
func (n *noOpCacheService) GetCachedMerchantSummary(key string) (*models.MerchantSummary, error) { return nil, nil }
func (n *noOpCacheService) SetCachedMerchantSummary(key string, summary *models.MerchantSummary, ttl time.Duration) error { return nil }
func (n *noOpCacheService) Get(key string, dest interface{}) error { return nil }
func (n *noOpCacheService) Set(key string, value interface{}, ttl time.Duration) error { return nil }
func (n *noOpCacheService) Delete(key string) error { return nil }
func (n *noOpCacheService) DeletePattern(pattern string) error { return nil }
func (n *noOpCacheService) Ping() error { return nil }
func (n *noOpCacheService) Close() error { return nil }
