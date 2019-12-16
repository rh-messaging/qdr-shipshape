package testing

import (
	"flag"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/ginkgowrapper"
	"os"
	"testing"
)

// Initialize provides a common procedure for initializing shipshape
// as well as qdr-shipshape (ginkgo, command line flags and etc).
func Initialize(m *testing.M) {
	// Register shipshape flags
	framework.RegisterFlags()

	// Register qdr-shipshape flags
	RegisterFlags()

	// Parse command line flags
	flag.Parse()

	// Using shipshape fail handler
	gomega.RegisterFailHandler(ginkgowrapper.Fail)

	// Running the tests now
	os.Exit(m.Run())
}

// RunSpecs will use Ginkgo to run your test specs, and eventually setup
// report dir accordingly. The uniqueId is used to help composing the generated
// JUnit file name (when --report-dir is specified when running your tests).
func RunSpecs(t *testing.T, uniqueId string, description string) {
	// If any ginkgoReporter has been defined, use them.
	if framework.TestContext.ReportDir != "" {
		ginkgo.RunSpecsWithDefaultAndCustomReporters(t, description, generateReporter(uniqueId))
	} else {
		ginkgo.RunSpecs(t, description)
	}
}

// Common ginkgo initialization
// Before suite validation setup (happens only once per test suite)
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	// Unique initialization (node 1 only)
	return nil
}, func(data []byte) {
	// Initialization for each parallel node
}, TimeoutSetupTeardown)

// After suite validation teardown (happens only once per test suite)
var _ = ginkgo.SynchronizedAfterSuite(func() {
	// All nodes tear down
}, func() {
	// Node1 only tear down
	framework.RunCleanupActions(framework.AfterEach)
	framework.RunCleanupActions(framework.AfterSuite)
}, TimeoutSetupTeardown)
