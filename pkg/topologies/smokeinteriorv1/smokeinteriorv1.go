package smokeinteriorv1

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
	"os"
	"time"
)

const (
	nameIcInteriorEast = "interior-east"
	nameIcInteriorWest = "interior-west"
)

func CreateSmokeInteriorRouterOnlyTopology(config *testcommon.Config, builder framework.Builder) *SmokeInteriorRouterOnlyTopology {
	return &SmokeInteriorRouterOnlyTopology{
		FrameworkSmoke: builder.Build(),
		config:         config,
	}
}

type SmokeInteriorRouterOnlyTopology struct {
	FrameworkSmoke           *framework.Framework
	IcInteriorEast           *v1alpha1.InterconnectSpec
	IcInteriorWest           *v1alpha1.InterconnectSpec
	config                   *testcommon.Config
}

func (s *SmokeInteriorRouterOnlyTopology) Deploy() error {
	ctx := s.FrameworkSmoke.GetFirstContext()

	//
	// Initializing the specs
	//
	s.initializeInteriors()

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

	return nil
}

func (s *SmokeInteriorRouterOnlyTopology) ValidateDeployment() error {
	for _, icName := range s.InteriorRouterNames() {
		if err := s.validateNetwork(icName); err != nil {
			return err
		}
	}
	// validate all services have been created
	ctx := s.FrameworkSmoke.GetFirstContext()
	for _, svcName := range s.AllRouterNames() {
		ginkgo.By("Validating availability of service: " + svcName)
		_, err := ctx.WaitForService(svcName, time.Minute, time.Second*10)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SmokeInteriorRouterOnlyTopology) AllowedProperties() []string {
	panic("implement me")
}

func (s *SmokeInteriorRouterOnlyTopology) AllRouterNames() []string {
	return []string{nameIcInteriorEast, nameIcInteriorWest}
}

func (s *SmokeInteriorRouterOnlyTopology) InteriorRouterNames() []string {
	return s.AllRouterNames()
}

func (s *SmokeInteriorRouterOnlyTopology) EdgeRouterNames() []string {
	return []string{}
}

// validateNetwork retrieves router nodes from all interior pods
// as well as the number of edge connections to each. Then it
// validates number of interior nodes and edge connections match
// expected values.
func (s *SmokeInteriorRouterOnlyTopology) validateNetwork(interiorName string) error {

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
				return len(es) == 4
			})
		if err != nil {
			return err
		}
		nodesCount := len(nodes)
		if nodesCount != 4 {
			return fmt.Errorf("expected 4 interior nodes, found: %d", nodesCount)
		}
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

func (s *SmokeInteriorRouterOnlyTopology) initializeInteriors() {

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

func (s *SmokeInteriorRouterOnlyTopology) defaultInteriorSpec() *v1alpha1.InterconnectSpec {
	return &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      2,
			Image:     s.config.GetEnvProperty(testcommon.PropertyImageQPIDDispatch, "quay.io/interconnectedcloud/qdrouterd:latest"),
			Role:      "interior",
			Placement: "Any",
		},
	}
}