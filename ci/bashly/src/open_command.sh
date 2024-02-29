# Set foundation based on command line arg or passed in name
if [[ ${args[foundation]} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

# Set lockfile path
LOCKFILE_PATH="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json"

# Open 1 or all (long-lived) foundations
if [[ $TMP_FOUNDATION_NAME == "all" ]]; then
	echo -e "FIXME"
	# CI_ENVS=(production-jammy acceptance-jammy staging-jammy production-xenial acceptance-xenial staging-xenial)

	# for my_env in "${CI_ENVS[@]}"; do
	# 	smith open --lockfile="${PWD}/shepherd_envs/$my_env-metadata.json"
	# done
else
	if [ -f "$LOCKFILE_PATH" ]; then
		smith open --lockfile="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json"
	else
		echo -e "No lockfile found for $TMP_FOUNDATION_NAME"
	fi
fi
