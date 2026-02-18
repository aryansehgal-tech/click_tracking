# API Entry Point – Session Creation

## Purpose

This is the first entry point of the tracking system.

It:

- Connects to PostgreSQL
- Exposes health endpoint
- Creates tracking sessions

This is the anchor for all future events.

---

## Endpoint: POST /sessions

### Request

{
  "campaign_id": "uuid"
}

### Response

{
  "session_id": "uuid"
}

---

## Why Validate UUID at API Layer?

We validate campaign_id:

binding:"required,uuid"

Because:

- Prevents invalid data reaching DB
- Protects DB from unnecessary load
- Fails fast

Always validate at edge.

---

## Why Generate UUID in API?

session.ID = uuid.New()

Why not DB?

- Avoid DB coordination
- Faster writes
- Supports horizontal scaling
- Enables stateless API servers

---

## Why Use UTC?

time.Now().UTC()

Tracking systems must:

- Avoid timezone bugs
- Keep analytics consistent
- Ensure multi-region compatibility

Never store local time.

---

## Why Health Endpoint?

GET /health

Used by:

- Load balancers
- Kubernetes probes
- Monitoring systems

Even small systems should have it.

---

## What Would Break at Scale Without This Design?

If we generate IDs in DB:
- Sequence contention
- Harder batching later

If no validation:
- Bad data enters system
- Reporting corruption

If no health check:
- Harder deployment automation

If no UTC:
- Analytics bugs
- Time range queries fail

---

## Current Flow

Client
   ↓
POST /sessions
   ↓
DB insert
   ↓
Return session_id

This is synchronous write.

Event ingestion will later become asynchronous.
