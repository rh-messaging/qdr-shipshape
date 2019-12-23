package interioredge

import (
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/messaging"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/deployment"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/operators"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	FrameworkSmoke *framework.Framework

	IcInteriorEast           *v1alpha1.InterconnectSpec
	IcEdgeEast1, IcEdgeEast2 *v1alpha1.InterconnectSpec

	IcInteriorWest           *v1alpha1.InterconnectSpec
	IcEdgeWest1, IcEdgeWest2 *v1alpha1.InterconnectSpec

	ConfigMap *v1.ConfigMap
	AllRouterNames = []string{NameIcInteriorEast, NameIcInteriorWest, NameEdgeEast1, NameEdgeEast2, NameEdgeWest1, NameEdgeWest2}
)

const (
	NameIcInteriorEast = "interior-east"
	NameIcInteriorWest = "interior-west"
	NameEdgeEast1      = "edge-east-1"
	NameEdgeEast2      = "edge-east-2"
	NameEdgeWest1      = "edge-west-1"
	NameEdgeWest2      = "edge-west-2"
	ConfigMapName      = "messaging-files"
)

// Creates a unique namespace prefixed as "e2e-tests-smoke"
var _ = ginkgo.BeforeEach(func() {
	// Initializes using only Qdr Operator
	FrameworkSmoke = framework.NewFrameworkBuilder("smoke").WithBuilders(operators.SupportedOperators[operators.OperatorTypeQdr]).Build()
})

// Initializes the Interconnect (CR) specs to be deployed
var _ = ginkgo.JustBeforeEach(func() {

	ctx := FrameworkSmoke.GetFirstContext()

	// Generates a config map with messaging files (content) to be
	// used by the AMQP QE Clients
	generateMessagingFilesConfigMap()

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
	validateNetwork(NameIcInteriorEast)
	validateNetwork(NameIcInteriorWest)
})

// generateMessagingFilesConfigMap creates a new config map that holds messaging
// files to be used by QE clients. It generates a 1kb, 100kb and 500kb files.
// Note: there is a threshold on the ConfigMap size of 1mb. If larger messages are
//       needed within QE Clients, we should change container init strategy when
//       defining the pod, so it downloads files during initialization (not sure if
//       that is a good idea.
func generateMessagingFilesConfigMap() {
	var err error
	ctx := FrameworkSmoke.GetFirstContext()
	ConfigMap, err = ctx.Clients.KubeClient.CoreV1().ConfigMaps(ctx.Namespace).Create(&v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name: ConfigMapName,
		},
		Data: map[string]string{
			"small-message.txt":  messaging.GenerateMessageContent("ThisIsARepeatableMessage", 1024),
			"medium-message.txt": messaging.GenerateMessageContent("ThisIsARepeatableMessage", 1024*100),
			"large-message.txt":  messaging.GenerateMessageContent("ThisIsARepeatableMessage", 1024*500),
		},
	})
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(ConfigMap).NotTo(gomega.BeNil())
}

// validateNetwork retrieves router nodes from all interior pods
// as well as the number of edge connections to each. Then it
// validates number of interior nodes and edge connections match
// expected values.
func validateNetwork(interiorName string) {

	ginkgo.By("Validating network on " + interiorName)
	ctx := FrameworkSmoke.GetFirstContext()

	podList, err := ctx.ListPodsForDeploymentName(interiorName)
	gomega.Expect(err).To(gomega.BeNil())

	connectedEdges := 0
	for _, pod := range podList.Items {
		// Expect that the 4 interior routers are showing up
		nodes, err := qdrmanagement.QdmanageQueryWithRetries(*ctx, pod.Name, entities.Node{}, 10,
			60, nil, func(es []entities.Entity, err error) bool {
				return err != nil || len(es) == 4
			})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(len(nodes)).To(gomega.Equal(4))

		// Expect that edge routers are connected
		conns, err := qdrmanagement.QdmanageQuery(*ctx, pod.Name, entities.Connection{}, func(e entities.Entity) bool {
			c := e.(entities.Connection)
			return c.Role == "edge"
		})
		gomega.Expect(err).To(gomega.BeNil())
		connectedEdges += len(conns)
	}
	gomega.Expect(connectedEdges).To(gomega.Equal(2))
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
