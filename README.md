# Onboarding guide

[![general-service CI](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/general-ci.yaml/badge.svg)](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/general-ci.yaml)
[![ticket-service CI](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/ticket-ci.yaml/badge.svg)](https://github.com/SoltuneMontepre/Fuvekonse/actions/workflows/ticket-ci.yaml)

## Overview

This document expands the basic steps to get the author-service running locally. Each step below preserves the original items and adds brief explanations and copy-paste commands where applicable.

## Environment setup

### Requirements

- go >= 1.25 [[Download here]](https://go.dev/doc/install)
- Docker Engine [[Download here]](https://www.docker.com/get-started/)
- Node 18+ [[Download here]](https://nodejs.org/en/download)

### 1. Set up git convention linting

Run the command:

```bash
npm i
```

This will install the `node_modules` required for git commit/branch linting.

### 2. Switch to your working directory

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

### 3. Install dependencies

After switching to the appropriate directory, ensures all Go module dependencies are downloaded and the go.mod/go.sum files are consistent.

Run:

```bash
go mod tidy
```

---

### 4. Install Documentation CLI & Hot reloading CLI

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

### 5. Creating environment variables file

Clone `.env.example` to `.env` using the following command:

Unix/macOS:

```bash
cp .env.example .env
```

Windows (PowerShell):

```pwsh
Copy-Item .env.example .env
```

Windows (CMD):

```cmd
copy .env.example .env
```

---

### 6. Start the databases via Docker

Starts required infrastructure (databases, caches, etc.) as defined in the repository's docker-compose files. Running in detached mode is common for local development.

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

---

### 7. Run database migration

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

### 8. Start the development server

Run this from the service directory (general-service or ticket-service) with a valid .env and DBs up. air is the hot-reload dev server: it watches source files, rebuilds on change, and runs pre_cmd hooks (the repo uses swag generation there).

```bash
air
```

---

to start the development server!

### 9. Start coding

Congratulations — onboarding complete.

---

### Development Flow

1. Create a feature branch

   ```bash
   git checkout -b feat/short-description
   ```

2. Commit using the repository's commit convention and push:

   ```bash
   git add .
   ```

   ```bash
   git commit -m "feat(module): short description"
   ```

   ```bash
   git push --set-upstream origin feat/short-description
   ```

3. Open a PR linking the relevant issue and follow the repo's review checklist.
