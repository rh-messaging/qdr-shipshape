package python

import (
	"bytes"
	"fmt"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"os"
	"os/exec"
	"strconv"
	"text/template"
)

func DeployPythonClientByYaml(namespace, clientName, pythonImg, command, url string, msgCount, msgSize, clientCounter, timeout int) error {

	type PythonClientStruct struct {
		Command       string
		ClientName    string
		ClientCounter int
		PythonImage   string
		MsgSize       int
		Timeout       int
		Url           string
		MsgCount      int
	}

	YAMLFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-*.yaml", clientName))
	if err != nil {
		os.Exit(1)
	}

	yamlTemplate, err := template.New("pythonClientYamlTemplate").Parse(`
apiVersion: v1
kind: Pod
metadata:
 labels:
  amqp-client-impl: {{ .Command }}
 name: {{ .ClientName }}-{{ .ClientCounter }}
spec:
 containers:
 - image: {{ .PythonImage }}
   imagePullPolicy: IfNotPresent
   name: {{ .ClientName }}-{{ .ClientCounter }}
   args:
   - /opt/client/{{ .Command }}
   command:
   - python3
   env:
   - name: MSG_SIZE
     value: "{{ .MsgSize }}"
   - name: AMQP_TIMEOUT
     value: "{{ .Timeout }}"
   - name: AMQP_URL
     value: {{ .Url }}
   - name: MSG_COUNT
     value: "{{ .MsgCount }}"
   securityContext:
     allowPrivilegeEscalation: false
     runAsNonRoot: true
     seccompProfile:
       type: "RuntimeDefault"
 restartPolicy: Never`)
	if err != nil {
		panic(err)
	}

	pythonClientElement := PythonClientStruct{
		Command:       command,
		ClientName:    clientName,
		ClientCounter: clientCounter,
		PythonImage:   pythonImg,
		MsgSize:       msgSize,
		Timeout:       timeout,
		Url:           url,
		MsgCount:      msgCount,
	}

	err = yamlTemplate.Execute(YAMLFile, pythonClientElement)
	if err != nil {
		panic(err)
	}

	err = YAMLFile.Close()
	if err != nil {
		panic(err)
	}

	YAMLfileName := YAMLFile.Name()
	cmdToExec := exec.Command("kubectl", "apply", "-f", YAMLfileName, "-n", namespace)
	var cmdOut bytes.Buffer
	var cmdError bytes.Buffer
	cmdToExec.Stdout = &cmdOut
	cmdToExec.Stderr = &cmdError

	if err := cmdToExec.Run(); err != nil {
		log.Logf("\n\nError while applying YAML for python-client %s : %s - STDOUT = %s - STDERR = %s\n\n", YAMLfileName, err, cmdOut, cmdError)
		panic(err)
	}
	return nil
}

func DeployPythonClient(ctx *framework.ContextData, icName, clientName, address string, debug bool, command PythonClientCmd, numberOfClients, msgCount, msgSize, timeout int) []*PythonClient {

	var res []*PythonClient
	url := fmt.Sprintf("amqp://%s:5672/%s", interconnect.GetDefaultServiceName(icName, ctx.Namespace), address)
	log.Logf("Deploying client: [%s] using URL = [%s]", clientName, url)

	for clientCounter := 0; clientCounter < numberOfClients; clientCounter++ {
		DeployPythonClientByYaml(ctx.Namespace, clientName, PythonClientImage, string(command), url, msgCount, msgSize, clientCounter, timeout)

		builder := NewClientBuilder(clientName+"-"+strconv.Itoa(clientCounter), command, *ctx, url)
		builder.Timeout(timeout)
		c := builder.Build()
		res = append(res, c)
	}

	// Returning deployed receivers
	return res
}
