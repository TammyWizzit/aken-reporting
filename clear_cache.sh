#!/bin/bash

# Clear Redis cache for the reporting service

echo "Clearing Redis cache for AKEN Reporting Service..."

# Check if docker-compose is available
if command -v docker-compose &> /dev/null; then
    echo "Using docker-compose..."
    docker-compose exec redis redis-cli FLUSHDB
elif command -v docker &> /dev/null; then
    # Try to find redis container
    REDIS_CONTAINER=$(docker ps --format "table {{.Names}}" | grep redis | head -n 1)
    if [ ! -z "$REDIS_CONTAINER" ]; then
        echo "Using docker with container: $REDIS_CONTAINER"
        docker exec $REDIS_CONTAINER redis-cli FLUSHDB
    else
        echo "No Redis container found"
        exit 1
    fi
else
    echo "Neither docker-compose nor docker found"
    exit 1
fi

echo "Redis cache cleared successfully!"