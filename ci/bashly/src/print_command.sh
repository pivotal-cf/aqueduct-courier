if [[ ${args[foundation]} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

tpi_get_command $TMP_FOUNDATION_NAME
ENV_DESCRIPTION=$TMP_FOUNDATION_NAME

echo -e "\n\n********** CLI COMMANDS: USERNAME / PASSWORD **********"

# CLI STRINGS
echo -e "\n** CEIP ONLY **"
mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/username-password/ceip"
echo "telemetry-collector collect \\
    --url ${OPS_MANAGER_URL} \\
    --username pivotalcf \\
    --password ${OPS_MANAGER_PASSWORD} \\
    --env-type development \\
    --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/ceip"

if [[ ${args[--collect]} ]]; then
	eval "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --username pivotalcf \\
        --password ${OPS_MANAGER_PASSWORD} \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/ceip"
fi

if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
	echo -e "\n\n** OPERATIONAL DATA ONLY - WITH USAGE **"
	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/username-password/operational-data-only-with-usage"
	echo -e "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --username pivotalcf \\
        --password ${OPS_MANAGER_PASSWORD} \\
        --usage-service-url https://app-usage.${SYS_DOMAIN} \\
        --usage-service-client-id usage_service \\
        --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
        --cf-api-url https://api.${SYS_DOMAIN} \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/operational-data-only-with-usage \\
        --operational-data-only"

	if [[ ${args[--collect]} ]]; then
		eval "telemetry-collector collect \\
            --url ${OPS_MANAGER_URL} \\
            --username pivotalcf \\
            --password ${OPS_MANAGER_PASSWORD} \\
            --usage-service-url https://app-usage.${SYS_DOMAIN} \\
            --usage-service-client-id usage_service \\
            --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
            --cf-api-url https://api.${SYS_DOMAIN} \\
            --env-type development \\
            --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/operational-data-only-with-usage \\
            --operational-data-only"
	fi
fi

echo -e "\n\n** OPERATIONAL DATA ONLY - NO USAGE **"
mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/username-password/operational-data-only-no-usage"
echo "telemetry-collector collect \\
    --url ${OPS_MANAGER_URL} \\
    --username pivotalcf \\
    --password ${OPS_MANAGER_PASSWORD} \\
    --env-type development \\
    --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/operational-data-only-no-usage \\
    --operational-data-only"

if [[ ${args[--collect]} ]]; then
	eval "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --username pivotalcf \\
        --password ${OPS_MANAGER_PASSWORD} \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/operational-data-only-no-usage \\
        --operational-data-only"
fi

if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
	echo -e "\n\n** ALL **"
	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/username-password/all"
	echo "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --username pivotalcf \\
        --password ${OPS_MANAGER_PASSWORD} \\
        --usage-service-url https://app-usage.${SYS_DOMAIN} \\
        --usage-service-client-id usage_service \\
        --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
        --cf-api-url https://api.${SYS_DOMAIN} \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/all"

	if [[ ${args[--collect]} ]]; then
		eval "telemetry-collector collect \\
            --url ${OPS_MANAGER_URL} \\
            --username pivotalcf \\
            --password ${OPS_MANAGER_PASSWORD} \\
            --usage-service-url https://app-usage.${SYS_DOMAIN} \\
            --usage-service-client-id usage_service \\
            --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
            --cf-api-url https://api.${SYS_DOMAIN} \\
            --env-type development \\
            --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/all"
	fi
fi

if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
	echo -e "\n\n********** CLI COMMANDS: CLIENT ID / CLIENT SECRET **********"
	# CLIENT ID / CLIENT SECRET

	echo -e "\n** CEIP ONLY **"
	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/client-id-client-secret/ceip"
	echo "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --client-id restricted_view_api_access \\
        --client-secret ${UAA_CLIENT_SECRET} \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/ceip"

	if [[ ${args[--collect]} ]]; then
		eval "telemetry-collector collect \\
            --url ${OPS_MANAGER_URL} \\
            --client-id restricted_view_api_access \\
            --client-secret ${UAA_CLIENT_SECRET} \\
            --env-type development \\
            --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/ceip"
	fi

	if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
		echo -e "\n\n** OPERATIONAL DATA ONLY - WITH USAGE **"
		mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/client-id-client-secret/operational-data-only-with-usage"
		echo -e "telemetry-collector collect \\
            --url ${OPS_MANAGER_URL} \\
            --client-id restricted_view_api_access \\
            --client-secret ${UAA_CLIENT_SECRET} \\
            --usage-service-url https://app-usage.${SYS_DOMAIN} \\
            --usage-service-client-id usage_service \\
            --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
            --cf-api-url https://api.${SYS_DOMAIN} \\
            --env-type development \\
            --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/operational-data-only-with-usage \\
            --operational-data-only"

		if [[ ${args[--collect]} ]]; then
			eval "telemetry-collector collect \\
                --url ${OPS_MANAGER_URL} \\
                --client-id restricted_view_api_access \\
                --client-secret ${UAA_CLIENT_SECRET} \\
                --usage-service-url https://app-usage.${SYS_DOMAIN} \\
                --usage-service-client-id usage_service \\
                --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
                --cf-api-url https://api.${SYS_DOMAIN} \\
                --env-type development \\
                --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/operational-data-only-with-usage \\
                --operational-data-only"
		fi
	fi

	echo -e "\n\n** OPERATIONAL DATA ONLY - NO USAGE **"
	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/client-id-client-secret/operational-data-only-no-usage"
	echo "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --client-id restricted_view_api_access \\
        --client-secret ${UAA_CLIENT_SECRET} \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/operational-data-only-no-usage \\
        --operational-data-only"

	if [[ ${args[--collect]} ]]; then
		eval "telemetry-collector collect \\
            --url ${OPS_MANAGER_URL} \\
            --client-id restricted_view_api_access \\
            --client-secret ${UAA_CLIENT_SECRET} \\
            --env-type development \\
            --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/operational-data-only-no-usage \\
            --operational-data-only"
	fi

	if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
		echo -e "\n\n** ALL **"
		mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/client-id-client-secret/all"
		echo -e "telemetry-collector collect \\
            --url ${OPS_MANAGER_URL} \\
            --client-id restricted_view_api_access \\
            --client-secret ${UAA_CLIENT_SECRET} \\
            --usage-service-url https://app-usage.${SYS_DOMAIN} \\
            --usage-service-client-id usage_service \\
            --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
            --cf-api-url https://api.${SYS_DOMAIN} \\
            --env-type development \\
            --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/all"

		if [[ ${args[--collect]} ]]; then
			eval "telemetry-collector collect \\
                --url ${OPS_MANAGER_URL} \\
                --client-id restricted_view_api_access \\
                --client-secret ${UAA_CLIENT_SECRET} \\
                --usage-service-url https://app-usage.${SYS_DOMAIN} \\
                --usage-service-client-id usage_service \\
                --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
                --cf-api-url https://api.${SYS_DOMAIN} \\
                --env-type development \\
                --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/all"
		fi
	fi
fi
