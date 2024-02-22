export ENVS_JSON=$(tpi_list_command --json)
export CI_ENVS=(production-jammy acceptance-jammy staging-jammy production-xenial acceptance-xenial staging-xenial)
mkdir -p "${PWD}/shepherd_envs"

# Function that takes in a string and returns a boolean.
# Checks to see if if there is a JSON object within ENVS_JSON
# where the value of the "description" key is set to the value of
# the string we're processing from CI_ENVS
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
	export TELEMETRY_TILE_GUID=$(smith om --lockfile="$LOCKFILE_PATH" -- curl -s --path /api/v0/deployed/products | jq -r '.[] | select(.type == "pivotal-telemetry-om").guid') || ""
	if [[ -z $TELEMETRY_TILE_GUID ]]; then
		return 1
	else
		return 0
	fi
}

check_all_envs_exist() {
	echo -e "Checking that all enivronments exist..."
	for my_env in "${CI_ENVS[@]}"; do
		extract_env_details "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)

		if ! env_exists "$my_env"; then
			printf "No match for %s: creating new env...\n" "$my_env"
			tpi_create_command "$my_env"
			exit 0
		fi
	done

	echo "*** All environments exist. ***"
}

check_all_envs_ready() {
	echo -e "Checking that all enivronments are ready..."
	for my_env in "${CI_ENVS[@]}"; do
		tpi_get_command "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)
		if ! env_ready; then
			printf "%s not ready. Please wait.\n" "$my_env"
			exit 0
		fi
	done

	echo "*** All environments ready. ***"
}

check_telemetry_tile_installed() {
	echo -e "Checking that Telemetry Tile is installed..."
	for my_env in "${CI_ENVS[@]}"; do
		extract_env_details "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)
		if [[ $ENV_TYPE != "staging" ]]; then
			printf "\nNOT a staging env..."
			if ! telemetry_installed; then
				printf "\nInstalling the Telemetry Tile"
				tpi_install_command "$my_env"
				exit 0
			fi
		else
			printf "\nJust a staging env...\n"
		fi
	done

	echo -e "*** All (necessary) Telemetry Tiles installed. ***"
}

upload_secrets_to_vault() {
	echo -e "Uploading secrets to Vault..."
	for my_env in "${CI_ENVS[@]}"; do
		extract_env_details "$my_env"
		export LOCKFILE_PATH=${PWD}/shepherd_envs/$my_env-metadata.json
		export ENV_TYPE=$(echo "$my_env" | cut -d '-' -f 1)

		tpi_update_command "$my_env"
	done

	echo "*** Secrets uploaded to Vault. ***"
}

check_all_envs_exist
check_all_envs_ready
check_telemetry_tile_installed
upload_secrets_to_vault
