package opsmanager

import (
	"github.com/pkg/errors"

	"io"

	"fmt"

	"github.com/pivotal-cf/om/api"
)

const (
	PendingChangesExistsMessage   = "There are pending changes on this Ops Manager, please apply them or revert them."
	PendingChangesFailedMessage   = "Failed to retrieve pending change list from Ops Manager"
	DeployedProductsFailedMessage = "Failed to retrieve deployed products list from Ops Manager"
	RequestorFailureErrorFormat   = "Failed retrieving %s %s"

	OpsManagerName      = "ops_manager"
	DirectorProductType = "p-bosh"

	PropertiesDataType       = "properties"
	ResourcesDataType        = "resources"
	VmTypesDataType          = "vm_types"
	DiagnosticReportDataType = "diagnostic_report"
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
	DirectorProperties() (io.Reader, error)
	VmTypes() (io.Reader, error)
	DiagnosticReport() (io.Reader, error)
}

type DataRetriever func() (io.Reader, error)

type AqueductData struct {
	Data io.Reader
	Name string
	Type string
}

type DataCollector struct {
	omService             OmService
	requestService        Requestor
	pendingChangesService PendingChangesLister
	deployProductsService DeployedProductsLister
}

type DataCollectorBuilder struct {
	OmService             OmService
	RequestService        Requestor
	PendingChangesService PendingChangesLister
	DeployProductsService DeployedProductsLister
}

func NewDataCollector(builder DataCollectorBuilder) *DataCollector {
	return &DataCollector{
		omService:             builder.OmService,
		requestService:        builder.RequestService,
		pendingChangesService: builder.PendingChangesService,
		deployProductsService: builder.DeployProductsService,
	}
}

func (dc *DataCollector) Collect() ([]AqueductData, error) {
	pc, err := dc.pendingChangesService.List()
	if err != nil {
		return []AqueductData{}, errors.Wrap(err, PendingChangesFailedMessage)
	}

	if len(pc.ChangeList) > 0 {
		return []AqueductData{}, errors.New(PendingChangesExistsMessage)
	}

	pl, err := dc.deployProductsService.List()
	if err != nil {
		return []AqueductData{}, errors.Wrap(err, DeployedProductsFailedMessage)
	}

	var data []AqueductData

	for _, product := range pl {
		if product.Type == DirectorProductType {
			data, err = appendRetrievedData(data, dc.omService.DirectorProperties, product.Type, PropertiesDataType)
		} else {
			data, err = appendRetrievedData(data, dc.productResourcesCaller(product.GUID), product.Type, ResourcesDataType)
		}
		if err != nil {
			return []AqueductData{}, err
		}
	}

	data, err = appendRetrievedData(data, dc.omService.VmTypes, OpsManagerName, VmTypesDataType)
	if err != nil {
		return []AqueductData{}, err
	}

	data, err = appendRetrievedData(data, dc.omService.DiagnosticReport, OpsManagerName, DiagnosticReportDataType)
	if err != nil {
		return []AqueductData{}, err
	}

	return data, nil
}

func (dc *DataCollector) productResourcesCaller(guid string) DataRetriever {
	return func() (io.Reader, error) {
		return dc.omService.ProductResources(guid)
	}
}

func appendRetrievedData(data []AqueductData, retriever DataRetriever, productType, dataType string) ([]AqueductData, error) {
	output, err := retriever()
	if err != nil {
		return data, errors.Wrap(err, fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))
	}

	return append(data, AqueductData{Data: output, Name: productType, Type: dataType}), nil
}
