package interioredge

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/clients/python"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"io"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	AnycastDebug = true
	SnapshotRouterLinks = true
	SnapshotDelay = 10 * time.Second
	WG sync.WaitGroup
)

var _ = Describe("Exchange AnyCast messages across all nodes", func() {

	const (
		numberClients = 1
		totalSmall    = 100000
		totalMedium   = 10000
		totalLarge    = 1000
	)

	var (
		//allRouterNames = []string{"interior-east"}
		allRouterNames = TopologySmoke.AllRouterNames()
		totalSenders   = numberClients * len(allRouterNames)
	)

	It(fmt.Sprintf("exchanges %d small messages with 1kb using %d senders and receivers", totalSmall, totalSenders), func() {
		runAnycastTest(totalSmall, 1024, numberClients, allRouterNames)
	})

	It(fmt.Sprintf("exchanges %d medium messages with 100kb using %d senders and receivers", totalMedium, totalSenders), func() {
		runAnycastTest(totalMedium, 1024*100, numberClients, allRouterNames)
	})

	It(fmt.Sprintf("exchanges %d large messages with 500kb using %d senders and receivers", totalLarge, totalSenders), func() {
		runAnycastTest(totalLarge, 1024*500, numberClients, allRouterNames)
	})

})

func runAnycastTest(msgCount int, msgSize int, numClients int, allRouterNames []string) {

	const (
		anycastAddress = "anycast/smoke/interioredge"
		//timeout        = 180
		timeout        = 600
	)
	ctx := TopologySmoke.FrameworkSmoke.GetFirstContext()

	// Deploying all senders across all nodes
	By("Deploying senders across all router nodes")
	senders := []*python.PythonClient{}
	for _, routerName := range allRouterNames {
		sndName := fmt.Sprintf("sender-pythonbasic-%s", routerName)
		senders = append(senders, python.DeployPythonClient(ctx, routerName, sndName, anycastAddress, python.BasicSender, numClients, msgCount, msgSize, timeout)...)
	}

	// Deploying all receivers across all nodes
	By("Deploying receivers across all router nodes")
	receivers := []*python.PythonClient{}
	for _, routerName := range allRouterNames {
		rcvName := fmt.Sprintf("receiver-pythonbasic-%s", routerName)
		receivers = append(receivers, python.DeployPythonClient(ctx, routerName, rcvName, anycastAddress, python.BasicReceiver, numClients, msgCount, msgSize, timeout)...)
	}

	//TODO maybe remove this or make it a reusable component
	if SnapshotRouterLinks {
		snapshotRouters(allRouterNames, ctx, anycastAddress, WG)
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
		log.Logf("Sender %s - Status: %v - Results - Delivered: %d - Released: %d - Rejected: %d - Modified: %d - Accepted: %d",
			s.Name, s.Status(), res.Delivered, res.Released, res.Rejected, res.Modified, res.Accepted)

		// Adding sender results
		totalSent := res.Delivered - res.Rejected - res.Released
		sndResults = append(sndResults, results{delivered: totalSent, success: s.Status() == amqp.Success})

		if AnycastDebug {
			saveLogs(s.Pod.Name)
		}
	}

	By("Collecting receivers results")
	for _, r := range receivers {
		log.Logf("Waiting receiver: %s", r.Name)
		r.Wait()

		log.Logf("Parsing receiver results")
		res := r.Result()
		log.Logf("Receiver %s - Status: %v - Results - Delivered: %d", r.Name, r.Status(), res.Delivered)

		// Adding receiver results
		rcvResults = append(rcvResults, results{delivered: res.Delivered, success: r.Status() == amqp.Success})

		if AnycastDebug {
			saveLogs(r.Pod.Name)
		}
	}

	if AnycastDebug {
		for _, r := range allRouterNames {
			pods, err := ctx.ListPodsForDeploymentName(r)
			if err != nil {
				log.Logf("Error retrieving pods: %v", err)
				continue
			}
			for _, p := range pods.Items {
				saveLogs(p.Name)
			}
		}
	}

	// At this point, we should stop snapshoting the routers
	SnapshotRouterLinks = false
	WG.Wait()

	// Validating total number of messages sent/received
	By("Validating sender results")
	for _, s := range sndResults {
		gomega.Expect(s.success).To(gomega.BeTrue())
		gomega.Expect(s.delivered).To(gomega.Equal(msgCount))
	}
	By("Validating receiver results")
	for _, r := range rcvResults {
		gomega.Expect(r.success).To(gomega.BeTrue())
		gomega.Expect(r.delivered).To(gomega.Equal(msgCount))
	}

}

func snapshotRouters(allRouterNames []string, ctx *framework.ContextData, anycastAddress string, wg sync.WaitGroup) {
	for _, routerName := range allRouterNames {
		pods, err := ctx.ListPodsForDeploymentName(routerName)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		for _, pod := range pods.Items {
			wg.Add(1)
			go func(podName string) {
				defer wg.Done()
				logFile := fmt.Sprintf("/tmp/ns_%s_pod_%s_linkStats.log", ctx.Namespace, podName)
				log.Logf("Saving linkStatus to: %s", logFile)
				f, _ := os.Create(logFile)
				defer f.Close()
				w := bufio.NewWriter(f)

				// Once all senders/receivers stopped, this should become false
				for SnapshotRouterLinks {
					// Retrieve links from router for the related address
					links, err := qdrmanagement.QdmanageQuery(*ctx, podName, entities.Link{}, func(entity entities.Entity) bool {
						log.Logf("Entity: %v", entity)
						l := entity.(entities.Link)
						return strings.HasSuffix(l.OwningAddr, anycastAddress)
					})
					gomega.Expect(err).NotTo(gomega.HaveOccurred())

					if len(links) > 0 {
						for _, e := range links {
							stat, err := json.Marshal(e)
							gomega.Expect(err).NotTo(gomega.HaveOccurred())
							w.Write(stat)
							w.WriteString("\n")
							w.Flush()
						}
					}

					time.Sleep(SnapshotDelay)
				}

				log.Logf("Finished snapshoting router - Logfile: %s", logFile)
			}(pod.Name)
		}
	}
}

func saveLogs(podName string) {

	// Wait so pod has enough time to finish properly
	time.Sleep(5 * time.Second)

	ctx := TopologySmoke.FrameworkSmoke.GetFirstContext()
	request := ctx.Clients.KubeClient.CoreV1().Pods(ctx.Namespace).GetLogs(podName, &v1.PodLogOptions{})
	logs, err := request.Stream()

	// Close when done reading
	defer logs.Close()

	// Reading logs into buf
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)

	// Allows reading line by line
	reader := bufio.NewReader(buf)

	// Iterate through lines\
	logFile := fmt.Sprintf("/tmp/ns_%s_pod_%s.log", ctx.Namespace, podName)
	log.Logf("Saving pod logs to: %s", logFile)
	f, _ := os.Create(logFile)
	w := bufio.NewWriter(f)
	outer: for {
		var line, partLine []byte
		var fullLine = true

		// ReadLine may not return the full line when it exceeds 4096 bytes,
		// so we need to keep reading till fullLine is false or eof is found
		for fullLine {
			partLine, fullLine, err = reader.ReadLine()
			line = append(line, partLine...)
			if err == io.EOF {
				break outer
			}
		}

		// write line
		w.Write(line)
		w.WriteString("\n")

	}
	w.Flush()
	f.Close()
}
