package python

import (
	"github.com/rh-messaging/shipshape/pkg/framework"
	"strconv"
)

type ClientBuilder struct {
	client  *PythonClient
	command PythonClientCmd
	image   string
	env     map[string]string
}

func NewClientBuilder(name string, command PythonClientCmd, ctx framework.ContextData, url string) *ClientBuilder {
	sb := &ClientBuilder{
		client:  &PythonClient{},
		command: command,
		env:     map[string]string{},
	}

	// PythonClient initialization
	sb.client.Name = name
	sb.client.Context = ctx
	sb.client.Url = url

	// URL to be used within the container
	sb.EnvVar("AMQP_URL", url)

	return sb
}

func (s *ClientBuilder) EnableTracing() *ClientBuilder {
	s.EnvVar("PN_TRACE_FRM", "1")
	return s
}

func (s *ClientBuilder) EnvVar(variable, value string) *ClientBuilder {
	s.env[variable] = value
	return s
}

func (s *ClientBuilder) ImageCustom(image string) *ClientBuilder {
	s.image = image
	return s
}

func (s *ClientBuilder) Timeout(timeout int) *ClientBuilder {
	s.client.Timeout = timeout
	s.EnvVar("AMQP_TIMEOUT", strconv.Itoa(timeout))
	return s
}

func (s *ClientBuilder) Build() *PythonClient {

	// Preparing Pod, Container (commands and args) and etc
	podBuilder := framework.NewPodBuilder(s.client.Name, s.client.Context.Namespace)
	podBuilder.AddLabel("amqp-client-impl", string(s.command))
	podBuilder.RestartPolicy("Never")

	//
	// Helps building the container for client pod
	//
	image := PythonClientImage
	if s.image != "" {
		image = s.image
	}
	cBuilder := framework.NewContainerBuilder(s.client.Name, image)
	cBuilder.WithCommands("python3")
	cBuilder.AddArgs("/opt/client/" + string(s.command))

	//
	// Populating environment variables to the container
	//
	for variable, value := range s.env {
		cBuilder.EnvVar(variable, value)
	}

	// Retrieving container and adding to pod
	c := cBuilder.Build()
	podBuilder.AddContainer(c)
	pod := podBuilder.Build()
	s.client.Pod = pod

	return s.client

}
