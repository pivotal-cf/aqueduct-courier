# Create correct environmnt if it doesn't already exist
if [[ ${args[foundation]} == "" ]]; then
	export ENV_DESCRIPTION=$1
else
	export ENV_DESCRIPTION=${args[foundation]}
fi

ENV_MATCHES=$(shepherd list lease --desc-search "$ENV_DESCRIPTION" --namespace tpi-telemetry --json)
array_length=$(jq '. | length' <<<"$ENV_MATCHES")
TELEMETRY_TILE_INSTALL_REQUIRED=false

if [ "$array_length" -eq 0 ]; then
	echo -e "$ENV_DESCRIPTION not found\n"

	# Create TAS 5 if we need a jammy env
	if [[ $ENV_DESCRIPTION == "production-jammy" || $ENV_DESCRIPTION == "acceptance-jammy" || $ENV_DESCRIPTION == "staging-jammy" ]]; then
		echo -e "Creating $ENV_DESCRIPTION"

		if [[ $ENV_DESCRIPTION == "production-jammy" || $ENV_DESCRIPTION == "acceptance-jammy" ]]; then
			TELEMETRY_TILE_INSTALL_REQUIRED=true
		fi

		shepherd create lease --pool "tas-5_0" --duration 168h --namespace tpi-telemetry --description "$ENV_DESCRIPTION" --pool-namespace official
	fi

	# Create TAS 2.13 if we need a xenial env
	if [[ $ENV_DESCRIPTION == "production-xenial" || $ENV_DESCRIPTION == "acceptance-xenial" || $ENV_DESCRIPTION == "staging-xenial" ]]; then
		echo -e "Creating $ENV_DESCRIPTION"

		if [[ $ENV_DESCRIPTION == "production-xenial" || $ENV_DESCRIPTION == "acceptance-xenial" ]]; then
			TELEMETRY_TILE_INSTALL_REQUIRED=true
		fi

		shepherd create lease --pool "tas-2_13" --duration 168h --namespace tpi-telemetry --description "$ENV_DESCRIPTION" --pool-namespace official
	fi

	if [[ $TELEMETRY_TILE_INSTALL_REQUIRED == "true" ]]; then
		echo -e "Install Telemetry Tile before updating Vault variables\n"
	fi

	# Remove old metadata file
	rm -rf "${PWD}/shepherd_envs/$ENV_DESCRIPTION-metadata.json"

	# Remove old smith-data
	rm -rf "${PWD}/smith-data/$ENV_DESCRIPTION"
else
	echo -e "$ENV_DESCRIPTION already exists\n"
fi
