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
	Retry() (int, error)
}

type aggregator struct {
	config  *Config
	metrics map[string]Metric
	client  Graphite
}

// GetMetrics retuns the metrics stored till this point in the aggregator.
func (a *aggregator) GetMetrics() map[string]Metric {
	return a.metrics
}

// Retry tries to retry the flush of metrics in case something went wrong. If this
// retry went wrong it won't try a third time.
func (a *aggregator) Retry() (int, error) {
	a.client.Reconnect()
	return a.Flush()
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

// Flush forces sending the current stored metrics to graphite.
func (a *aggregator) Flush() (int, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if len(a.metrics) > 0 {
		buffer := bytes.NewBufferString("")
		timestamp := time.Now().Unix()
		for path, metric := range a.metrics {
			buffer.WriteString(fmt.Sprintf("%s %s %d\n", path, metric.Calculate(), timestamp))
		}
		n, err := a.client.SendBuffer(buffer)
		if err == nil {
			a.metrics = map[string]Metric{}
		}
		return n, err
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
				if _, err := a.Retry(); err != nil {
					log.Printf("Unable to send metrics after reconnecting neither: %s\n", err.Error())
				}
			}
		case <-stopSendingMetrics:
			return
		}
	}
}

// AddSum initialises a metric where the final value sent to graphite will be the addition
// of all the values passed to the aggregator. So if we call `AddSum` with a specific metric path and
// values 5, 10, 15 and then we `Flush`, we will be sending a final value of 30 to graphite.
func (a *aggregator) AddSum(path string, value interface{}) {
	a.updateMetric(path, value, &MetricSum{})
}

// Increase is used as an alias of `AddSum` where the value incremented is always 1. Useful for giving
// a comprehensive behaviour to the metric.
func (a *aggregator) Increase(path string) {
	a.updateMetric(path, 1, &MetricSum{})
}

// AddAverage initialises a metric where the final value sent to graphite will be the average
// of all the values passed to the aggregator. So if we call `AddAverage` with a specific metric path
// and values 2, 10, 10 and then we `Flush`, we will be sending a final value of 7.333333 to graphite. The
// maximum decimals allowed is 6.
func (a *aggregator) AddAverage(path string, value interface{}) {
	a.updateMetric(path, value, &MetricAverage{})
}

// SetActive initialises a boolean metric where the final value sent to graphite will
// be 1, representing an `active` status.
func (a *aggregator) SetActive(path string) {
	a.updateMetric(path, true, &MetricActive{})
}

// SetInactive initialises a boolean metric where the final value sent to graphite will
// be 0, representing an `inactive` status. It's inteded to be used with
func (a *aggregator) SetInactive(path string) {
	a.updateMetric(path, false, &MetricActive{})
}

// Run starts a go routine to periodically flush the values stored in the aggregator to graphite.
// Useful if we don't want to manually call `Flush` every time.
func (a *aggregator) Run(period time.Duration, stopSendingMetrics chan bool) Aggregator {
	go a.run(period, stopSendingMetrics)
	return a
}
