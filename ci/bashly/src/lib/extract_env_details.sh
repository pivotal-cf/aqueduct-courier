extract_env_details() {
	# SET ENV VARIABLES
	export TPI_ENV_TYPE=$(echo "$1" | cut -d '-' -f 1)
	export ENV_STEMCELL=$(echo "$1" | cut -d '-' -f 2)

	if [[ $(echo "$1" | cut -d '-' -f 3) == "test" ]]; then
		TPI_ENV_TYPE="development"
		ENV_STEMCELL="jammy"
	fi

	# GET ARRAY OF MATCHES ENVIRONMENTS
	env_array=$($SHEPHERD_BINARY_PATH list lease --namespace tpi-telemetry --desc-search "$1" --json)

	if [[ $(echo "$env_array" | jq '. | length') == 0 ]]; then
		echo -e "\nFoundation $1 does not exist\n"
		exit 1
	else
		# EXTRACT IDENTIFIER
		export ENV_IDENTIFIER=$(echo "$env_array" | jq -r '.[0].identifier')

		# GET LEASE
		echo -e "\nGetting $1 lease..."
		export ENV_LEASE=$($SHEPHERD_BINARY_PATH get lease "$ENV_IDENTIFIER" --namespace tpi-telemetry --json)

		# CHECK STATUS
		if [[ $(echo "$ENV_LEASE" | jq -r '.status') == "CREATING" ]]; then
			echo -e "\n*** Env still being created; try again later. ***\n"
			exit 1
		fi

		export LOCKFILE_DATA=$(echo "$ENV_LEASE" | jq -r .output)
	fi
}
