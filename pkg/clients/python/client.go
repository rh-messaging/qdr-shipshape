package python

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp"
	"io"
	v1 "k8s.io/api/core/v1"
	"time"
)

type PythonClientCmd string

const (
	PythonClientImage string          = "docker.io/qdrshipshape/clients-python:latest"
	BasicSender       PythonClientCmd = "basic_sender.py"
	BasicReceiver     PythonClientCmd = "basic_receiver.py"
)

type PythonClient struct {
	amqp.AmqpClientCommon
}

func (p *PythonClient) Result() amqp.ResultData {
	// If still running, just return an empty structure
	if p.Running() {
		return amqp.ResultData{}
	}

	// If client is not longer running and finalResult already set, return it
	if p.FinalResult != nil {
		return *p.FinalResult
	}

	// Wait for 5 seconds as python client delays before printing the result data
	time.Sleep(5 * time.Second)

	// Tail last line to see if it contains the result
	linesToTail := int64(1)
	request := p.Context.Clients.KubeClient.CoreV1().Pods(p.Context.Namespace).GetLogs(p.Pod.Name, &v1.PodLogOptions{
		TailLines:    &linesToTail,
	})
	logs, err := request.Stream()
	gomega.Expect(err).To(gomega.BeNil())

	// Close when done reading
	defer logs.Close()

	// Reading logs into buf
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	gomega.Expect(err).To(gomega.BeNil())

	// Allows reading line by line
	reader := bufio.NewReader(buf)

	// ReadLine may not return the full line when it exceeds 4096 bytes,
	// so we need to keep reading till fullLine is false or eof is found
	var fullLine = true
	var line, partLine []byte
	for fullLine {
		partLine, fullLine, err = reader.ReadLine()
		line = append(line, partLine...)
		if err == io.EOF {
			break
		}
		gomega.Expect(err).To(gomega.BeNil())
	}

	// Unmarshalling PythonClientResult
	var cliResult PythonClientResult
	err = json.Unmarshal([]byte(line), &cliResult)
	gomega.Expect(err).To(gomega.BeNil())

	// Generating result data
	result := amqp.ResultData{
		Delivered: cliResult.Delivered,
		Released:  cliResult.Released,
		Modified:  cliResult.Modified,
		Accepted:  cliResult.Accepted,
		Rejected:  cliResult.Rejected,
	}

	// Locking to set finalResults
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if p.FinalResult == nil {
		p.FinalResult = &result
	}

	return result
}

type PythonClientResult struct {
	Delivered int    `json:"delivered"`
	Released  int    `json:"released"`
	Rejected  int    `json:"rejected"`
	Modified  int    `json:"modified"`
	Accepted  int    `json:"accepted"`
	ErrorMsg  string `json:"errormsg"`
}
