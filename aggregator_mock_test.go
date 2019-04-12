package graphite

import (
	"time"
)

// MockAggregator implements the Aggregator interface and it's ready to be used to mock it.
type MockAggregator struct {
	Data              map[string]int
	MethodAddSum      func(*MockAggregator, string, interface{})
	MethodIncrease    func(*MockAggregator, string)
	MethodAddAverage  func(*MockAggregator, string, interface{})
	MethodSetActive   func(*MockAggregator, string)
	MethodSetInactive func(*MockAggregator, string)
	MethodRun         func(*MockAggregator, time.Duration, chan bool) Aggregator
	MethodFlush       func(*MockAggregator) (int, error)
}

// AddSum is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) AddSum(path string, value interface{}) {
	if m.MethodAddSum != nil {
		m.MethodAddSum(m, path, value)
		return
	}
	m.Data[path] = value.(int)
}

// Increase is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) Increase(path string) {
	if m.MethodIncrease != nil {
		m.MethodIncrease(m, path)
		return
	}
	m.Data[path] = 1
}

// AddAverage is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) AddAverage(path string, value interface{}) {
	if m.MethodAddAverage != nil {
		m.MethodAddAverage(m, path, value)
		return
	}
	m.Data[path] = value.(int)
}

// SetActive is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) SetActive(path string) {
	if m.MethodSetActive != nil {
		m.MethodSetActive(m, path)
		return
	}
	m.Data[path] = 1
}

// SetInactive is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) SetInactive(path string) {
	if m.MethodSetInactive != nil {
		m.MethodSetInactive(m, path)
		return
	}
	m.Data[path] = 0
}

// Run is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) Run(period time.Duration, stop chan bool) Aggregator {
	if m.MethodRun != nil {
		return m.MethodRun(m, period, stop)
	}
	return m
}

// Flush is an implementation of Aggregator interface to be used with the mocking object.
func (m *MockAggregator) Flush() (int, error) {
	if m.MethodFlush != nil {
		return m.MethodFlush(m)
	}
	return 0, nil
}
