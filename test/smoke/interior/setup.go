package interior

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/testcommon"
	"github.com/rh-messaging/qdr-shipshape/pkg/topologies/smokeinteriorv1"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/operators"
)

var (
	TopologySmoke *smokeinteriorv1.SmokeInteriorRouterOnlyTopology
	//ConfigMap     *v1.ConfigMap
	Config        *testcommon.Config
)

const (
	ConfigMapName      = "messaging-files"
)

// Creates a unique namespace prefixed as "e2e-tests-smoke"
var _ = ginkgo.BeforeEach(func() {
	// Loading configuration (ini or environment)
	Config = testcommon.LoadConfig("smoke/interior")

	// Initializes using only Qdr Operator
	qdrOperator := operators.SupportedOperators[operators.OperatorTypeQdr]
	qdrOperator.WithImage(Config.GetEnvProperty(testcommon.PropertyImageQDrOperator,"quay.io/interconnectedcloud/qdr-operator:latest"))
	builder := framework.
		NewFrameworkBuilder("smoke").
		WithBuilders(qdrOperator)

	// Create framework instance and topology
	TopologySmoke = smokeinteriorv1.CreateSmokeInteriorRouterOnlyTopology(Config, builder)
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

// IsDebugEnabled returns true if DEBUG env var or property is true
func IsDebugEnabled() bool {
	return Config.GetEnvPropertyBool("DEBUG", false)
}
