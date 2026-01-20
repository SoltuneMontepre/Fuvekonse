# Seed Test Users

This script creates test users with different roles in the database for testing.

## Usage

Run the seed script:

```bash
cd services/general-service
go run cmd/seed/main.go
```

## Test User Credentials

After running the seed script, you can login with:

### Admin User
```json
{
  "email": "admin@fuve.com",
  "password": "admin123"
}
```

### Regular User
```json
{
  "email": "user@fuve.com",
  "password": "user123"
}
```

### Dealer User
```json
{
  "email": "dealer@fuve.com",
  "password": "dealer123"
}
```

### Legacy Test User
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

## User Details

| Email | Password | Role | Fursona Name |
|-------|----------|------|--------------|
| admin@fuve.com | admin123 | Admin | AdminFox |
| user@fuve.com | user123 | User | UserWolf |
| dealer@fuve.com | dealer123 | Dealer | DealerCat |
| user@example.com | password123 | User | TestFursona |

## Testing the Login API

### Using cURL:

```bash
# Login as Admin
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@fuve.com",
    "password": "admin123"
  }'

# Login as User
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@fuve.com",
    "password": "user123"
  }'
```

### Using PowerShell:

```powershell
# Login as Admin
$body = @{
    email = "admin@fuve.com"
    password = "admin123"
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

## Role Permissions

| Role | Can Access Admin Panel | Can Purchase Tickets | Can Manage Dealer Booth |
|------|------------------------|---------------------|------------------------|
| Admin | Yes | Yes | Yes |
| User | No | Yes | No |
| Dealer | No | Yes | Yes |

## Notes

- The script will update users if they already exist
- Password is hashed using bcrypt
- All users are automatically verified (is_verified = true)
- Admin users can access `/admin/*` routes
