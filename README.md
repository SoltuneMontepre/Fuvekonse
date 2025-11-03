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
- [Testing](#testing)

## Prerequisites

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

#### 2. Creating environment variables file

Clone the example env into each service’s .env file.

Unix / macOS (bash):

```bash
cp .env.example ./services/general-service/.env
cp .env.example ./services/ticket-service/.env
```

Windows (PowerShell):

```powershell
Copy-Item .env.example .\services\general-service\.env -Force
Copy-Item .env.example .\services\ticket-service\.env -Force
```

Windows (CMD):

```cmd
copy .env.example services\general-service\.env
copy .env.example services\ticket-service\.env
```

#### 3. Switch to your working directory

Go to either of the services in `src/services/` using these command:

General service:

```bash
cd .\services\general-service\
```

Ticket service:

```bash
cd .\services\ticket-service\
```

---

#### 4. Install dependencies

After switching to the appropriate directory, ensures all Go module dependencies are downloaded and the go.mod/go.sum files are consistent.

Run:

```bash
go mod tidy
```

---

#### 5. Install Documentation CLI & Hot reloading CLI

Installs the Swagger documentation generator CLI tool required for development builds and `air-cli` globally to allow hot-reload.

Bash / macOS / Linux:

```bash
go install github.com/swaggo/swag/cmd/swag@latest && go install github.com/air-verse/air@latest
```

PowerShell (pwsh):

```powershell
go install github.com/swaggo/swag/cmd/swag@latest; go install github.com/air-verse/air@latest
```

Windows CMD:

```cmd
go install github.com/swaggo/swag/cmd/swag@latest && go install github.com/air-verse/air@latest
```

Or run them one by one:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

```bash
go install github.com/air-verse/air@latest
```

Note:

- Ensure your Go bin directory (GOBIN or $(go env GOPATH)/bin) is on PATH. Verify with `swag --version` and `air --version`.
- These will be installed globally so next time you don't have to do this again for other projects.

---

#### 6. Start the required services via Docker

Starts required infrastructure (databases, caches, etc.) as defined in the repository's `docker-compose` files. Running in detached mode is common for local development.

First, go back to the root directory (where the `docker-compose.yml` file is) with the command:

```bash
cd ../..
```

Detatched mode:

```bash
docker compose up -d
```

Run normally:

```bash
docker compose up
```

Services being started up includes:

- `Redis`
- `PostgreSQL`
- `S3 Bucket` (https://docs.localstack.cloud/aws/services/s3/)
- `Simple Queue Service` (https://docs.localstack.cloud/aws/services/sqs/)
- `Simple Message Service` (https://docs.localstack.cloud/aws/services/ses/)

---

#### 7. Run database migration

Applies database migrations so the service has the required schema/tables before running.

First, go back to general service:

```bash
cd .\services\general-service\
```

- Run:

```bash
go run ./cmd/migrate
```

---

#### 8. Start the development server

Run this from the service directory (general-service or ticket-service) with a valid .env and DBs up. air is the hot-reload dev server: it watches source files, rebuilds on change, and runs pre_cmd hooks (the repo uses swag generation there).

```bash
air
```

---

to start the development server!

#### 9. Start coding

Congratulations — onboarding complete.

---

### Development Flow

#### 1. Create a feature branch

```bash
   git checkout -b feat/short-description
```

#### 2. Commit using the repository's commit convention and push:

```bash
   git add .
```

```bash
   git commit -m "feat(module): short description"
```

```bash
   git push --set-upstream origin feat/short-description
```

#### 3. Open a PR linking the relevant issue and follow the repo's review checklist.

## LocalStack Guide

After running `docker compose up`, one can go to:

```
http://localhost.localstack.cloud:4566/_localstack/swagger
```

to view what API for localstack is available
