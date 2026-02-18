# Event Ingestion Handler

## Purpose

Handles POST /events.

This endpoint is write-heavy and must be extremely fast.

It does NOT write to PostgreSQL directly.

Instead:

API → Redis Stream → Worker → PostgreSQL

This decouples ingestion from storage.

---

## Flow

1. Validate input
2. Validate event type
3. Generate event_id
4. Push to Redis Stream
5. Increment real-time counters
6. Return 202 Accepted

---

## Why 202 Accepted?

Because:

- Event is queued
- Not yet persisted to DB
- Asynchronous pipeline

Returning 201 would imply DB write completed.

---

## Why Validate Event Types?

Prevents:

- Garbage data
- Broken reporting
- Counter pollution

Tracking systems must control event taxonomy.

---

## Why Use Redis Stream?

XAdd writes to persistent stream.

Benefits:

- Durable queue
- Supports consumer groups
- Supports ACK
- Supports replay
- Handles worker crashes safely

---

## Why Increment Counters Here?

We want real-time campaign metrics.

If we waited for DB:

- Slow
- Expensive COUNT queries
- No instant reporting

Redis counters give sub-millisecond updates.

---

## What Would Break Without This Design?

If writing directly to DB:

- DB overload during traffic bursts
- API latency spikes
- No buffering
- Hard crash recovery

If no event validation:

- Dirty data
- Broken aggregation

If no async queue:

- System tightly coupled
- Poor scalability

---

## Important Note

This endpoint does not guarantee durability.

Durability is guaranteed when:

Worker reads from stream
AND
Writes to PostgreSQL
AND
ACKs message

That part comes next.
