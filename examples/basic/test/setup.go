package test

import (
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/deployment"
	"github.com/rh-messaging/shipshape/pkg/framework"
)

var (
	Framework *framework.Framework
)

const (
	InterconnectName = "basic-ic"
)

// Creates a new instance of the Framework before each test
var _ = ginkgo.BeforeEach(func() {
	// Creates an instance of the framework
	Framework = framework.NewFrameworkBuilder("examples-basic").Build()
	ctx := Framework.GetFirstContext()

	// Deploys a minimum interconnect
	ic, err := deployment.CreateInterconnectFromSpec(*ctx, 1, InterconnectName, v1alpha1.InterconnectSpec{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(ic).NotTo(gomega.BeNil())

	// Wait till interconnect gets deployed
	err = framework.WaitForDeployment(ctx.Clients.KubeClient, ctx.Namespace, InterconnectName, int(1), framework.RetryInterval, framework.Timeout)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

})

// Destroys after each
var _ = ginkgo.AfterEach(func() {
	Framework.AfterEach()
})

