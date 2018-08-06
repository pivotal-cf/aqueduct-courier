package integration

import (
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"fmt"
)

func TestAqueductCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(45 * time.Second)
	RunSpecs(t, "Integration Suite")
}

var (
	aqueductBinaryPath string
	testVersion        = "0.0.1-test-version"
)

var _ = BeforeSuite(func() {
	var err error
	aqueductBinaryPath, err = gexec.Build(
		"github.com/pivotal-cf/aqueduct-courier",
		"-ldflags",
		fmt.Sprintf("-X github.com/pivotal-cf/aqueduct-courier/cmd.version=%s", testVersion),
	)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func escapeWindowsPathRegex(path string) string {
	return strings.Replace(path, `\`, `\\`, -1)
}
