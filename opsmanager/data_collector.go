package opsmanager

import (
	"fmt"
	"io"

	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pivotal-cf/om/api"
	"github.com/pkg/errors"
)

const (
	PendingChangesExistsMessage   = "There are pending changes on this Operations Manager, please apply them or revert them."
	PendingChangesFailedMessage   = "Failed to retrieve pending change list from Operations Manager"
	DeployedProductsFailedMessage = "Failed to retrieve deployed products list from Operations Manager"
	RequestorFailureErrorFormat   = "Failed retrieving %s %s"
)

//go:generate counterfeiter . PendingChangesLister
type PendingChangesLister interface {
	ListStagedPendingChanges() (api.PendingChangesOutput, error)
}

//go:generate counterfeiter . DeployedProductsLister
type DeployedProductsLister interface {
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
}

//go:generate counterfeiter . OmService
type OmService interface {
	ProductResources(guid string) (io.Reader, error)
	ProductProperties(guid string) (io.Reader, error)
	VmTypes() (io.Reader, error)
	DiagnosticReport() (io.Reader, error)
	DeployedProducts() (io.Reader, error)
	Installations() (io.Reader, error)
	Certificates() (io.Reader, error)
	CertificateAuthorities() (io.Reader, error)
}

type dataRetriever func() (io.Reader, error)

type DataCollector struct {
	omService             OmService
	pendingChangesService PendingChangesLister
	deployProductsService DeployedProductsLister
}

func NewDataCollector(oms OmService, pcs PendingChangesLister, dps DeployedProductsLister) *DataCollector {
	return &DataCollector{
		omService:             oms,
		pendingChangesService: pcs,
		deployProductsService: dps,
	}
}

func (dc *DataCollector) Collect() ([]Data, string, error) {
	var foundationId string
	pc, err := dc.pendingChangesService.ListStagedPendingChanges()
	if err != nil {
		return []Data{}, "", errors.Wrap(err, PendingChangesFailedMessage)
	}

	if hasPendingChanges(pc.ChangeList) {
		return []Data{}, "", errors.New(PendingChangesExistsMessage)
	}

	pl, err := dc.deployProductsService.ListDeployedProducts()
	if err != nil {
		return []Data{}, "", errors.Wrap(err, DeployedProductsFailedMessage)
	}

	var d []Data

	d, err = appendRetrievedData(d, dc.omService.DeployedProducts, data.OpsManagerProductType, data.DeployedProductsDataType)
	if err != nil {
		return []Data{}, "", err
	}

	for _, product := range pl {
		if product.Type != data.DirectorProductType {
			d, err = appendRetrievedData(d, dc.productResourcesCaller(product.GUID), product.Type, data.ResourcesDataType)
			if err != nil {
				return []Data{}, "", err
			}

			d, err = appendRetrievedData(d, dc.productPropertiesCaller(product.GUID), product.Type, data.PropertiesDataType)
			if err != nil {
				return []Data{}, "", err
			}
		} else {
			foundationId = product.GUID
		}
	}

	d, err = appendRetrievedData(d, dc.omService.VmTypes, data.OpsManagerProductType, data.VmTypesDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.DiagnosticReport, data.OpsManagerProductType, data.DiagnosticReportDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.Installations, data.OpsManagerProductType, data.InstallationsDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.Certificates, data.OpsManagerProductType, data.CertificatesDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.CertificateAuthorities, data.OpsManagerProductType, data.CertificateAuthoritiesDataType)
	if err != nil {
		return []Data{}, "", err
	}

	return d, foundationId, nil
}

func (dc DataCollector) productResourcesCaller(guid string) dataRetriever {
	return func() (io.Reader, error) {
		return dc.omService.ProductResources(guid)
	}
}

func (dc DataCollector) productPropertiesCaller(guid string) dataRetriever {
	return func() (io.Reader, error) {
		return dc.omService.ProductProperties(guid)
	}
}

func hasPendingChanges(changeList []api.ProductChange) bool {
	for _, change := range changeList {
		if change.Action != "unchanged" {
			return true
		}
	}
	return false
}

func appendRetrievedData(d []Data, retriever dataRetriever, productType, dataType string) ([]Data, error) {
	output, err := retriever()
	if err != nil {
		return d, errors.Wrap(err, fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))
	}

	return append(d, NewData(output, productType, dataType)), nil
}
