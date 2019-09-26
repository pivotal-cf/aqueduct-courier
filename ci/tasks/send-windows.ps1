$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

./binary/telemetry-collector-windows-amd64.exe --version
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

$TAR_FILE = (dir collected-data\*.tar).Name

./binary/telemetry-collector-windows-amd64.exe send --path collected-data\$TAR_FILE
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}
