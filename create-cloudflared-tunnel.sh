#!/bin/bash

set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <subdomain> <domain>"
    echo "Example: $0 myapp term.nz"
    echo "This will create a tunnel for myapp.term.nz"
    exit 1
fi

SUBDOMAIN="$1"
DOMAIN="$2"
TUNNEL_NAME="${SUBDOMAIN}-tunnel"
HOSTNAME="${SUBDOMAIN}.${DOMAIN}"

echo "Creating cloudflared tunnel for ${HOSTNAME}..."

# Check if tunnel already exists
TUNNEL_ID=$(cloudflared tunnel list | grep "${TUNNEL_NAME}" | awk '{print $1}')

if [ -z "$TUNNEL_ID" ]; then
    echo "Creating new tunnel ${TUNNEL_NAME}..."
    cloudflared tunnel create "${TUNNEL_NAME}"
    TUNNEL_ID=$(cloudflared tunnel list | grep "${TUNNEL_NAME}" | awk '{print $1}')

    if [ -z "$TUNNEL_ID" ]; then
        echo "Error: Failed to create tunnel or retrieve tunnel ID"
        exit 1
    fi
    echo "Tunnel created with ID: ${TUNNEL_ID}"
else
    echo "Tunnel ${TUNNEL_NAME} already exists with ID: ${TUNNEL_ID}"
fi

CRED_FILE="$HOME/.cloudflared/${TUNNEL_ID}.json"

if [ ! -f "$CRED_FILE" ]; then
    echo "Error: Credential file not found at ${CRED_FILE}"
    exit 1
fi

echo "Creating Kubernetes secret with credential file..."

# Check if secret already exists
if kubectl get secret "${SUBDOMAIN}-cloudflared-file" --namespace=personal >/dev/null 2>&1; then
    echo "Secret ${SUBDOMAIN}-cloudflared-file already exists in personal namespace, updating..."
else
    echo "Creating new secret ${SUBDOMAIN}-cloudflared-file..."
fi

kubectl create secret generic "${SUBDOMAIN}-cloudflared-file" \
    --from-file=credentials.json="${CRED_FILE}" \
    --namespace=personal \
    --dry-run=client -o yaml | kubectl apply -f -

echo "Configuring DNS routing..."

# Check if DNS route already exists
if cloudflared tunnel route show | grep -q "${HOSTNAME}"; then
    echo "DNS route for ${HOSTNAME} already exists, skipping..."
else
    echo "Creating DNS route for ${HOSTNAME}..."
    cloudflared tunnel route dns "${TUNNEL_NAME}" "${HOSTNAME}"
fi

echo "Creating tunnel config file..."

CONFIG_FILE="$HOME/.cloudflared/config-${SUBDOMAIN}.yml"

if [ -f "${CONFIG_FILE}" ]; then
    echo "Config file ${CONFIG_FILE} already exists, overwriting..."
else
    echo "Creating new config file at ${CONFIG_FILE}..."
fi

cat > "${CONFIG_FILE}" << EOF
tunnel: ${TUNNEL_ID}
credentials-file: ${CRED_FILE}

ingress:
  - hostname: ${HOSTNAME}
    service: http://localhost:8080
  - service: http_status:404
EOF

echo "Successfully created and configured tunnel ${TUNNEL_NAME} for ${HOSTNAME}"
echo "Tunnel ID: ${TUNNEL_ID}"
echo "DNS routing configured"
echo "Config file created at: ${CONFIG_FILE}"
echo "Credential file copied to k8s secret: ${SUBDOMAIN}-cloudflared-file in personal namespace"
echo ""
echo "To run the tunnel locally: cloudflared tunnel --config ${CONFIG_FILE} run"
echo "To deploy in k8s, use the ${SUBDOMAIN}-cloudflared-file secret in your deployment"
