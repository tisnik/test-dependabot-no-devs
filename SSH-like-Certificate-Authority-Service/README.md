# CA Service (SSH Certificate Authority)

HTTPS service that signs **ED25519** SSH public keys and returns short-lived SSH user certificates.

## Quickstart (Docker Compose)

From the repo root:

```bash
# 1) CA keypair (used to sign user certs)
ssh-keygen -t ed25519 -f ./secrets/ssh/ca_key -N ""

# 2) TLS cert for the HTTPS server (example using mkcert)
mkcert -key-file ./secrets/https/ca-service-local.key.pem \
  -cert-file ./secrets/https/ca-service-local.cert.pem \
  localhost 127.0.0.1 ::1

# 3) Configure env (see Configuration section below)
cp ./secrets/.env.example ./secrets/.env

# 4) Run
docker compose up --build app
```

## Authentication

Protected endpoint `POST /sign` requires:

```text
Authorization: Bearer <token>
```

There are **two** valid token types:

- **Static access token**: matches a token in `CA_ACCESS_TOKEN` and maps to a list of principals.
- **Google SSO app JWT**: obtained via Google login and verified with `APP_JWT_SECRET`.

## Configuration

### Static token auth (`CA_ACCESS_TOKEN`)

Format:

```text
CA_ACCESS_TOKEN=token:principal1,principal2;token2:principal3
```

### Google SSO auth (OAuth)

Required env vars:

```text
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# App JWT signing (generate for example with: openssl rand -base64 32)
APP_JWT_SECRET=                 # signs/verifies the app JWT (HMAC)

# Restrict SSO to these email domains only (comma-separated, no @). Empty = reject all.
SSO_ALLOWED_DOMAINS=example.com,gmail.com,redhat.com
```

Notes:
- The app JWT minted by SSO expires after **1 hour**.

## Google SSO: how to get a Bearer token

1. Create a Google OAuth client (Web application) - see https://developers.google.com/identity/protocols/oauth2.
2. Add an **Authorized redirect URI**: project currently uses `https://localhost:8443/auth/google/callback`.
3. Start the service and open `GET /auth/google/login` in your browser.
4. After login, Google redirects to `GET /auth/google/callback`, which returns JSON containing your app token:

```json
{ "token": "<app_jwt>", "email": "user@example.com" }
```

Use that `token` as the Bearer token for `POST /sign`.

`POST /auth/logout` exists, but it is effectively “client-side logout” (delete the token client-side).

## API

### Sign an SSH public key

`POST /sign`

- Requires `Authorization: Bearer ...`
- Request JSON body:
  - `public_key`: OpenSSH authorized_keys line
  - Only **ed25519** keys are accepted
- Principals used for the issued cert come from the auth method:
  - static token → configured principals
  - SSO JWT → principals from the token claim (the user email)

Example:

```bash
curl -k -X POST "https://localhost:8443/sign" \
  -H "Authorization: Bearer <static_token_or_sso_jwt>" \
  -H "Content-Type: application/json" \
  -d '{"public_key":"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... user@host"}'
```

## Tests

```bash
docker compose run --rm --build test
```

SSO integration tests will auto-skip unless `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `APP_JWT_SECRET`, and `SSO_ALLOWED_DOMAINS` are set.
