# Create correct environmnt if it doesn't already exist
if [[ ${args[foundation]:-} == "" ]]; then
	export ENV_DESCRIPTION=$1
else
	export ENV_DESCRIPTION=${args[foundation]}
fi

ENV_MATCHES=$($SHEPHERD_BINARY_PATH list lease --desc-search "$ENV_DESCRIPTION" --namespace tpi-telemetry --json)
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

		#$SHEPHERD_BINARY_PATH create lease --pool "tas-5_0" --duration 168h --namespace tpi-telemetry --description "$ENV_DESCRIPTION" --pool-namespace official
		# Use Custom Env for up to date Ops Man (supporting Core Consumption API)

		# FIXME: dynamically populate with latest Ops Man & TAS
		# FIXME: specify up to date stemcell
		# $SHEPHERD_BINARY_PATH create lease --template-namespace official --template-name gcp-tas-template --template-revision 2.1 --template-argument '{"configuration_folder": "3.0", "opsman_version": "3.0.30+LTS-T", "product_type": "srt*",  "tas_version": "6.0.4+LTS-T"}' --namespace tpi-telemetry --duration 168h --json --description "$ENV_DESCRIPTION"
		$SHEPHERD_BINARY_PATH create lease --pool "tas-6_0" --duration 168h --namespace tpi-telemetry --description "$ENV_DESCRIPTION" --pool-namespace official

		# Remove old metadata file
		rm -rf "${PWD}/shepherd_envs/$ENV_DESCRIPTION-metadata.json"

		# Remove old smith-data
		rm -rf "${PWD}/smith-data/$ENV_DESCRIPTION"
	fi

	# # Create TAS 2.11 if we need a xenial env
	# if [[ $ENV_DESCRIPTION == "production-xenial" || $ENV_DESCRIPTION == "acceptance-xenial" || $ENV_DESCRIPTION == "staging-xenial" ]]; then
	# 	echo -e "Creating $ENV_DESCRIPTION"

	# 	if [[ $ENV_DESCRIPTION == "production-xenial" || $ENV_DESCRIPTION == "acceptance-xenial" ]]; then
	# 		TELEMETRY_TILE_INSTALL_REQUIRED=true
	# 	fi

	# 	#$SHEPHERD_BINARY_PATH create lease --pool "tas-2_11" --duration 168h --namespace tpi-telemetry --description "$ENV_DESCRIPTION" --pool-namespace official
	# 	# Use Custom Env for up to date Ops Man (supporting Core Consumption API)

	# 	# FIXME: dynamically populate with latest Ops Man & TAS
	# 	# FIXME: specify up to date stemcell
	# 	$SHEPHERD_BINARY_PATH create lease --template-namespace official --template-name gcp-tas-template --template-revision 2.1 --template-argument '{"configuration_folder": "2.7", "opsman_version": "2.10.73", "product_type": "srt*",  "tas_version": "2.11.57"}' --namespace tpi-telemetry --duration 168h --json --description "$ENV_DESCRIPTION"

	# 	# Remove old metadata file
	# 	rm -rf "${PWD}/shepherd_envs/$ENV_DESCRIPTION-metadata.json"

	# 	# Remove old smith-data
	# 	rm -rf "${PWD}/smith-data/$ENV_DESCRIPTION"
	# fi

	if [[ $TELEMETRY_TILE_INSTALL_REQUIRED == "true" ]]; then
		echo -e "Install Telemetry Tile before updating Vault variables\n"
	fi
else
	echo -e "$ENV_DESCRIPTION already exists"
fi
