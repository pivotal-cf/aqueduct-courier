tpi_get_command ${args[foundation]}
ENV_DESCRIPTION=${args[foundation]}

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

if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
	echo -e "\n\n** OPERATIONAL DATA ONLY - WITH USAGE **"
	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/username-password/operational-data-only-with-usage"
	echo -e "telemetry-collector collect \\
    --url ${OPS_MANAGER_URL} \\
    --username pivotalcf \\
    --password ${OPS_MANAGER_PASSWORD} \\
    --usage-service-url https://app-usage.sys.${ENV_DESCRIPTION}.cf-app.com \\
    --usage-service-client-id usage_service \\
    --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
    --cf-api-url https://api.sys.${ENV_DESCRIPTION}.cf-app.com \
    --env-type development \\
    --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/operational-data-only-with-usage \\
    --operational-data-only"
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

if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
	echo -e "\n\n** ALL **"
	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/username-password/all"
	echo "telemetry-collector collect \\
    --url ${OPS_MANAGER_URL} \\
    --username pivotalcf \\
    --password ${OPS_MANAGER_PASSWORD} \\
    --usage-service-url https://app-usage.sys.${ENV_DESCRIPTION}.cf-app.com \\
    --usage-service-client-id usage_service \\
    --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
    --cf-api-url https://api.sys.${ENV_DESCRIPTION}.cf-app.com \\
    --env-type development \\
    --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/username-password/all"
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

    if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
    	echo -e "\n\n** OPERATIONAL DATA ONLY - WITH USAGE **"
    	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/client-id-client-secret/operational-data-only-with-usage"
    	echo -e "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --client-id restricted_view_api_access \\
        --client-secret ${UAA_CLIENT_SECRET} \\
        --usage-service-url https://app-usage.sys.${ENV_DESCRIPTION}.cf-app.com \\
        --usage-service-client-id usage_service \\
        --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
        --cf-api-url https://api.sys.${ENV_DESCRIPTION}.cf-app.com \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/client-id-client-secret/operational-data-only-with-usage \\
        --operational-data-only"
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

    if [ -n "$TELEMETRY_USAGE_SERVICE_PASSWORD" ]; then
    	echo -e "\n\n** ALL **"
    	mkdir -p "${PWD}/smith-data/$ENV_DESCRIPTION/client-id-client-secret/all"
    	echo -e "telemetry-collector collect \\
        --url ${OPS_MANAGER_URL} \\
        --client-id restricted_view_api_access \\
        --client-secret ${UAA_CLIENT_SECRET} \\
        --usage-service-url https://app-usage.sys.${ENV_DESCRIPTION}.cf-app.com \\
        --usage-service-client-id usage_service \\
        --usage-service-client-secret ${TELEMETRY_USAGE_SERVICE_PASSWORD} \\
        --cf-api-url https://api.sys.${ENV_DESCRIPTION}.cf-app.com \\
        --env-type development \\
        --output-dir ${PWD}/smith-data/${ENV_DESCRIPTION}/all"
    fi
fi
