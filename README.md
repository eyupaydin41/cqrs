# 🚀 CQRS + Event Sourcing + gRPC Microservices

A **production-ready** CQRS (Command Query Responsibility Segregation) implementation with **Event Sourcing**, **gRPC**, and **Time Travel** capabilities using Go, Kafka, ClickHouse, and PostgreSQL.

## ⭐ Highlights

- ✅ **CQRS Pattern** - Separate read/write models
- ✅ **Event Sourcing** - Events as source of truth
- ✅ **gRPC Communication** - Type-safe inter-service communication
- ✅ **Time Travel** - Query historical states at any point in time
- ✅ **Event Store** - ClickHouse for immutable event storage
- ✅ **Kafka Streaming** - Asynchronous event propagation
- ✅ **Snapshots** - Performance optimization for large event streams
- ✅ **Replay Capability** - Rebuild state from events
- ✅ **Postman Collection** - 26 ready-to-use API endpoints

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CQRS + Event Sourcing                            │
└─────────────────────────────────────────────────────────────────────────┘

                    COMMAND SIDE                                QUERY SIDE
                         │                                          │
    ┌────────────────────┼────────────────────┐                   │
    │                    │                    │                   │
    │          ┌─────────▼─────────┐          │                   │
    │          │  Auth Service     │          │                   │
    │          │  (Port 8088)      │          │                   │
    │          │                   │          │                   │
    │          │  • Register       │          │         ┌─────────▼─────────┐
    │          │  • Change Pass    │──gRPC───▶│         │  Query Service    │
    │          │  • Change Email   │          │         │  (Port 8089)      │
    │          └─────────┬─────────┘          │         │                   │
    │                    │                    │         │  • Get Users      │
    │                    │ Kafka Events       │         │  • Login          │
    │                    ▼                    │         └─────────┬─────────┘
    │          ┌─────────────────────┐        │                   │
    │          │   Kafka (9092)      │────────┼───────────────────┤
    │          │                     │        │                   │
    │          │  Topic:             │        │                   │
    │          │  - user-events      │        │                   │
    │          └──────────┬──────────┘        │                   │
    │                     │                   │                   │
    │                     ▼                   │                   │
    │          ┌─────────────────────┐        │                   │
    │          │   Event Store       │        │                   │
    │          │   (Port 8090)       │────────┼───────────────────┘
    │          │                     │        │
    │          │  • HTTP: 8090       │        │
    │          │  • gRPC: 9090       │◀───────┘ (gRPC GetAggregateEvents)
    │          │                     │
    │          │  Features:          │
    │          │  • Time Travel      │
    │          │  • Snapshots        │
    │          │  • Event Replay     │
    │          └──────────┬──────────┘
    │                     │
    │                     ▼
    │          ┌─────────────────────┐
    │          │  ClickHouse (9000)  │
    │          │                     │
    │          │  Immutable Events   │
    │          │  Columnar Storage   │
    │          └─────────────────────┘
    │
    └────────────────────────────────────────────┘

         DATA STORES

    ┌────────────────┐              ┌────────────────┐
    │ PostgreSQL     │              │ PostgreSQL     │
    │ Auth DB (5432) │              │ Query DB (5433)│
    │                │              │                │
    │ (Not used -    │              │ • users        │
    │  Event Sourced)│              │ • auth_proj.   │
    └────────────────┘              │ • login_hist.  │
                                    └────────────────┘
```

### 🔄 Data Flow

**1. Command Flow (Write):**
```
User → Auth Service → Publish Event → Kafka → Event Store → ClickHouse
                                           │
                                           └──→ Query Service → PostgreSQL
```

**2. Query Flow (Read with gRPC):**
```
User → Auth Service → gRPC Call → Event Store → ClickHouse
                         ↓
                   Load Events → Reconstruct Aggregate → Process Command
```

**3. Query Flow (Simple Read):**
```
User → Query Service → PostgreSQL (Read Model)
```

---

## 🧩 Components

| Service | Port | Technology | Purpose |
|---------|------|------------|---------|
| **Auth Service** | 8088 | Go + Gin | Command side - Write operations |
| **Query Service** | 8089 | Go + Gin | Query side - Read operations |
| **Event Store** | 8090 (HTTP)<br>9090 (gRPC) | Go + ClickHouse | Event storage, Time Travel, gRPC server |
| **Kafka** | 9092 | Confluent | Event streaming |
| **ClickHouse** | 9000, 8123 | ClickHouse | Immutable event storage |
| **PostgreSQL Auth** | 5432 | PostgreSQL | Command DB (not used - Event Sourced) |
| **PostgreSQL Query** | 5433 | PostgreSQL | Read model |
| **Zookeeper** | 2181 | Zookeeper | Kafka coordination |

---

## 🚀 Features

### 🎯 CQRS (Command Query Responsibility Segregation)

- **Command Side (Auth Service):** Write operations only
  - Register User
  - Change Password (with Event Sourcing + gRPC)
  - Change Email (with Event Sourcing + gRPC)

- **Query Side (Query Service):** Read operations only
  - Get All Users
  - Login with JWT
  - Optimized read models

### 📦 Event Sourcing

- **Events as Source of Truth:** No traditional CRUD, only events
- **Event Store:** All events stored in ClickHouse
- **Aggregate Reconstruction:** Rebuild state from events
- **Immutability:** Events are never modified or deleted

**Supported Events:**
- `user.created`
- `user.password.changed`
- `user.email.changed`
- `user.deactivated`
- `user.login.recorded`

### ⚡ gRPC Communication

**Why gRPC?**
- Type-safe communication
- Binary serialization (faster than JSON)
- Compile-time validation
- Better performance

**gRPC Endpoint:**
```protobuf
service EventStoreService {
  rpc GetAggregateEvents(GetAggregateEventsRequest) returns (GetAggregateEventsResponse);
}
```

**Usage:**
- Auth Service calls Event Store via gRPC
- Loads aggregate event history
- Reconstructs user state
- Processes commands (Change Password, Change Email)

### ⏰ Time Travel

Query historical states at any point in time!

```bash
# Current state
GET /replay/user/{id}/state

# State at specific time
GET /replay/user/{id}/state-at?timestamp=2025-01-15T10:00:00Z

# Compare two states
GET /replay/user/{id}/compare?time1=2025-01-01T00:00:00Z&time2=2025-01-15T00:00:00Z

# Full history
GET /replay/user/{id}/history
```

**Use Cases:**
- Audit: "What was the user's email on January 1st?"
- Debugging: "What state caused the bug?"
- Compliance: "Show user data as of this date"

### 📸 Snapshots

Performance optimization for aggregates with many events.

**Without Snapshot:** Replay 1000 events (~100ms)
**With Snapshot:** Load 1 snapshot + 10 recent events (~5ms)

```bash
# Create snapshot
POST /snapshots/{id}

# Get state using snapshot
GET /snapshots/{id}/state
```

### 🔄 Event Replay

Rebuild read models from events.

```bash
# Get events since timestamp
GET /events/replay?since=2025-01-01T00:00:00Z
```

---

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Postman (for API testing)
- Protocol Buffers compiler (for gRPC code generation)

---

## 🛠️ Quick Start

### 1. Clone & Setup

```bash
git clone <repository-url>
cd cqrs
```

### 2. Create .env File

Create a `.env` file in the root directory:

```env
# PostgreSQL Auth
POSTGRES_AUTH_USER=postgres
POSTGRES_AUTH_PASSWORD=postgres
POSTGRES_AUTH_DB=auth_db

# PostgreSQL Query
POSTGRES_QUERY_USER=postgres
POSTGRES_QUERY_PASSWORD=postgres
POSTGRES_QUERY_DB=query_db

# ClickHouse
CLICKHOUSE_HOST=clickhouse:9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=mypass
CLICKHOUSE_DB=events

# JWT
JWT_SECRET=supersecretkey

# Kafka
KAFKA_BROKER=kafka:29092
KAFKA_TOPIC=user-events
KAFKA_GROUP=query-group
```

### 3. Start Services

```bash
docker-compose up --build -d
```

This starts all 8 services:
- ✅ Zookeeper
- ✅ Kafka
- ✅ PostgreSQL (Auth)
- ✅ PostgreSQL (Query)
- ✅ ClickHouse
- ✅ Auth Service
- ✅ Query Service
- ✅ Event Store

### 4. Verify Services

```bash
# Check all containers
docker-compose ps

# Check logs
docker-compose logs -f

# Health checks
curl http://localhost:8088/health  # Auth Service
curl http://localhost:8089/health  # Query Service
curl http://localhost:8090/health  # Event Store
```

### 5. Import Postman Collection

**📮 Postman Collection Included!**

Import `CQRS-EventSourcing.postman_collection.json` into Postman.

**26 Ready-to-Use Endpoints:**
- 🔐 Auth Service (4 endpoints)
- 🔍 Query Service (3 endpoints)
- 📦 Event Store (5 endpoints)
- ⏰ Time Travel (4 endpoints)
- 📸 Snapshots (3 endpoints)
- 🎯 Complete User Journey (7 steps)

**See `POSTMAN_GUIDE.md` for detailed usage!**

---

## 📡 API Endpoints

### 🔐 Auth Service (Port 8088) - COMMAND

| Method | Endpoint | Description | gRPC Used |
|--------|----------|-------------|-----------|
| GET | `/health` | Health check | ❌ |
| POST | `/register` | Register new user | ❌ |
| PUT | `/users/:id/password` | Change password | ✅ Yes! |
| PUT | `/users/:id/email` | Change email | ✅ Yes! |

**Example: Register User**
```bash
curl -X POST http://localhost:8088/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

**Example: Change Password (gRPC Demo!)**
```bash
curl -X PUT http://localhost:8088/users/{USER_ID}/password \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

**What happens:**
1. Auth Service calls Event Store via **gRPC**
2. Loads user's event history
3. Reconstructs aggregate from events
4. Validates old password
5. Changes password
6. Publishes `user.password.changed` event to Kafka

### 🔍 Query Service (Port 8089) - QUERY

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/users` | Get all users (read model) |
| POST | `/login` | Login with JWT |

**Example: Get All Users**
```bash
curl http://localhost:8089/users
```

**Example: Login**
```bash
curl -X POST http://localhost:8089/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### 📦 Event Store (Port 8090)

#### Basic Event Queries

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check + event count |
| GET | `/events` | Get all events (with filters) |
| GET | `/events/aggregate/:id` | Get events for aggregate |
| GET | `/events/count` | Total event count |
| GET | `/events/replay?since=<timestamp>` | Get events since timestamp |

**Example: Get User Events**
```bash
curl http://localhost:8090/events/aggregate/{USER_ID}
```

**Response:**
```json
{
  "aggregate_id": "abc-123",
  "events": [
    {
      "event_type": "user.created",
      "version": 1,
      "timestamp": "2025-10-25T19:57:00Z"
    },
    {
      "event_type": "user.password.changed",
      "version": 2,
      "timestamp": "2025-10-25T19:58:00Z"
    }
  ],
  "count": 2
}
```

#### Time Travel Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/replay/user/:id/state` | Current state from events |
| GET | `/replay/user/:id/state-at?timestamp=<time>` | State at specific time |
| GET | `/replay/user/:id/history` | Full change history |
| GET | `/replay/user/:id/compare?time1=<t1>&time2=<t2>` | Compare states |

**Example: Time Travel**
```bash
# See user state on January 1st, 2025
curl "http://localhost:8090/replay/user/{USER_ID}/state-at?timestamp=2025-01-01T00:00:00Z"
```

#### Snapshot Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/snapshots/:id` | Create snapshot |
| GET | `/snapshots/:id` | Get latest snapshot |
| GET | `/snapshots/:id/state` | Get state (snapshot + events) |

---

## 🧪 Testing Scenarios

### Scenario 1: Basic CQRS Flow (5 min)

```bash
# 1. Register user
curl -X POST http://localhost:8088/register \
  -d '{"email":"test@example.com","password":"pass123"}'

# Response: {"id":"abc-123","message":"User registered successfully"}

# 2. Wait 2-3 seconds for Kafka processing

# 3. Query users (read model)
curl http://localhost:8089/users

# Should see the new user!
```

### Scenario 2: Event Sourcing + gRPC (10 min)

```bash
# 1. Register
curl -X POST http://localhost:8088/register \
  -d '{"email":"test@example.com","password":"pass123"}'
# Save the user_id from response

# 2. Change Password (triggers gRPC!)
curl -X PUT http://localhost:8088/users/{user_id}/password \
  -d '{"old_password":"pass123","new_password":"newpass456"}'

# 3. Check logs to see gRPC calls
docker logs cqrs-auth-service-1 | grep "gRPC"
# Should see: "gRPC Call: GetAggregateEvents"
# Should see: "gRPC Response: Received 1 events"

# 4. View events in Event Store
curl http://localhost:8090/events/aggregate/{user_id}
# Should see: user.created, user.password.changed
```

### Scenario 3: Time Travel (10 min)

```bash
# 1. Register user at 19:57:00
USER_ID=$(curl -X POST http://localhost:8088/register \
  -d '{"email":"test@example.com","password":"pass123"}' | jq -r '.id')

# 2. Change password at 19:58:00
curl -X PUT http://localhost:8088/users/$USER_ID/password \
  -d '{"old_password":"pass123","new_password":"newpass"}'

# 3. Change email at 19:59:00
curl -X PUT http://localhost:8088/users/$USER_ID/email \
  -d '{"new_email":"newemail@example.com"}'

# 4. Time travel to 19:57:30 (only registered)
curl "http://localhost:8090/replay/user/$USER_ID/state-at?timestamp=2025-10-25T19:57:30Z"
# Shows: original email, original password

# 5. Time travel to 19:58:30 (password changed)
curl "http://localhost:8090/replay/user/$USER_ID/state-at?timestamp=2025-10-25T19:58:30Z"
# Shows: original email, NEW password

# 6. Compare states
curl "http://localhost:8090/replay/user/$USER_ID/compare?time1=2025-10-25T19:57:00Z&time2=2025-10-25T20:00:00Z"
# Shows: diff between two states
```

### Scenario 4: Snapshots (10 min)

```bash
# 1. Register and make 5 password changes
USER_ID=$(curl -X POST http://localhost:8088/register \
  -d '{"email":"test@example.com","password":"pass1"}' | jq -r '.id')

for i in {1..5}; do
  curl -X PUT http://localhost:8088/users/$USER_ID/password \
    -d "{\"old_password\":\"pass$i\",\"new_password\":\"pass$((i+1))\"}"
done

# 2. Get state (replays 6 events - slow)
time curl http://localhost:8090/replay/user/$USER_ID/state

# 3. Create snapshot
curl -X POST http://localhost:8090/snapshots/$USER_ID

# 4. Make 2 more changes
curl -X PUT http://localhost:8088/users/$USER_ID/password \
  -d '{"old_password":"pass6","new_password":"pass7"}'

# 5. Get state with snapshot (snapshot + 1 event - fast!)
time curl http://localhost:8090/snapshots/$USER_ID/state
```

---

## 🗂️ Project Structure

```
cqrs/
├── proto/
│   └── event_store.proto              # gRPC service definition
│
├── auth-service/                       # COMMAND Service
│   ├── api/
│   │   └── auth_controller.go         # HTTP handlers
│   ├── command/
│   │   ├── commands.go                # Command definitions
│   │   └── handler.go                 # Command handlers (uses gRPC!)
│   ├── domain/
│   │   ├── user_aggregate.go          # User aggregate
│   │   └── events.go                  # Domain events
│   ├── event/
│   │   └── producer.go                # Kafka producer
│   ├── grpc/
│   │   └── event_store_client.go      # gRPC client wrapper
│   ├── proto/
│   │   ├── event_store.pb.go          # Generated protobuf code
│   │   └── event_store_grpc.pb.go     # Generated gRPC code
│   ├── Dockerfile.dev
│   ├── go.mod
│   └── main.go
│
├── query-service/                      # QUERY Service
│   ├── api/
│   │   ├── user_handler.go            # User queries
│   │   └── auth_handler.go            # Login
│   ├── event/
│   │   └── consumer.go                # Kafka consumer
│   ├── model/
│   │   ├── user.go
│   │   ├── login_history.go
│   │   └── auth_projection.go         # Read model
│   ├── repository/
│   │   ├── user_repo.go
│   │   └── auth_projection_repo.go
│   ├── service/
│   │   ├── user_service.go
│   │   └── auth_service.go
│   ├── Dockerfile.dev
│   ├── go.mod
│   └── main.go
│
├── event-store/                        # EVENT STORE Service
│   ├── api/
│   │   ├── handler.go                 # Event queries
│   │   ├── replay_handler.go          # Time travel
│   │   └── snapshot_handler.go        # Snapshots
│   ├── consumer/
│   │   └── kafka_consumer.go          # Kafka → ClickHouse
│   ├── grpc/
│   │   └── server.go                  # gRPC server implementation
│   ├── model/
│   │   ├── user_event.go
│   │   ├── user_aggregate.go
│   │   └── snapshot.go
│   ├── proto/
│   │   ├── event_store.pb.go          # Generated code
│   │   └── event_store_grpc.pb.go
│   ├── repository/
│   │   ├── event_repository.go        # ClickHouse operations
│   │   └── snapshot_repository.go
│   ├── service/
│   │   ├── event_service.go
│   │   ├── replay_service.go          # Time travel logic
│   │   └── snapshot_service.go
│   ├── Dockerfile.dev
│   ├── go.mod
│   └── main.go
│
├── integration-tests/
│   └── time_travel_test.go
│
├── docker-compose.yml
├── .env
├── CQRS-EventSourcing.postman_collection.json  # 26 endpoints!
├── POSTMAN_GUIDE.md                            # Detailed guide
├── ARCHITECTURE.md                             # Architecture details
└── README.md
```

---

## 🔧 Configuration

### Environment Variables

All services are configured via `.env` file:

```env
# PostgreSQL
POSTGRES_AUTH_USER=postgres
POSTGRES_AUTH_PASSWORD=postgres
POSTGRES_AUTH_DB=auth_db
POSTGRES_QUERY_USER=postgres
POSTGRES_QUERY_PASSWORD=postgres
POSTGRES_QUERY_DB=query_db

# ClickHouse (Event Store)
CLICKHOUSE_HOST=clickhouse:9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=mypass
CLICKHOUSE_DB=events

# JWT
JWT_SECRET=supersecretkey

# Kafka
KAFKA_BROKER=kafka:29092
KAFKA_TOPIC=user-events
KAFKA_GROUP=query-group
```

### Docker Compose Ports

| Service | Internal Port | External Port |
|---------|--------------|---------------|
| Auth Service | 8088 | 8088 |
| Query Service | 8089 | 8089 |
| Event Store HTTP | 8090 | 8090 |
| Event Store gRPC | 9090 | 9090 |
| Kafka | 29092 | 9092 |
| PostgreSQL Auth | 5432 | 5432 |
| PostgreSQL Query | 5432 | 5433 |
| ClickHouse HTTP | 8123 | 8123 |
| ClickHouse Native | 9000 | 9000 |
| Zookeeper | 2181 | 2181 |

---

## 🐳 Docker Commands

```bash
# Start all services
docker-compose up -d

# Start with rebuild
docker-compose up --build -d

# View logs
docker-compose logs -f

# View specific service logs
docker logs cqrs-auth-service-1 -f
docker logs cqrs-event-store-1 -f

# View gRPC logs
docker logs cqrs-auth-service-1 | grep "gRPC"

# Stop all services
docker-compose down

# Stop and remove volumes (deletes data!)
docker-compose down -v

# Restart specific service
docker-compose restart auth-service
```

---

## 🗄️ Database Access

### ClickHouse (Event Store)

```bash
# Connect to ClickHouse
docker exec -it cqrs-clickhouse-1 clickhouse-client

# View events
SELECT event_type, aggregate_id, version, timestamp
FROM events.events
ORDER BY timestamp DESC
LIMIT 10;

# Count events by type
SELECT event_type, COUNT(*) as count
FROM events.events
GROUP BY event_type;

# Get user events
SELECT * FROM events.events
WHERE aggregate_id = 'USER_ID_HERE'
ORDER BY version;
```

### PostgreSQL Query DB

```bash
# Connect
docker exec -it cqrs-postgres-query-1 psql -U postgres -d query_db

# View users
SELECT * FROM users;

# View auth projections
SELECT * FROM auth_projections;

# View login history
SELECT * FROM login_histories ORDER BY login_at DESC LIMIT 10;
```

---

## 🔍 Monitoring & Debugging

### Check Kafka Topics

```bash
# List topics
docker exec cqrs-kafka-1 kafka-topics \
  --list --bootstrap-server localhost:9092

# View messages
docker exec cqrs-kafka-1 kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic user-events \
  --from-beginning
```

### Check gRPC Communication

```bash
# Auth Service logs (gRPC client)
docker logs cqrs-auth-service-1 --tail 50 | grep "gRPC"
# Look for: "gRPC Call: GetAggregateEvents"
# Look for: "gRPC Response: Received X events"

# Event Store logs (gRPC server)
docker logs cqrs-event-store-1 --tail 50 | grep "gRPC"
# Look for: "gRPC: GetAggregateEvents called"
# Look for: "retrieved X events for aggregate"
```

### Health Checks

```bash
# All services
curl http://localhost:8088/health  # Auth
curl http://localhost:8089/health  # Query
curl http://localhost:8090/health  # Event Store

# Event count
curl http://localhost:8090/events/count
```

---

## 🚦 Troubleshooting

### Problem: Services not starting

```bash
# Check container status
docker-compose ps

# Check logs
docker-compose logs [service-name]

# Restart
docker-compose restart
```

### Problem: Kafka not processing events

```bash
# Wait 30-60 seconds after startup
# Check Kafka logs
docker logs cqrs-kafka-1

# Verify topic exists
docker exec cqrs-kafka-1 kafka-topics \
  --list --bootstrap-server localhost:9092
```

### Problem: gRPC connection failed

```bash
# Check Event Store gRPC server is running
docker logs cqrs-event-store-1 | grep "gRPC server starting"
# Should see: "🚀 gRPC server starting on port 9090"

# Check Auth Service can connect
docker logs cqrs-auth-service-1 | grep "Connecting to Event-Store"
```

### Problem: ClickHouse connection failed

```bash
# Check ClickHouse is healthy
docker ps | grep clickhouse

# Test connection
docker exec cqrs-clickhouse-1 clickhouse-client --query "SELECT 1"
```

---

## 🎓 Development

### Local Development (without Docker)

```bash
# Start infrastructure only
docker-compose up -d zookeeper kafka postgres-auth postgres-query clickhouse

# Run auth-service locally
cd auth-service
go run main.go

# Run query-service locally
cd query-service
go run main.go

# Run event-store locally
cd event-store
go run main.go
```

### gRPC Code Generation

If you modify `proto/event_store.proto`:

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
cd proto
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       event_store.proto

# Copy to services
cp *.pb.go ../auth-service/proto/
cp *.pb.go ../event-store/proto/
```

### Running Integration Tests

```bash
cd integration-tests
go test -v ./...
```

---

## 📚 Technologies

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.23+ | Programming language |
| **Gin** | Latest | HTTP web framework |
| **gRPC** | 1.76.0 | RPC framework |
| **Protocol Buffers** | 3 | Serialization |
| **PostgreSQL** | 15 | Read model storage |
| **ClickHouse** | Latest | Event storage (columnar) |
| **Apache Kafka** | 7.5.0 | Event streaming |
| **Zookeeper** | 7.5.0 | Kafka coordination |
| **GORM** | Latest | ORM (Query Service) |
| **bcrypt** | Latest | Password hashing |
| **JWT** | Latest | Authentication |
| **Docker** | Latest | Containerization |

---

## 📖 Key Concepts

### CQRS (Command Query Responsibility Segregation)

Separate models for reading and writing data:
- **Commands:** Write operations (Auth Service)
- **Queries:** Read operations (Query Service)

### Event Sourcing

Store state as sequence of events:
- ✅ Events are immutable
- ✅ Complete audit trail
- ✅ Time travel capability
- ✅ Replay events to rebuild state

### gRPC

Remote Procedure Call framework:
- ✅ Type-safe communication
- ✅ Binary protocol (fast)
- ✅ Generated client/server code
- ✅ HTTP/2 based

### Domain-Driven Design

- **Aggregates:** User aggregate
- **Events:** Domain events (UserCreatedEvent, etc.)
- **Commands:** User intentions (RegisterUser, ChangePassword)

---

## 🎯 Use Cases

### 1. Audit & Compliance
- "Show me all changes to this user account"
- "What was the user's email on January 1st?"
- Complete audit trail with timestamps

### 2. Debugging
- Replay events to reproduce bugs
- Time travel to see state when bug occurred
- Compare states before/after

### 3. Analytics
- Query events from ClickHouse
- Analyze user behavior patterns
- Generate reports from event history

### 4. Data Recovery
- Rebuild read models from events
- Fix corrupted data by replaying events
- Migrate to new read model structure

---

## 🔗 Resources

### Documentation
- [POSTMAN_GUIDE.md](./POSTMAN_GUIDE.md) - API testing guide
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed architecture
- Proto files in `proto/` directory

### External Resources
- [CQRS Pattern - Martin Fowler](https://martinfowler.com/bliki/CQRS.html)
- [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html)
- [gRPC Documentation](https://grpc.io/docs/)
- [Apache Kafka](https://kafka.apache.org/documentation/)
- [ClickHouse](https://clickhouse.com/docs)

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

---

## 📝 License

MIT License - see LICENSE file for details

---

## 👨‍💻 Author

**Eyup Aydin**

Created with ❤️ using Go, gRPC, Kafka, ClickHouse, and PostgreSQL

---

## 🎉 Quick Start Summary

```bash
# 1. Start services
docker-compose up -d

# 2. Import Postman collection
# File: CQRS-EventSourcing.postman_collection.json

# 3. Test basic flow
curl -X POST http://localhost:8088/register \
  -d '{"email":"test@example.com","password":"pass123"}'

# 4. Test gRPC (Change Password)
curl -X PUT http://localhost:8088/users/{USER_ID}/password \
  -d '{"old_password":"pass123","new_password":"newpass"}'

# 5. Test Time Travel
curl "http://localhost:8090/replay/user/{USER_ID}/state-at?timestamp=2025-10-25T19:58:00Z"

# 6. View logs
docker logs cqrs-auth-service-1 | grep "gRPC"
```

**That's it! You're ready to explore CQRS + Event Sourcing + gRPC! 🚀**
