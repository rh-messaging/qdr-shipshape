package python

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"strconv"
)

func DeployPythonClient(ctx *framework.ContextData, icName, clientName, address string, debug bool, command PythonClientCmd, numberOfClients, msgCount, msgSize, timeout int) []*PythonClient {
	res := []*PythonClient{}
	url := fmt.Sprintf("amqp://%s:5672/%s", interconnect.GetDefaultServiceName(icName, ctx.Namespace), address)
	log.Logf("Deploying client: [%s] using URL = [%s]", clientName, url)
	for i := 0; i < numberOfClients; i++ {
		builder := NewClientBuilder(clientName + "-" + strconv.Itoa(i), command, *ctx, url)
		// In case debugging is true, enable proton tracing frame
		if debug {
			builder.EnableTracing()
		}
		builder.EnvVar("MSG_COUNT", strconv.Itoa(msgCount))
		builder.EnvVar("MSG_SIZE", strconv.Itoa(msgSize))
		builder.Timeout(timeout)
		builder.ImageCustom("docker.io/qdrshipshape/clients-python:dispatch-1626")
		c := builder.Build()
		gomega.Expect(c).NotTo(gomega.BeNil())
		res = append(res, c)
	}

	// Deploying
	for _, r := range res {
		err := r.Deploy()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	// Returning deployed receivers
	return res
}
