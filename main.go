package main

import (
	_ "net/http/pprof"
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"flag"
)

const (
	GetRouteSchemaPath    = "%s/jolokia/exec/org.apache.camel:context=%s,type=routes,name=\"%s\"/dumpRouteAsXml(boolean)/true"
	GetRouteEndpointsPath = "%s/jolokia/exec/org.apache.camel:context=%s,type=routes,name=\"%s\"/createRouteStaticEndpointJson(boolean)/true"
	GetRoutesPath         = "/jolokia/read/org.apache.camel:type=routes,*"
)

var (
	httpPort                     = flag.String("httpPort", "8080", "http server port")
	serviceUpdateIntervalSeconds = flag.Int("serviceUpdateIntervalSeconds", 60, "update interval for infos")
	routeUpdateIntervalSeconds   = flag.Int("routeUpdateIntervalSeconds", 60, "update interval for infos")
	graphiteUrl                  = flag.String("graphiteUrl", "", "host and port to send plaint text metrics to graphite")
	graphiteRepeatSendOnFail     = flag.Bool("graphiteRepeatSendOnFail", false, "repeat send metrcis to graphite on fail")
)

func main() {
	flag.Parse()

	config, err := ReadConfig("services.json")
	if err != nil {
		panic(fmt.Sprintf("Error during configuration %v", err))
	}

	var metricConsumer MetricConsumer
	if *graphiteUrl != "" {
		metricConsumer = NewGraphite(*graphiteUrl, *graphiteRepeatSendOnFail)
		log.Println("Metrics will be passed to graphite: " + *graphiteUrl)
	} else {
		metricConsumer = &MetricConsumerStub{}
	}

	instance, err := NewInstance(config, &metricConsumer)
	if err != nil {
		panic(fmt.Sprintf("Error during configuration %v", err))
	}

	// proxy stuff
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		var js []byte
		var err error
		envName := r.URL.Query().Get("env")
		var environmentToReturn *Environment
		if envName != "" {
			for _, environment := range instance.Environments {
				if environment.Name == envName {
					environmentToReturn = environment
					break
				}
			}
		}
		// marshal
		if envName != "" && environmentToReturn != nil {
			js, err = json.Marshal(environmentToReturn)
		} else {
			js, err = json.Marshal(instance)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	log.Println("Http server started on port " + *httpPort)
	http.ListenAndServe(":" + *httpPort, nil)
}
