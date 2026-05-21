# MuchToDo API

A robust RESTful API for a ToDo application built with Go (Golang). This project features user authentication, JWT-based session management, CRUD operations for ToDo items, and an optional Redis caching layer.

The API is built with a clean, layered architecture to separate concerns, making it scalable and easy to maintain. It includes a full suite of unit and integration tests and provides interactive API documentation via Swagger.

## Features

* **User Management**: Secure user registration, login, update, and deletion.
* **Authentication**: JWT-based authentication that supports both `httpOnly` cookies (for web clients) and `Authorization` headers.
* **CRUD for ToDos**: Full create, read, update, and delete functionality for user-specific ToDo items.
* **Structured Logging**: Configurable, structured JSON logging with request context for production-ready monitoring.
* **Optional Caching**: Redis-backed caching layer that can be toggled on or off via environment variables.
* **API Documentation**: Auto-generated interactive Swagger documentation.
* **Testing**: Comprehensive unit and integration test suites.
* **Graceful Shutdown**: The server shuts down gracefully, allowing active requests to complete.

## Prerequisites

To run this project locally, you will need the following installed:

* **Go**: Version 1.21 or later.
* **Swag CLI**: To generate the Swagger API documentation.
* **Make** (optional, for easier command execution):

  On macOS, you can install `make` via Homebrew if it's not already available:

  ```bash
  brew install make
  ```

  On Linux, `make` is usually pre-installed or available via your package manager.

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Using Make

This project includes a `Makefile` to simplify common development tasks. You can use `make <target>` to run commands such as starting the server, building, running tests, and managing Docker containers.

## Getting Started

### 1. Clone the Repository

```bash
git clone <your-repository-url>
cd starttech-application/backend
```

### 2. Configure Environment Variables

Create a `.env` file in the root of the project by copying the example.

```bash
cp .env.example .env
```

Now, open the `.env` file and **change the** `JWT_SECRET_KEY` to a new, long, random string.

Also, ensure that the `MONGO_URI` and `DB_NAME` points to your local MongoDB instance and db.

You can leave the other variables as they are for local development.

### 3. Start Local Dependencies

With Docker running, start the MongoDB and Redis containers using Docker Compose.

```bash
docker-compose up -d
```
**Or using Make:**
```bash
make dc-up
```

### 4. Install Go Dependencies

Download the necessary Go modules.

```bash
go mod tidy
```
**Or using Make:**
```bash
make tidy
```

### 5. Generate API Documentation

Generate the Swagger/OpenAPI documentation from the code comments.

```bash
swag init -g cmd/api/main.go
```
**Or using Make:**
```bash
make generate-docs
```

### 6. Run the Application

You can now run the API server.

```bash
go run ./cmd/api/main.go
```
**Or using Make (also generates docs first):**
```bash
make run
```

The server will start, and you should see log output in your terminal.

* The API will be available at `http://localhost:8080`.
* The interactive Swagger documentation will be at `http://localhost:8080/swagger/index.html`.

## Running Tests

The project includes both unit and integration tests.

### Run Unit Tests

These tests are fast and do not require any external dependencies.

```bash
go test ./...
```
**Or using Make:**
```bash
make unit-test
```

### Run Integration Tests

These tests require Docker to be running as they spin up their own temporary database and cache containers.

```bash
INTEGRATION=true go test -v --tags=integration ./...
```
**Or using Make:**
```bash
make integration-test
```

The `INTEGRATION=true` environment variable is required to explicitly enable these tests. The `-v` flag provides verbose output.

## Other Useful Make Commands

- **Build the binary:**  
  ```bash
  make build
  ```
- **Clean build artifacts:**  
  ```bash
  make clean
  ```
- **Stop Docker containers:**  
  ```bash
  make dc-down
  ```
- **Restart Docker containers:**  
  ```bash
  make dc-restart
  ```

Refer to the `Makefile` for more available commands.
