ENV_IDENTIFIER=${args[id]}

export ENV_LEASE=$($SHEPHERD_BINARY_PATH get lease "$ENV_IDENTIFIER" --namespace tpi-telemetry --json)
export LOCKFILE_DATA=$(echo "$ENV_LEASE" | jq -r .output)

echo "$LOCKFILE_DATA" >/tmp/lease.json
export LOCKFILE_PATH=/tmp/lease.json

smith cf-login --lockfile=$LOCKFILE_PATH
cf create-space system
cf target -o "system" -s "system"

smith om -l "$LOCKFILE_PATH" >/tmp/bashly_om_env.sh
chmod +x /tmp/bashly_om_env.sh
source /tmp/bashly_om_env.sh

eval $(smith om -l "$LOCKFILE_PATH")
eval $(smith om -l "$LOCKFILE_PATH")

smith bosh -l "$LOCKFILE_PATH" >/tmp/bashly_bosh_env.sh
chmod +x /tmp/bashly_bosh_env.sh
source /tmp/bashly_bosh_env.sh

eval $(smith bosh -l "$LOCKFILE_PATH")
eval $(smith bosh -l "$LOCKFILE_PATH")

smith open --lockfile=$LOCKFILE_PATH
