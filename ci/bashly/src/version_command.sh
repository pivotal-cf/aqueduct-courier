OM_VERSION=${args[tag]}

cd ~/workspace/ops-manager
git pull ~/workspace/ops-manager &>/dev/null
git checkout $OM_VERSION &>/dev/null
echo -e "\n**************************"
tile_version=$(cat web/config/auto-versions.yml | yq .telemetry_tile.version)
echo -e "Ops Manager:\t$OM_VERSION"
echo -e "Telemetry:\tv$tile_version"
echo -e "**************************\n"
git switch - &>/dev/null
cd ~/workspace/tile/aqueduct-courier/ci/bashly
