# go-certdist

`go-certdist` is a simple and secure tool for distributing SSL/TLS certificates from a central server to multiple clients. It uses `age` for end-to-end encryption, ensuring that certificates are only readable by the intended client.
Allowlisting age public keys is implemented to prevent unauthorized access to certificates.

## Features

- **Secure by Design**: Leverages `age` encryption for strong, modern, and easy-to-use public-key cryptography. This allows even secure transmissions via http, but still a reverse proxy for https is encouraged.
- **Efficient Distribution**: Clients only updates the certificate when the client's version is expired or missing.
- **Simple Configuration**: Uses straightforward YAML files for both server and client configuration.
- **Automated Renewals**: Automatically execute shell commands on the client after a new certificate is successfully downloaded.
- **Cross-Platform**: Builds and runs on Linux, macOS, and Windows.

## Getting Started

### 1. Build from Source

First, clone the repository and build the binary:

```bash
go build
```

This will create a `go-certdist` (or `go-certdist.exe`) binary in the project root.

### 2. Generate Keys

`go-certdist` uses `age` key pairs for encryption. The server needs the client's *public key* to encrypt the certificate, and the client uses its *private key* to decrypt it.
It's perfectly fine to also use age-keygen to generate these keys, if installed.

Generate a new key pair:

```bash
./go-certdist keygen
# AGE_PRIVATE_KEY=AGE-SECRET-KEY-1...
# AGE_PUBLIC_KEY=age1...
```

Save these keys. The `AGE_PRIVATE_KEY` is a secret and should be protected on the client machine. The `AGE_PUBLIC_KEY` is safe to share and has to be added to the server's configuration.

### 3. Create Configuration Files

You can generate dummy configuration files to get started quickly.

**Server Config:**

```bash
./go-certdist config server > server.yml
```

**Client Config:**

```bash
./go-certdist config client > client.yml
```

Now, edit these files with your specific settings.

## Usage

### Server

The server is responsible for serving certificates to authorized clients.

**Example `server.yml`:**

```yaml
server:
  port: 8080  # Use a reverse proxy to serve https
  certificate_directories:
    - "/etc/letsencrypt/live"
public_age_keys:
  - "age1..." # Client 1's public key
  - "age1..." # Client 2's public key
```

- `port`: The port the server will listen on.
- `certificate_directories`: A list of directories where the server will look for certificates.
- `public_age_keys`: A allowlist of client public keys that are authorized to request certificates.

**To start the server:**

```bash
./go-certdist server server.yml
```

### Client

The client requests certificates from the server for specific domains.

**Example `client.yml`:**

```yaml
connection:
  server: "https://your-server.com"
age_key:
  # public_key: <optional, auto-generated from private-key>
  private_key: "AGE-SECRET-KEY-1..."
certificate:
  - domain: "example.com"
    directory: "/path/to/client/certs/example.com"
    renew_commands:
      - "systemctl reload nginx"
      - "systemctl reload postfix"
```

- `server`: The URL of the `go-certdist` server.
- `private_key`: The client's secret `age` private key.
- `domain`: The domain for which to request a certificate.
- `directory`: The directory where the downloaded certificate files will be saved.
- `renew_commands`: A list of shell commands to execute after a new certificate is successfully downloaded.

**To run the client:**

```bash
./go-certdist client client.yml
```

The client will check for an existing certificate. If one exists and is not expiring soon, it will send its expiration date to the server. The server will only send a new certificate if the client's version is expired or missing. Otherwise, it returns a `304 Not Modified` and the client exits gracefully.

## Development

To run the end-to-end integration test:

```bash
go build -o certdist
./scripts/run-integration-test.sh
```

