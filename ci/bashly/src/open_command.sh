# Set foundation based on command line arg or passed in name
if [[ ${args[foundation]} == "" ]]; then
	export TMP_FOUNDATION_NAME=$1
else
	export TMP_FOUNDATION_NAME=${args[foundation]}
fi

# Set lockfile path
LOCKFILE_PATH="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json"

if [ -f "$LOCKFILE_PATH" ]; then
	smith open --lockfile="${PWD}/shepherd_envs/$TMP_FOUNDATION_NAME-metadata.json"
else
	echo -e "No lockfile found for $TMP_FOUNDATION_NAME"
fi
