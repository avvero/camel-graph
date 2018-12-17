package model

import (
	"log"
	"fmt"
	"net"
	"time"
	"github.com/fatih/pool"
)

type Graphite struct {
	pool             *pool.Pool
	metrics          chan *Metric
	repeatSendOnFail bool
}

const (
	StopCharacter = "\r\n\r\n"
	SleepDuration = 5
)

func NewGraphite(url string, repeatSendOnFail bool) *Graphite {
	//pool
	factory := func() (net.Conn, error) { return net.Dial("tcp", url) }
	pool, err := pool.NewChannelPool(5, 30, factory)
	if err != nil {
		panic(fmt.Sprintf("Error during creating new pool for graphite %v", err))
	}

	graphite := &Graphite{pool: &pool, metrics: make(chan *Metric, 1000), repeatSendOnFail: repeatSendOnFail}
	go func() {
		for {
			select {
			case metric := <-graphite.metrics:
				times := 1
				for err := graphite.send(metric); err != nil && times < 4; {
					sleepTime := SleepDuration * time.Duration(times)
					log.Println(fmt.Sprintf("Will sleep for %v (%v attempt of 3)", sleepTime, times))
					time.Sleep(sleepTime * time.Second)
					times++
				}
			}
		}
	}()
	return graphite
}
func (graphite *Graphite) send(metric *Metric) error {
	message := fmt.Sprintf("%s %v %v", metric.name, metric.value, metric.time.Unix())

	conn, err := (*graphite.pool).Get()
	if err != nil {
		if graphite.repeatSendOnFail {
			log.Println(fmt.Sprintf("Could not connected to graphite"))
			return err
		} else {
			log.Println(fmt.Sprintf("Could not connected to graphite, metrics will be lost"))
			return nil
		}
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message + StopCharacter))
	if err != nil {
		log.Println(fmt.Sprintf("Something wrong with the connection it will be closed"))
		if pc, ok := conn.(*pool.PoolConn); ok {
			pc.MarkUnusable()
			pc.Close()
		}
	}
	//log.Println(fmt.Sprintf("Do send metrics: %s", message))
	//log.Println(fmt.Sprintf("available connections in the pool: %s", (*graphite.pool).Len()))
	return nil
}

func (graphite *Graphite) consumeMetric(metric *Metric) {
	graphite.metrics <- metric
}
