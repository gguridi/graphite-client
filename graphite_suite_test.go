package graphite_test

import (
	"fmt"
	"net"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGraphiteClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graphite client suite")
}

const (
	// MaxBuffer specifies a maximum buffer to use during the tests, to store whatever
	// we receive through the different test servers we create.
	MaxBuffer = 2048
)

func tcpHandler(received chan string, listener net.Listener) {
	fmt.Printf("Listening to connections to %s...\n", listener.Addr().String())
	for {
		connection, err := listener.Accept()
		go func(r chan string) {
			if err != nil {
				return
			}
			defer connection.Close()

			buffer := make([]byte, MaxBuffer)
			_, err = connection.Read(buffer)
			if err != nil {
				return
			}

			message := string(buffer)
			if message != "" {
				r <- message
			}
		}(received)
	}
}

func udpHandler(received chan string, listener net.PacketConn) {
	fmt.Println("Listening to udp connections...")
	for {
		buffer := make([]byte, MaxBuffer)
		listener.ReadFrom(buffer)
		message := string(buffer)
		if message != "" {
			received <- message
		}
	}
}

func createTCPServer(endpoint string) (net.Listener, chan string) {
	result := make(chan string)
	listener, err := net.Listen("tcp", endpoint)
	Expect(err).To(BeNil(), "Expected TCP server at %s", endpoint)
	go tcpHandler(result, listener)
	return listener, result
}

func createUDPServer(endpoint string) (net.PacketConn, chan string) {
	result := make(chan string)
	listener, err := net.ListenPacket("udp", endpoint)
	Expect(err).To(BeNil(), "Expected UDP server at %s", endpoint)
	go udpHandler(result, listener)
	return listener, result
}
