package opsmanager_test

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/onsi/gomega/gbytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager/opsmanagerfakes"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"

	"github.com/pkg/errors"

	. "github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("DataCollector", func() {
	var (
		logger                 *log.Logger
		bufferedOutput         *gbytes.Buffer
		omService              *opsmanagerfakes.FakeOmService
		omURL                  string
		pendingChangesLister   *opsmanagerfakes.FakePendingChangesLister
		deployedProductsLister *opsmanagerfakes.FakeDeployedProductsLister

		dataCollector *DataCollector
	)

	BeforeEach(func() {
		bufferedOutput = gbytes.NewBuffer()
		logger = log.New(bufferedOutput, "", 0)
		omService = new(opsmanagerfakes.FakeOmService)
		omURL = "some-opsmanager-url"
		pendingChangesLister = new(opsmanagerfakes.FakePendingChangesLister)
		deployedProductsLister = new(opsmanagerfakes.FakeDeployedProductsLister)

		dataCollector = NewDataCollector(*logger, omService, omURL, pendingChangesLister, deployedProductsLister)
	})

	It("does not return an error if there are pending changes with an action other than unchanged", func() {
		nonEmptyPendingChanges := api.PendingChangesOutput{
			ChangeList: []api.ProductChange{
				{
					Action: "unchanged",
					GUID:   "some-guid",
				},
				{
					Action: "totally-changed",
					GUID:   "some-changed-guid",
				},
			},
		}
		pendingChangesLister.ListStagedPendingChangesReturns(nonEmptyPendingChanges, nil)

		data, foundationId, err := dataCollector.Collect()
		Expect(data).To(ConsistOf(
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.DeployedProductsDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.VmTypesDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.DiagnosticReportDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.InstallationsDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.CertificatesDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.CertificateAuthoritiesDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.PendingChangesDataType),
		))
		Expect(foundationId).To(BeEmpty())
		Expect(err).ToNot(HaveOccurred())
		Eventually(bufferedOutput).Should(gbytes.Say(fmt.Sprintf(PendingChangesExistsFormat, "")))
		Eventually(bufferedOutput).Should(gbytes.Say("some-changed-guid: totally-changed"))
	})

	It("returns an error if listing pending changes errors", func() {
		pendingChangesLister.ListStagedPendingChangesReturns(api.PendingChangesOutput{}, errors.New("Listing things is hard"))

		data, foundationId, err := dataCollector.Collect()
		Expect(data).To(BeEmpty())
		Expect(foundationId).To(BeEmpty())
		Expect(err).To(MatchError(ContainSubstring(PendingChangesFailedMessage)))
		Expect(err).To(MatchError(ContainSubstring("Listing things is hard")))
	})

	It("returns an error if listing deployed products errors", func() {
		deployedProductsLister.ListDeployedProductsReturns([]api.DeployedProductOutput{}, errors.New("Listing things is hard"))

		data, foundationId, err := dataCollector.Collect()
		Expect(data).To(BeEmpty())
		Expect(foundationId).To(BeEmpty())
		Expect(err).To(MatchError(ContainSubstring(DeployedProductsFailedMessage)))
		Expect(err).To(MatchError(ContainSubstring("Listing things is hard")))
	})

	It("returns an error when omService.ProductResources errors", func() {
		deployedProductsLister.ListDeployedProductsReturns(
			[]api.DeployedProductOutput{
				{Type: collector_tar.DirectorProductType, GUID: "p-bosh-always-first"},
				{Type: "best-product-1", GUID: "p1-guid"},
			},
			nil,
		)
		omService.ProductResourcesReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, "best-product-1", collector_tar.ResourcesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.ProductProperties errors", func() {
		deployedProductsLister.ListDeployedProductsReturns(
			[]api.DeployedProductOutput{
				{Type: collector_tar.DirectorProductType, GUID: "p-bosh-always-first"},
				{Type: "best-product-1", GUID: "p1-guid"},
			},
			nil,
		)
		omService.ProductPropertiesReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, "best-product-1", collector_tar.PropertiesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.VmTypes errors", func() {
		omService.VmTypesReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.VmTypesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.DiagnosticReport errors", func() {
		omService.DiagnosticReportReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.DiagnosticReportDataType, "Requesting things is hard")
	})

	It("returns an error when omService.DeployedProducts errors", func() {
		omService.DeployedProductsReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.DeployedProductsDataType, "Requesting things is hard")
	})

	It("returns an error when omService.Installations errors", func() {
		omService.InstallationsReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.InstallationsDataType, "Requesting things is hard")
	})

	It("returns an error when omService.Certificates errors", func() {
		omService.CertificatesReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.CertificatesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.CertificateAuthorities errors", func() {
		omService.CertificateAuthoritiesReturns(nil, errors.New("Requesting things is hard"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.CertificateAuthoritiesDataType, "Requesting things is hard")
	})

	It("succeeds", func() {
		resourcesReaders := []io.Reader{
			strings.NewReader("r1 data"),
			strings.NewReader("r2 data"),
		}
		propertiesReaders := []io.Reader{
			strings.NewReader("p1 data"),
			strings.NewReader("p2 data"),
		}
		vmTypesReader := strings.NewReader("vm_types data")
		diagnosticReportReader := strings.NewReader("diagnostic data")
		deployedProductsReader := strings.NewReader("deployed products data")
		installationsReader := strings.NewReader("installations data")
		certificatesReader := strings.NewReader("certificates data")
		certificateAuthoritiesReader := strings.NewReader("certificate authorities data")
		pendingChangesReader := strings.NewReader("pending_changes")
		directorProduct := api.DeployedProductOutput{Type: collector_tar.DirectorProductType, GUID: "p-bosh-always-first"}
		deployedProducts := []api.DeployedProductOutput{
			{Type: "best-product-1", GUID: "p1-guid"},
			{Type: "best-product-2", GUID: "p2-guid"},
		}
		deployedProductsLister.ListDeployedProductsReturns(append([]api.DeployedProductOutput{directorProduct}, deployedProducts...), nil)
		for i, r := range resourcesReaders {
			omService.ProductResourcesReturnsOnCall(i, r, nil)
		}
		for i, r := range propertiesReaders {
			omService.ProductPropertiesReturnsOnCall(i, r, nil)
		}
		omService.VmTypesReturns(vmTypesReader, nil)
		omService.DiagnosticReportReturns(diagnosticReportReader, nil)
		omService.DeployedProductsReturns(deployedProductsReader, nil)
		omService.InstallationsReturns(installationsReader, nil)
		omService.CertificatesReturns(certificatesReader, nil)
		omService.CertificateAuthoritiesReturns(certificateAuthoritiesReader, nil)
		omService.PendingChangesReturns(pendingChangesReader, nil)

		collectedData, foundationId, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(bufferedOutput).To(gbytes.Say("Collecting data from Operations Manager at some-opsmanager-url"))
		Expect(foundationId).To(Equal("p-bosh-always-first"))
		Expect(collectedData).To(ConsistOf(
			NewData(
				deployedProductsReader,
				collector_tar.OpsManagerProductType,
				collector_tar.DeployedProductsDataType,
			),
			NewData(
				resourcesReaders[0],
				deployedProducts[0].Type,
				collector_tar.ResourcesDataType,
			),
			NewData(
				resourcesReaders[1],
				deployedProducts[1].Type,
				collector_tar.ResourcesDataType,
			),
			NewData(
				propertiesReaders[0],
				deployedProducts[0].Type,
				collector_tar.PropertiesDataType,
			),
			NewData(
				propertiesReaders[1],
				deployedProducts[1].Type,
				collector_tar.PropertiesDataType,
			),
			NewData(
				vmTypesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.VmTypesDataType,
			),
			NewData(
				diagnosticReportReader,
				collector_tar.OpsManagerProductType,
				collector_tar.DiagnosticReportDataType,
			),
			NewData(
				installationsReader,
				collector_tar.OpsManagerProductType,
				collector_tar.InstallationsDataType,
			),
			NewData(
				certificatesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.CertificatesDataType,
			),
			NewData(
				certificateAuthoritiesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.CertificateAuthoritiesDataType,
			),
			NewData(
				pendingChangesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.PendingChangesDataType,
			),
		))
	})

	It("succeeds when there is a deployed product in a delete state", func() {
		//If the product is in a delete state, there will be no properties or resources, but there will be a deployed product
		deletePendingChanges := api.PendingChangesOutput{
			ChangeList: []api.ProductChange{
				{
					Action: "unchanged",
					GUID:   "p1-guid",
				},
				{
					Action: "unchanged",
					GUID:   "p2-guid",
				},
				{
					Action: "delete",
					GUID:   "p3-guid",
				},
			},
		}
		pendingChangesLister.ListStagedPendingChangesReturns(deletePendingChanges, nil)

		resourcesReaders := []io.Reader{
			strings.NewReader("r1 data"),
			strings.NewReader("r2 data"),
		}
		propertiesReaders := []io.Reader{
			strings.NewReader("p1 data"),
			strings.NewReader("p2 data"),
		}
		vmTypesReader := strings.NewReader("vm_types data")
		diagnosticReportReader := strings.NewReader("diagnostic data")
		deployedProductsReader := strings.NewReader("deployed products data")
		installationsReader := strings.NewReader("installations data")
		certificatesReader := strings.NewReader("certificates data")
		certificateAuthoritiesReader := strings.NewReader("certificate authorities data")
		pendingChangesReader := strings.NewReader("pending_changes")
		directorProduct := api.DeployedProductOutput{Type: collector_tar.DirectorProductType, GUID: "p-bosh-always-first"}
		deployedProducts := []api.DeployedProductOutput{
			{Type: "best-product-1", GUID: "p1-guid"},
			{Type: "best-product-2", GUID: "p2-guid"},
			{Type: "deleted-product", GUID: "p3-guid"},
		}
		deployedProductsLister.ListDeployedProductsReturns(append([]api.DeployedProductOutput{directorProduct}, deployedProducts...), nil)
		for i, r := range resourcesReaders {
			omService.ProductResourcesReturnsOnCall(i, r, nil)
		}
		for i, r := range propertiesReaders {
			omService.ProductPropertiesReturnsOnCall(i, r, nil)
		}
		omService.VmTypesReturns(vmTypesReader, nil)
		omService.DiagnosticReportReturns(diagnosticReportReader, nil)
		omService.DeployedProductsReturns(deployedProductsReader, nil)
		omService.InstallationsReturns(installationsReader, nil)
		omService.CertificatesReturns(certificatesReader, nil)
		omService.CertificateAuthoritiesReturns(certificateAuthoritiesReader, nil)
		omService.PendingChangesReturns(pendingChangesReader, nil)

		collectedData, foundationId, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(bufferedOutput).To(gbytes.Say("Collecting data from Operations Manager at some-opsmanager-url"))
		Expect(foundationId).To(Equal("p-bosh-always-first"))
		Expect(collectedData).To(ConsistOf(
			NewData(
				deployedProductsReader,
				collector_tar.OpsManagerProductType,
				collector_tar.DeployedProductsDataType,
			),
			NewData(
				resourcesReaders[0],
				deployedProducts[0].Type,
				collector_tar.ResourcesDataType,
			),
			NewData(
				resourcesReaders[1],
				deployedProducts[1].Type,
				collector_tar.ResourcesDataType,
			),
			NewData(
				propertiesReaders[0],
				deployedProducts[0].Type,
				collector_tar.PropertiesDataType,
			),
			NewData(
				propertiesReaders[1],
				deployedProducts[1].Type,
				collector_tar.PropertiesDataType,
			),
			NewData(
				vmTypesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.VmTypesDataType,
			),
			NewData(
				diagnosticReportReader,
				collector_tar.OpsManagerProductType,
				collector_tar.DiagnosticReportDataType,
			),
			NewData(
				installationsReader,
				collector_tar.OpsManagerProductType,
				collector_tar.InstallationsDataType,
			),
			NewData(
				certificatesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.CertificatesDataType,
			),
			NewData(
				certificateAuthoritiesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.CertificateAuthoritiesDataType,
			),
			NewData(
				pendingChangesReader,
				collector_tar.OpsManagerProductType,
				collector_tar.PendingChangesDataType,
			),
		))
	})

	It("succeeds if there are no deployed products", func() {
		collectedData, foundationId, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(foundationId).To(Equal(""))
		Expect(collectedData).To(ConsistOf(
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.DeployedProductsDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.VmTypesDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.DiagnosticReportDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.InstallationsDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.CertificatesDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.CertificateAuthoritiesDataType),
			NewData(nil, collector_tar.OpsManagerProductType, collector_tar.PendingChangesDataType),
		))
	})

	It("returns an error when omService.PendingChanges errors", func() {
		omService.PendingChangesReturns(nil, errors.New("I broke when detecting stuff I should have detected"))
		collectedData, foundationId, err := dataCollector.Collect()
		assertOmServiceFailure(collectedData, foundationId, err, collector_tar.OpsManagerProductType, collector_tar.PendingChangesDataType, "I broke when detecting stuff I should have detected")
	})
})

func assertOmServiceFailure(d []Data, foundationId string, err error, productType, dataType, causeErrorMessage string) {
	Expect(d).To(BeEmpty())
	Expect(foundationId).To(BeEmpty())
	Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))))
	Expect(err).To(MatchError(ContainSubstring(causeErrorMessage)))
}
