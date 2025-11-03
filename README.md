# Fuvekonse

[![general-service CI](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/general-ci.yaml/badge.svg)](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/general-ci.yaml)
[![ticket-service CI](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/ticket-ci.yaml/badge.svg)](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/ticket-ci.yaml)

## Overview

Fuvekonse is a microservices-based application built with Go, featuring:

- **General Service**: Core application functionality including user management, roles, and permissions
- **Ticket Service**: Ticket management and processing system

The services use PostgreSQL for data persistence, Redis for caching, and LocalStack for local AWS services (S3, SQS, SES) development.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Environment Setup](#environment-setup)
- [Running the Services](#running-the-services)
- [Development Flow](#development-flow)
- [LocalStack Guide](#localstack-guide)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **Go >= 1.25** [[Download here]](https://go.dev/doc/install)
- **Docker Engine** [[Download here]](https://www.docker.com/get-started/)
- **Node.js 18+** [[Download here]](https://nodejs.org/en/download)

### Verify Your Installation

Before proceeding, verify that all prerequisites are installed correctly:

```bash
go version
docker --version
node --version
npm --version
```

Expected output should show versions matching or exceeding the requirements above.

## Quick Start

For experienced developers, here's the TL;DR:

```bash
# Install dependencies
npm i

# Copy environment files
cp .env.example ./services/general-service/.env
cp .env.example ./services/ticket-service/.env

# Install Go tools
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/air-verse/air@latest

# Start infrastructure
docker compose up -d

# Run migrations
cd services/general-service
go mod tidy
go run ./cmd/migrate

# Start development server
air
```

## Environment Setup

### Steps:

#### 1. Set up git convention linting

Install the Node.js dependencies required for git commit/branch linting:

```bash
npm i
```

This configures Husky hooks to enforce conventional commit messages.

---

#### 2. Create environment variables file

Copy the example environment file to each service directory:

**Unix / macOS / Linux (bash):**

```bash
cp .env.example ./services/general-service/.env
cp .env.example ./services/ticket-service/.env
```

**Windows (PowerShell):**

```powershell
Copy-Item .env.example .\services\general-service\.env -Force
Copy-Item .env.example .\services\ticket-service\.env -Force
```

**Windows (CMD):**

```cmd
copy .env.example services\general-service\.env
copy .env.example services\ticket-service\.env
```

**Environment Variables Reference:**

The `.env` file contains the following configuration:

| Variable             | Description             | Default Value                                                         |
| -------------------- | ----------------------- | --------------------------------------------------------------------- |
| `PORT`               | Service port            | `8085`                                                                |
| `DB_HOST`            | PostgreSQL host         | `localhost`                                                           |
| `DB_PORT`            | PostgreSQL port         | `5432`                                                                |
| `DB_USER`            | Database user           | `root`                                                                |
| `DB_PASSWORD`        | Database password       | `root`                                                                |
| `DB_NAME`            | Database name           | `fuvekon`                                                             |
| `REDIS_HOST`         | Redis host              | `localhost`                                                           |
| `REDIS_PORT`         | Redis port              | `6379`                                                                |
| `AWS_REGION`         | AWS region              | `ap-southeast-1`                                                      |
| `S3_BUCKET_URL`      | LocalStack S3 endpoint  | `http://localhost:4566/fuvekonse-bucket`                              |
| `SQS_QUEUE_URL`      | LocalStack SQS endpoint | `http://sqs.ap-southeast-1.localhost:4566/000000000000/fuvekon-queue` |
| `SES_EMAIL_IDENTITY` | SES email sender        | `fuveSupport@fuve.com`                                                |

---

#### 3. Install Go dependencies

Navigate to the service directory you want to work on:

**General service:**

```bash
cd services/general-service
```

**Ticket service:**

```bash
cd services/ticket-service
```

Then install dependencies:

```bash
go mod tidy
```

This ensures all Go module dependencies are downloaded and the `go.mod`/`go.sum` files are consistent.

---

#### 4. Install development tools

Install the Swagger documentation generator and Air hot-reload CLI globally:

**Bash / macOS / Linux:**

```bash
go install github.com/swaggo/swag/cmd/swag@latest && go install github.com/air-verse/air@latest
```

**PowerShell (pwsh):**

```powershell
go install github.com/swaggo/swag/cmd/swag@latest; go install github.com/air-verse/air@latest
```

**Windows CMD:**

```cmd
go install github.com/swaggo/swag/cmd/swag@latest && go install github.com/air-verse/air@latest
```

**Or install them one by one:**

```bash
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/air-verse/air@latest
```

**Important Notes:**

- Ensure your Go bin directory (`GOBIN` or `$(go env GOPATH)/bin`) is in your `PATH`
- Verify installation with `swag --version` and `air --version`
- These tools are installed globally and can be reused across projects

---

## Running the Services

#### 5. Start infrastructure services

Return to the project root directory:

```bash
cd ../..
```

Start the required infrastructure (PostgreSQL, Redis, LocalStack) using Docker Compose:

**Detached mode (recommended):**

```bash
docker compose up -d
```

**Foreground mode (to see logs):**

```bash
docker compose up
```

**Services started:**

- **PostgreSQL** (port 5432) - Main database
- **Redis** (port 6379) - Caching layer
- **LocalStack** (port 4566) - Local AWS services:
  - [S3 Bucket](https://docs.localstack.cloud/aws/services/s3/)
  - [Simple Queue Service (SQS)](https://docs.localstack.cloud/aws/services/sqs/)
  - [Simple Email Service (SES)](https://docs.localstack.cloud/aws/services/ses/)

---

#### 6. Run database migrations

Navigate to the general service directory:

```bash
cd services/general-service
```

Apply database migrations to create the required schema and tables:

```bash
go run ./cmd/migrate
```

This ensures your database schema is up-to-date before running the service.

---

#### 7. Start the development server

From the service directory (with `.env` configured and infrastructure running), start the hot-reload development server:

```bash
air
```

**What Air does:**

- Watches source files for changes
- Automatically rebuilds and restarts the service
- Generates Swagger documentation on each rebuild
- Provides live feedback during development

**Service URLs:**

- **General Service**: `http://localhost:8085` (or your configured `PORT`)
- **API Documentation**: `http://localhost:8085/swagger/index.html`

---

#### 8. Start coding! 🎉

Congratulations — onboarding complete. Your development environment is ready.

---

## Development Flow

### Git Workflow

#### 1. Create a feature branch

Follow the conventional branch naming pattern:

```bash
git checkout -b feat/short-description      # For new features
git checkout -b fix/short-description       # For bug fixes
git checkout -b docs/short-description      # For documentation
git checkout -b refactor/short-description  # For refactoring
```

#### 2. Commit using conventional commits

Use semantic commit messages that follow the pattern: `type(scope): description`

```bash
git add .
git commit -m "feat(auth): add JWT token validation"
git commit -m "fix(api): resolve null pointer in user handler"
git commit -m "docs(readme): update environment setup instructions"
```

**Common commit types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

#### 3. Push and create a Pull Request

```bash
git push --set-upstream origin feat/short-description
```

Then:

1. Open a PR on GitHub
2. Link relevant issues using keywords (e.g., `Closes #123`)
3. Fill out the PR template
4. Request reviews from team members
5. Address review feedback
6. Wait for CI checks to pass

---

## LocalStack Guide

LocalStack provides a local AWS cloud stack for development and testing.

### Accessing LocalStack

After running `docker compose up`, you can access LocalStack at:

**API Endpoint:**

```
http://localhost:4566
```

**Swagger UI (API Documentation):**

```
http://localhost.localstack.cloud:4566/_localstack/swagger
```

### Interacting with LocalStack Services

#### Using AWS CLI

Configure AWS CLI to use LocalStack:

```bash
aws configure set aws_access_key_id test
aws configure set aws_secret_access_key test
aws configure set region ap-southeast-1
```

#### S3 Examples

**List buckets:**

```bash
aws --endpoint-url=http://localhost:4566 s3 ls
```

**Upload a file:**

```bash
aws --endpoint-url=http://localhost:4566 s3 cp myfile.txt s3://fuvekonse-bucket/
```

**List objects in bucket:**

```bash
aws --endpoint-url=http://localhost:4566 s3 ls s3://fuvekonse-bucket/
```

#### SQS Examples

**List queues:**

```bash
aws --endpoint-url=http://localhost:4566 sqs list-queues
```

**Send a message:**

```bash
aws --endpoint-url=http://localhost:4566 sqs send-message \
  --queue-url http://sqs.ap-southeast-1.localhost:4566/000000000000/fuvekon-queue \
  --message-body "Test message"
```

**Receive messages:**

```bash
aws --endpoint-url=http://localhost:4566 sqs receive-message \
  --queue-url http://sqs.ap-southeast-1.localhost:4566/000000000000/fuvekon-queue
```

#### SES Examples

**Verify email identity:**

```bash
aws --endpoint-url=http://localhost:4566 ses verify-email-identity \
  --email-address fuveSupport@fuve.com
```

**Send test email:**

```bash
aws --endpoint-url=http://localhost:4566 ses send-email \
  --from fuveSupport@fuve.com \
  --destination "ToAddresses=recipient@example.com" \
  --message "Subject={Data=Test},Body={Text={Data=Hello from LocalStack}}"
```

---

## Troubleshooting

### Common Issues and Solutions

#### 1. Port already in use

**Symptoms:** Docker fails to start with "port is already allocated" error

**Solution:**

```bash
# Check what's using the port (example for port 5432)
# Windows PowerShell:
netstat -ano | findstr :5432

# Kill the process or stop the conflicting service
# Then restart Docker Compose
docker compose down
docker compose up -d
```

#### 2. Go tools not found (swag/air)

**Symptoms:** `swag: command not found` or `air: command not found`

**Solution:**

Check if Go bin is in your PATH:

```bash
# Check Go bin path
go env GOPATH

# Add to PATH (PowerShell - add to $PROFILE for persistence)
$env:PATH += ";$(go env GOPATH)\bin"

# Verify installation
swag --version
air --version
```

#### 3. Docker containers not starting

**Symptoms:** Services fail to start or immediately exit

**Solution:**

```bash
# Check container logs
docker compose logs

# Restart specific service
docker compose restart fuvekon-db

# Clean restart
docker compose down -v
docker compose up -d
```

#### 4. Database connection errors

**Symptoms:** Service can't connect to PostgreSQL

**Solution:**

1. Verify database is running: `docker compose ps`
2. Check `.env` credentials match `docker-compose.yml`
3. Ensure database has been migrated: `go run ./cmd/migrate`
4. Test connection:
   ```bash
   docker exec -it fuvekon-db psql -U root -d fuvekon
   ```

#### 5. Redis connection errors

**Symptoms:** Cache operations fail

**Solution:**

```bash
# Test Redis connection
docker exec -it fuvekon-cache redis-cli ping
# Should return: PONG

# Check Redis logs
docker compose logs fuvekon-cache
```

#### 6. Migration fails

**Symptoms:** Database migration errors

**Solution:**

```bash
# Ensure you're in the correct directory
cd services/general-service

# Check database is accessible
docker compose ps fuvekon-db

# Try running migration with verbose output
go run ./cmd/migrate
```

#### 7. Hot reload not working (Air)

**Symptoms:** Changes don't trigger rebuild

**Solution:**

1. Ensure you're in the service directory (not root)
2. Check `.air.toml` configuration exists
3. Restart Air: `Ctrl+C` then `air`
