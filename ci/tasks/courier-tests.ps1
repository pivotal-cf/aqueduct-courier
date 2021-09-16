$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

$env:GOPATH = "$PWD/go"
$env:PATH += ";${env:GOPATH}/bin"

go version
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

go install github.com/onsi/ginkgo/ginkgo@latest
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

pushd go/src/github.com/pivotal-cf/aqueduct-courier
  ginkgo -failOnPending -race -randomizeAllSpecs -randomizeSuites -r .
popd
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}
