ensure_shepherd_login() {
	#export SHEPHERD_TOKEN=$(vault kv get -format=json /runway_concourse/tanzu-portfolio-insights/shepherd/tpi-telemetry | jq -r .data.secret)

	if ! $SHEPHERD_BINARY_PATH login user --json | jq -e '.login.User == "success"' &>/dev/null; then
		echo "Failed to login to shepherd" >&2
	fi
}
