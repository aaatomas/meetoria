# Meetoria Architecture

## Domain Hierarchy

```
Platform
  └── Organization (one Meetoria customer / tenant)
        └── Branch (physical location)
              └── Employee
                    └── Service (org catalog, enabled per branch)
                          └── Booking
```

- **Customers** belong to the **organization** (shared across branches).
- **Services** belong to the **organization**; **branch_services** defines which services each branch offers.
- **Employees** belong to a **branch** (`employees.branch_id`).
- **Bookings** are scoped to organization + branch + customer + employee + service.

## Layer Responsibilities

| Layer | Responsibility |
|-------|---------------|
| **Handler** | HTTP request/response, input validation, auth context extraction |
| **Service** | Business rules, orchestration, tenant authorization |
| **Repository** | Data access, organization-scoped queries |
| **Database** | PostgreSQL with UUID keys, soft deletes, exclusion constraints |

## Multi-Tenancy

Every query includes `organization_id` filtering. The service layer calls `VerifyMembership()` before any tenant operation. The `X-Organization-ID` header identifies the active organization. The `X-Branch-ID` header (or query param) scopes branch-level operations within that organization.

Users may belong to multiple organizations via `organization_users`. Use **branches** for multiple locations within one business — not multiple organizations.

## Event Flow

```
Booking Created (API)
    → booking.created event (RabbitMQ)
    → notification record created (CREATED → QUEUED)
    → notification.sms / notification.email events
    → SMS Worker / Email Worker (separate DBs)
    → Provider (Twilio, SMTP, etc.)
```

All events carry `correlation_id` for end-to-end tracing. Meetoria does not send SMS/email directly.

## Authentication

Keycloak handles all identity operations. Meetoria stores `keycloak_id` (from JWT `sub` claim) and manages organization-level roles in `organization_users`.

## Analytics

Statistics are pre-aggregated in dedicated analytics tables (organization, branch, employee, customer), updated asynchronously via booking events.

## Migrations

Schema is defined in a single file: `backend/migrations/001_schema.sql`. The backend migrator applies it on startup. Postgres init only runs `000_create_databases.sql` for worker/Keycloak databases.
