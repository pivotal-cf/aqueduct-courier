tpi_get_command ${args[foundation]}




update_vault() {
	ENV_DESCRIPTION=${args[foundation]}

	# Write the lockfile to a secret
	vault kv put runway_concourse/tanzu-portfolio-insights/toolsmiths/"${ENV_DESCRIPTION}"-lockfile @"${PWD}/shepherd_envs/$ENV_DESCRIPTION-metadata.json"

	# Set path
	VAULT_PATH="${ENV_DESCRIPTION}"
	if [[ $ENV_STEMCELL == "jammy" ]]; then
		VAULT_PATH="${ENV_TYPE}"
	fi

	# Do the updating
	if [[ $ENV_TYPE == "staging" ]]; then
		vault kv put runway_concourse/tanzu-portfolio-insights/toolsmiths/"${VAULT_PATH}" \
			env-name="${NAME}" \
			p-bosh-id-guid="${P_BOSH_ID}"
		echo -e "\nUpdated 2 Vault variables for ${ENV_DESCRIPTION}\n"
	else
		vault kv put runway_concourse/tanzu-portfolio-insights/toolsmiths/"${VAULT_PATH}" \
			cf-api-url="https://api.sys.${NAME}.cf-app.com" \
			env-name="${NAME}" \
			gcp-project-id="${GCP_PROJECT_ID}" \
			gcp-zone="us-central1-f" \
			iaas_type="google" \
			ops-manager-hostname="pcf.${NAME}.cf-app.com" \
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
			usage-service-url="https://app-usage.sys.${NAME}.cf-app.com"

		echo -e "\nUpdated 17 Vault variables for ${ENV_DESCRIPTION}\n"
	fi
}




update_vault
