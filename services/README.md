Onboarding guide

requirements:
go >= 1.25
swag CLI (for Swagger documentation generation)

This document expands the basic steps to get the author-service running locally. Each step below preserves the original items and adds brief explanations and copy-paste commands where applicable.

1. go mod tidy

- Ensures all Go module dependencies are downloaded and the go.mod/go.sum files are consistent.
- Run:

```
go mod tidy
```

2. Install swag CLI

- Installs the Swagger documentation generator CLI tool required for development builds.
- Run:

```
go install github.com/swaggo/swag/cmd/swag@latest
```

3. Docker compose up

- Starts required infrastructure (databases, caches, etc.) as defined in the repository's docker-compose files. Running in detached mode is common for local development.
- Example (detached):

```
docker compose up -d
```

- Example (foreground, useful to watch logs):

```
docker compose up
```

4. clone .env.example to .env using command

- Copy the example environment file to a working .env file and update any values (DB credentials, ports, API keys) as needed.
- Unix/macOS:

```
cp .env.example .env
```

- Windows (PowerShell):

```
Copy-Item .env.example .env
```

5. migrate: go run ./cmd/migrate

- Applies database migrations so the service has the required schema/tables before running.
- Run:

```
go run ./cmd/migrate
```

6. run dev: air

- Starts the development server with live reload (using air). Ensure air is installed and configured for this project.
- If air is installed globally:

```
air
```

- If using go run for air (if not installed globally):

```
go run github.com/cosmtrek/air@latest
```

Notes:

- Verify Docker and Docker Compose are installed and running before step 3.
- Ensure your .env values match any requirements from the docker-compose services (ports, credentials).
- If ports or services conflict, stop other local services or adjust the .env/docker-compose settings.
- Follow repository README or docs for any additional environment-specific settings.
- The swag CLI is required for development builds as it generates Swagger documentation automatically via air's pre_cmd hook.
