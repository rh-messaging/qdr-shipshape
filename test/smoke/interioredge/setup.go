package interioredge

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/messaging"
	"github.com/rh-messaging/qdr-shipshape/pkg/topologies/smokev1"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/operators"
	v1 "k8s.io/api/core/v1"
)

var (
	TopologySmoke *smokev1.SmokeRouterOnlyTopology
	ConfigMap     *v1.ConfigMap
)

const (
	ConfigMapName      = "messaging-files"
)

// Creates a unique namespace prefixed as "e2e-tests-smoke"
var _ = ginkgo.BeforeEach(func() {
	// Initializes using only Qdr Operator
	builder := framework.
		NewFrameworkBuilder("smoke").
		WithBuilders(operators.SupportedOperators[operators.OperatorTypeQdr])

	// Create framework instance and topology
	TopologySmoke = smokev1.CreateSmokeRouterOnlyTopology(builder)

	// Generates a config map with messaging files (content) to be
	// used by the AMQP QE Clients
	ConfigMap = messaging.GenerateSmallMediumLargeMessagesConfigMap(TopologySmoke.FrameworkSmoke, ConfigMapName)
})

// Initializes the Interconnect (CR) specs to be deployed
var _ = ginkgo.JustBeforeEach(func() {

	// Deploying topology
	err := TopologySmoke.Deploy()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	//
	// Asserting that network is properly formed
	//
	err = TopologySmoke.ValidateDeployment()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

})

// After each test completes, run cleanup actions to save resources (otherwise resources will remain till
// all specs from this suite are done.
var _ = ginkgo.AfterEach(func() {
	TopologySmoke.FrameworkSmoke.AfterEach()
})
