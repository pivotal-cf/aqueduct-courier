$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

$env:GOPATH = "$PWD/go"
$env:PATH += ";${env:GOPATH}/bin"

go version
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

go get github.com/onsi/ginkgo/ginkgo
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}

ginkgo -failOnPending -race -randomizeAllSpecs -randomizeSuites -r -nodes=4 go/src/github.com/pivotal-cf/aqueduct-courier
if ($LastExitCode -ne 0) {
  Exit $LastExitCode
}
