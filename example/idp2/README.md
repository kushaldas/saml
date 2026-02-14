# Example SAML Identity Provider (idp2)

A SAML 2.0 Identity Provider built on the [crewjam/saml](https://github.com/crewjam/saml) library, designed to run behind a reverse proxy (e.g. Caddy) without TLS.

I wrote this for my book [Learning SAML](https://kushaldas.in/learningsaml/) as a simple SAML Idp.

## Quick Start with Docker

### 1. Generate SAML signing certificates

```bash
mkdir -p data
openssl req -x509 -newkey rsa:2048 -keyout tmp.key -out data/idp.pem -days 3650 -nodes -subj '/CN=your.domain'
openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in tmp.key -out data/idp.key
rm tmp.key
```

The key must be in PKCS8 format (not PKCS1).

### 2. Configure the IDP base URL

Create a `.env` file:

```
IDP_BASE_URL=https://your-idp-domain.example.com
```

If omitted, it defaults to `https://localhost:9090`.

### 3. Build and run

```bash
docker compose build
docker compose up -d
```

The IDP listens on `127.0.0.1:9090` (plain HTTP). Point your reverse proxy at this address for TLS termination.

## Building without Docker

The `go.mod` lives at the repository root (`../../`), not in this directory.

```bash
cd /path/to/saml
go build -o example/idp2/idp2 ./example/idp2/
cd example/idp2
./idp2 -idp https://your-idp-url -bind :8080
```

The binary expects `idp.key` and `idp.pem` in the working directory, and `users.json` in either the working directory or `/etc/idp2/`.

## Test Users

Run `python3 create_users.py` to generate `users.json` with two test users (password for both: `hunter2`):

| User  | Email            | Groups                |
|-------|------------------|-----------------------|
| alice | alice@example.se | Administrators, Users |
| bob   | bob@example.se   | Users                 |

## Endpoints

| Path | Description |
|------|-------------|
| `/` | Welcome / index page |
| `/metadata` | SAML IdP metadata XML |
| `/sso` | Single Sign-On endpoint |
| `/login` | Login form |

## Registering a Service Provider

Download the SP metadata and upload it to the IdP. In real production software you will always set this via configuration and you will also make sure that no one should be able to change any of these settings/values.

```bash
# Download the SP metadata (from the SP running in another terminal)
curl http://localhost:5000/metadata/ -o sp.xml

# Upload it to the IdP
curl -i -L -X PUT --data-binary @sp.xml http://localhost:8080/services/localhost:5000
```

The last part of the URL (`localhost:5000`) is just a key for this test IdP.

## Learn More


