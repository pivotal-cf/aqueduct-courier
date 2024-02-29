export ALL_ENVS_READY=true
export NOT_READY_ENVS=()
export ENVS_JSON=$(tpi_list_command --json)
export CI_ENVS=(production-jammy acceptance-jammy staging-jammy production-xenial acceptance-xenial staging-xenial)
mkdir -p "${PWD}/shepherd_envs"

##############
# The idea of this command is that it can be run continuously
# and it will maintain the necessary long-lived environments
# as well as installing the Telemetry Tile and uploading the
# secrets to Vault.

# TODO: have script run every 6 hours
##############

# Function that takes in a string and returns a boolean.
# Checks to see if if there is a JSON object within ENVS_JSON
# where the value of the "description" key is set to the value of
# the string we're processing from CI_ENVS
# FIXME: function is currently un-used
env_exists() {
	if shepherd list lease --json --wide --namespace=tpi-telemetry | jq -r --arg env "$1" 'any(.[]; .description == $env)' | grep -q true; then
		return 0
	else
		printf "%s\n" "$1"
		return 1
	fi

}

# Function that takes in a string and returns a boolean.
# Checks to see if environment is ready by checking for
# existance of lockfile.
env_ready() {
	if [[ $LOCKFILE_DATA == null ]]; then
		return 1
	else
		return 0
	fi
}

# Function to get Telemetry Tile guid
telemetry_installed() {
	export TELEMETRY_TILE_GUID=""
	if [ -f "$LOCKFILE_PATH" ]; then
		export TELEMETRY_TILE_GUID=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path /api/v0/deployed/products | jq -r '.[] | select(.type == "pivotal-telemetry-om").guid') || ""
		if [[ -z $TELEMETRY_TILE_GUID ]]; then
			return 1
		else
			return 0
		fi
	else
		return 1
	fi
}

check_all_envs_exist() {
	echo -e "Checking that all enivronments exist..."
	for my_env in "${CI_ENVS[@]}"; do
		# If env already exists, this is a no-op
		tpi_create_command "$my_env"
	done

	echo "*** All environments exist. ***"
}

# Will exit as soon as it finds an
# enviroment that isn't ready
# FIXME: function currently un-used
check_all_envs_ready() {
	echo -e "Checking that all enivronments are ready..."
	for my_env in "${CI_ENVS[@]}"; do
		tpi_get_command "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export TPI_ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)
		if ! env_ready; then
			printf "%s not ready. Please wait.\n" "$my_env"
			exit 0
		fi
	done

	echo "*** All environments ready. ***"
}

# Will exit as soon as it finds an
# environment that needs the Telemety
# Tile but doesn't have it installed.
# Will kick off installation of Tile.
check_telemetry_tile_installed() {
	for my_env in "${CI_ENVS[@]}"; do
		extract_env_details "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export TPI_ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)
		if [[ $TPI_ENV_TYPE != "staging" ]]; then
			if ! telemetry_installed; then
				printf "\nInstalling the Telemetry Tile"
				tpi_install_command "$my_env"
				# TODO: can this be removed (will we have the Vault secrets??)
				# exit 0
			else
				printf "\nTelemetry Tile is installed on $my_env"
			fi
		else
			printf "\nStaging Env: Telemetry Tile not necessary...\n"
		fi
	done

	echo -e "*** All (necessary) Telemetry Tiles installed. ***"
}

# Note: will continue to upload secrets even if they already exist
upload_secrets_to_vault() {
	echo -e "Uploading secrets to Vault..."
	for my_env in "${CI_ENVS[@]}"; do
		extract_env_details "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export TPI_ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)

		tpi_update_command "$my_env"
	done

	echo "*** Secrets uploaded to Vault. ***"
}

check_all_envs_exist
check_telemetry_tile_installed

if [[ $ALL_ENVS_READY == true ]]; then
	upload_secrets_to_vault
else
	echo -e "The following environments are not ready:"
	for not_ready_env in "${NOT_READY_ENVS[@]}"; do
		echo -e "$not_ready_env"
	done
fi
