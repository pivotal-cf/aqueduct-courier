package opsmanager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager/opsmanagerfakes"

	"github.com/pkg/errors"

	"strings"

	"fmt"
	"io"

	. "github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("DataCollector", func() {
	var (
		omService              *opsmanagerfakes.FakeOmService
		pendingChangesLister   *opsmanagerfakes.FakePendingChangesLister
		deployedProductsLister *opsmanagerfakes.FakeDeployedProductsLister

		dataCollector *DataCollector
	)

	BeforeEach(func() {
		omService = new(opsmanagerfakes.FakeOmService)
		pendingChangesLister = new(opsmanagerfakes.FakePendingChangesLister)
		deployedProductsLister = new(opsmanagerfakes.FakeDeployedProductsLister)

		dataCollector = NewDataCollector(DataCollectorBuilder{
			OmService:             omService,
			PendingChangesService: pendingChangesLister,
			DeployProductsService: deployedProductsLister,
		})
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

	It("returns an error when omService.DirectorProperties errors", func() {
		deployedProductsLister.ListReturns(
			[]api.DeployedProductOutput{
				{Type: DirectorProductType, GUID: "p-bosh-always-first"},
				{Type: "best-product-1", GUID: "p1-guid"},
			},
			nil,
		)
		omService.DirectorPropertiesReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, DirectorProductType, PropertiesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.VmTypes errors", func() {
		omService.VmTypesReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, OpsManagerName, VmTypesDataType, "Requesting things is hard")
	})

	It("returns an error when omService.DiagnosticReport errors", func() {
		omService.DiagnosticReportReturns(nil, errors.New("Requesting things is hard"))
		data, err := dataCollector.Collect()
		assertOmServiceFailure(data, err, OpsManagerName, DiagnosticReportDataType, "Requesting things is hard")
	})

	It("succeeds", func() {
		directorReader := strings.NewReader("director data")
		readers := []io.Reader{
			strings.NewReader("p1 data"),
			strings.NewReader("p2 data"),
		}
		vmTypesReader := strings.NewReader("vm_types data")
		diagnosticReportReader := strings.NewReader("diagnostic data")
		directorProduct := api.DeployedProductOutput{Type: DirectorProductType, GUID: "p-bosh-always-first"}
		deployedProducts := []api.DeployedProductOutput{
			{Type: "best-product-1", GUID: "p1-guid"},
			{Type: "best-product-2", GUID: "p2-guid"},
		}
		deployedProductsLister.ListReturns(append([]api.DeployedProductOutput{directorProduct}, deployedProducts...), nil)
		omService.DirectorPropertiesReturns(directorReader, nil)
		for i, r := range readers {
			omService.ProductResourcesReturnsOnCall(i, r, nil)
		}
		omService.VmTypesReturns(vmTypesReader, nil)
		omService.DiagnosticReportReturns(diagnosticReportReader, nil)

		data, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(ConsistOf(
			NewData(
				directorReader,
				directorProduct.Type,
				PropertiesDataType,
			),
			NewData(
				readers[0],
				deployedProducts[0].Type,
				ResourcesDataType,
			),
			NewData(
				readers[1],
				deployedProducts[1].Type,
				ResourcesDataType,
			),
			NewData(
				vmTypesReader,
				OpsManagerName,
				VmTypesDataType,
			),
			NewData(
				diagnosticReportReader,
				OpsManagerName,
				DiagnosticReportDataType,
			),
		))
	})

	It("succeeds if there are no deployed products", func() {
		data, err := dataCollector.Collect()
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(ConsistOf(
			NewData(nil, OpsManagerName, VmTypesDataType),
			NewData(nil, OpsManagerName, DiagnosticReportDataType),
		))
	})
})

func assertOmServiceFailure(d []Data, err error, productType, dataType, causeErrorMessage string) {
	Expect(d).To(BeEmpty())
	Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))))
	Expect(err).To(MatchError(ContainSubstring(causeErrorMessage)))
}
