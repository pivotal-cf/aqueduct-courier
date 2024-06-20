if [[ ${args[foundation]:-} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

tpi_get_command $TMP_FOUNDATION_NAME

renew_foundation() {
	$SHEPHERD_BINARY_PATH update lease "$ENV_IDENTIFIER" --namespace tpi-telemetry --expire-in 168h

	echo -e "$TMP_FOUNDATION_NAME lease extended for 2 weeks"
}

renew_foundation
