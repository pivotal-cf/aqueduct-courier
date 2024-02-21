ensure_pivnet_login

if [[ ${args[foundation]} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

tpi_get_command $TMP_FOUNDATION_NAME

LOCKFILE_PATH="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json"

# Check that this foundation doesn't already have the Telemetry Tile installed
export TELEMETRY_TILE_GUID=$(smith om --lockfile="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json" -- curl -s --path /api/v0/deployed/products | jq -r '.[] | select(.type == "pivotal-telemetry-om").guid') || ""
if [[ -n $TELEMETRY_TILE_GUID ]]; then
	echo -e "\nThe Telemetry Tile is already installed\n"
	exit 0
fi

smith cf-login --lockfile="$LOCKFILE_PATH"
eval $(smith om -l "$LOCKFILE_PATH")
eval $(smith bosh -l "$LOCKFILE_PATH")

# Enable the Usage Service Errand
export TAS_TILE_GUID=$(om curl --silent --path /api/v0/deployed/products | jq -r '.[] | select(.type == "cf").guid')
om curl --silent --path "/api/v0/staged/products/$TAS_TILE_GUID/errands" -x PUT -d '{"errands":[{"name":"push-usage-service","post_deploy":true}]}'

# errand_exists=$(bosh errands --json | jq -r '.["Tables"]' | jq -r '.[0]' | jq -r '.["Rows"]' | jq 'any(.[]; .name == "push-usage-service")')
# CF_DEPLOYMENT="$(bosh deployments --json | jq -r '.Tables[] | select(.Content="deployments") | .Rows[] | select(.name | test("^cf-")) | .name')"
# bosh run-errand -d "$CF_DEPLOYMENT" push-usage-service

# Get details of latest Telemetry Tile
# PRODUCT_NAME="pivotal-telemetry-om"
# TILE_INFO_JSON=$(pivnet releases --format=json -p "$PRODUCT_NAME" -l 1)
# TILE_VERSION=$(echo $TILE_INFO_JSON | jq -r '.[0].version')

# FIXME: this doesn't work for Xenial environments until
# they switch to bundling our 1.3.x line by default. For
# now, manually upload the 1.3.x tile and re-run this command.
TILE_FULL_VERSION=$(om products --format=json | jq '.[] | select(.name == "pivotal-telemetry-om")' | jq -r .available | jq -r '.[0]')

# Download latest Telemetry Tile
# rm -rf /tmp/tile
# mkdir /tmp/tile
# pivnet download-product-files --accept-eula --download-dir=/tmp/tile -p "$PRODUCT_NAME" -r "$TILE_VERSION" -g "*.pivotal"

network="$(smith read --lockfile=$LOCKFILE_PATH | jq -r .ert_subnet)"
az="us-central1-f"

function retry {
	local retries=$1
	local count=0
	shift

	until "$@"; do
		exit=$?
		count=$((count + 1))
		if [ $count -lt $retries ]; then
			echo "Attempt $count/$retries ended with exit $exit"
			# Need a short pause between smith CLI executions or it fails unexpectedly.
			sleep 30
		else
			echo "Attempted $count/$retries times and failed."
			return $exit
		fi
	done
	return 0
}

export LOADER_API_KEY=$(vault kv get -format=json /runway_concourse/tanzu-portfolio-insights/aqueduct-loader | jq -r '.data' | jq -r '.["production-loader-api-keys"]' | jq -r '.["Telemetry and Insights"]' | jq -r '.[0]')
echo -e "Loader API Key:\t$LOADER_API_KEY"

# echo "Uploading tile..."
# retry 5 smith om --lockfile=$LOCKFILE_PATH -- upload-product --product "/tmp/tile/pivotal-telemetry-om-$TILE_VERSION.pivotal"

echo "Staging tile..."
retry 5 om stage-product \
	--product-name "pivotal-telemetry-om" \
	--product-version "$TILE_FULL_VERSION"

FOUNDATION_NICKNAME=$(cat $LOCKFILE_PATH | jq .name)
TILE_ENV_TYPE=$ENV_TYPE
if [[ $ENV_TYPE == "acceptance" ]]; then
	TILE_ENV_TYPE="qa"
	FOUNDATION_NICKNAME="best-acceptance-env"
fi

echo "Configuring tile..."
retry 5 om configure-product \
	--config "$HOME/workspace/tile/p-telemetry/ci/tasks/product-config.yml" \
	--var audit-mode="false" \
	--var env-type="$TILE_ENV_TYPE" \
	--var foundation-nickname="$FOUNDATION_NICKNAME" \
	--var loader-endpoint="https://telemetry.pivotal.io" \
	--var telemetry-api-key="$LOADER_API_KEY" \
	--var flush-interval=10 \
	--var collector-cron-schedule="0 8 * * *" \
	--var network="$network" \
	--var az="$az" \
	--var ops-manager-timeout=30 \
	--var ops-manager-request-timeout=15

# echo "Uploading stemcell..."
# retry 5 smith om -- upload-stemcell --stemcell stemcell/*.tgz

# echo "Assigning stemcell to tile..."
# retry 5 smith om -- assign-stemcell \
#    --product pivotal-telemetry-om

smith open --lockfile="$LOCKFILE_PATH"

echo "Applying changes..."
rm -rf /tmp/bashly/logs
mkdir -p /tmp/bashly/logs
LOG_FILE=/tmp/bashly/logs/$TMP_FOUNDATION_NAME.log
touch $LOG_FILE
nohup om apply-changes --reattach >$LOG_FILE 2>&1 &

echo -e "Run the following command to watch the install:\n"
echo -e "tail -f ${LOG_FILE}"
