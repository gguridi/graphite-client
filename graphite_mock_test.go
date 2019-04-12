package graphite

import (
	"bytes"
	"fmt"
)

// MockGraphite implements the interface Graphite
type MockGraphite struct {
	Data                map[string]string
	MethodSend          func(*MockGraphite, string, string) (int, error)
	MethodSendBuffer    func(*MockGraphite, *bytes.Buffer) (int, error)
	MethodNewAggregator func(*MockGraphite) Aggregator
	MethodConnect       func(*MockGraphite) error
	MethodReconnect     func(*MockGraphite) error
	MethodDisconnect    func(*MockGraphite) error
}

// Send is an implementation of Graphite interface to be used with the mocking object.
func (m *MockGraphite) Send(path string, value string) (int, error) {
	if m.MethodSend != nil {
		return m.MethodSend(m, path, value)
	}
	m.Data[path] = fmt.Sprintf("%s:%s", path, value)
	return 0, nil
}

// SendBuffer is an implementation of Graphite interface to be used with the mocking object.
func (m *MockGraphite) SendBuffer(buffer *bytes.Buffer) (int, error) {
	if m.MethodSendBuffer != nil {
		return m.MethodSendBuffer(m, buffer)
	}
	m.Data["buffer"] = buffer.String()
	return 0, nil
}

// NewAggregator is an implementation of Graphite interface to be used with the mocking object.
func (m *MockGraphite) NewAggregator() Aggregator {
	if m.MethodNewAggregator != nil {
		return m.MethodNewAggregator(m)
	}
	return &MockAggregator{}
}

// Connect is an implementation of Graphite interface to be used with the mocking object.
func (m *MockGraphite) Connect() error {
	if m.MethodConnect != nil {
		return m.MethodConnect(m)
	}
	return nil
}

// Reconnect is an implementation of Graphite interface to be used with the mocking object.
func (m *MockGraphite) Reconnect() error {
	if m.MethodReconnect != nil {
		return m.MethodReconnect(m)
	}
	return nil
}

// Disconnect is an implementation of Graphite interface to be used with the mocking object.
func (m *MockGraphite) Disconnect() error {
	if m.MethodDisconnect != nil {
		return m.MethodDisconnect(m)
	}
	return nil
}
