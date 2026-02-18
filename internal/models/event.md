# Event Model

## Purpose

The Event model represents a single tracked action inside a Session.

Examples:
- link_opened
- page_view
- product_view
- add_to_cart
- checkout_started
- purchase_completed

Each Session can generate 20–30 Events.

This table is append-only and write-heavy.

---

## Field Breakdown

### ID (UUID)

Primary key of the event.

Why UUID?

- Generated at API layer
- Avoids DB coordination
- Supports horizontal scaling
- Required for idempotency

Later, this will also act as our idempotency key.

---

### SessionID (UUID)

Links event to a Session.

Indexed because:
- Session-level queries exist
- Debugging often traces by session

---

### CampaignID (UUID)

Denormalized from Session.

Why duplicate this field?

Without it:
- Every aggregation requires join with sessions
- Joins become expensive at scale
- Reporting slows significantly

This is intentional denormalization for performance.

---

### EventType (string)

Represents event category.

Examples:
- page_view
- product_view
- purchase_completed

Indexed because:
- Common filter in aggregation
- Used for real-time counters

We keep it string for flexibility.
In larger systems this could be enum.

---

### EventTime (time.Time)

Time the event occurred.

Important:
This is NOT the same as CreatedAt.

EventTime:
- Used for analytics
- May come from client or API layer

CreatedAt:
- When stored in DB

These can differ if ingestion is delayed.

---

### Metadata (jsonb)

Flexible event data.

Example:
{
  "product_id": "123",
  "price": 199.99,
  "currency": "USD"
}

Why jsonb?

- Tracking systems evolve constantly
- Avoids frequent schema migrations
- Allows indexing specific JSON fields later
- Efficient storage in PostgreSQL

---

### CreatedAt (time.Time)

Timestamp when event was inserted into database.

Used for:
- Debugging ingestion delay
- Monitoring pipeline lag

---

## Index Strategy

We index:
- SessionID
- CampaignID
- EventType
- EventTime

Why?

Because common queries are:

SELECT COUNT(*)
FROM events
WHERE campaign_id = ?
AND event_type = ?
AND event_time BETWEEN ?

Without proper indexing:
- Full table scans
- Slow reporting
- DB CPU spikes

---

## Design Philosophy

Events table is:

- Append-only
- Never updated
- Rarely deleted
- Extremely write-heavy

We optimize for:
- Fast inserts
- Fast filtering
- Horizontal scalability

We avoid:
- Foreign key constraints (optional at scale)
- Complex joins
- Heavy normalization

---

## What Would Break at Scale Without This Design?

If CampaignID not duplicated:
- Expensive joins
- Aggregation slowdown

If no JSONB:
- Constant schema migrations
- DevOps overhead

If no indexes:
- Slow campaign reporting
- DB meltdown during traffic spikes

If EventTime not separated:
- Analytics inaccuracies
- Impossible to measure ingestion lag

---

## Production Considerations

- Table may later require time-based partitioning
- Index bloat must be monitored
- WAL size grows with heavy writes
- Connection pooling must be tuned

This table will grow fastest in the system.
Design it carefully.
