package test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/examples/basic/test/javaclient"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
)

var _ = Describe("Basic", func() {

	It("Sends 1000 anycast messages", func() {
		// Defines the URL to use
		url := fmt.Sprintf("amqp://%s:5672", InterconnectName)

		// Creating the sender
		sender := javaclient.NewClientBuilder(
			"java-sender",
			javaclient.JavaBasicSender,
			*Framework.GetFirstContext(),
			url).Timeout(60).Build()

		// Creating the receiver
		receiver := javaclient.NewClientBuilder(
			"java-receiver",
			javaclient.JavaBasicReceiver,
			*Framework.GetFirstContext(),
			url).Timeout(60).Build()

		// Deploy sender and receiver
		sender.Deploy()
		receiver.Deploy()

		// Wait for sender and receiver to finish
		sender.Wait()
		receiver.Wait()

		//
		// Validate results
		//
		gomega.Expect(sender.Status()).To(gomega.Equal(amqp.Success))
		gomega.Expect(receiver.Status()).To(gomega.Equal(amqp.Success))

		// Messages accepted
		gomega.Expect(sender.Result().Accepted).To(gomega.Equal(1000))
		gomega.Expect(receiver.Result().Accepted).To(gomega.Equal(1000))
	})

})
