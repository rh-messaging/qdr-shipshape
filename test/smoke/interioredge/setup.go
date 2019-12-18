package interioredge

import (
	"bytes"
	"fmt"
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/api/client/amqp/qeclients"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/deployment"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"github.com/rh-messaging/shipshape/pkg/framework/operators"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

var (
	FrameworkSmoke *framework.Framework

	IcInteriorEast           *v1alpha1.InterconnectSpec
	IcEdgeEast1, IcEdgeEast2 *v1alpha1.InterconnectSpec

	IcInteriorWest           *v1alpha1.InterconnectSpec
	IcEdgeWest1, IcEdgeWest2 *v1alpha1.InterconnectSpec
)

const (
	NameIcInteriorEast = "interior-east"
	NameIcInteriorWest = "interior-west"
	NameEdgeEast1      = "edge-east-1"
	NameEdgeEast2      = "edge-east-2"
	NameEdgeWest1      = "edge-west-1"
	NameEdgeWest2      = "edge-west-2"
)

// Creates a unique namespace prefixed as "e2e-tests-smoke"
var _ = ginkgo.BeforeEach(func() {
	// Initializes using only Qdr Operator
	FrameworkSmoke = framework.NewFrameworkBuilder("smoke").WithBuilders(operators.SupportedOperators[operators.OperatorTypeQdr]).Build()
})

// Initializes the Interconnect (CR) specs to be deployed
var _ = ginkgo.JustBeforeEach(func() {

	ctx := FrameworkSmoke.GetFirstContext()

	cfgMap, err := ctx.Clients.KubeClient.CoreV1().ConfigMaps(ctx.Namespace).Create(&v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name: "messaging-files",
		},
		Data: map[string]string{
			"small-message.txt":  generateMessageContent("ThisIsARepeatableMessage", 1024),
			"medium-message.txt": generateMessageContent("ThisIsARepeatableMessage", 1024 * 100),
			"large-message.txt":  generateMessageContent("ThisIsARepeatableMessage", 1024 * 500),
		},
	})
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(cfgMap).NotTo(gomega.BeNil())

	//
	// Initializing the specs
	//
	initializeInteriors()
	initializeEdges()

	//
	// Deploy Interior Routers
	//
	deployInterconnect(ctx, NameIcInteriorEast, IcInteriorEast)
	deployInterconnect(ctx, NameIcInteriorWest, IcInteriorWest)

	// Creating a python sender manually
	url := fmt.Sprintf("amqp://%s:5672/anycast/messaging", NameIcInteriorEast)
	sdr, err := qeclients.NewAmqpSender(qeclients.Python, "sender", *ctx, url, 1, "abcdef")
	pSdr := sdr.(*qeclients.AmqpPythonSender)
	pSdr.Pod.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{Name: "messaging-files", ReadOnly: true, MountPath: "/opt/messaging-files"},
	}
	pSdr.Pod.Spec.Volumes = []v1.Volume{
		{Name: "messaging-files", VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{Name:"messaging-files"},
			},
		}},
	}

	// workaround to change cmd line args
	pSdr.Pod.Spec.Containers[0].Args = []string{
		"--count",
		strconv.Itoa(10),
		"--timeout",
		strconv.Itoa(600),
		"--broker-url",
		url,
		"--msg-content-from-file",
		"/opt/messaging-files/medium-message.txt",
		"--log-msgs",
		"json",
		"--on-release",
		"retry",
	}
	err = pSdr.Deploy()
	gomega.Expect(err).To(gomega.BeNil())

	rcv, err := qeclients.NewAmqpReceiver(qeclients.Python, "receiver", *ctx, url, 10)
	pRcv := rcv.(*qeclients.AmqpPythonReceiver)
	pRcv.Pod.Spec.Containers[0].Args = []string{
		"--count",
		strconv.Itoa(10),
		"--timeout",
		strconv.Itoa(600),
		"--broker-url",
		url,
		"--log-msgs",
		"json",
	}
	err = pRcv.Deploy()
	gomega.Expect(err).To(gomega.BeNil())

	pSdr.Wait()
	pRcv.Wait()

	//log.Logf("\n\nGO CHECK LOGS:\n\nkubectl -n %s logs sender\n\n", ctx.Namespace)
	//time.Sleep(5 * time.Minute)
	//
	// Validating results
	senderResult := pSdr.Result()
	receiverResult := pRcv.Result()

	// Ensure results obtained
	gomega.Expect(senderResult).NotTo(gomega.BeNil())
	gomega.Expect(receiverResult).NotTo(gomega.BeNil())

	// Validate sent/received messages
	gomega.Expect(senderResult.Delivered).To(gomega.Equal(10))
	gomega.Expect(receiverResult.Delivered).To(gomega.Equal(10))

	//
	// Deploy Edge Routers
	//
	deployInterconnect(ctx, NameEdgeEast1, IcEdgeEast1)
	deployInterconnect(ctx, NameEdgeEast2, IcEdgeEast2)
	deployInterconnect(ctx, NameEdgeWest1, IcEdgeWest1)
	deployInterconnect(ctx, NameEdgeWest2, IcEdgeWest2)

	//
	// Asserting that network is properly formed
	//
	validateNetwork()
})

func generateMessageContent(pattern string, size int) string {
	var buf bytes.Buffer
	patLen := len(pattern)
	times := size / patLen
	rem := size % patLen
	for i := 0; i < times; i++ {
		buf.WriteString(pattern)
	}
	if rem > 0 {
		buf.WriteString(pattern[:rem])
	}
	return buf.String()
}

func validateNetwork() {

	ctx := FrameworkSmoke.GetFirstContext()

	podList, err := ctx.ListPodsForDeploymentName(NameIcInteriorEast)
	gomega.Expect(err).To(gomega.BeNil())

	for _, pod := range podList.Items {
		nodes, err := qdrmanagement.QdmanageQuery(*ctx, pod.Name, entities.Node{}, nil)
		gomega.Expect(err).To(gomega.BeNil())
		log.Logf("Nodes = %v", nodes)
	}
}

func deployInterconnect(ctx *framework.ContextData, icName string, icSpec *v1alpha1.InterconnectSpec) {
	// Deploying Interconnect using provided context
	ic, err := deployment.CreateInterconnectFromSpec(*ctx, icSpec.DeploymentPlan.Size, icName, *icSpec)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(ic).NotTo(gomega.BeNil())

	// Wait for Interconnect deployment
	err = framework.WaitForDeployment(ctx.Clients.KubeClient, ctx.Namespace, icName, int(icSpec.DeploymentPlan.Size), framework.RetryInterval, framework.Timeout)
	gomega.Expect(err).To(gomega.BeNil())
}

func initializeInteriors() {

	// Initializing the interior routers
	IcInteriorEast = defaultInteriorSpec()

	// TODO Discuss about it.
	//      If we change it to deploy using two distinct namespaces, we will need
	//      to install certificates (secret) in the WEST interior namespace that is
	//      generated by EAST interior CA, otherwise it won't work as external
	//      Routes/Ingress won't be available.
	ctx := FrameworkSmoke.GetFirstContext()
	IcInteriorWest = defaultInteriorSpec()
	IcInteriorWest.InterRouterConnectors = []v1alpha1.Connector{
		{
			Host:           interconnect.GetDefaultServiceName(NameIcInteriorEast, ctx.Namespace),
			Port:           55672,
			RouteContainer: false,
			VerifyHostname: false,
			SslProfile:     "",
		},
	}

}

func defaultInteriorSpec() *v1alpha1.InterconnectSpec {
	// TODO Define a standard configuration file that allows images to be customized
	return &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      2,
			Image:     "quay.io/interconnectedcloud/qdrouterd:latest",
			Role:      "interior",
			Placement: "Any",
		},
	}
}

func initializeEdges() {
	ctx := FrameworkSmoke.GetFirstContext()
	IcEdgeEast1 = defaultEdgeSpec(NameIcInteriorEast, ctx.Namespace)
	IcEdgeEast2 = defaultEdgeSpec(NameIcInteriorEast, ctx.Namespace)
	IcEdgeWest1 = defaultEdgeSpec(NameIcInteriorWest, ctx.Namespace)
	IcEdgeWest2 = defaultEdgeSpec(NameIcInteriorWest, ctx.Namespace)
}

func defaultEdgeSpec(interiorIcName, ns string) *v1alpha1.InterconnectSpec {
	return &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      1,
			Image:     "quay.io/interconnectedcloud/qdrouterd:latest",
			Role:      "edge",
			Placement: "Any",
		},
		EdgeConnectors: []v1alpha1.Connector{
			{
				Host: interconnect.GetDefaultServiceName(interiorIcName, ns),
				Port: 45672,
			},
		},
	}
}
