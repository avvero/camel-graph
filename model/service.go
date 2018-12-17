package model

import (
	"time"
	"sync"
	"errors"
	"encoding/json"
	"log"
	"fmt"
	"strings"
	"github.com/antchfx/xquery/xml"
)

const (
	UPDATE_STATE_IN_PROCESS = "in_process"
	UPDATE_STATE_DONE       = "done"
	UPDATE_STATE_FAILED     = "failed"

	NONE = "None"
)

const (
	GetRouteSchemaPath    = "%s/jolokia/exec/org.apache.camel:context=%s,type=routes,name=\"%s\"/dumpRouteAsXml(boolean)/true"
	GetRouteEndpointsPath = "%s/jolokia/exec/org.apache.camel:context=%s,type=routes,name=\"%s\"/createRouteStaticEndpointJson(boolean)/true"
	GetRoutesPath         = "/jolokia/read/org.apache.camel:type=routes,*"
)

type Instance struct {
	Environments []*Environment `json:"environments,omitempty"`
}

type Environment struct {
	Name       string              `json:"name,omitempty"`
	ServiceMap map[string]*Service `json:"serviceMap,omitempty"`
}

type JsonTime time.Time

func (t JsonTime)MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

type Service struct {
	Name        string            `json:"name,omitempty"`
	Url         string            `json:"url,omitempty"`
	RouteMap    map[string]*Route `json:"routeMap,omitempty"`
	LastUpdated JsonTime          `json:"lastUpdated,"`
	Error       string            `json:"error,omitempty"`
	Color       string            `json:"color,omitempty"`
	// IN_PROCESS, DONE, FAILED
	UpdatingState string     `json:"updatingState,omitempty"`

	updateMutex    sync.Mutex
	metricConsumer *MetricConsumer
	config         *ServiceConfig
	environment    *Environment
	upd            chan time.Time
}

func NewInstance(config *InstanceConfig, metricConsumer *MetricConsumer) (*Instance, error) {
	instance := &Instance{Environments: make([]*Environment, len(config.Environments))}
	for i, environmentConfig := range config.Environments {
		environment, err := NewEnvironment(config, environmentConfig, metricConsumer)
		if err != nil {
			return nil, err
		}
		instance.Environments[i] = environment
	}
	return instance, nil
}

func NewEnvironment(instanceConfig *InstanceConfig, envConfig *EnvironmentConfig, metricConsumer *MetricConsumer) (*Environment, error) {
	if envConfig.Name == "" {
		return nil, errors.New("environment name must not be empty")
	}
	environment := &Environment{
		Name:       envConfig.Name,
		ServiceMap: make(map[string]*Service)}
	for _, serviceConfig := range envConfig.Services {
		service, err := NewService(instanceConfig, serviceConfig, environment, metricConsumer)
		if err != nil {
			return nil, err
		}
		environment.ServiceMap[service.Name] = service
	}
	return environment, nil
}

//Creates new service from config
func NewService(instanceConfig *InstanceConfig, config *ServiceConfig, environment *Environment, metricConsumer *MetricConsumer) (*Service, error) {
	if config.Name == "" {
		return nil, errors.New("service name must not be empty")
	}
	if config.Url == "" {
		return nil, errors.New("service url must not be empty")
	}
	service := &Service{
		config:             config,
		Name:               config.Name,
		Url:                config.Url,
		Color:              config.Color,
		upd:                make(chan time.Time, 100),
		RouteMap:           make(map[string]*Route),
		UpdatingState:      UPDATE_STATE_IN_PROCESS,
		environment:        environment,
		metricConsumer:     metricConsumer}
	go service.doUpdate(instanceConfig.ServiceUpdateIntervalSeconds, instanceConfig.RouteUpdateIntervalSeconds)
	return service, nil
}


func (service *Service) doUpdate(serviceUpdateIntervalSeconds int, routeUpdateIntervalSeconds int) {
	ticker := time.NewTicker(time.Duration(serviceUpdateIntervalSeconds) * time.Second)
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
			err := service.update(t, routeUpdateIntervalSeconds)
			if err != nil {
				service.UpdatingState = UPDATE_STATE_FAILED
				service.Error = fmt.Sprintf("%s", err)
			} else {
				service.Error = ""
				service.UpdatingState = UPDATE_STATE_DONE
				service.LastUpdated = JsonTime(t)
			}
		}
	}
}

func (service *Service) update(t time.Time, routeUpdateIntervalSeconds int) error {
	url := service.config.Url + GetRoutesPath
	body, err := callEndpoint(url, service.config.Authorization)
	if err != nil {
		log.Printf("error: %s:%s error during getting routes from %s: %s", service.environment.Name, service.Name,
			service.config.Url, err)
		return err
	} else {
		for _, r := range service.RouteMap {
			r.State = NONE
		}
		// Merge from object
		response := &ReadRouteResponse{}
		json.Unmarshal(body, response)
		//log.Printf("info:  %s:%s read route response %v", service.environment.Name, service.Name, response)
		for _, v := range response.Value {
			properContext := strings.Replace(v.CamelManagementName, " ", "_", -1)
			properRouteId := strings.Replace(v.RouteId, " ", "_", -1)
			routeName := fmt.Sprintf("%s.%s", properContext, properRouteId)
			route, exists := service.RouteMap[routeName]
			if !exists {
				route = &Route{
					Context:     v.CamelManagementName,
					Name:        v.RouteId,
					EndpointUri: v.EndpointUri,
					Endpoints: &Endpoints{
						Inputs:  make([]string, 0),
						Outputs: make([]string, 0),
					},
					service:       service,
					UpdatingState: UPDATE_STATE_IN_PROCESS,

					upd:         make(chan time.Time, 100),
					metrics: make(chan *Metric, 1000)}
				service.RouteMap[routeName] = route
				// Add first input
				route.Endpoints.Inputs = append(route.Endpoints.Inputs, cleanEndpoint(route, v.EndpointUri))
				go route.doUpdate(routeUpdateIntervalSeconds)
				go route.sendMetrics()
			}
			route.State = v.State
			route.Uptime = v.Uptime
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
		return nil
	}
}

func (route *Route) doUpdate(routeUpdateIntervalSeconds int) {
	ticker := time.NewTicker(time.Duration(routeUpdateIntervalSeconds) * time.Second)
	route.upd <- time.Now()

	for {
		select {
		case <-ticker.C:
			if route.State == NONE {
				log.Printf("info:  %s:%s:%s route is outed of service", route.service.environment.Name,
					route.service.Name, route.Name)
				return
			}
			if route.UpdatingState == UPDATE_STATE_IN_PROCESS {
				//log.Printf("info:  %s:%s:%s route is still has been updating", route.service.environment.Name,
				//	route.service.Name, route.Name)
			} else {
				//log.Printf("info:  %s:%s:%s route is updating", route.service.environment.Name,
				//	route.service.Name, route.Name)
				route.upd <- time.Now()
			}
		case t := <-route.upd:
			route.UpdatingState = UPDATE_STATE_IN_PROCESS
			err:= route.update()
			if err != nil {
				route.UpdatingState = UPDATE_STATE_FAILED
				route.Error = fmt.Sprintf("%s", err)
				if route.State == NONE {
					//close(route.upd)
					//close(route.metrics)
				}
			} else {
				route.Error = ""
				route.UpdatingState = UPDATE_STATE_DONE
				route.LastUpdated = JsonTime(t)
			}
		}
	}
}

func (route *Route) update() error {
	// Endpoints
	url := fmt.Sprintf(GetRouteEndpointsPath, route.service.config.Url, route.Context, route.Name)

	//log.Printf("indo: %s:%s:%s getting route endoints from %s",
	//	route.service.environment.Name, route.service.Name, route.Name, url)

	body, err := callEndpoint(url, route.service.config.Authorization)
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
			if endpointsEntry != nil && *endpointsEntry.Routes != nil{
				for _, v := range *endpointsEntry.Routes {
					if v.Outputs == nil {
						continue
					}
					for _, o := range v.Outputs {

						// skip configured and not filled endpoints
						if len(o.Uri) == 0 && strings.Contains(o.Uri, "{{") {
							continue
						}
						candidate := cleanEndpoint(route, o.Uri)
						if !contains(route.Endpoints.Outputs, candidate) {
							route.Endpoints.Outputs = append(route.Endpoints.Outputs, candidate)
						}
					}
				}
			}
		}
	}
	// Schema
	url = fmt.Sprintf(GetRouteSchemaPath, route.service.config.Url, route.Context, route.Name)
	//log.Printf("info:  %s:%s:%s getting schema %s", route.service.environment.Name, route.service.Name, route.Name, url)
	body, err = callEndpoint(url, route.service.config.Authorization)
	if err != nil {
		log.Printf("error: %s:%s:%s error during getting schema from %s: %s", route.service.environment.Name,
			route.service.Name, route.Name, route.service.config.Url, err)
		return err
	} else {
		response := &ReadResponse{}
		json.Unmarshal(body, response)
		if response != nil && len(response.Value) > 0 {
			route.Schema = response.Value
			xmlNode, _ := xmlquery.Parse(strings.NewReader(route.Schema))
			// to endpoints
			toEndpoints := xmlquery.Find(xmlNode, "//to")
			for _, toEndpoint := range toEndpoints {
				uri := toEndpoint.SelectAttr("uri")
				if len(uri) > 0 {
					candidate := cleanEndpoint(route, uri)
					if !contains(route.Endpoints.Outputs, candidate) {
						route.Endpoints.Outputs = append(route.Endpoints.Outputs, candidate)
					}
				}
			}
			// from endpoints
			fromEndpoints := xmlquery.Find(xmlNode, "//from")
			for _, fromEndpoint := range fromEndpoints {
				uri := fromEndpoint.SelectAttr("uri")
				if len(uri) > 0 {
					candidate := cleanEndpoint(route, uri)
					if !contains(route.Endpoints.Inputs, candidate) {
						route.Endpoints.Inputs = append(route.Endpoints.Inputs, candidate)
					}
				}
			}
		}
		return nil
	}
}

func contains(list []string, entry string) bool {
	if list == nil || len(list) == 0 {
		return false
	}
	for _, v := range list {
		if v == entry {
			return true
		}
	}
	return false
}

func cleanEndpoints(route *Route, v []*RouteEndpointEntry) (inputs []string) {
	inputs = make([]string, 0)
	for _, i := range v {
		url := i.Uri
		inputs = append(inputs, cleanEndpoint(route, url))
	}
	return
}

func cleanEndpoint(route *Route, endpoint string) (result string) {
	endpoint = strings.Replace(endpoint, "%7B", "{", -1)
	endpoint = strings.Replace(endpoint, "%7D", "}", -1)
	endpoint = strings.Replace(endpoint, "://", ":", -1)
	endpoint = strings.Replace(endpoint, "activemq:", "jms:", -1)
	//take left part before ?
	endpoint = strings.Split(endpoint, "?")[0]
	if strings.Contains(endpoint, "VirtualTopic") {
		topicName := strings.Split(endpoint, "VirtualTopic")[1]
		endpoint = "VirtualTopic" + topicName
	}
	if strings.HasPrefix(endpoint, "direct") {
		endpoint = route.service.Name + ":" + endpoint
	}
	if strings.HasPrefix(endpoint, "timer") {
		endpoint = route.service.Name + ":" + endpoint
	}
	return endpoint
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
