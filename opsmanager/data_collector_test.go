package opsmanager_test

import (
	"fmt"
	"io"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager/opsmanagerfakes"

	"github.com/pkg/errors"

	. "github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("DataCollector", func() {
	var (
		omService              *opsmanagerfakes.FakeOmService
		pendingChangesLister   *opsmanagerfakes.FakePendingChangesLister
		deployedProductsLister *opsmanagerfakes.FakeDeployedProductsLister

		dataCollector DataCollector
	)

	BeforeEach(func() {
		omService = new(opsmanagerfakes.FakeOmService)
		pendingChangesLister = new(opsmanagerfakes.FakePendingChangesLister)
		deployedProductsLister = new(opsmanagerfakes.FakeDeployedProductsLister)

		dataCollector = NewDataCollector(omService, pendingChangesLister, deployedProductsLister)
	})

	It("returns an error if there are pending changes", func() {
		nonEmptyPendingChanges := api.PendingChangesOutput{ChangeList: []api.ProductChange{{}}}
		pendingChangesLister.ListReturns(nonEmptyPendingChanges, nil)

		data, err := dataCollector.Collect()
		Expect(data).To(BeEmpty())
		Expect(err).To(MatchError(PendingChangesExistsMessage))
	})

	It("returns an error if listing pending changes errors", func() {
		pendingChangesLister.ListReturns(api.PendingChangesOutput{}, errors.New("Listing things is hard"))

		data, err := dataCollector.Collect()
		Expect(data).To(BeEmpty())
		Expect(err).To(MatchError(ContainSubstring(PendingChangesFailedMessage)))
		Expect(err).To(MatchError(ContainSubstring("Listing things is hard")))
	})

	It("returns an error if listing deployed products errors", func() {
		deployedProductsLister.ListReturns([]api.DeployedProductOutput{}, errors.New("Listing things is hard"))

		data, err := dataCollector.Collect()
		Expect(data).To(BeEmpty())
		Expect(err).To(MatchError(ContainSubstring(DeployedProductsFailedMessage)))
		Expect(err).To(MatchError(ContainSubstring("Listing things is hard")))
	})

	It("returns an error when omService.ProductResources errors", func() {
		deployedProductsLister.ListReturns(
			[]api.DeployedProductOutput{
				{Type: DirectorProductType, GUID: "p-bosh-always-first"},
				{Type: "best-product-1", GUID: "p1-guid"},
			},
			nil,
		)
		omService.ProductResourcesReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, "best-product-1", ResourcesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.ProductProperties errors", func() {
		deployedProductsLister.ListReturns(
			[]api.DeployedProductOutput{
				{Type: DirectorProductType, GUID: "p-bosh-always-first"},
				{Type: "best-product-1", GUID: "p1-guid"},
			},
			nil,
		)
		omService.ProductPropertiesReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, "best-product-1", PropertiesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.VmTypes errors", func() {
		omService.VmTypesReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, OpsManagerProductType, VmTypesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.DiagnosticReport errors", func() {
		omService.DiagnosticReportReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, OpsManagerProductType, DiagnosticReportDataType, "Requesting things is hard")
	})

	It("returns an error when omService.DeployedProducts errors", func() {
		omService.DeployedProductsReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, OpsManagerProductType, DeployedProductsDataType, "Requesting things is hard")
	})

	It("returns an error when omService.Installations errors", func() {
		omService.InstallationsReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, OpsManagerProductType, InstallationsDataType, "Requesting things is hard")
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
		directorProduct := api.DeployedProductOutput{Type: DirectorProductType, GUID: "p-bosh-always-first"}
		deployedProducts := []api.DeployedProductOutput{
			{Type: "best-product-1", GUID: "p1-guid"},
			{Type: "best-product-2", GUID: "p2-guid"},
		}
		deployedProductsLister.ListReturns(append([]api.DeployedProductOutput{directorProduct}, deployedProducts...), nil)
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

		data, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(ConsistOf(
			NewData(
				deployedProductsReader,
				OpsManagerProductType,
				DeployedProductsDataType,
			),
			NewData(
				resourcesReaders[0],
				deployedProducts[0].Type,
				ResourcesDataType,
			),
			NewData(
				resourcesReaders[1],
				deployedProducts[1].Type,
				ResourcesDataType,
			),
			NewData(
				propertiesReaders[0],
				deployedProducts[0].Type,
				PropertiesDataType,
			),
			NewData(
				propertiesReaders[1],
				deployedProducts[1].Type,
				PropertiesDataType,
			),
			NewData(
				vmTypesReader,
				OpsManagerProductType,
				VmTypesDataType,
			),
			NewData(
				diagnosticReportReader,
				OpsManagerProductType,
				DiagnosticReportDataType,
			),
			NewData(
				installationsReader,
				OpsManagerProductType,
				InstallationsDataType,
			),
		))
	})

	It("succeeds if there are no deployed products", func() {
		data, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(ConsistOf(
			NewData(nil, OpsManagerProductType, DeployedProductsDataType),
			NewData(nil, OpsManagerProductType, VmTypesDataType),
			NewData(nil, OpsManagerProductType, DiagnosticReportDataType),
			NewData(nil, OpsManagerProductType, InstallationsDataType),
		))
	})
})

func assertOmServiceFailure(d []Data, err error, productType, dataType, causeErrorMessage string) {
	Expect(d).To(BeEmpty())
	Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))))
	Expect(err).To(MatchError(ContainSubstring(causeErrorMessage)))
}
