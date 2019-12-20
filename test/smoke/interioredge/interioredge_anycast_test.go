package interioredge

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp/qeclients"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"strconv"
)

var _ = Describe("Exchanges AnyCast messages across the nodes", func() {

	It("exchanges 100 small messages", func() {
		// TODO Sample implementation
		//      Need to investigate why it is failing when connecting to more than 1 router
		ctx := FrameworkSmoke.GetFirstContext()

		// Deploying all senders across all nodes
		By("Deploying senders across all router nodes")
		senders := []*qeclients.AmqpQESender{}
		for _, name := range AllRouterNames {
			senders = append(senders, deploySenders(ctx, name, "sender-python-" + name, qeclients.Python, 2, 100, "small-message.txt")...)
		}

		// Deploying all receivers across all nodes
		By("Deploying receivers across all router nodes")
		receivers := []*qeclients.AmqpQEReceiver{}
		for _, name := range AllRouterNames {
			receivers = append(receivers, deployReceivers(ctx, name, "receiver-python-" + name, qeclients.Python, 2, 100)...)
		}

		// Wait on all senders and receivers to finish (or timeout)
		totalSent := 0
		totalReceived := 0
		By("Waiting validating senders statuses")
		for _, s := range senders {
			s.Wait()
			gomega.Expect(s.Status()).To(gomega.Equal(amqp.Success))
			gomega.Expect(s.Status()).To(gomega.Equal(amqp.Success))
			log.Logf("Sender %s - Results - Delivered: %d - Released: %d - Modified: %d",
				s.Name, s.Result().Delivered, s.Result().Released, s.Result().Modified)
			totalSent += s.Result().Delivered
		}
		By("Waiting validating receivers statuses")
		for _, r := range receivers {
			r.Wait()
			gomega.Expect(r.Status()).To(gomega.Equal(amqp.Success))
			gomega.Expect(r.Status()).To(gomega.Equal(amqp.Success))
			log.Logf("Receiver %s - Results - Delivered: %d - Released: %d - Modified: %d",
				r.Name, r.Result().Delivered, r.Result().Released, r.Result().Modified)
			totalReceived += r.Result().Delivered
		}

		// Validating total number of messages sent/received
		By("Waiting validating number of sent and received messages")
		//Cannot validate properly because senders are not logging number of released or modified msgs
		//gomega.Expect(totalSent).To(gomega.Equal(100 * 2 * len(AllRouterNames)))
		gomega.Expect(totalReceived).To(gomega.Equal(100 * 2 * len(AllRouterNames)))

	})
})

// deployReceivers creates a slice of receivers and deploy all of them
func deployReceivers(ctx *framework.ContextData, icName string, receiverName string, impl qeclients.AmqpQEClientImpl, numberOfReceivers int, messages int) []*qeclients.AmqpQEReceiver {
		res := []*qeclients.AmqpQEReceiver{}
		url := fmt.Sprintf("amqp://%s:5672/anycastAddress", interconnect.GetDefaultServiceName(icName, ctx.Namespace))
		for i := 0; i < numberOfReceivers; i++ {
		rBuilder := qeclients.NewReceiverBuilder(receiverName + "-" + strconv.Itoa(i+1), impl, *ctx, url)
		rBuilder.Messages(messages)
		rcv, err := rBuilder.Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(rcv).NotTo(gomega.BeNil())
		res = append(res, rcv)
	}

	// Deploying
	for _, r := range res {
		err := r.Deploy()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	// Returning deployed receivers
	return res
}

// deploySenders creates a slice of senders and deploy all of them
func deploySenders(ctx *framework.ContextData, icName string, senderName string, impl qeclients.AmqpQEClientImpl,
	numberOfSenders int, messages int, contentFile string) []*qeclients.AmqpQESender {
	res := []*qeclients.AmqpQESender{}
	url := fmt.Sprintf("amqp://%s:5672/anycastAddress", interconnect.GetDefaultServiceName(icName, ctx.Namespace))
	for i := 0; i < numberOfSenders; i++ {
		psBuilder := qeclients.NewSenderBuilder(senderName + "-" + strconv.Itoa(i+1), impl, *ctx, url)
		psBuilder.Messages(messages)
		psBuilder.MessageContentFromFile(ConfigMapName, contentFile)
		sdr, err := psBuilder.Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sdr).NotTo(gomega.BeNil())
		res = append(res, sdr)
	}

	// Deploying
	for _, s := range res {
		err := s.Deploy()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	// Returning deployed senders
	return res
}
