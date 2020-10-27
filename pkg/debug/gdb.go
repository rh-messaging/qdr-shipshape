package debug

import (
	"bufio"
	"fmt"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"os"
	"time"
)

// BackTrace runs gdb on pod to collect a backtrace of PID 1 and save to /tmp
func BackTrace(ctx *framework.ContextData, podName string) {

	// Reading backtrace's output
	log.Logf("-- saving backtrace for: %s", podName)
	kubeexec := framework.NewKubectlExecCommand(*ctx, podName, 20*time.Second, "bash", "-c", "gdb /usr/sbin/qdrouterd 1 -ex 'set pagination off' -ex 'thread apply all bt' -ex 'set confirm off' -ex 'quit'")
	output, err := kubeexec.Exec()

	// In case kubectl exec fails
	if err != nil {
		log.Logf("-- unable to collect backtrace: %s", err)
	}

	logFile := fmt.Sprintf("/tmp/ns_%s_pod_%s.backtrace.log", ctx.Namespace, podName)
	if len(os.Getenv("LOGFOLDER")) > 0 {
		logFile = fmt.Sprintf("%s/ns_%s_pod_%s.backtrace.log", os.Getenv("LOGFOLDER"), ctx.Namespace, podName)
	}
	log.Logf("-- writing backtrace logs to: %s", logFile)
	f, _ := os.Create(logFile)
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString(output)
	w.Flush()

}

// Runs pstack on PID 1 and save to /tmp
func Pstack(ctx *framework.ContextData, podName string) {

	// Reading pstack's output
	log.Logf("-- saving pstack for: %s", podName)
	kubeexec := framework.NewKubectlExecCommand(*ctx, podName, 20*time.Second, "bash", "-c", "pstack 1")
	output := kubeexec.ExecOrDie()

	// Iterate through lines\
	logFile := fmt.Sprintf("/tmp/ns_%s_pod_%s.pstack.log", ctx.Namespace, podName)
	if len(os.Getenv("LOGFOLDER")) > 0 {
		logFile = fmt.Sprintf("%s/ns_%s_pod_%s.pstack.log", os.Getenv("LOGFOLDER"), ctx.Namespace, podName)
	}
	log.Logf("-- writing pstack logs to: %s", logFile)
	f, _ := os.Create(logFile)
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString(output)
	w.Flush()

}
