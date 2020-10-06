package graphite_test

import (
	"bytes"
	"net"

	. "github.com/gguridi/graphite-client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("graphite client", func() {

	var (
		client       Graphite
		result       chan string
		resultString string
	)

	Context("tcp protocol", func() {

		var (
			listener net.Listener
		)

		BeforeEach(func() {
			listener, result = createTCPServer(":3000")
			client = NewGraphiteTCP(&Config{
				Host: "localhost",
				Port: 3000,
			})
		})

		AfterEach(func() {
			listener.Close()
		})

		It("connects successfully if graphite is listening", func() {
			err := client.Connect()
			Expect(err).ToNot(HaveOccurred())
		})

		It("connects returns an error if graphite host is not reacheable", func() {
			client = NewGraphiteTCP(&Config{
				Host: "unknown",
				Port: 3000,
			})
			err := client.Connect()
			Expect(err).To(HaveOccurred())
		})

		It("connects returns an error if graphite is not listening", func() {
			client = NewGraphiteTCP(&Config{
				Host: "localhost",
				Port: 3010,
			})
			err := client.Connect()
			Expect(err).To(HaveOccurred())
		})

		It("reconnects successfully if graphite is listening", func() {
			err := client.Reconnect()
			Expect(err).ToNot(HaveOccurred())
		})

		It("disconnects with an error if never was connected", func() {
			err := client.Disconnect()
			Expect(err).To(HaveOccurred())
		})

		It("disconnects without errors if a connection was previously established", func() {
			client.Connect()
			err := client.Disconnect()
			Expect(err).ToNot(HaveOccurred())
		})

		It("reconnects automatically when sending metrics if connection hasn't been set", func() {
			n, err := client.Send("test", "1")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(18))
		})

		It("send a message with metric and value to graphite", func() {
			n, err := client.Send("metricA", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(22))
			Eventually(result).Should(Receive(&resultString))
			Expect(resultString).To(MatchRegexp(`metricA 10 \d{10}\n`))
		})

		It("reconnects automatically when sending a buffer if connection hasn't been set", func() {
			n, err := client.SendBuffer(bytes.NewBufferString("metric 10 1554992147\n"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(21))
		})

		It("send a whole buffer to graphite", func() {
			client.Connect()
			n, err := client.SendBuffer(bytes.NewBufferString("metric 10 1554992147\n"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(21))
			Eventually(result).Should(Receive(&resultString))
			Expect(resultString).To(ContainSubstring("metric 10 1554992147\n"))
		})

		It("returns an error if it can't deliver the metric to graphite", func() {
			listener.Close()
			n, err := client.Send("metricA", "10")
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("udp protocol", func() {

		var (
			listener net.PacketConn
		)

		BeforeEach(func() {
			listener, result = createUDPServer(":3001")
			client = NewGraphiteUDP(&Config{
				Host: "localhost",
				Port: 3001,
			})
		})

		AfterEach(func() {
			listener.Close()
		})

		It("connects successfully if graphite is listening", func() {
			err := client.Connect()
			Expect(err).ToNot(HaveOccurred())
		})

		It("connects returns an error if graphite host is not reacheable", func() {
			client = NewGraphiteUDP(&Config{
				Host: "unknown",
				Port: 103006,
			})
			err := client.Connect()
			Expect(err).To(HaveOccurred())
		})

		It("connects returns an error if graphite is not listening", func() {
			client = NewGraphiteTCP(&Config{
				Host: "localhost",
				Port: 3010,
			})
			err := client.Connect()
			Expect(err).To(HaveOccurred())
		})

		It("reconnects successfully if graphite is listening", func() {
			err := client.Reconnect()
			Expect(err).ToNot(HaveOccurred())
		})

		It("disconnects with an error if never was connected", func() {
			err := client.Disconnect()
			Expect(err).To(HaveOccurred())
		})

		It("disconnects without errors if a connection was previously established", func() {
			client.Connect()
			err := client.Disconnect()
			Expect(err).ToNot(HaveOccurred())
		})

		It("reconnects automatically when sending metrics if connection hasn't been set", func() {
			n, err := client.Send("test", "1")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(18))
		})

		It("send a message with metric and value to graphite", func() {
			client.Connect()
			n, err := client.Send("metricA", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(22))
			Eventually(result).Should(Receive(&resultString))
			Expect(resultString).To(MatchRegexp(`metricA 10 \d{10}\n`))
		})

		It("reconnects automatically when sending a buffer if connection hasn't been set", func() {
			n, err := client.SendBuffer(bytes.NewBufferString("metric 10 1554992147\n"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(21))
		})

		It("send a whole buffer to graphite", func() {
			client.Connect()
			n, err := client.SendBuffer(bytes.NewBufferString("metric 10 1554992147\n"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(21))
			Eventually(result).Should(Receive(&resultString))
			Expect(resultString).To(ContainSubstring("metric 10 1554992147\n"))
		})

		It("doesn't return an error if can't deliver the metric to graphite because it's UDP", func() {
			listener.Close()
			n, err := client.Send("metricA", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(22))
		})
	})
})
