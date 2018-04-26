package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestAqueductCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var aqueductBinaryPath string

var _ = BeforeSuite(func() {
	var err error
	aqueductBinaryPath, err = gexec.Build("github.com/pivotal-cf/aqueduct-courier")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
