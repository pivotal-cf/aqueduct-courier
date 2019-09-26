$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

./binary/telemetry-collector-windows-amd64.exe --version
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

./binary/telemetry-collector-windows-amd64.exe send --path collected-data\*.tar
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}
