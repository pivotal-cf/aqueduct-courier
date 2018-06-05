package data

const (
	OpsManagerProductType = "ops_manager"
	DirectorProductType   = "p-bosh"

	ResourcesDataType        = "resources"
	VmTypesDataType          = "vm_types"
	DiagnosticReportDataType = "diagnostic_report"
	DeployedProductsDataType = "deployed_products"
	InstallationsDataType    = "installations"
	PropertiesDataType       = "properties"
)

type Metadata struct {
	EnvType      string
	CollectedAt  string
	CollectionId string
	FileDigests  []FileDigest
}
type FileDigest struct {
	Name        string
	MimeType    string
	MD5Checksum string
	ProductType string
	DataType    string
}
