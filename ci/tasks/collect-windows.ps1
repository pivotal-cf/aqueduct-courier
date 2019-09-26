$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

./binary/telemetry-collector-windows-amd64.exe --version
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

./binary/telemetry-collector-windows-amd64.exe collect --output-dir ./collected-data
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}
