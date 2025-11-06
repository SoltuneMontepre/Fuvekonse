# Seed Test User

This script creates a test user in the database for testing the authentication API.

## Usage

Run the seed script:

```bash
cd services/general-service
go run cmd/seed/main.go
```

## Test User Credentials

After running the seed script, you can login with:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

## User Details

- **Email**: user@example.com
- **Password**: password123
- **Fursona Name**: TestFursona
- **First Name**: Test
- **Last Name**: User
- **Role**: User
- **Is Verified**: true

## Testing the Login API

### Using cURL:

```bash
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Using PowerShell:

```powershell
$body = @{
    email = "user@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8085/api/v1/auth/login" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

### Using Swagger UI:

Visit: http://localhost:8085/swagger/index.html

1. Navigate to the `/auth/login` endpoint
2. Click "Try it out"
3. Enter the credentials
4. Click "Execute"

## Expected Response

```json
{
  "user": {
    "id": "06313f91-9d9a-482d-a6aa-59a2f9fad3f7",
    "email": "user@example.com",
    "fursona_name": "TestFursona",
    "first_name": "Test",
    "last_name": "User",
    "role": "User",
    "avatar": "https://via.placeholder.com/150",
    "is_verified": true
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

## Notes

- The script will update the user if it already exists
- Password is hashed using bcrypt
- The user is automatically verified (is_verified = true)
