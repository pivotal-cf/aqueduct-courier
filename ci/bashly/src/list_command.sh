if [[ ${args[--json]} == "1" ]]; then
	$SHEPHERD_BINARY_PATH list lease --json --wide --namespace=tpi-telemetry
else
	$SHEPHERD_BINARY_PATH list lease --wide --namespace=tpi-telemetry
fi
