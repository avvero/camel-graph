package main

import (
	"time"
	"log"
	"fmt"
)

type Metric struct {
	name    string
	time    time.Time
	value   interface{}
}

type MetricConsumer interface {
	consumeMetric(metric *Metric)
}

type MetricConsumerStub struct {
}

func (it *MetricConsumerStub) consumeMetric(metric *Metric) {
	message := fmt.Sprintf("%s %v %v", metric.name, metric.value, metric.time.Unix())
	if false {
		log.Println(fmt.Sprintf("log metrics: %s", message))
	}
}

func NewMetric(metricName string, value interface{}, t time.Time) *Metric {
	return &Metric{name: metricName, value: value, time: t}
}