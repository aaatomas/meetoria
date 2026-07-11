# Meetoria Architecture

## Overview

Meetoria follows Clean Architecture with Domain-Driven Design principles. Business logic lives exclusively in the service layer.

## Layer Responsibilities

| Layer | Responsibility |
|-------|---------------|
| **Handler** | HTTP request/response, input validation, auth context extraction |
| **Service** | Business rules, orchestration, tenant authorization |
| **Repository** | Data access, organization-scoped queries |
| **Database** | PostgreSQL with UUID keys, soft deletes, exclusion constraints |

## Multi-Tenancy

Every query includes `organization_id` filtering. The service layer calls `VerifyMembership()` before any tenant operation. The `X-Organization-ID` header identifies the active tenant context.

## Event Flow

```
Booking Created (API)
    → booking.created event (RabbitMQ)
    → notification record created (CREATED → QUEUED)
    → notification.sms / notification.email events
    → SMS Worker / Email Worker (separate DBs)
    → Provider (Twilio, SMTP, etc.)
    → Delivery status (future: delivery confirmation events)
```

All events carry `correlation_id` for end-to-end tracing.

## Authentication

Keycloak handles all identity operations. Meetoria stores `keycloak_id` (from JWT `sub` claim) and manages organization-level roles in `organization_users`.

## Analytics

Statistics are pre-aggregated in dedicated analytics tables, updated asynchronously via booking events. Dashboard queries read from aggregated tables, not transactional booking data.
