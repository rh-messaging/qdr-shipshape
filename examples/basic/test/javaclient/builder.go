package javaclient

import (
	"github.com/rh-messaging/shipshape/pkg/framework"
	"strconv"
	"strings"
)

type ClientBuilder struct {
	client  *JavaClient
	command JavaClientCmd
	image   string
	env     map[string]string
}

func NewClientBuilder(name string, command JavaClientCmd, ctx framework.ContextData, url string) *ClientBuilder {
	sb := &ClientBuilder{
		client:  &JavaClient{},
		command: command,
		env:     map[string]string{},
	}

	// JavaClient initialization
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

// Address to use
func (s *ClientBuilder) Address(addr string) *ClientBuilder {
	s.EnvVar("AMQP_ADDR", addr)
	return s
}

// Timeout sets the amount of secs to wait before timing out
func (s *ClientBuilder) Timeout(timeout int) *ClientBuilder {
	s.client.Timeout = timeout
	s.EnvVar("AMQP_TIMEOUT", strconv.Itoa(timeout*1000))
	return s
}

func (s *ClientBuilder) Build() *JavaClient {

	// Preparing Pod, Container (commands and args) and etc
	podBuilder := framework.NewPodBuilder(s.client.Name, s.client.Context.Namespace)
	classNameSlice := strings.Split(string(s.command), ".")
	podBuilder.AddLabel("amqp-client-impl", classNameSlice[len(classNameSlice)-1])
	podBuilder.RestartPolicy("Never")

	//
	// Helps building the container for client pod
	//
	image := JavaClientImage
	if s.image != "" {
		image = s.image
	}
	cBuilder := framework.NewContainerBuilder(s.client.Name, image)
	cBuilder.WithCommands("java")
	cBuilder.AddArgs("-cp", "/opt/client/*")
	cBuilder.AddArgs(string(s.command))

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
