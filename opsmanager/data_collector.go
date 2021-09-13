package opsmanager

import (
	"fmt"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"
	"github.com/pkg/errors"
	"io"
	"log"
	"strings"
)

const (
	PendingChangesExistsMessage   = "There are pending changes on this Operations Manager, please apply them or revert them."
	PendingChangesFailedMessage   = "Failed to retrieve pending change list from Operations Manager"
	DeployedProductsFailedMessage = "Failed to retrieve deployed products list from Operations Manager"
	RequestorFailureErrorFormat   = "Failed retrieving %s %s"
	PendingChangesExistsFormat    = "Warning: This foundation has pending changes. The collector will continue to collect but reports from the Tanzu team may represent products with pending changes and therefore staged data, rather than deployed data. List of changes:\n%s"
)

var PendingChangesExistsError = errors.New(PendingChangesExistsMessage)

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
	logger                log.Logger
	omService             OmService
	opsManagerURL         string
	pendingChangesService PendingChangesLister
	deployProductsService DeployedProductsLister
}

func NewDataCollector(logger log.Logger, oms OmService, omURL string, pcs PendingChangesLister, dps DeployedProductsLister) *DataCollector {
	return &DataCollector{
		logger:                logger,
		omService:             oms,
		opsManagerURL:         omURL,
		pendingChangesService: pcs,
		deployProductsService: dps,
	}
}

func (dc *DataCollector) Collect() ([]Data, string, error) {
	dc.logger.Printf("Collecting data from Operations Manager at %s", dc.opsManagerURL)

	var foundationId string
	pc, err := dc.pendingChangesService.ListStagedPendingChanges()
	if err != nil {
		return []Data{}, "", errors.Wrap(err, PendingChangesFailedMessage)
	}

	if hasPendingChanges(pc.ChangeList) {
		dc.logger.Print(createPendingChangesWarningMessage(pc.ChangeList))
	}

	pl, err := dc.deployProductsService.ListDeployedProducts()
	if err != nil {
		return []Data{}, "", errors.Wrap(err, DeployedProductsFailedMessage)
	}

	var d []Data

	d, err = appendRetrievedData(d, dc.omService.DeployedProducts, collector_tar.OpsManagerProductType, collector_tar.DeployedProductsDataType)
	if err != nil {
		return []Data{}, "", err
	}

	for _, product := range pl {
		if product.Type != collector_tar.DirectorProductType {
			d, err = appendRetrievedData(d, dc.productResourcesCaller(product.GUID), product.Type, collector_tar.ResourcesDataType)
			if err != nil {
				return []Data{}, "", err
			}

			d, err = appendRetrievedData(d, dc.productPropertiesCaller(product.GUID), product.Type, collector_tar.PropertiesDataType)
			if err != nil {
				return []Data{}, "", err
			}
		} else {
			foundationId = product.GUID
		}
	}

	d, err = appendRetrievedData(d, dc.omService.VmTypes, collector_tar.OpsManagerProductType, collector_tar.VmTypesDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.DiagnosticReport, collector_tar.OpsManagerProductType, collector_tar.DiagnosticReportDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.Installations, collector_tar.OpsManagerProductType, collector_tar.InstallationsDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.Certificates, collector_tar.OpsManagerProductType, collector_tar.CertificatesDataType)
	if err != nil {
		return []Data{}, "", err
	}

	d, err = appendRetrievedData(d, dc.omService.CertificateAuthorities, collector_tar.OpsManagerProductType, collector_tar.CertificateAuthoritiesDataType)
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

func createPendingChangesWarningMessage(changeList []api.ProductChange) string{
	var changesList []string
	for _, change := range changeList {
		if change.Action != "unchanged" {
			changesList = append(changesList, fmt.Sprintf("%s: %s", change.GUID, change.Action))
		}
	}
	return fmt.Sprintf(PendingChangesExistsFormat, strings.Join(changesList, "\n"))
}

func appendRetrievedData(d []Data, retriever dataRetriever, productType, dataType string) ([]Data, error) {
	output, err := retriever()
	if err != nil {
		return d, errors.Wrap(err, fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))
	}

	return append(d, NewData(output, productType, dataType)), nil
}
