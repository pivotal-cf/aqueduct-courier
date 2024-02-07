ensure_pivnet_login() {
  # Get Pivnet token
  export PIVNET_REFRESH_TOKEN=$(vault kv get -format=json /runway_concourse/tanzu-portfolio-insights/pivnet | jq -r '.data' | jq -r '.["refresh-token"]')

  # export PIVNET_ACCESS_TOKEN=$(curl -X POST https://network.tanzu.vmware.com/api/v2/authentication/access_tokens -d '{"refresh_token":$PIVNET_REFRESH_TOKEN}' | jq -r .access_token)
  # curl -X GET https://network.tanzu.vmware.com/api/v2/authentication -H "Authorization: Bearer $PIVNET_ACCESS_TOKEN"

  if ! pivnet login --api-token=$PIVNET_REFRESH_TOKEN &>/dev/null; then
    echo "Failed to login to pivnet" >&2
    echo "Attempting again..."
    if ! pivnet login --api-token=$PIVNET_REFRESH_TOKEN &>/dev/null; then
      echo "Failed to login to pivnet a 2nd time" >&2
      exit 0
    fi
  fi
}
