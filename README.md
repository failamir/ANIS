# Authentication Service

This is a high-performance Authentication Service built with Go, PostgreSQL, and Redis, designed according to the PRD for the Network Integrated System.

## Features

- **Authentication**: Login, Registration, Refresh Token (with Rotation).
- **Security**: Argon2id Password Hashing, JWT Access Tokens (Short-lived), Refresh Tokens (Long-lived, Revocable).
- **Infrastructure**: Dockerized (Postgres + Redis included).
- **Performance**: Built with Go, efficient DB/Cache usage.

## Prerequisites

- Docker & Docker Compose

## Quick Start

1. **Clone and Setup**
   Ensure you are in the project directory.

2. **Run with Docker**
   ```bash
   docker-compose up --build
   ```
   This will start:
   - Postgres (Database): Port 5432
   - Redis (Cache): Port 6379
   - Auth Service: Port 8080

## API Endpoints

### 1. Register
**POST** `/api/v1/auth/register`
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

### 2. Login
**POST** `/api/v1/auth/login`
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```
**Response:**
```json
{
  "AccessToken": "...",
  "RefreshToken": "...",
  ...
}
```

### 3. Refresh Token
**POST** `/api/v1/auth/refresh`
```json
{
  "refresh_token": "..."
}
```

### 4. Protected Route (Example)
**GET** `/api/v1/protected/profile`
**Headers:** `Authorization: Bearer <AccessToken>`

## Configuration

Environment variables are set in `docker-compose.yml`. For local development without Docker, copy the values to a `.env` file.
