package interioredge

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/clients/python"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
)

var _ = Describe("Exchange MultiCast messages across all nodes", func() {

	//const (
	//	numberClients = 1
	//	totalSmall    = 100000
	//	totalMedium   = 10000
	//	totalLarge    = 1000
	//)
	//
	//var (
	//	//allRouterNames = []string{"interior-east"}
	//	allRouterNames = TopologySmoke.AllRouterNames()
	//	totalSenders   = numberClients * len(allRouterNames)
	//)
	//
	//It(fmt.Sprintf("exchanges %d small messages with 1kb using %d senders and receivers", totalSmall, totalSenders), func() {
	//	runMulticastTest(totalSmall, 1024, numberClients, allRouterNames)
	//})

})

func runMulticastTest(msgCount int, msgSize int, numClients int, allRouterNames []string) {

	const (
		multicastAddress = "multicast/smoke/interioredge"
		timeout          = 600
	)
	ctx := TopologySmoke.FrameworkSmoke.GetFirstContext()
	perReceiverMsgCount := msgCount * numClients * len(allRouterNames)

	// Deploying all senders across all nodes
	By("Deploying senders across all router nodes")
	senders := []*python.PythonClient{}
	for _, routerName := range allRouterNames {
		sndName := fmt.Sprintf("sender-pythonbasic-%s", routerName)
		senders = append(senders, python.DeployPythonClient(ctx, routerName, sndName, multicastAddress, python.BasicSender, numClients, msgCount, msgSize, timeout)...)
	}

	// Deploying all receivers across all nodes
	By("Deploying receivers across all router nodes")
	receivers := []*python.PythonClient{}
	for _, routerName := range allRouterNames {
		rcvName := fmt.Sprintf("receiver-pythonbasic-%s", routerName)
		receivers = append(receivers, python.DeployPythonClient(ctx, routerName, rcvName, multicastAddress, python.BasicReceiver, numClients, perReceiverMsgCount, msgSize, timeout)...)
	}

	//TODO split this function into more cohesive ones
	type results struct {
		delivered int
		success   bool
	}
	sndResults := []results{}
	rcvResults := []results{}

	By("Collecting senders results")
	for _, s := range senders {
		log.Logf("Waiting sender: %s", s.Name)
		s.Wait()

		log.Logf("Parsing sender results")
		res := s.Result()
		log.Logf("Sender %s - Results - Delivered: %d - Released: %d - Rejected: %d - Modified: %d",
			s.Name, res.Delivered, res.Released, res.Rejected, res.Modified)

		// Adding sender results
		totalSent := res.Delivered - res.Rejected - res.Released
		sndResults = append(sndResults, results{delivered: totalSent, success: s.Status() == amqp.Success})
	}

	By("Collecting receiver results")
	for _, r := range receivers {
		log.Logf("Waiting receiver: %s", r.Name)
		r.Wait()

		log.Logf("Parsing receiver results")
		res := r.Result()
		log.Logf("Receiver %s - Results - Delivered: %d", r.Name, res.Delivered)

		// Adding receiver results
		rcvResults = append(rcvResults, results{delivered: res.Delivered, success: r.Status() == amqp.Success})
	}

	// Validating total number of messages sent/received
	By("Validating sender results")
	for _, s := range sndResults {
		gomega.Expect(s.success).To(gomega.BeTrue())
		gomega.Expect(s.delivered).To(gomega.Equal(msgCount))
	}
	By("Validating receiver results")
	for _, r := range rcvResults {
		gomega.Expect(r.success).To(gomega.BeTrue())
		gomega.Expect(r.delivered).To(gomega.Equal(perReceiverMsgCount))
	}

}
