package send_receive

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	qdr_python_client "github.com/rh-messaging/qdr-shipshape/pkg/clients/python"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp/qeclients"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
)

const (
	MessageCount int = 100
)

var _ = Describe("Exchanges AnyCast messages across the nodes", func() {

	It("Exchanges small messages", func() {
		var (
			pythonSender   *qeclients.AmqpQEClientCommon
			pythonReceiver *qeclients.AmqpQEClientCommon
			err            error
		)

		ctx := travisFramework.GetFirstContext()
		base_url := fmt.Sprintf("amqp://%s:5672/", interconnect.GetDefaultServiceName(IcInteriorRouterName, ctx.Namespace))
		url := base_url + "anycastAddress"

		By("Deploying one python-qeclient sender and one python-qeclient receiver")

		pythonSenderBuilder := qeclients.NewSenderBuilder("sender-"+IcInteriorRouterName, qeclients.Python, *ctx, url)
		pythonSenderBuilder.Messages(MessageCount)
		pythonSenderBuilder.MessageContentFromFile(ConfigMapName, "small-message.txt")
		pythonSender, err = pythonSenderBuilder.Build()
		Expect(err).NotTo(HaveOccurred())
		Expect(pythonSender).NotTo(BeNil())

		pythonReceiverBuilder := qeclients.NewReceiverBuilder("receiver-"+IcInteriorRouterName, qeclients.Python, *ctx, url)
		pythonReceiverBuilder.Messages(MessageCount)
		pythonReceiver, err = pythonReceiverBuilder.Build()
		Expect(err).NotTo(HaveOccurred())
		Expect(pythonReceiver).NotTo(BeNil())

		err = pythonSender.Deploy()
		Expect(err).NotTo(HaveOccurred())

		err = pythonReceiver.Deploy()
		Expect(err).NotTo(HaveOccurred())

		pythonSender.Wait()
		Expect(pythonSender.Status()).To(Equal(amqp.Success))

		pythonReceiver.Wait()
		Expect(pythonReceiver.Status()).To(Equal(amqp.Success))

		log.Logf("Sender %s - Results - Delivered: %d - Released: %d - Modified: %d",
			pythonSender.Name, pythonSender.Result().Delivered, pythonSender.Result().Released, pythonSender.Result().Modified)

		log.Logf("Receiver %s - Results - Delivered: %d - Released: %d - Modified: %d",
			pythonReceiver.Name, pythonReceiver.Result().Delivered, pythonReceiver.Result().Released, pythonReceiver.Result().Modified)

		Expect(pythonReceiver.Result().Delivered).To(Equal(MessageCount))
		Expect(pythonSender.Result().Delivered).To(Equal(MessageCount))

		By("Deploying one qdr-python-sender and one qdr-python-receiver")
		sender := qdr_python_client.DeployPythonClient(ctx, IcInteriorRouterName, "sender-"+IcInteriorRouterName, "anycast/some", true, qdr_python_client.BasicSender, 1, 1, 128, 60)[0]
		receiver := qdr_python_client.DeployPythonClient(ctx, IcInteriorRouterName, "receiver-"+IcInteriorRouterName, "anycast/some", true, qdr_python_client.BasicReceiver, 1, 1, 128, 60)[0]

		sender.Wait()
		receiver.Wait()
		sender_result := sender.Result()
		receiver_result := receiver.Result()

		Expect(sender.Name).To(Equal("sender-interior-0"))
		Expect(sender.Status()).To(Equal(amqp.Success))
		Expect(sender_result.Delivered).To(Equal(1))
		Expect(sender_result.Released).To(Equal(0))
		Expect(sender_result.Rejected).To(Equal(0))
		Expect(sender_result.Modified).To(Equal(0))
		Expect(sender_result.Accepted).To(Equal(1))

		Expect(receiver.Name).To(Equal("receiver-interior-0"))
		Expect(receiver.Status()).To(Equal(amqp.Success))
		Expect(receiver_result.Delivered).To(Equal(1))
		Expect(receiver_result.Released).To(Equal(0))
		Expect(receiver_result.Rejected).To(Equal(0))
		Expect(receiver_result.Modified).To(Equal(0))
		Expect(receiver_result.Accepted).To(Equal(0))
	})
})
