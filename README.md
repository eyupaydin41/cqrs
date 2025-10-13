# CQRS Event-Driven Microservices with Go

A production-ready CQRS (Command Query Responsibility Segregation) implementation using Go, Kafka, and PostgreSQL. This project demonstrates event-driven architecture with separate read and write models.

## 🏗️ Architecture

```
┌─────────────────┐      ┌──────────────┐      ┌─────────────────┐
│                 │      │              │      │                 │
│  Auth Service   │─────▶│    Kafka     │─────▶│ Query Service   │
│  (Write/CMD)    │      │   Events     │      │  (Read/Query)   │
│                 │      │              │      │                 │
└────────┬────────┘      └──────────────┘      └────────┬────────┘
         │                                               │
         │                                               │
         ▼                                               ▼
┌─────────────────┐                            ┌─────────────────┐
│   PostgreSQL    │                            │   PostgreSQL    │
│    (Auth DB)    │                            │   (Query DB)    │
│                 │                            │                 │
│  - users        │                            │  - users        │
│                 │                            │  - login_hist.  │
└─────────────────┘                            └─────────────────┘
```

### Components

- **Auth Service** (`:8088`): Handles write operations (Register, Login) and publishes events to Kafka
- **Query Service** (`:8089`): Handles read operations and consumes events from Kafka
- **Kafka** (`:9092`): Event streaming platform for asynchronous communication
- **PostgreSQL**: Separate databases for command and query sides
- **Zookeeper** (`:2181`): Kafka coordination service

## 🚀 Features

- ✅ CQRS Pattern Implementation
- ✅ Event-Driven Architecture
- ✅ Microservices with Go
- ✅ Apache Kafka for Event Streaming
- ✅ Separate Read/Write Databases
- ✅ JWT Authentication
- ✅ Docker Compose Setup
- ✅ Login History Tracking
- ✅ Health Check Endpoints

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Postman or curl (for testing)

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd cqrs
```

### 2. Start All Services

```bash
docker-compose up --build
```

This command will start:
- Zookeeper
- Kafka
- PostgreSQL (Auth DB)
- PostgreSQL (Query DB)
- Auth Service
- Query Service

### 3. Verify Services are Running

```bash
# Check all containers
docker-compose ps

# Check logs
docker-compose logs -f

# Check specific service logs
docker-compose logs -f auth-service
docker-compose logs -f query-service
```

## 📡 API Endpoints

### Auth Service (Port 8088)

#### Health Check
```http
GET http://localhost:8088/health
```

#### Register User
```http
POST http://localhost:8088/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**
```json
{
  "id": "uuid-here",
  "message": "User registered successfully"
}
```

#### Login User
```http
POST http://localhost:8088/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "jwt-token-here"
}
```

### Query Service (Port 8089)

#### Get All Users
```http
GET http://localhost:8089/users
```

**Response:**
```json
[
  {
    "ID": "uuid",
    "Email": "user@example.com",
    "CreatedAt": "2025-10-14T01:30:00Z"
  }
]
```

## 🧪 Testing

### Using curl

```bash
# Health Check
curl http://localhost:8088/health

# Register
curl -X POST http://localhost:8088/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"12345"}'

# Login
curl -X POST http://localhost:8088/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"12345"}'

# Get Users
curl http://localhost:8089/users
```

### Using Postman

Import the following collection or create requests manually:

1. **Health Check**: `GET http://localhost:8088/health`
2. **Register**: `POST http://localhost:8088/register`
3. **Login**: `POST http://localhost:8088/login`
4. **Get Users**: `GET http://localhost:8089/users`

## 🔍 Event Flow

### Registration Flow

1. Client sends POST request to `/register`
2. Auth Service creates user in Auth DB
3. Auth Service publishes `UserRegistered` event to Kafka
4. Query Service consumes event
5. Query Service stores user in Query DB

### Login Flow

1. Client sends POST request to `/login`
2. Auth Service validates credentials
3. Auth Service publishes `UserLoggedIn` event to Kafka
4. Query Service consumes event
5. Query Service stores login history in Query DB

## 📊 Database Schema

### Auth Database (Command Side)

**users table:**
```sql
- id (UUID, PK)
- email (VARCHAR, UNIQUE)
- password (VARCHAR, hashed)
- created_at (TIMESTAMP)
```

### Query Database (Query Side)

**users table:**
```sql
- id (UUID, PK)
- email (VARCHAR)
- created_at (TIMESTAMP)
```

**login_histories table:**
```sql
- id (UUID, PK)
- user_id (UUID, indexed)
- email (VARCHAR)
- login_at (TIMESTAMP, indexed)
- created_at (TIMESTAMP)
```

## 🗂️ Project Structure

```
cqrs/
├── auth-service/
│   ├── api/
│   │   └── handler.go          # HTTP handlers
│   ├── config/
│   │   ├── config.go           # Database config
│   │   ├── env.go              # Environment variables
│   │   └── jwt.go              # JWT utilities
│   ├── event/
│   │   └── producer.go         # Kafka producer
│   ├── model/
│   │   └── user.go             # User model
│   ├── repository/
│   │   └── user_repo.go        # Database operations
│   ├── service/
│   │   └── user_service.go     # Business logic
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
│
├── query-service/
│   ├── api/
│   │   └── user_handler.go     # HTTP handlers
│   ├── config/
│   │   ├── db.go               # Database config
│   │   └── env.go              # Environment variables
│   ├── event/
│   │   └── consumer.go         # Kafka consumer
│   ├── model/
│   │   ├── user.go             # User model
│   │   └── login_history.go   # Login history model
│   ├── repository/
│   │   ├── user_repo.go        # User operations
│   │   └── login_history_repo.go # Login history operations
│   ├── service/
│   │   └── user_service.go     # Event handlers
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
│
├── docker-compose.yml
└── README.md
```

## 🔧 Configuration

### Environment Variables

Services are configured via Docker Compose, but you can also use `.env` files:

**Auth Service:**
```env
DB_HOST=postgres-auth
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_db
DB_PORT=5432
DB_SSLMODE=disable
KAFKA_BROKER=kafka:29092
KAFKA_TOPIC=user-events
JWT_SECRET=your-secret-key-here
```

**Query Service:**
```env
DB_HOST=postgres-query
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=query_db
DB_PORT=5432
DB_SSLMODE=disable
KAFKA_BROKER=kafka:29092
KAFKA_TOPIC=user-events
KAFKA_GROUP=query-group
```

## 🐳 Docker Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Stop and remove volumes (deletes data)
docker-compose down -v

# Rebuild services
docker-compose up --build

# View logs
docker-compose logs -f [service-name]

# Restart specific service
docker-compose restart [service-name]

# Scale query service
docker-compose up -d --scale query-service=3
```

## 🗄️ Database Access

### Access Auth Database
```bash
docker exec -it cqrs-postgres-auth-1 psql -U postgres -d auth_db
```

### Access Query Database
```bash
docker exec -it cqrs-postgres-query-1 psql -U postgres -d query_db
```

### Useful SQL Commands
```sql
-- List tables
\dt

-- View users
SELECT * FROM users;

-- View login history
SELECT * FROM login_histories ORDER BY login_at DESC;

-- Count logins per user
SELECT user_id, email, COUNT(*) as login_count
FROM login_histories
GROUP BY user_id, email;
```

## 🔍 Monitoring

### Check Kafka Topics
```bash
docker exec -it cqrs-kafka-1 kafka-topics --list --bootstrap-server localhost:9092
```

### View Kafka Messages
```bash
docker exec -it cqrs-kafka-1 kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic user-events \
  --from-beginning
```

## 🚦 Troubleshooting

### Services Not Starting
```bash
# Check container status
docker-compose ps

# Check logs for errors
docker-compose logs [service-name]

# Restart services
docker-compose restart
```

### Kafka Connection Issues
- Ensure Kafka and Zookeeper are fully started (can take 30-60 seconds)
- Check Kafka logs: `docker-compose logs kafka`
- Verify topic creation: `docker exec -it cqrs-kafka-1 kafka-topics --list --bootstrap-server localhost:9092`

### Database Connection Issues
- Verify PostgreSQL containers are running
- Check credentials in docker-compose.yml
- Test connection: `docker exec -it cqrs-postgres-auth-1 psql -U postgres -d auth_db`

## 🎯 Development

### Local Development Setup

```bash
# Install dependencies for auth-service
cd auth-service
go mod download
go run main.go

# Install dependencies for query-service
cd query-service
go mod download
go run main.go
```

### Running Tests
```bash
# Run tests for auth-service
cd auth-service
go test ./...

# Run tests for query-service
cd query-service
go test ./...
```

## 📚 Technologies Used

- **Go 1.23**: Programming language
- **Gin**: HTTP web framework
- **GORM**: ORM library
- **PostgreSQL**: Relational database
- **Apache Kafka**: Event streaming
- **Confluent Kafka Go**: Kafka client
- **JWT**: Authentication tokens
- **Docker & Docker Compose**: Containerization
- **bcrypt**: Password hashing

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 👨‍💻 Author

Created with ❤️ by Eyup Aydin

## 🔗 Resources

- [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html)
- [Event-Driven Architecture](https://martinfowler.com/articles/201701-event-driven.html)
- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Go Documentation](https://golang.org/doc/)
