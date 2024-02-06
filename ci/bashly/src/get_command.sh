extract_env_details ${args[foundation]}




mkdir -p "${PWD}/shepherd_envs"
LOCKFILE_PATH=${PWD}/shepherd_envs/${args[foundation]}-metadata.json
if [[ $LOCKFILE_DATA == null ]]; then
	echo -e "\n*** Env not yet ready; try again later. ***\n"
	exit 0
else
	echo -e "Writing lockfile..."
	echo "$LOCKFILE_DATA" >"$LOCKFILE_PATH"
fi

# Source envs
echo -e "Targeting ${args[foundation]}..."
eval "$(smith om --lockfile="$LOCKFILE_PATH")"
smith cf-login --lockfile="$LOCKFILE_PATH"




######################
### FETCH ENV VARS ###
######################
echo -e "\nGetting env variables...\n"

# Function to get ops_manager password
fetch_ops_manager_pw() {
	export OPS_MANAGER_PASSWORD=$(smith read --lockfile="$LOCKFILE_PATH" | jq -r .ops_manager.password)
	echo -e "OPS_MANAGER_PASSWORD:\t\t\t$OPS_MANAGER_PASSWORD"
}

# Function to get ops_manager username
fetch_ops_manager_username() {
	export OPS_MANAGER_USERNAME=$(smith read --lockfile="$LOCKFILE_PATH" | jq -r .ops_manager.username)
	echo -e "OPS_MANAGER_USERNAME:\t\t\t$OPS_MANAGER_USERNAME"
}

# Function to get ops_manager url
fetch_ops_manager_url() {
	export OPS_MANAGER_URL=$(smith read --lockfile="$LOCKFILE_PATH" | jq -r .ops_manager.url)
	echo -e "OPS_MANAGER_URL:\t\t\t$OPS_MANAGER_URL"
}

# Function to get p_bosh_id
fetch_p_bosh_id() {
	export P_BOSH_ID=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path=/api/v0/deployed/products | jq -r ".[].guid" | grep bosh)
	echo -e "P_BOSH_ID:\t\t\t\t$P_BOSH_ID"
}

# Function to get Telemetry Tile guid
fetch_telemetry_tile_guid() {
	if [[ $ENV_TYPE != "staging" ]]; then
		export TELEMETRY_TILE_GUID=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path /api/v0/deployed/products | jq -r '.[] | select(.type == "pivotal-telemetry-om").guid') || ""

		if [[ -z "$TELEMETRY_TILE_GUID" ]]; then
			echo -e "\n*** YOU MUST INSTALL THE TELEMETRY TILE ***\n"
			exit 0
		fi

		echo -e "TELEMETRY_TILE_GUID:\t\t\t$TELEMETRY_TILE_GUID"
	fi
}

# Function to get uaa_client_secret
fetch_uaa_client_secret() {
	if [[ $ENV_TYPE != "staging" ]]; then
		export UAA_CLIENT_SECRET=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path /api/v0/deployed/products/"${TELEMETRY_TILE_GUID}"/manifest | jq -r '.instance_groups[] | select(.name == "telemetry-centralizer").jobs[] | select(.name == "telemetry-collector").properties.opsmanager.auth.uaa_client_secret')
		echo -e "UAA_CLIENT_SECRET:\t\t\t$UAA_CLIENT_SECRET"
	fi
}

# Function to get cf_guid
fetch_cf_guid() {
	export CF_GUID=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path /api/v0/deployed/products | jq -r '.[] | select(.type == "cf").guid')
	if [ -n "$CF_GUID" ]; then
		echo -e "CF_GUID:\t\t\t\t$CF_GUID"
	else
		echo -e "CF_GUID:"
	fi
}

# Function to get GPC project id
fetch_gcp_project_id() {
	#GCP_PROJECT_ID=$(smith om --lockfile=$LOCKFILE_PATH -- curl -s --path /api/v0/staged/director/iaas_configurations | jq -r '.iaas_configurations[0].project')
	export GCP_PROJECT_ID=$(smith read --lockfile="$LOCKFILE_PATH" | jq -r .project)
	echo -e "GCP_PROJECT_ID:\t\t\t\t$GCP_PROJECT_ID"
}

# Function to make an API call using the cf_guid and print the telemetry_usage_service_password
fetch_telemetry_usage_service_password() {
	export TELEMETRY_USAGE_SERVICE_PASSWORD=""

	if [[ $ENV_TYPE != "staging" ]]; then
		if [ -n "$CF_GUID" ]; then
			TELEMETRY_USAGE_SERVICE_PASSWORD=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path /api/v0/deployed/products/"${CF_GUID}"/credentials/.uaa.usage_service_client_credentials | jq -r .credential.value.password)
			echo -e "TELEMETRY_USAGE_SERVICE_PASSWORD:\t$TELEMETRY_USAGE_SERVICE_PASSWORD"
		else
			echo -e "\n*** YOU MUST INSTALL THE TELEMETRY TILE ***\n"
			exit 0
		fi
	else
		echo -e "TELEMETRY_USAGE_SERVICE_PASSWORD:"
	fi
}

# Function to get name
fetch_name() {
	export NAME=$(smith read --lockfile="$LOCKFILE_PATH" | jq -r .name)
	echo -e "NAME:\t\t\t\t\t$NAME"
}




fetch_ops_manager_pw
fetch_ops_manager_username
fetch_ops_manager_url
fetch_p_bosh_id
fetch_telemetry_tile_guid
fetch_uaa_client_secret
fetch_cf_guid
fetch_telemetry_usage_service_password
fetch_gcp_project_id
fetch_name
