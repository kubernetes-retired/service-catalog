package brokerapi

type EmptyResponse struct{}

type ErrorResponse struct {
	Description string `json:"description"`
}

type CatalogResponse struct {
	Services []Service `json:"services"`
}

type ProvisioningResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
}

type BindingResponse struct {
	Credentials interface{} `json:"credentials"`
}
