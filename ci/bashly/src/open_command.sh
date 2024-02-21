if [[ ${args[foundation]} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

# FIXME: won't work if lockfile doesn't exist
smith open --lockfile="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json"
