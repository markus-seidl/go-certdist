#!/bin/bash

set -e

BINARY_PATH=${1:-"./certdist"}

# Function to clean up temporary files and directories
cleanup() {
    echo "Cleaning up..."
    rm -rf certs client-out server.yml client.yml keys.env ./renew_command_executed
}

# Trap to ensure cleanup is called on script exit
trap cleanup EXIT

if [ ! -f "$BINARY_PATH" ]; then
    echo "Binary not found at $BINARY_PATH. Please build the project first or provide the correct path."
    exit 1
fi

chmod +x $BINARY_PATH

echo "--- Setting up test environment ---"
mkdir -p certs client-out
DOMAIN="certdist.example.com"

# Create a temporary openssl config for SAN
cat > certs/openssl.cnf <<EOL
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no
[req_distinguished_name]
CN = $DOMAIN
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = $DOMAIN
EOL

openssl req -x509 -newkey rsa:2048 -keyout certs/test.key -out certs/test.pem -sha256 -days 3 -nodes -config certs/openssl.cnf -extensions v3_req

echo "--- Setting up port for integration test ---"
PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()')
echo "Using port $PORT for integration test"

echo "--- Generating keys ---"
$BINARY_PATH keygen > keys.env
source keys.env

echo "--- Configuring server and client ---"
cat << EOF > server.yml
server:
  port: $PORT
  certificate_directories:
    - "certs"
public_age_keys:
  - "$AGE_PUBLIC_KEY"
EOF
echo "Server configuration:"
cat server.yml

cat << EOF > client.yml
connection:
  server: "http://localhost:$PORT"
age_key:
  private_key: "$AGE_PRIVATE_KEY"
certificate:
  - domain: "$DOMAIN"
    directory: "client-out"
    renew_commands:
      - "touch ./renew_command_executed"
EOF
echo "Client configuration:"
cat client.yml

echo "--- Running integration test ---"
$BINARY_PATH server server.yml &
SERVER_PID=$!

echo "Wait for server to start..."
sleep 1

$BINARY_PATH client client.yml

kill $SERVER_PID

echo "--- Verifying result ---"
if diff -q certs/test.pem client-out/test.pem && diff -q certs/test.key client-out/test.key; then
  echo "* PASS: Certificate and private key are identical."
else
  diff certs/test.pem client-out/test.pem
  diff certs/test.key client-out/test.key

  echo "* FAIL: Files differ or are not found."
  exit 1
fi

if [ ! -f "./renew_command_executed" ]; then
  echo "FAIL: Renew command was not executed."
  exit 1
else
  echo "* PASS: Renew command was executed."
fi

echo "Integration test passed!"
