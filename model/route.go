package model

import "time"

type Route struct {
	Context     string     `json:"context,omitempty"`
	Name        string     `json:"name,omitempty"`
	Error       string     `json:"error,omitempty"`
	LastUpdated JsonTime   `json:"lastUpdated"`
	State       string     `json:"state,omitempty"`
	Uptime      string     `json:"uptime,omitempty"`
	Schema      string     `json:"schema,omitempty"`
	EndpointUri string     `json:"endpointUri,omitempty"`
	Endpoints   *Endpoints `json:"endpoints,omitempty"`
	// metrics
	ExchangesTotal      int `json:"exchangesTotal,omitempty"`
	ExchangesCompleted  int `json:"exchangesCompleted,omitempty"`
	ExchangesFailed     int `json:"exchangesFailed,omitempty"`
	ExchangesInflight   int `json:"exchangesInflight,omitempty"`
	MaxProcessingTime   int `json:"maxProcessingTime,omitempty"`
	MinProcessingTime   int `json:"minProcessingTime,omitempty"`
	LastProcessingTime  int `json:"lastProcessingTime,omitempty"`
	MeanProcessingTime  int `json:"meanProcessingTime,omitempty"`
	TotalProcessingTime int `json:"totalProcessingTime,omitempty"`
	FailuresHandled     int `json:"failuresHandled,omitempty"`
	Redeliveries        int `json:"redeliveries,omitempty"`
	StartTimestamp      string`json:"startTimestamp,omitempty"`
	// meta

	// DONE, FAILED
	UpdatingState string `json:"updatingState,omitempty"`
	metrics chan *Metric
	service *Service
	upd     chan time.Time
}

type Context struct {
	Name   string   `json:"name,omitempty"`
	Routes []*Route `json:"routes,omitempty"`
}

type Endpoints struct {
	Inputs  []string `json:"inputs,omitempty"`
	Outputs []string `json:"outputs,omitempty"`
}


