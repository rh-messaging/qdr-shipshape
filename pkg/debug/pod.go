package debug

import (
	"bufio"
	"fmt"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"io"
	v1 "k8s.io/api/core/v1"
	"os"
)

// SavePodLogs saves the pod/container logs to /tmp for debugging purposes
func SavePodLogs(ctx *framework.ContextData, podName string, container string) {

	log.Logf("-- saving pod logs for: %s", podName)
	request := ctx.Clients.KubeClient.CoreV1().Pods(ctx.Namespace).GetLogs(podName, &v1.PodLogOptions{Container: container})
	logs, err := request.Stream()
	if err != nil {
		log.Logf("ERROR getting stream - %v", err)
		return
	}

	// Close when done reading
	defer logs.Close()

	// Iterate through lines\
	logFile := fmt.Sprintf("/tmp/ns_%s_pod_%s.log", ctx.Namespace, podName)
	log.Logf("-- writing pod logs to: %s", logFile)
	f, _ := os.Create(logFile)
	defer f.Close()
	w := bufio.NewWriter(f)

	// Saving logs
	_, err = io.Copy(w, logs)
	if err != nil {
		log.Logf("error writing logs to: %s - %s", logFile, err)
	}
	w.Flush()

}
