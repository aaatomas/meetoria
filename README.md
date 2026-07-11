# Meetoria

**Schedule Smarter. Grow Faster.**

Meetoria is a scalable, multi-tenant appointment scheduling SaaS platform designed for service businesses worldwide — starting with hair salons and beauty studios, expanding to medical clinics, fitness, consultants, and any appointment-based business.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Frontend  │────▶│ Meetoria API │────▶│ PostgreSQL  │
│  React/MUI  │     │   Go/Gin     │     └─────────────┘
└─────────────┘     └──────┬───────┘     ┌─────────────┐
                           │             │    Redis    │
                    ┌──────▼───────┐     └─────────────┘
                    │  Keycloak    │
                    │  (Auth/IdP)  │     ┌─────────────┐
                    └──────────────┘     │  RabbitMQ   │
                           │             └──────┬──────┘
                    ┌──────▼───────┐            │
                    │   RabbitMQ   │     ┌──────┴──────┐
                    └──────┬───────┘     │             │
                           │        ┌────▼────┐  ┌─────▼─────┐
                    ┌──────┴───────┐│SMS Worker│  │Email Worker│
                    │  Events      │└─────────┘  └───────────┘
                    └──────────────┘
```

### Key Design Decisions

- **Clean Architecture**: Handler → Service → Repository → Database. No business logic in HTTP handlers.
- **Multi-Tenant**: Every entity is scoped by `organization_id`. Cross-tenant access is blocked at the service layer.
- **Keycloak Auth**: Authentication delegated entirely to Keycloak (OIDC/JWT/PKCE). Meetoria handles organization-level RBAC.
- **Event-Driven Notifications**: Core publishes events to RabbitMQ. SMS/Email workers are separate deployable services with their own databases.
- **Double Booking Prevention**: PostgreSQL exclusion constraints + Redis distributed locks + transactional checks.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24+, Gin, GORM, PostgreSQL |
| Frontend | React, TypeScript, MUI, TanStack Query |
| Auth | Keycloak (OIDC, JWT, PKCE) |
| Cache | Redis |
| Messaging | RabbitMQ |
| Workers | Go (SMS, Email — separate projects) |
| Infra | Docker Compose, Kubernetes, Helm |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for local backend development)
- Node.js 22+ (for local frontend development)

### Run with Docker Compose

```bash
# Start all services
docker compose up -d

# Services:
# - Frontend:    http://localhost:3000
# - API:         http://localhost:8081
# - Keycloak:    http://localhost:8080  (admin/admin)
# - RabbitMQ UI: http://localhost:15672 (meetoria/meetoria)
# - PostgreSQL:  localhost:5432
# - Redis:       localhost:6379
```

### Local Development

**Backend:**
```bash
cd backend
go run ./cmd/api
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
# Opens at http://localhost:5173
```

**Workers:**
```bash
cd workers/sms-worker && go run ./cmd/worker
cd workers/email-worker && go run ./cmd/worker
```

## Project Structure

```
meetoria/
├── backend/                  # Meetoria Core API
│   ├── cmd/api/              # Application entrypoint
│   ├── internal/
│   │   ├── auth/             # Keycloak JWT validation, middleware
│   │   ├── organization/     # Multi-tenant organizations
│   │   ├── user/             # User management (Keycloak-linked)
│   │   ├── customer/         # Customer CRM
│   │   ├── employee/         # Staff management
│   │   ├── service/          # Business services (haircut, etc.)
│   │   ├── booking/          # Appointment booking engine
│   │   ├── schedule/         # Working hours, breaks, holidays
│   │   ├── notification/     # Notification status tracking
│   │   ├── analytics/        # Pre-aggregated statistics
│   │   └── common/           # Shared config, errors, redis, rabbitmq
│   ├── migrations/           # PostgreSQL schema
│   └── pkg/events/           # Event definitions
├── frontend/                 # React web application
├── workers/
│   ├── sms-worker/           # SMS notification worker (separate DB)
│   └── email-worker/         # Email notification worker (separate DB)
├── infrastructure/
│   ├── keycloak/             # Realm configuration
│   └── helm/                 # Kubernetes Helm charts
└── docker-compose.yml
```

## API Overview

All endpoints require `Authorization: Bearer <jwt>` header.

Organization-scoped endpoints also require `X-Organization-ID` header.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/me` | Current user profile |
| POST | `/api/v1/organizations` | Create organization |
| GET | `/api/v1/organizations/:id/customers` | List customers |
| POST | `/api/v1/organizations/:id/bookings` | Create booking |
| GET | `/api/v1/organizations/:id/bookings/availability` | Get available slots |
| GET | `/api/v1/organizations/:id/analytics/dashboard` | Dashboard stats |

Swagger docs available at `http://localhost:8081/swagger/index.html`

## Roles

**Keycloak Global Roles:** `platform_admin`, `customer`

**Organization Roles:** `organization_owner`, `manager`, `employee`, `customer`

Users can hold different roles in different organizations.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | postgres://... | PostgreSQL connection |
| `REDIS_URL` | redis://localhost:6379/0 | Redis connection |
| `RABBITMQ_URL` | amqp://... | RabbitMQ connection |
| `KEYCLOAK_URL` | http://localhost:8080 | Keycloak server |
| `KEYCLOAK_REALM` | meetoria | Keycloak realm |
| `JWT_ISSUER` | http://localhost:8080/realms/meetoria | JWT issuer |

## License

Proprietary — All rights reserved.
