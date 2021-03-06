package graphite

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	// ProtocolTCP is a constant to specify the protocol TCP
	ProtocolTCP = "tcp"
	// ProtocolUDP is a constant to specify the protocol UDP
	ProtocolUDP = "udp"
)

// Graphite is an interface for a graphite client
type Graphite interface {
	Send(string, string) (int, error)
	SendBuffer(*bytes.Buffer) (int, error)
	NewAggregator() Aggregator
	Connect() error
	Reconnect() error
	Disconnect() error
}

type graphite struct {
	config     *Config
	protocol   string
	connection net.Conn
}

func newGraphite(config *Config, protocol string) Graphite {
	return &graphite{
		config:   config,
		protocol: protocol,
	}
}

// NewGraphiteTCP creates a new graphite client based on TCP.
func NewGraphiteTCP(config *Config) Graphite {
	return newGraphite(config, ProtocolTCP)
}

// NewGraphiteUDP creates a new graphite client based on UDP.
func NewGraphiteUDP(config *Config) Graphite {
	return newGraphite(config, ProtocolUDP)
}

// Connect establishes a connection with the graphite server, returning an error if something happened.
func (graphite *graphite) Connect() error {
	connection, err := graphite.connect(graphite.protocol)
	if err == nil {
		graphite.connection = connection
	}
	return err
}

// Reconnect tries to close a previous connection and reconnect with the graphite server.
func (graphite *graphite) Reconnect() error {
	graphite.Disconnect()
	return graphite.Connect()
}

// Disconnect tries to close a previous connection, returning an error if it can't.
func (graphite *graphite) Disconnect() error {
	if graphite.connection != nil {
		err := graphite.connection.Close()
		graphite.connection = nil
		return err
	}
	return fmt.Errorf("Connection was previously disconnected or never established")
}

// NewAggregator returns a new aggregator that will use the created client.
func (graphite *graphite) NewAggregator() Aggregator {
	return &aggregator{
		config:  graphite.config,
		client:  graphite,
		metrics: map[string]Metric{},
	}
}

func (graphite *graphite) getConnection() (net.Conn, error) {
	if graphite.config.ForceReconnect || graphite.connection == nil {
		if err := graphite.Reconnect(); err != nil {
			return nil, fmt.Errorf("Unable to connect/reconnect before sending metrics: %s", err.Error())
		}
	}
	return graphite.connection, nil
}

// Send is used to immediately send a metric to graphite, without having to specify a timestamp
// that will be acquired from the current datetime.
//
//         import graphite "github.com/gguridi/graphite-client"
//
//         client := graphite.NewGraphiteTCP(graphite.Config{
//             Host: "example.com",
//             Port: 2003,
//         })
//         client.Send("files.processed.count", 15)
func (graphite *graphite) Send(path, value string) (int, error) {
	metric := graphite.format(path, value, time.Now().Unix())
	return graphite.SendBuffer(bytes.NewBufferString(metric))
}

func (graphite *graphite) format(path string, value string, timestamp int64) string {
	return fmt.Sprintf("%s %s %d\n", path, value, timestamp)
}

// SendBuffer is used to immediately send a while buffer to graphite.
//
//         import graphite "github.com/gguridi/graphite-client"
//         import bytes
//
//         client := graphite.NewGraphiteTCP(graphite.Config{
//             Host: "example.com",
//             Port: 2003,
//         })
//         client.SendBuffer(bytes.NewBufferString(`
//             files.processed.count 15 1554992147
//             files.unprocessed.count 35 1554992147
//         `))
func (graphite *graphite) SendBuffer(buffer *bytes.Buffer) (int, error) {
	connection, err := graphite.getConnection()
	if err == nil {
		return connection.Write(buffer.Bytes())
	}
	return 0, err
}

func (graphite *graphite) connect(protocol string) (net.Conn, error) {
	switch protocol {
	case ProtocolUDP:
		return graphite.connectUDP()
	default:
		return graphite.connectTCP()
	}
}

func (graphite *graphite) connectTCP() (net.Conn, error) {
	address := graphite.config.getAddress()
	log.Printf("Graphite: connecting to %s via TCP\n", address)
	return net.DialTimeout("tcp", address, graphite.config.getTimeout())
}

func (graphite *graphite) connectUDP() (net.Conn, error) {
	address := graphite.config.getAddress()
	log.Printf("Graphite: connecting to %s via UDP\n", address)
	udpAddress, err := net.ResolveUDPAddr("udp", address)
	if err == nil {
		return net.DialUDP("udp", nil, udpAddress)
	}
	return nil, err
}
