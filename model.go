package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"sync"
)

const (
	UPDATE_STATE_IN_PROCESS = "in_process"
	UPDATE_STATE_DONE       = "done"
	UPDATE_STATE_FAILED     = "failed"
)

type Instance struct {
	Environments []*Environment `json:"environments,omitempty"`
}

type Environment struct {
	Name       string              `json:"name,omitempty"`
	ServiceMap map[string]*Service `json:"serviceMap,omitempty"`
}

type Service struct {
	Name        string            `json:"name,omitempty"`
	Url         string            `json:"url,omitempty"`
	RouteMap    map[string]*Route `json:"routeMap,omitempty"`
	LastUpdated *time.Time        `json:"lastUpdated"`
	Error       string            `json:"error,omitempty"`
	Color       string            `json:"color,omitempty"`
	// IN_PROCESS, DONE, FAILED
	UpdatingState string     `json:"updatingState,omitempty"`
	// DONE, FAILED
	FinalUpdatingState string     `json:"finalUpdatingState,omitempty"`

	updateMutex    sync.Mutex
	metricConsumer *MetricConsumer
	config         *ServiceConfig
	environment    *Environment
	upd            chan time.Time
}

type Context struct {
	Name   string   `json:"name,omitempty"`
	Routes []*Route `json:"routes,omitempty"`
}

type Route struct {
	Context     string     `json:"context,omitempty"`
	Name        string     `json:"name,omitempty"`
	Error       string     `json:"error,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated"`
	State       string     `json:"state,omitempty"`
	Uptime      string     `json:"uptime,omitempty"`
	Schema      string     `json:"schema,omitempty"`
	EndpointUri string     `json:"endpointUri,omitempty"`
	Endpoints   *Endpoints `json:"endpoints,omitempty"`
	// IN_PROCESS, DONE, FAILED
	UpdatingState string     `json:"updatingState,omitempty"`
	// DONE, FAILED
	FinalUpdatingState string     `json:"finalUpdatingState,omitempty"`
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
	metrics chan *Metric
	service *Service
	upd     chan time.Time
}

type Endpoints struct {
	Inputs  []string `json:"inputs,omitempty"`
	Outputs []string `json:"outputs,omitempty"`
}

func NewInstance(config *Config, metricConsumer *MetricConsumer) (*Instance, error) {
	instance := &Instance{Environments: make([]*Environment, len(config.Environments))}
	for i, environmentConfig := range config.Environments {
		environment, err := NewEnvironment(environmentConfig, metricConsumer)
		if err != nil {
			return nil, err
		}
		instance.Environments[i] = environment
	}
	return instance, nil
}

func NewEnvironment(config *EnvironmentConfig, metricConsumer *MetricConsumer) (*Environment, error) {
	if config.Name == "" {
		return nil, errors.New("Environment name must not be empty")
	}
	environment := &Environment{
		Name:       config.Name,
		ServiceMap: make(map[string]*Service)}
	for _, serviceConfig := range config.Services {
		service, err := NewService(serviceConfig, environment, metricConsumer)
		if err != nil {
			return nil, err
		}
		environment.ServiceMap[service.Name] = service
	}
	return environment, nil
}

//Creates new service from config
func NewService(config *ServiceConfig, environment *Environment, metricConsumer *MetricConsumer) (*Service, error) {
	if config.Name == "" {
		return nil, errors.New("Service name must not be empty")
	}
	if config.Url == "" {
		return nil, errors.New("Service url must not be empty")
	}
	service := &Service{
		config:             config,
		Name:               config.Name,
		Url:                config.Url,
		Color:              config.Color,
		upd:                make(chan time.Time, 100),
		RouteMap:           make(map[string]*Route),
		UpdatingState:      UPDATE_STATE_IN_PROCESS,
		FinalUpdatingState: UPDATE_STATE_IN_PROCESS,
		environment:        environment,
		metricConsumer:     metricConsumer}
	go service.doUpdate()
	return service, nil
}

func (service *Service) doUpdate() {
	ticker := time.NewTicker(time.Duration(*serviceUpdateIntervalSeconds) * time.Second)
	service.upd <- time.Now()

	for {
		select {
		case <-ticker.C:
			if service.UpdatingState == UPDATE_STATE_IN_PROCESS {
				//log.Printf("info:  %s:%s is still has been updating", service.environment.Name, service.Name)
			} else {
				service.upd <- time.Now()
			}
		case t := <-service.upd:
			service.UpdatingState = UPDATE_STATE_IN_PROCESS
			err := service.update(t)
			if err != nil {
				service.UpdatingState = UPDATE_STATE_FAILED
				service.FinalUpdatingState = UPDATE_STATE_FAILED
				service.Error = fmt.Sprintf("%s", err)
			} else {
				service.Error = ""
				service.UpdatingState = UPDATE_STATE_DONE
				service.FinalUpdatingState = UPDATE_STATE_DONE
				service.LastUpdated = &t
			}
		}
	}
}

func (service *Service) update(t time.Time) error {
	//service.updateMutex.Lock()
	//defer service.updateMutex.Unlock()

	url := service.config.Url + GetRoutesPath
	//log.Printf("info:  %s:%s getting routes from %s", service.environment.Name, service.Name, url)
	body, err := callEndpoint(url, service.config.Authorization)
	if err != nil {
		log.Printf("error: %s:%s error during getting routes from %s: %s", service.environment.Name, service.Name,
			service.config.Url, err)
		return err
	} else {
		// make all routes ready_to_update
		//for _, r := range service.RouteMap {
		//	r.UpdateState = "ready_to_update"
		//}
		// Merge from object
		response := &ReadRouteResponse{}
		json.Unmarshal(body, response)
		//log.Printf("info:  %s:%s read route response %v", service.environment.Name, service.Name, response)
		for _, v := range response.Value {
			properContext := strings.Replace(v.CamelManagementName, " ", "_", -1)
			properRouteId := strings.Replace(v.RouteId, " ", "_", -1)
			routeName := fmt.Sprintf("%s.%s", properContext, properRouteId)
			route, exists := service.RouteMap[routeName]
			// collect metrics
			if !exists {
				log.Printf("info:  %s:%s:%s register new route", service.environment.Name, service.Name, routeName)
				route = &Route{
					Context:     v.CamelManagementName,
					Name:        v.RouteId,
					State:       v.State,
					EndpointUri: v.EndpointUri,
					Endpoints: &Endpoints{
						Inputs:  []string{v.EndpointUri},
						Outputs: make([]string, 0),
					},
					Uptime:              v.Uptime,
					service:             service,
					UpdatingState:       UPDATE_STATE_IN_PROCESS,
					FinalUpdatingState:  UPDATE_STATE_IN_PROCESS,

					upd:         make(chan time.Time, 100),
					metrics: make(chan *Metric, 1000)}
				service.RouteMap[routeName] = route
				go route.doUpdate()
				go route.sendMetrics()
			} else {
				//log.Printf("info:  %s:%s:%s route is allready registered", service.environment.Name, service.Name, routeName)
			}
			// Metrics
			route.ExchangesTotal = v.ExchangesTotal
			route.ExchangesCompleted = v.ExchangesCompleted
			route.ExchangesFailed = v.ExchangesFailed
			route.ExchangesInflight = v.ExchangesInflight
			route.MaxProcessingTime = v.MaxProcessingTime
			route.MinProcessingTime = v.MinProcessingTime
			route.LastProcessingTime = v.LastProcessingTime
			route.MeanProcessingTime = v.MeanProcessingTime
			route.TotalProcessingTime = v.TotalProcessingTime
			route.FailuresHandled = v.FailuresHandled
			route.Redeliveries = v.Redeliveries

			route.collectMetrics(&v, t)
		}
		// stop update
		//for _, route := range service.RouteMap {
		//	if route.UpdateState == "ready_to_update" {
		//		log.Printf("%s: route %s does not exist in service, stop updating it", service.Name, route.Name)
		//		close(route.upd)
		//	}
		//}
		return nil
	}
}
func (route *Route) collectMetrics(e *ReadRouteEntry, t time.Time) {
	properContext := strings.Replace(e.CamelManagementName, " ", "_", -1)
	properRouteId := strings.Replace(e.RouteId, " ", "_", -1)
	routeName := fmt.Sprintf("%s_%s", properContext, properRouteId)
	properRouteName := strings.Replace(routeName, ".", "_", -1)

	serviceName := fmt.Sprintf("camel-graph.%s.%s.%s", route.service.environment.Name, route.service.Name, properRouteName)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "exchanges_total"), e.ExchangesTotal, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "exchanges_completed"), e.ExchangesCompleted, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "exchanges_failed"), e.ExchangesFailed, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "exchanges_inflight"), e.ExchangesInflight, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "max_processing_time"), e.MaxProcessingTime, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "min_processing_time"), e.MinProcessingTime, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "last_processing_time"), e.LastProcessingTime, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "mean_processing_time"), e.MeanProcessingTime, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "total_processing_time"), e.TotalProcessingTime, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "failures_handled"), e.FailuresHandled, t)
	route.metrics <- NewMetric(fmt.Sprintf("%s.%s", serviceName, "redeliveries"), e.Redeliveries, t)
}

func (route *Route) sendMetrics() {
	for {
		select {
		case metric := <-route.metrics:
			(*route.service.metricConsumer).consumeMetric(metric)
		}
	}
}

func (route *Route) doUpdate() {
	ticker := time.NewTicker(time.Duration(*routeUpdateIntervalSeconds) * time.Second)
	route.upd <- time.Now()

	for {
		select {
		case <-ticker.C:
			if route.UpdatingState == UPDATE_STATE_IN_PROCESS {
				//log.Printf("info:  %s:%s:%s route is still has been updating", route.service.environment.Name,
				//	route.service.Name, route.Name)
			} else {
				route.upd <- time.Now()
			}
		case t := <-route.upd:
			route.UpdatingState = UPDATE_STATE_IN_PROCESS
			err:= route.update()
			if err != nil {
				route.UpdatingState = UPDATE_STATE_FAILED
				route.FinalUpdatingState = UPDATE_STATE_FAILED
				route.Error = fmt.Sprintf("%s", err)
			} else {
				route.Error = ""
				route.UpdatingState = UPDATE_STATE_DONE
				route.FinalUpdatingState = UPDATE_STATE_DONE
				route.LastUpdated = &t
			}
		}
	}
}

func (route *Route) update() error {
	// Schema
	url := fmt.Sprintf(GetRouteSchemaPath, route.service.config.Url, route.Context, route.Name)
	//log.Printf("info:  %s:%s:%s getting schema %s", route.service.environment.Name, route.service.Name, route.Name, url)
	body, err := callEndpoint(url, route.service.config.Authorization)
	if err != nil {
		log.Printf("error: %s:%s:%s error during getting schema from %s: %s", route.service.environment.Name,
			route.service.Name, route.Name, route.service.config.Url, err)
		return err
	} else {
		response := &ReadResponse{}
		json.Unmarshal(body, response)
		route.Schema = response.Value

		// Endpoints
		url = fmt.Sprintf(GetRouteEndpointsPath, route.service.config.Url, route.Context, route.Name)
		//log.Printf("info:  %s:%s:%s getting endoints %s", route.service.environment.Name, route.service.Name, route.Name, url)
		body, err = callEndpoint(url, route.service.config.Authorization)
		if err != nil {
			log.Printf("error: %s:%s:%s error during getting route endoints from %s: %s",
				route.service.environment.Name, route.service.Name, route.Name, route.service.config.Url, err)
			return err
		} else {
			response := &ReadResponse{}
			json.Unmarshal(body, response)
			if response.Status != 200 {
				log.Printf("error: %s:%s:%s error during getting route endoints from %s: %s",
					route.service.environment.Name, route.service.Name, route.Name, route.service.config.Url, response.Error)
				return errors.New(response.Error)
			} else {
				endpointsEntry := &ReadRoutesEndpointsEntry{}
				json.Unmarshal([]byte(response.Value), endpointsEntry)
				// do with endpoints
				//inputs := make([]string, 0)
				outputs := make([]string, 0)
				if endpointsEntry != nil {
					if *endpointsEntry.Routes != nil {
						for _, v := range *endpointsEntry.Routes {
							//if v.Inputs != nil {
							//	inpavvero649uts = extractEndpointUrlsAsStrings(v.Inputs)
							//}
							if v.Outputs != nil {
								outputs = extractEndpointUrlsAsStrings(v.Outputs)
							}
						}
					}
				}
				//route.Endpoints = &Endpoints{Inputs: inputs, Outputs: outputs}
				route.Endpoints.Outputs = outputs
				return nil
			}
		}
	}
}

func extractEndpointUrlsAsStrings(v []*RouteEndpointEntry) (inputs []string) {
	inputs = make([]string, 0)
	for _, i := range v {
		url := i.Uri
		//todo
		url = strings.Replace(url, "%7B", "{", -1)
		url = strings.Replace(url, "%7D", "}", -1)

		inputs = append(inputs, url)
	}
	return
}
