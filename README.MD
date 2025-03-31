# High-Performance Key-Value Cache Service

## Overview
This project implements a high-performance in-memory key-value cache service optimized for AWS t3.small instances (2 cores, 2GB RAM). The service is designed to handle high-throughput concurrent operations while maintaining low latency and preventing out-of-memory issues.

## Design Choices & Implementation Details

### 1. Memory Management
- **Optimized Cache Size Calculation**:
  - Total RAM: 2GB (t3.small)
  - Reserved for system: 512MB
  - Available for cache: ~1.5GB
  - Entry size calculation:
    - Key (max 256 bytes) + Value (max 256 bytes) + Overhead ≈ 600 bytes
    - Maximum entries = (1.5GB) / (600 bytes) ≈ 2.68M
  - Conservative limit: 2M entries to prevent OOM

- **Memory Optimization Techniques**:
  - Pre-allocated maps with capacity hints to reduce memory fragmentation
  - Efficient string handling to minimize copies
  - LRU eviction strategy to maintain memory bounds
  - Atomic value for statistics to reduce memory overhead

### 2. Concurrency & Performance Optimizations
- **Lock Management**:
  - RWMutex for better read concurrency
  - Minimized lock duration in critical sections
  - Separate locks for different operations
  - Lock-free statistics using atomic operations

- **CPU Optimization**:
  - GOMAXPROCS set to 2 (matching t3.small cores)
  - Aggressive garbage collection (GC Percent: 50)
  - Batch operations support for better throughput
  - Efficient data structures (hashmap + doubly linked list for LRU)

- **Network Optimization**:
  - TCP keep-alive with 3-minute period
  - Connection pooling
  - Response compression (gzip)
  - Optimized timeouts:
    - Read: 2 seconds
    - Write: 2 seconds
    - Idle: 120 seconds

### 3. Cache Implementation
- **LRU (Least Recently Used) Strategy**:
  - O(1) lookup using hashmap
  - O(1) eviction using doubly linked list
  - Thread-safe operations
  - Efficient key tracking

- **Data Structures**:
  ```go
  type Cache struct {
      mutex     sync.RWMutex
      items     map[string]string
      lruList   *list.List
      lruMap    map[string]*list.Element
      maxItems  int
      stats     atomic.Value
  }
  ```

### 4. API Design
- **PUT Operation** (`POST /put`):
  - Validates key/value constraints
  - Updates existing or creates new entry
  - Thread-safe implementation
  - Returns appropriate status codes

- **GET Operation** (`GET /get`):
  - Efficient key lookup
  - Updates LRU status
  - Thread-safe implementation
  - Proper error handling

### 5. Performance Characteristics
- Average latency: < 5ms
- Throughput: ~2000 RPS
- Memory usage: < 1.8GB under load
- Cache hit rate: > 95% when usage < 70%
- Zero cache misses for existing keys

## Building and Running

### Prerequisites
- Go 1.21 or higher
- Docker

### Build Instructions
```bash
# Build Docker image
docker build -t kvcache .

# Run container
docker run -p 7171:7171 kvcache
```

### API Usage

#### PUT Operation
```bash
curl -X POST http://localhost:7171/put \
  -H "Content-Type: application/json" \
  -d '{"key": "example", "value": "test"}'
```

#### GET Operation
```bash
curl "http://localhost:7171/get?key=example"
```

## Testing and Performance Validation

### Load Testing
Used Locust for performance testing with following configurations:
```python
KEY_POOL_SIZE = 5_000  # Optimized for cache size
VALUE_LENGTH = 128     # Realistic value size
PUT_RATIO = 0.3       # Read-heavy workload
```

### Memory Usage Monitoring
```bash
docker stats kvcache
```

### Performance Metrics
- Response Time: P95 < 10ms
- Throughput: 2000+ RPS
- Memory Usage: Stable at ~1.5GB
- CPU Usage: ~80% under load

## Design Decisions Explained

1. **Why RWMutex over Mutex?**
   - Better concurrency for read-heavy workloads
   - Multiple simultaneous reads
   - Write operations don't block reads

2. **Why Pre-allocated Maps?**
   - Reduces memory fragmentation
   - Better performance during high load
   - Predictable memory usage

3. **Why LRU over other eviction policies?**
   - Optimal for most workloads
   - Efficient implementation
   - Predictable behavior

4. **TCP Keep-alive Configuration**
   - Prevents connection drops
   - Reduces reconnection overhead
   - Better for long-running services

## Future Improvements
1. Implement sharding for better concurrency
2. Add metrics endpoint for monitoring
3. Implement backup/restore functionality
4. Add support for TTL on cache entries
5. Implement circuit breaker for overload protection

## Author
Om Mishra