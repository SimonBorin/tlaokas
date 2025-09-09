# The Lost Art of Keeping a Secret

A minimal Go web service for securely sharing secrets.  
Secrets are encrypted server-side using AES-256 and can be accessed **once** or until **12 hours have passed**, whichever comes first.

## Features

- 🔐 AES-256 encryption with unique IV per secret
- 🧾 Secrets stored in PostgreSQL with optional expiration
- 🕓 Secrets expire after 12 hours or after first view
- 🔁 Automatic cleanup of viewed/expired secrets
- 🌐 Simple HTTP API with CORS enabled

## API

### POST `/secret`

Creates a new encrypted secret.

**Request:**

```json
{
  "secret": "This is my top secret message"
}
```

**Response:**

```json
{
  "url": "http://localhost:8080/secret/<uuid>"
}
```

---

### GET `/secret/{id}`

Retrieves and decrypts a secret (only once).  
Returns 404 if the secret is expired or already viewed.

**Response:**

Plaintext secret in `text/plain`.

---

## Setup

### Requirements

- Go 1.20+
- PostgreSQL
- `DB_DSN` environment variable set, for example:

```bash
export DB_DSN="postgres://user:password@localhost:5432/secretdb"
```

### Run

```bash
go run main.go
```

Server will start on `http://localhost:8080`.

## Example

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"secret": "Hello, world!"}' \
     http://localhost:8080/secret

# -> {"url":"http://localhost:8080/secret/abc123-uuid"}
```

Then open the provided URL to view the secret.

## Notes

- Secrets are deleted from the database once viewed or expired.
- Make sure your encryption key is kept secure and never hardcoded in production!

## License

MIT
