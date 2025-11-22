# Project Architecture & Structure

## 1. Architecture Overview

The **Agnos Demo Service** follows a standard **Layered Architecture** (Clean Architecture principles) typical for Go microservices. It is designed to be modular, testable, and secure.

### High-Level Diagram
```mermaid
graph TD
    Client[Client / External] -->|HTTP :80| Nginx[Nginx Load Balancer]
    Nginx -->|HTTP :8080| App[Go Application]
    App -->|TCP :5432| DB[(PostgreSQL)]
    
    subgraph "Go Application Internal"
        Router[Gin Router] --> Middleware[Middleware (Auth/Log)]
        Middleware --> Handlers[HTTP Handlers]
        Handlers --> Models[Data Models]
        Handlers --> Database[Database Layer]
    end
```

### Key Components
1.  **Entry Point (Nginx)**: Acts as a reverse proxy and load balancer, exposing only port 80 to the outside world. It forwards requests to the internal Go application.
2.  **Application Layer (Go)**:
    *   **Framework**: Built using [Gin Web Framework](https://github.com/gin-gonic/gin) for high performance.
    *   **Security**: Implements JWT-based authentication and strict Hospital-Based Access Control (HBAC).
    *   **Database Access**: Uses `pgx` driver for efficient PostgreSQL interaction.
3.  **Data Layer (PostgreSQL)**: Stores staff and patient data with relational integrity and unique constraints.

---

## 2. Project Structure

The project follows the standard [Go Project Layout](https://github.com/golang-standards/project-layout).

```
agnos_demo/
├── cmd/                    # Main applications for this project
│   ├── server/             # Entry point for the HTTP API server
│   └── migrate.go          # CLI command for database migrations
├── internal/               # Private application and library code
│   ├── database/           # Database interfaces and connection logic
│   ├── handlers/           # HTTP request handlers (Controllers)
│   ├── middleware/         # HTTP middleware (Auth, Logging)
│   ├── migrations/         # Go-based database migration logic
│   ├── mocks/              # Mock implementations for unit testing
│   ├── models/             # Domain models and data structures
│   └── routes/             # Router setup and URL mapping
├── cfg/                    # Configuration files (config.yaml)
├── docs/                   # Documentation (API Spec, ER Diagram, Architecture)
├── docker-compose.yml      # Docker services orchestration
├── Dockerfile              # Multi-stage Docker build definition
├── nginx.conf              # Nginx configuration
├── go.mod                  # Go module definition
└── Makefile                # Build and utility commands
```

### Detailed Directory Description

*   **`cmd/`**: Contains the `main` packages.
    *   `server/main.go`: Initializes the application, connects to the DB, and starts the HTTP server.
*   **`internal/handlers/`**: Contains the business logic for each endpoint (e.g., `SearchPatient`, `LoginStaff`). This layer parses requests, calls the DB, and formats responses.
*   **`internal/models/`**: Defines the Go structs that map to database tables and JSON requests/responses.
*   **`internal/middleware/`**:
    *   `auth.go`: Validates JWT tokens and extracts user context (Hospital ID).
    *   `logging.go`: Structured logging using `slog`.
*   **`internal/migrations/`**: Contains the migration logic. We use a custom Go-based migration system to ensure schema consistency on startup.

---

## 3. Security Architecture

### Authentication & Authorization
*   **JWT (JSON Web Tokens)**: Used for stateless authentication. The token contains the staff's ID and **Hospital Code**.
*   **Hospital Isolation**: A strict policy where staff can *only* access data belonging to their assigned hospital.
    *   **Search**: Automatically filters SQL queries by the staff's hospital prefix.
    *   **Direct Access**: Verifies that the requested patient's hospital matches the staff's hospital before returning data.

### Data Protection
*   **Password Hashing**: Staff passwords are hashed using **bcrypt** before storage.
*   **Network Isolation**: The Go application and Database run on an internal Docker network (`hospital-net`) and are not directly exposed to the host. Only Nginx is accessible.

---

## 4. Infrastructure

The system is containerized using Docker Compose:

| Service | Image | Description |
|---------|-------|-------------|
| **nginx** | `nginx:alpine` | Reverse proxy, listens on port 80. |
| **app** | `golang:1.24-alpine` | The main API service. Internal port 8080. |
| **db** | `postgres:16-alpine` | PostgreSQL database. Internal port 5432. |
| **migrate** | *Custom Build* | Ephemeral container that runs migrations on startup. |

### Boot Sequence
1.  **db** starts and waits for health check.
2.  **migrate** starts, connects to **db**, applies any pending schema changes, and exits.
3.  **app** starts (depends on **migrate** completion), initializes, and listens on 8080.
4.  **nginx** starts, ready to accept traffic on port 80.
