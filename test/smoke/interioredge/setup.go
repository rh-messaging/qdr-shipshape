package interioredge

import (
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/rh-messaging/qdr-shipshape/pkg/spec/interconnect"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/operators"
)

var (
	FrameworkWest *framework.Framework
	FrameworkEast *framework.Framework

	IcInteriorEast *v1alpha1.InterconnectSpec
	IcEdgeEast1, IcEdgeEast2 *v1alpha1.InterconnectSpec

	IcInteriorWest *v1alpha1.InterconnectSpec
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

// Creates two distinct namespaces prefixed as "smoke-"
var _ = ginkgo.BeforeEach(func() {
	// Initializes using only Qdr Operator
	FrameworkEast = framework.NewFrameworkBuilder("smoke").WithBuilders(operators.SupportedOperators[operators.OperatorTypeQdr]).Build()
	FrameworkWest = framework.NewFrameworkBuilder("smoke").WithBuilders(operators.SupportedOperators[operators.OperatorTypeQdr]).Build()
})

// Initializes the Interconnect (CR) specs to be deployed
var _ = ginkgo.JustBeforeEach(func() {
	contextEast := FrameworkEast.GetFirstContext()
	contextWest := FrameworkWest.GetFirstContext()

	// Initializing the interior routers
	IcInteriorEast := &v1alpha1.InterconnectSpec{
		DeploymentPlan:        v1alpha1.DeploymentPlanType{
			Size: 2,
			Image: "quay.io/interconnectedcloud/qdrouterd:latest",
			Role: "interior",
			Placement: "Any",
		},
	}

	// TODO Discuss about it. Using as is (in two distinct namespaces) it
	//      won't work if cert-manager is installed as secrets won't be shared.
	//      If we need to support cert-manager, then we have to use a single namespace
	//      or add some further code to copy CA from one namespace to the other.
	IcInteriorWest := &v1alpha1.InterconnectSpec{
		DeploymentPlan:        v1alpha1.DeploymentPlanType{
			Size: 2,
			Image: "quay.io/interconnectedcloud/qdrouterd:latest",
			Role: "interior",
			Placement: "Any",
		},
		InterRouterConnectors: []v1alpha1.Connector{
			{
				Host:           interconnect.GetDefaultServiceName(NameIcInteriorEast, contextEast.Namespace),
				Port:           55672,
				RouteContainer: false,
				VerifyHostname: false,
				SslProfile:     "",
			},
		},
	}

})