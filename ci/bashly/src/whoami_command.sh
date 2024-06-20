# FIXME: only works for environments with TAS, won't detect if environment has been reaped
export SYS_DOMAIN=$(cf api | grep 'API endpoint' | awk '{print $3}' | cut -d'/' -f3 | sed 's/^api\.//') || ""

if [[ -z $SYS_DOMAIN ]]; then
	echo -e "No foundations targeted"
	exit 0
fi

if [ $(cat shepherd_envs/acceptance-jammy-metadata.json | jq -r .sys_domain) = $SYS_DOMAIN ]; then
	echo -e "acceptance-jammy"
	exit 0
fi

if [ $(cat shepherd_envs/production-jammy-metadata.json | jq -r .sys_domain) = $SYS_DOMAIN ]; then
	echo -e "production-jammy"
	exit 0
fi

if [ $(cat shepherd_envs/staging-jammy-metadata.json | jq -r .sys_domain) = $SYS_DOMAIN ]; then
	echo -e "staging-jammy"
	exit 0
fi

echo -e "Targeted foundation does not match a Shepherd Env"
