package opsmanager

import (
	"fmt"
	"io"

	"github.com/pkg/errors"

	"github.com/pivotal-cf/om/api"
)

const (
	PendingChangesExistsMessage   = "There are pending changes on this Operations Manager, please apply them or revert them."
	PendingChangesFailedMessage   = "Failed to retrieve pending change list from Operations Manager"
	DeployedProductsFailedMessage = "Failed to retrieve deployed products list from Operations Manager"
	RequestorFailureErrorFormat   = "Failed retrieving %s %s"

	OpsManagerName      = "ops_manager"
	DirectorProductType = "p-bosh"

	ResourcesDataType        = "resources"
	VmTypesDataType          = "vm_types"
	DiagnosticReportDataType = "diagnostic_report"
	DeployedProductsDataType = "deployed_products"
	InstallationsDataType    = "installations"
)

//go:generate counterfeiter . PendingChangesLister
type PendingChangesLister interface {
	List() (api.PendingChangesOutput, error)
}

//go:generate counterfeiter . DeployedProductsLister
type DeployedProductsLister interface {
	List() ([]api.DeployedProductOutput, error)
}

//go:generate counterfeiter . OmService
type OmService interface {
	ProductResources(guid string) (io.Reader, error)
	VmTypes() (io.Reader, error)
	DiagnosticReport() (io.Reader, error)
	DeployedProducts() (io.Reader, error)
	Installations() (io.Reader, error)
}

type dataRetriever func() (io.Reader, error)

type DataCollector struct {
	omService             OmService
	pendingChangesService PendingChangesLister
	deployProductsService DeployedProductsLister
}

func NewDataCollector(oms OmService, pcs PendingChangesLister, dps DeployedProductsLister) DataCollector {
	return DataCollector{
		omService:             oms,
		pendingChangesService: pcs,
		deployProductsService: dps,
	}
}

func (dc DataCollector) Collect() ([]Data, error) {
	pc, err := dc.pendingChangesService.List()
	if err != nil {
		return []Data{}, errors.Wrap(err, PendingChangesFailedMessage)
	}

	if len(pc.ChangeList) > 0 {
		return []Data{}, errors.New(PendingChangesExistsMessage)
	}

	pl, err := dc.deployProductsService.List()
	if err != nil {
		return []Data{}, errors.Wrap(err, DeployedProductsFailedMessage)
	}

	var d []Data

	d, err = appendRetrievedData(d, dc.omService.DeployedProducts, OpsManagerName, DeployedProductsDataType)
	if err != nil {
		return []Data{}, err
	}

	for _, product := range pl {
		if product.Type != DirectorProductType {
			d, err = appendRetrievedData(d, dc.productResourcesCaller(product.GUID), product.Type, ResourcesDataType)
		}
		if err != nil {
			return []Data{}, err
		}
	}

	d, err = appendRetrievedData(d, dc.omService.VmTypes, OpsManagerName, VmTypesDataType)
	if err != nil {
		return []Data{}, err
	}

	d, err = appendRetrievedData(d, dc.omService.DiagnosticReport, OpsManagerName, DiagnosticReportDataType)
	if err != nil {
		return []Data{}, err
	}

	d, err = appendRetrievedData(d, dc.omService.Installations, OpsManagerName, InstallationsDataType)
	if err != nil {
		return []Data{}, err
	}

	return d, nil
}

func (dc DataCollector) productResourcesCaller(guid string) dataRetriever {
	return func() (io.Reader, error) {
		return dc.omService.ProductResources(guid)
	}
}

func appendRetrievedData(d []Data, retriever dataRetriever, productType, dataType string) ([]Data, error) {
	output, err := retriever()
	if err != nil {
		return d, errors.Wrap(err, fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))
	}

	return append(d, NewData(output, productType, dataType)), nil
}
