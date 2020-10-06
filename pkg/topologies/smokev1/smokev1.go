package smokev1

import (
	"fmt"
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/qdr-shipshape/pkg/testcommon"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/deployment"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
    "os"
	"time"
)

const (
	nameIcInteriorEast = "interior-east"
	nameIcInteriorWest = "interior-west"
	nameEdgeEast1      = "edge-east-1"
	nameEdgeEast2      = "edge-east-2"
	nameEdgeWest1      = "edge-west-1"
	nameEdgeWest2      = "edge-west-2"
)

func CreateSmokeRouterOnlyTopology(config *testcommon.Config, builder framework.Builder) *SmokeRouterOnlyTopology {
	return &SmokeRouterOnlyTopology{
		FrameworkSmoke: builder.Build(),
		config: config,
	}
}

type SmokeRouterOnlyTopology struct {
	FrameworkSmoke           *framework.Framework
	IcInteriorEast           *v1alpha1.InterconnectSpec
	IcEdgeEast1, IcEdgeEast2 *v1alpha1.InterconnectSpec
	IcInteriorWest           *v1alpha1.InterconnectSpec
	IcEdgeWest1, IcEdgeWest2 *v1alpha1.InterconnectSpec
	config                   *testcommon.Config
}

func (s *SmokeRouterOnlyTopology) Deploy() error {
	ctx := s.FrameworkSmoke.GetFirstContext()

	//
	// Initializing the specs
	//
	s.initializeInteriors()
	s.initializeEdges()

	//
	// Deploy Interior Routers
	//
	deployInteriorMap := map[string]*v1alpha1.InterconnectSpec{
		nameIcInteriorEast: s.IcInteriorEast,
		nameIcInteriorWest: s.IcInteriorWest,
	}
	for name, spec := range deployInteriorMap {
		if err := deployInterconnect(ctx, name, spec); err != nil {
			return err
		}
	}

	//
	// Deploy Edge Routers
	//
	deployEdgeMap := map[string]*v1alpha1.InterconnectSpec{
		nameEdgeEast1: s.IcEdgeEast1,
		nameEdgeEast2: s.IcEdgeEast2,
		nameEdgeWest1: s.IcEdgeWest1,
		nameEdgeWest2: s.IcEdgeWest2,
	}

	// Deploy all interconnect instances
	for name, spec := range deployEdgeMap {
		if err := deployInterconnect(ctx, name, spec); err != nil {
			return err
		}
	}

	return nil
}

func (s *SmokeRouterOnlyTopology) ValidateDeployment() error {
	for _, icName := range s.InteriorRouterNames() {
		if err := s.validateNetwork(icName); err != nil {
			return err
		}
	}
	// validate all services have been created
	ctx := s.FrameworkSmoke.GetFirstContext()
	for _, svcName := range s.AllRouterNames() {
		ginkgo.By("Validating availability of service: " + svcName)
		_, err := ctx.WaitForService(svcName, time.Minute, time.Second * 10)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SmokeRouterOnlyTopology) AllowedProperties() []string {
	panic("implement me")
}

func (s *SmokeRouterOnlyTopology) AllRouterNames() []string {
	return []string{nameIcInteriorEast, nameIcInteriorWest, nameEdgeEast1, nameEdgeEast2, nameEdgeWest1, nameEdgeWest2}
}

func (s *SmokeRouterOnlyTopology) InteriorRouterNames() []string {
	return s.AllRouterNames()[0:2]
}

func (s *SmokeRouterOnlyTopology) EdgeRouterNames() []string {
	return s.AllRouterNames()[2:]
}

// validateNetwork retrieves router nodes from all interior pods
// as well as the number of edge connections to each. Then it
// validates number of interior nodes and edge connections match
// expected values.
func (s *SmokeRouterOnlyTopology) validateNetwork(interiorName string) error {

	ginkgo.By("Validating network on " + interiorName)
	ctx := s.FrameworkSmoke.GetFirstContext()

	podList, err := ctx.ListPodsForDeploymentName(interiorName)
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		// Expect that the 4 interior routers are showing up
		nodes, err := qdrmanagement.QdmanageQueryWithRetries(*ctx, pod.Name, entities.Node{}, 10,
			60, nil, func(es []entities.Entity, err error) bool {
				return err != nil || len(es) == 4
			})
		if err != nil {
			return err
		}
		nodesCount := len(nodes)
		if nodesCount != 4 {
			return fmt.Errorf("expected 4 interior nodes, found: %d", nodesCount)
		}
	}

	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Logf("retrieving edge connections from all pods of %s - attempt %d/%d", interiorName, attempt, maxAttempts)
		connectedEdges := 0
		for _, pod := range podList.Items {
			// Expect that edge routers are connected
			conns, err := qdrmanagement.QdmanageQuery(*ctx, pod.Name, entities.Connection{}, func(e entities.Entity) bool {
				c := e.(entities.Connection)
				return c.Role == "edge"
			})
			if err != nil {
				return err
			}
			connectedEdges += len(conns)
		}

		if connectedEdges == 2 {
			break
		} else if attempt == maxAttempts {
			return fmt.Errorf("expected 2 edge connections, found: %d", connectedEdges)
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}

func deployInterconnect(ctx *framework.ContextData, icName string, icSpec *v1alpha1.InterconnectSpec) error {
	// Deploying Interconnect using provided context
	if _, err := deployment.CreateInterconnectFromSpec(*ctx, icSpec.DeploymentPlan.Size, icName, *icSpec); err != nil {
		return err
	}
	// Wait for Interconnect deployment
	err := framework.WaitForDeployment(ctx.Clients.KubeClient, ctx.Namespace, icName, int(icSpec.DeploymentPlan.Size), framework.RetryInterval, framework.Timeout)
	return err
}

func (s *SmokeRouterOnlyTopology) initializeInteriors() {

	// Initializing the interior routers
	s.IcInteriorEast = s.defaultInteriorSpec()
    if os.Getenv("IMAGE_QDROUTERD_INTEROP") != "" {
        s.IcInteriorEast.DeploymentPlan.Image = os.Getenv("IMAGE_QDROUTERD_INTEROP")
    }

	// TODO Discuss about it.
	//      If we change it to deploy using two distinct namespaces, we will need
	//      to install certificates (secret) in the WEST interior namespace that is
	//      generated by EAST interior CA, otherwise it won't work as external
	//      Routes/Ingress won't be available.
	ctx := s.FrameworkSmoke.GetFirstContext()
	s.IcInteriorWest = s.defaultInteriorSpec()
	s.IcInteriorWest.InterRouterConnectors = []v1alpha1.Connector{
		{
			Host: interconnect.GetDefaultServiceName(nameIcInteriorEast, ctx.Namespace),
			Port: 55672,
		},
	}

}

func (s *SmokeRouterOnlyTopology) defaultInteriorSpec() *v1alpha1.InterconnectSpec {
	return &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      2,
			Image:     s.config.GetEnvProperty(testcommon.PropertyImageQPIDDispatch, "quay.io/interconnectedcloud/qdrouterd:latest"),
			Role:      "interior",
			Placement: "Any",
		},
	}
}

func (s *SmokeRouterOnlyTopology) initializeEdges() {
	ctx := s.FrameworkSmoke.GetFirstContext()
	s.IcEdgeEast1 = s.defaultEdgeSpec(nameIcInteriorEast, ctx.Namespace)
	s.IcEdgeEast2 = s.defaultEdgeSpec(nameIcInteriorEast, ctx.Namespace)
    // If we set the image for Interop mode
    if os.Getenv("IMAGE_QDROUTERD_INTEROP") != "" {
        s.IcEdgeEast1.DeploymentPlan.Image = os.Getenv("IMAGE_QDROUTERD_INTEROP")
        s.IcEdgeEast2.DeploymentPlan.Image = os.Getenv("IMAGE_QDROUTERD_INTEROP")
    }
	s.IcEdgeWest1 = s.defaultEdgeSpec(nameIcInteriorWest, ctx.Namespace)
	s.IcEdgeWest2 = s.defaultEdgeSpec(nameIcInteriorWest, ctx.Namespace)
}

func (s *SmokeRouterOnlyTopology) defaultEdgeSpec(interiorIcName, ns string) *v1alpha1.InterconnectSpec {
	return &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      1,
			Image:     s.config.GetEnvProperty(testcommon.PropertyImageQPIDDispatch, "quay.io/interconnectedcloud/qdrouterd:latest"),
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
