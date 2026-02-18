# Redis Layer

## Purpose

Redis is used for:

1. Event ingestion buffering (Redis Streams)
2. Real-time campaign counters
3. Idempotency key protection
4. Worker crash recovery

It is NOT used as a traditional cache.

---

## Why Redis Streams Instead of List or Pub/Sub?

Pub/Sub:
- Not durable
- Messages lost on crash

List:
- No consumer groups
- Hard crash recovery
- No pending tracking

Redis Streams:
- Persistent
- Supports consumer groups
- Supports ACK model
- Supports message claiming
- Perfect for async pipelines

---

## Pool Configuration

PoolSize: 50
MinIdleConns: 10

Why?

High write systems need:
- Many concurrent connections
- Avoid connection creation overhead

Timeouts are important to:
- Prevent hanging requests
- Detect Redis failures quickly

---

## Why Ping On Startup?

If Redis is down:
- Fail fast
- Do not start API
- Avoid partial system boot

Failing fast is better than degraded undefined behavior.

---

## Environment Variables Required

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

If Redis runs in Docker:
Make sure port 6379 is exposed.

---

## What Would Break Without Redis?

If events go directly to DB:

- API latency spikes during DB slowdown
- No buffering
- Harder crash recovery
- Harder horizontal scaling

Redis decouples ingestion from persistence.

---

## Production Considerations

In real deployment:

- Redis may run as managed service
- Monitor:
  - Memory usage
  - Stream length
  - Consumer lag
  - Blocked clients

This layer protects the database from burst traffic.
