package interioredge

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/clients/python"
	"github.com/rh-messaging/qdr-shipshape/pkg/debug"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"io"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	timeout = 1800
)

type results struct {
	delivered int
	success   bool
}

// runSmokeTest main function that coordinates the execution of the smoke test
//              and validates the expected received messages count based on the
//              given address name (if multicast expects msgCount * len(senders)
func runSmokeTest(address string, msgCount int, msgSize int, allRouterNames []string) {

	// Use for threads executed in debug mode to collect info
	var WG sync.WaitGroup

	ctx := TopologySmoke.FrameworkSmoke.GetFirstContext()
	doneSnapshoting := false

	// Reading number of clients from config or use default of 1
	numClients, _ := Config.GetEnvPropertyInt("NUMBER_CLIENTS", 1)

	// Deploying all senders and receivers across all nodes
	senders, receivers := DeployClients(allRouterNames, ctx, address, numClients, msgCount, msgSize, timeout)

	// If the test times out, check the routers status
	cc := make(chan struct{})
	go func(closech <-chan struct{}) {
		fmt.Printf("Waiting 10 minutes to gather results\n")
		ta := time.After(time.Minute * 10)
		select {
		case <-ta:
			fmt.Printf("==========\nResult gathering timed out, logging router links status\n==================\n")
			for _, podRouter := range TopologySmoke.AllRouterNames() {

				podList, err := ctx.Clients.KubeClient.CoreV1().Pods(ctx.Namespace).List(v12.ListOptions{
					LabelSelector: fmt.Sprintf("interconnect_cr=%s", podRouter),
				})

				if err != nil {
					panic(err)
				}

				for _, pod := range podList.Items {
					commandToRun := fmt.Sprintf("qdstat -l")
					fmt.Println("=============== Router Pod => ", pod.Name)

					kb := framework.NewKubectlExecCommand(*ctx, pod.Name, time.Minute, strings.Split(commandToRun, " ")...)
					out, err := kb.Exec()

					if err != nil {
						log.Logf("error: %v\n", err)
					}
					fmt.Println("--- stdout ---")
					fmt.Println(out)
				}
			}
		case <-cc:
			fmt.Println("Closing routine for results gathering before timeout")
		}
	}(cc)

	// If debug mode is enabled, snapshot router links
	if IsDebugEnabled() {
		debug.SnapshotRouters(allRouterNames, ctx, entities.Link{}, nil, &WG, &doneSnapshoting)
	}

	// Collecting results
	sndResults, rcvResults := CollectResults(senders, receivers)
	close(cc)

	// Trigger goroutines waiting on doneChannel
	if IsDebugEnabled() {
		WG.Add(3)
		go saveSenderLogs(senders, ctx, &WG)
		go saveReceiverLogs(receivers, ctx, &WG)
		go saveRouterLogs(allRouterNames, ctx, &WG)
	}

	// At this point, we should stop snapshoting the routers
	doneSnapshoting = true

	// Wait till all goroutines complete (saving logs)
	WG.Wait()

	//
	// Validating total number of messages sent/received
	//

	// Sender results skipped when validating multicast results
	if !strings.Contains(address, "multicast") {
		By("Validating sender results")
		for _, s := range sndResults {
			gomega.Expect(s.success).To(gomega.BeTrue())
		}
	}
	By("Validating receiver results")
	for _, r := range rcvResults {
		gomega.Expect(r.success).To(gomega.BeTrue())
	}

}

// DeployClients deploys both senders and receivers across all routers and return the
//               generated slices of senders and receivers.
func DeployClients(allRouterNames []string, ctx *framework.ContextData, address string, numClients int, msgCount int, msgSize int, timeout int) ([]*python.PythonClient, []*python.PythonClient) {
	var senders []*python.PythonClient
	var receivers []*python.PythonClient

	// when using anycast, deploy senders first (as they must be supressed)
	if strings.Contains(address, "anycast") {
		senders = deploySenders(allRouterNames, ctx, address, numClients, msgCount, msgSize, timeout)
		receivers = deployReceivers(allRouterNames, msgCount, address, len(allRouterNames), ctx, numClients, msgSize, timeout)
		// when using multicast start receivers first (as the sender here uses presettled delivery)
	} else {
		receivers = deployReceivers(allRouterNames, msgCount, address, len(allRouterNames), ctx, numClients, msgSize, timeout)
		// wait till receiver is running (otherwise senders will send msgs before receivers can consume them)
		for _, r := range receivers {
			r.WaitForStatus(60, amqp.Running)
		}
		senders = deploySenders(allRouterNames, ctx, address, numClients, msgCount, msgSize, timeout)
	}
	return senders, receivers
}

func deployReceivers(allRouterNames []string, msgCount int, address string, numSenders int, ctx *framework.ContextData, numClients int, msgSize int, timeout int) []*python.PythonClient {
	// Deploying all receivers across all nodes
	By("Deploying receivers across all router nodes")
	receivers := []*python.PythonClient{}
	for _, routerName := range allRouterNames {
		rcvName := fmt.Sprintf("receiver-pythonbasic-%s", routerName)
		rcvMsgCount := msgCount
		if strings.HasPrefix(address, "multicast") {
			rcvMsgCount *= numSenders
		}
		receivers = append(receivers, python.DeployPythonClient(ctx, routerName, rcvName, address, IsDebugEnabled(), python.BasicReceiver, numClients, rcvMsgCount, msgSize, timeout)...)
	}
	return receivers
}

func deploySenders(allRouterNames []string, ctx *framework.ContextData, address string, numClients int, msgCount int, msgSize int, timeout int) []*python.PythonClient {
	// Deploying all senders across all nodes
	By("Deploying senders across all router nodes")
	senders := []*python.PythonClient{}
	for _, routerName := range allRouterNames {
		sndName := fmt.Sprintf("sender-pythonbasic-%s", routerName)
		senders = append(senders, python.DeployPythonClient(ctx, routerName, sndName, address, IsDebugEnabled(), python.BasicSender, numClients, msgCount, msgSize, timeout)...)
	}
	return senders
}

// saveSenderLogs Iterates through the sender slice and save pods logs (used in debug mode)
func saveSenderLogs(senders []*python.PythonClient, ctx *framework.ContextData, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Logf("Saving sender logs")
	for _, s := range senders {
		debug.SavePodLogs(ctx, s.Pod.Name, s.Pod.Spec.Containers[0].Name)
	}
}

// saveReceiverLogs Iterates through the receiver slice and save pods logs (used in debug mode)
func saveReceiverLogs(receivers []*python.PythonClient, ctx *framework.ContextData, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Logf("Saving receiver logs")
	for _, r := range receivers {
		debug.SavePodLogs(ctx, r.Pod.Name, r.Pod.Spec.Containers[0].Name)
	}
}

// saveRouterLogs Iterates through the routers slice and save pods logs (used in debug mode)
func saveRouterLogs(routers []string, ctx *framework.ContextData, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Logf("Saving router logs")
	for _, r := range routers {
		pods, err := ctx.ListPodsForDeploymentName(r)
		if err != nil {
			log.Logf("Error retrieving pods: %v", err)
			continue
		}
		for _, p := range pods.Items {
			debug.SavePodLogs(ctx, p.Name, p.Spec.Containers[0].Name)
			debug.BackTrace(ctx, p.Name)
		}
	}

}

func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

// CollectResults generate results obtained from python's basic senders and receivers.
func CollectResults(senders []*python.PythonClient, receivers []*python.PythonClient) ([]results, []results) {
	sndResults := []results{}
	rcvResults := []results{}

	By("Collecting senders results")
	for _, s := range senders {
		if strings.Contains(s.Url, "multicast") {
			log.Logf("Skipping sender results on multicast (validate just receivers)")
			break
		}

		log.Logf("Waiting sender: %s", s.Name)
		s.Wait()

		log.Logf("Parsing sender results")
		res := s.Result()

		if s.Status() != amqp.Success {

			// Tail last line to see if it contains the result
			linesToTail := int64(20)
			request := s.Context.Clients.KubeClient.CoreV1().Pods(s.Context.Namespace).GetLogs(s.Pod.Name, &v1.PodLogOptions{
				TailLines: &linesToTail,
			})
			logs, err := request.Stream()

			// Close when done reading
			defer logs.Close()

			// Reading logs into buf
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, logs)

			if err != nil {
				panic("error in copy information from podLogs to buf")
			}

			logMsg := buf.String()

			fileName := fmt.Sprintf("LogLines-for-pod-%s", s.Name)
			err = WriteToFile(fileName, logMsg)
			if err != nil {
				panic("Error while saving pod logs")
			}
			fmt.Printf(logMsg)
		}

		log.Logf("Sender %s - Status: %v - Results - Delivered: %d - Released: %d - Rejected: %d - Modified: %d - Accepted: %d",
			s.Name, s.Status(), res.Delivered, res.Released, res.Rejected, res.Modified, res.Accepted)

		// Adding sender results
		totalSent := res.Accepted
		// This multicast test uses presettled, so there is no ack from receivers
		if strings.Contains(s.Url, "multicast") {
			totalSent = res.Delivered
		}
		sndResults = append(sndResults, results{delivered: totalSent, success: s.Status() == amqp.Success})
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
	}

	return sndResults, rcvResults
}
