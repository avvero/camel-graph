package model

type ReadRouteResponse struct {
	Value  map[string]ReadRouteEntry `json:"value,omitempty"`
	Status int                       `json:"status,omitempty"`
}

type ReadRouteEntry struct {
	EndpointUri         string
	CamelManagementName string
	Uptime              string
	CamelId             string
	RouteId             string
	State               string
	ExchangesTotal      int
	ExchangesCompleted  int
	ExchangesFailed     int
	ExchangesInflight   int
	MaxProcessingTime   int
	MinProcessingTime   int
	LastProcessingTime  int
	MeanProcessingTime  int
	TotalProcessingTime int
	FailuresHandled     int
	Redeliveries        int
	StartTimestamp      string
}

type ReadResponse struct {
	Value  string `json:"value,omitempty"`
	Error  string `json:"error,omitempty"`
	Status int    `json:"status,omitempty"`
}

type ReadRoutesEndpointsEntry struct {
	Routes *map[string]*ReadRouteEndpointsEntry `json:"routes,omitempty"`
}

type ReadRouteEndpointsEntry struct {
	Inputs  []*RouteEndpointEntry `json:"inputs,omitempty"`
	Outputs []*RouteEndpointEntry `json:"outputs,omitempty"`
}

type RouteEndpointEntry struct {
	Uri string `json:"uri,omitempty"`
}
