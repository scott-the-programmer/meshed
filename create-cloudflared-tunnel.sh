#!/bin/bash

set -euo pipefail

# Support creating a tunnel for either a subdomain or the root (apex) domain.
# Usage patterns:
#   1) $0 <subdomain> <domain> [port] [k8s-service-alias]     -> tunnel for <subdomain>.<domain>
#   2) $0 <domain> [port] [k8s-service-alias]                 -> tunnel for <domain> (root/apex)
# Special subdomain tokens treated as root: @, root, apex, -
# Examples:
#   $0 myapp term.nz                      => myapp.term.nz (port 80, service: myapp.term.nz)
#   $0 myapp term.nz 8080                 => myapp.term.nz (port 8080, service: myapp.term.nz)
#   $0 myapp term.nz 8080 myapp-svc       => myapp.term.nz (port 8080, k8s service: myapp-svc)
#   $0 term.nz                            => term.nz (root, port 80, service: term.nz)
#   $0 term.nz 3000                       => term.nz (root, port 3000, service: term.nz)
#   $0 term.nz 3000 api-svc               => term.nz (root, port 3000, k8s service: api-svc)

if [ $# -lt 1 ] || [ $# -gt 4 ]; then
  echo "Usage: $0 <subdomain> <domain> [port] [k8s-service-alias]" >&2
  echo "   or: $0 <domain> [port] [k8s-service-alias] (for root/apex)" >&2
  echo "Examples:" >&2
  echo "  $0 blog example.com" >&2
  echo "  $0 blog example.com 8080" >&2
  echo "  $0 blog example.com 8080 blog-svc" >&2
  echo "  $0 example.com" >&2
  echo "  $0 example.com 3000" >&2
  echo "  $0 example.com 3000 api-svc" >&2
  exit 1
fi

if [ $# -eq 1 ]; then
  SUBDOMAIN="" # root
  DOMAIN="$1"
  PORT="80"
  K8S_SERVICE_ALIAS=""
elif [ $# -eq 2 ]; then
  # Could be either: <domain> <port> OR <subdomain> <domain>
  # If second arg is numeric, treat as <domain> <port>
  if [[ "$2" =~ ^[0-9]+$ ]]; then
    SUBDOMAIN="" # root
    DOMAIN="$1"
    PORT="$2"
    K8S_SERVICE_ALIAS=""
  else
    SUBDOMAIN="$1"
    DOMAIN="$2"
    PORT="80"
    K8S_SERVICE_ALIAS=""
  fi
elif [ $# -eq 3 ]; then
  # Could be: <subdomain> <domain> <port> OR <subdomain> <domain> <alias> OR <domain> <port> <alias>
  # If third arg is numeric, it's <subdomain> <domain> <port>
  if [[ "$3" =~ ^[0-9]+$ ]]; then
    SUBDOMAIN="$1"
    DOMAIN="$2"
    PORT="$3"
    K8S_SERVICE_ALIAS=""
  # If second arg is numeric, it's <domain> <port> <alias>
  elif [[ "$2" =~ ^[0-9]+$ ]]; then
    SUBDOMAIN="" # root
    DOMAIN="$1"
    PORT="$2"
    K8S_SERVICE_ALIAS="$3"
  # Otherwise it's <subdomain> <domain> <alias>
  else
    SUBDOMAIN="$1"
    DOMAIN="$2"
    PORT="80"
    K8S_SERVICE_ALIAS="$3"
  fi
else
  # 4 args: <subdomain> <domain> <port> <alias>
  SUBDOMAIN="$1"
  DOMAIN="$2"
  PORT="$3"
  K8S_SERVICE_ALIAS="$4"
fi

# Normalise root indicators
if [[ -n "${SUBDOMAIN}" && "${SUBDOMAIN}" =~ ^(@|root|apex|-)$ ]]; then
  SUBDOMAIN=""
fi

# For root domains we need a stable resource basename. Previous version replaced dots with '-',
# but deployments expect concatenated form (e.g. term.nz -> termnz). We'll derive both forms:
DOMAIN_HYPHEN="${DOMAIN//./-}"
DOMAIN_COMPACT="${DOMAIN//./}" # remove dots entirely

if [ -z "${SUBDOMAIN}" ]; then
  HOSTNAME="${DOMAIN}"
  # Keep tunnel name with hyphenated form for readability, but secret uses compact form.
  TUNNEL_NAME="${DOMAIN_HYPHEN}-tunnel"
  SECRET_BASENAME="${DOMAIN_COMPACT}"
  CONFIG_ID="root" # for config file naming
else
  HOSTNAME="${SUBDOMAIN}.${DOMAIN}"
  TUNNEL_NAME="${SUBDOMAIN}-tunnel"
  SECRET_BASENAME="${SUBDOMAIN}"
  CONFIG_ID="${SUBDOMAIN}"
fi

echo "Creating cloudflared tunnel for ${HOSTNAME}..."

# Check if tunnel already exists (exact name match to avoid substring collisions)
# 'cloudflared tunnel list' output typically has columns: ID NAME CREATED ...
TUNNEL_ID=$(cloudflared tunnel list | awk -v name="${TUNNEL_NAME}" 'NR>1 { if ($2==name) { print $1; exit } }')

if [ -n "$TUNNEL_ID" ]; then
  echo "Tunnel ${TUNNEL_NAME} already exists with ID: ${TUNNEL_ID}, deleting..."
  cloudflared tunnel delete -f "${TUNNEL_NAME}"
  echo "Tunnel ${TUNNEL_NAME} deleted."
  TUNNEL_ID=""
fi

echo "Creating new tunnel ${TUNNEL_NAME}..."
cloudflared tunnel create "${TUNNEL_NAME}"
TUNNEL_ID=$(cloudflared tunnel list | awk -v name="${TUNNEL_NAME}" 'NR>1 { if ($2==name) { print $1; exit } }')

if [ -z "$TUNNEL_ID" ]; then
  echo "Error: Failed to create tunnel or retrieve tunnel ID"
  exit 1
fi
echo "Tunnel created with ID: ${TUNNEL_ID}"

HOST_CLOUDFLARED_DIR="$HOME/.cloudflared"
CRED_FILE="${HOST_CLOUDFLARED_DIR}/${TUNNEL_ID}.json"

if [ ! -f "$CRED_FILE" ]; then
  echo "Error: Credential file not found at ${CRED_FILE}"
  exit 1
fi

echo "Creating Kubernetes secret with credential file..."

# Check if secret already exists (namespace 'personal' is assumed to exist)
SECRET_NAME="${SECRET_BASENAME}-cloudflared-file"
if kubectl get secret "${SECRET_NAME}" --namespace=personal >/dev/null 2>&1; then
  echo "Secret ${SECRET_NAME} already exists in personal namespace, updating..."
else
  echo "Creating new secret ${SECRET_NAME}..."
fi

# We'll later append the in-cluster config; for now just create/update with credentials
kubectl create secret generic "${SECRET_NAME}" \
  --from-file=credentials.json="${CRED_FILE}" \
  --from-file=creds.json="${CRED_FILE}" \
  --namespace=personal \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Configuring DNS routing..."

# First, try to delete any existing tunnel route
echo "Checking for existing tunnel route for ${HOSTNAME}..."
DELETE_OUTPUT=$(cloudflared tunnel route dns delete "${HOSTNAME}" 2>&1 || true)
if echo "$DELETE_OUTPUT" | grep -qi "not found\|no route\|neither the ID nor the name"; then
  echo "No existing tunnel route found for ${HOSTNAME}."
elif echo "$DELETE_OUTPUT" | grep -qi "success\|deleted\|removed"; then
  echo "Existing tunnel route for ${HOSTNAME} deleted."
else
  # Surface any unexpected output only if it seems like an actual error
  if echo "$DELETE_OUTPUT" | grep -qi "error\|failed"; then
    echo "Tunnel route deletion output: $DELETE_OUTPUT"
  else
    echo "No existing tunnel route found for ${HOSTNAME}."
  fi
fi

# Create new DNS route
echo "Creating new DNS route for ${HOSTNAME}..."
CREATE_OUTPUT=$(cloudflared tunnel route dns "${TUNNEL_NAME}" "${HOSTNAME}" 2>&1 || true)
if echo "$CREATE_OUTPUT" | grep -qi "success\|added"; then
  echo "DNS route for ${HOSTNAME} created."
elif echo "$CREATE_OUTPUT" | grep -qi "record.*already exists"; then
  echo "Warning: DNS record for ${HOSTNAME} already exists in Cloudflare."
  echo "You may need to manually delete the existing DNS record and re-run this script."
  echo "Options:"
  echo "  1. Delete the DNS record via Cloudflare dashboard: https://dash.cloudflare.com/"
  echo "  2. Use the Cloudflare API or CLI to remove the conflicting record"
  echo ""
  echo "Continuing with tunnel setup (credential and config files will still be created)..."
else
  # If output contains error, surface it but do not hard-fail unless clearly fatal
  if echo "$CREATE_OUTPUT" | grep -qi "error\|failed"; then
    echo "Warning: DNS route creation may have failed:" >&2
    echo "$CREATE_OUTPUT" >&2
  else
    echo "DNS route creation output: $CREATE_OUTPUT"
  fi
fi

echo "Creating tunnel config file..."

CONFIG_FILE="$HOME/.cloudflared/config-${CONFIG_ID}.yml"
K8S_CONFIG_FILE="$HOME/.cloudflared/config-${CONFIG_ID}-k8s.yml"

if [ -f "${CONFIG_FILE}" ]; then
  echo "Config file ${CONFIG_FILE} already exists, overwriting..."
else
  echo "Creating new config file at ${CONFIG_FILE}..."
fi

cat >"${CONFIG_FILE}" <<EOF
tunnel: ${TUNNEL_ID}
credentials-file: ${CRED_FILE}

ingress:
    - hostname: ${HOSTNAME}
    service: http://${HOSTNAME}:${PORT}
    - service: http_status:404
EOF

# Kubernetes-friendly config referencing file paths as they will appear when the secret is mounted.
# Assumes secret mounted at /etc/cloudflared (common pattern) with credentials key named credentials.json
# If K8S_SERVICE_ALIAS is provided, use it for the service backend; otherwise use hostname
if [ -z "${K8S_SERVICE_ALIAS}" ]; then
  K8S_SERVICE_TARGET="${HOSTNAME}"
else
  K8S_SERVICE_TARGET="${K8S_SERVICE_ALIAS}"
fi

cat >"${K8S_CONFIG_FILE}" <<EOF
tunnel: ${TUNNEL_ID}
credentials-file: /etc/cloudflared/credentials.json

ingress:
    - hostname: ${HOSTNAME}
    service: http://${K8S_SERVICE_TARGET}:${PORT}
    - service: http_status:404
EOF

# Patch secret to include k8s config (idempotent apply)
kubectl create secret generic "${SECRET_NAME}" \
  --from-file=credentials.json="${CRED_FILE}" \
  --from-file=creds.json="${CRED_FILE}" \
  --from-file=config.yml="${K8S_CONFIG_FILE}" \
  --namespace=personal \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Successfully created and configured tunnel ${TUNNEL_NAME} for ${HOSTNAME}"
echo "Tunnel ID: ${TUNNEL_ID}"
echo "DNS routing configured"
echo "Local config file created at: ${CONFIG_FILE}"
echo "K8s config file created at: ${K8S_CONFIG_FILE} (embedded into secret as config.yml)"
echo "Credential + config stored in k8s secret: ${SECRET_NAME} (keys: credentials.json, creds.json, config.yml) in personal namespace"
echo ""
echo "To run the tunnel locally: cloudflared tunnel --config ${CONFIG_FILE} run"
echo "K8s: mount secret ${SECRET_NAME} at /etc/cloudflared and run: cloudflared tunnel --config /etc/cloudflared/config.yml run"
