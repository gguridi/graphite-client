package graphite

import (
	"fmt"
	"strconv"
)

// Metric is an interface to be able to create new metric types easily.
// Each metric must have some methods to be able to be used by the Aggregator.
type Metric interface {

	// Update receives a generic value through interface{} to update its internal value.
	Update(interface{})
	// Clear is used to reset the metric to the initial value.
	Clear()
	// Calculate is used to perform the necessary operations to retrieve the final value
	// that will be sent to graphite, standarised as a string.
	Calculate() string
}

// MetricSum creates a metric that contains a value that increases with time.
type MetricSum struct {
	Sum int64
}

// Update increases the value of the metric with the amount received.
func (metric *MetricSum) Update(value interface{}) {
	metric.Sum += int64(value.(int))
}

// Clear reinitiales the value to zero.
func (metric *MetricSum) Clear() {
	metric.Sum = 0
}

// Calculate calculates the value to send.
func (metric *MetricSum) Calculate() string {
	return strconv.FormatInt(metric.Sum, 10)
}

// MetricAverage creates a metric to store the average value between several values.
type MetricAverage struct {
	Sum   int64
	Count int64
}

// Update increases the components necessary to calculate afterwards the average value.
// Each time the metric is updated, the result of Calculate will change.
func (metric *MetricAverage) Update(value interface{}) {
	metric.Sum += int64(value.(int))
	metric.Count++
}

// Clear reinitiales the average value and counter.
func (metric *MetricAverage) Clear() {
	metric.Sum = 0
	metric.Count = 0
}

// Calculate calculates the value to send.
func (metric *MetricAverage) Calculate() string {
	if metric.Sum > 0 {
		return fmt.Sprintf("%.6f", float64(metric.Sum)/float64(metric.Count))
	}
	return "0"
}

// MetricActive creates a metric to set a boolean status in graphite.
type MetricActive struct {
	State bool
}

// Update sets the active/inactive status through a boolean.
func (metric *MetricActive) Update(value interface{}) {
	metric.State = value.(bool)
}

// Clear reinitiales the value to inactive.
func (metric *MetricActive) Clear() {
	metric.State = false
}

// Calculate calculates the value to send.
func (metric *MetricActive) Calculate() string {
	bool2integer := map[bool]int{false: 0, true: 1}
	return strconv.Itoa(bool2integer[metric.State])
}
