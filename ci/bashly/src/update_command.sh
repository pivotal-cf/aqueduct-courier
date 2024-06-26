if [[ ${args[foundation]:-} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

tpi_get_command $TMP_FOUNDATION_NAME

update_vault() {
	ENV_DESCRIPTION=$TMP_FOUNDATION_NAME

	# Write the lockfile to a secret
	vault kv put runway_concourse/tanzu-portfolio-insights/toolsmiths/"${ENV_DESCRIPTION}"-lockfile @"${PWD}/shepherd_envs/$ENV_DESCRIPTION-metadata.json"

	# Do the updating
	if [[ $TPI_ENV_TYPE == "staging" || $(echo $TPI_ENV_TYPE | grep -q 'p-telemetry') || $(echo $TPI_ENV_TYPE | grep -q 'telemetry-release') ]]; then
		vault kv put runway_concourse/tanzu-portfolio-insights/toolsmiths/"${ENV_DESCRIPTION}" \
			env-name="${NAME}" \
			p-bosh-id-guid="${P_BOSH_ID}"
		echo -e "\nUpdated 2 Vault variables for ${ENV_DESCRIPTION}\n"
	else
		vault kv put runway_concourse/tanzu-portfolio-insights/toolsmiths/"${ENV_DESCRIPTION}" \
			cf-api-url="https://api.${SYS_DOMAIN}" \
			env-name="${NAME}" \
			gcp-project-id="${GCP_PROJECT_ID}" \
			gcp-zone="us-central1-f" \
			iaas_type="google" \
			ops-manager-hostname="${OPS_MANAGER_DNS}" \
			ops-manager-url="${OPS_MANAGER_URL}" \
			opsman-client-id="restricted_view_api_access" \
			opsman-instance-name="${NAME}-ops-manager" \
			opsman-password="${OPS_MANAGER_PASSWORD}" \
			opsman-uaa-client-secret="${UAA_CLIENT_SECRET}" \
			opsman-username="${OPS_MANAGER_USERNAME}" \
			p-bosh-id-guid="${P_BOSH_ID}" \
			telemetry-tile-guid="${TELEMETRY_TILE_GUID}" \
			telemetry-usage-service-password="${TELEMETRY_USAGE_SERVICE_PASSWORD}" \
			usage-service-client-id="usage_service" \
			usage-service-url="https://app-usage.${SYS_DOMAIN}"

		echo -e "\nUpdated 17 Vault variables for ${ENV_DESCRIPTION}\n"
	fi
}

echo -e "ALL_ENVS_READY: $ALL_ENVS_READY"
if [ "$ALL_ENVS_READY" = true ]; then
	update_vault
else
	echo -e "The following environments are not ready:"
	for not_ready_env in "${NOT_READY_ENVS[@]}"; do
		echo -e "$not_ready_env"
	done
fi
