package graphite

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

// Aggregator is an interface exposing the methods that we can use to work with different kinds of metrics
// in a transparent way for the user.
type Aggregator interface {
	AddSum(string, interface{})
	Increase(string)
	AddAverage(string, interface{})
	SetActive(string)
	SetInactive(string)
	Run(time.Duration, chan bool) Aggregator
	Flush() (int, error)
}

type aggregator struct {
	config  *Config
	metrics map[string]Metric
	client  Graphite
}

func (a *aggregator) GetMetrics() map[string]Metric {
	return a.metrics
}

func (a *aggregator) getMetric(path string, defaultMetric Metric) Metric {
	metricPath := a.config.getMetricPath(path)
	if metric, exists := a.metrics[metricPath]; exists {
		return metric
	}
	return defaultMetric
}

func (a *aggregator) setMetric(path string, metric Metric) {
	metricPath := a.config.getMetricPath(path)
	a.metrics[metricPath] = metric
}

func (a *aggregator) updateMetric(path string, value interface{}, defaultMetric Metric) {
	mutex.Lock()
	defer mutex.Unlock()
	metric := a.getMetric(path, defaultMetric)
	metric.Update(value)
	a.setMetric(path, metric)
}

func (a *aggregator) Flush() (int, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if len(a.metrics) > 0 {
		buffer := bytes.NewBufferString("")
		timestamp := time.Now().Unix()
		for path, metric := range a.metrics {
			buffer.WriteString(fmt.Sprintf("%s %s %d\n", path, metric.Calculate(), timestamp))
		}
		a.metrics = map[string]Metric{}
		return a.client.SendBuffer(buffer)
	}
	return 0, nil
}

func (a *aggregator) run(period time.Duration, stopSendingMetrics chan bool) {
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			if _, err := a.Flush(); err != nil {
				log.Printf("Unable to send metrics: %s\n", err.Error())
			}
		case <-stopSendingMetrics:
			return
		}
	}
}

func (a *aggregator) AddSum(path string, value interface{}) {
	a.updateMetric(path, value, &MetricSum{})
}

func (a *aggregator) Increase(path string) {
	a.updateMetric(path, 1, &MetricSum{})
}

func (a *aggregator) AddAverage(path string, value interface{}) {
	a.updateMetric(path, value, &MetricAverage{})
}

func (a *aggregator) SetActive(path string) {
	a.updateMetric(path, true, &MetricActive{})
}

func (a *aggregator) SetInactive(path string) {
	a.updateMetric(path, false, &MetricActive{})
}

func (a *aggregator) Run(period time.Duration, stopSendingMetrics chan bool) Aggregator {
	go a.run(period, stopSendingMetrics)
	return a
}
