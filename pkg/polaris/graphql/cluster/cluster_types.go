package cluster

// TimeZone represents the valid cluster timezones.
type Timezone string

// Product represents the valid cluster products.
type Product string

const (
	CDM          Product = "CDM"
	CLOUD_DIRECT Product = "CLOUD_DIRECT"
	DATOS        Product = "DATOS"
	POLARIS      Product = "POLARIS"
)

// ProductType represents the valid cluster product types.
type ProductType string

const (
	CLOUD      ProductType = "Cloud"
	RSC        ProductType = "Polaris"
	EXOCOMPUTE ProductType = "ExoCompute"
	ONPREM     ProductType = "OnPrem"
	ROBO       ProductType = "Robo"
	UNKNOWN    ProductType = "Unknown"
)

// Status represents the valid cluster statuses.
type Status string

const (
	Connected    Status = "Connected"
	Disconnected Status = "Disconnected"
	Initializing Status = "Initializing"
)

// SystemStatus represents the valid cluster system statuses.
type SystemStatus string

const (
	SystemStatusOK      SystemStatus = "OK"
	SystemStatusWARNING SystemStatus = "WARNING"
	SystemStatusFATAL   SystemStatus = "FATAL"
)

// SearchFilter represents the valid cluster search filters.
type SearchFilter struct {
	ID              []string       `json:"id"`
	Name            []string       `json:"name"`
	Type            []ProductType  `json:"type"`
	ConnectionState []Status       `json:"connectionState"`
	SystemStatus    []SystemStatus `json:"systemStatus"`
	ProductType     []Product      `json:"productType"`
}

// SortByEnum represents the valid sort by values.
type SortBy string

const (
	SortByEstimatedRunway  SortBy = "ESTIMATED_RUNWAY"
	SortByInstalledVersion SortBy = "INSTALLED_VERSION"
	SortByClusterName      SortBy = "ClusterName"
	SortByClusterType      SortBy = "ClusterType"
	SortByClusterLocation  SortBy = "CLUSTER_LOCATION"
	SortByRegisteredAt     SortBy = "RegisteredAt"
)
