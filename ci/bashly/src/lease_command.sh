ENV_IDENTIFIER=${args[id]}

export ENV_LEASE=$($SHEPHERD_BINARY_PATH get lease "$ENV_IDENTIFIER" --namespace tpi-telemetry --json)
export LOCKFILE_DATA=$(echo "$ENV_LEASE" | jq -r .output)

echo "$LOCKFILE_DATA" >/tmp/lease.json

smith cf-login --lockfile="/tmp/lease.json"
eval $(smith om -l "/tmp/lease.json")
eval $(smith bosh -l "/tmp/lease.json")
smith open --lockfile="/tmp/lease.json"
