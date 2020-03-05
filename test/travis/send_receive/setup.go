package send_receive

import (
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/messaging"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/deployment"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/operators"
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	IcInteriorRouterName = "interior"
	ConfigMapName        = "messaging-files"
)

var IcInteriorRouterSpec *v1alpha1.InterconnectSpec = &v1alpha1.InterconnectSpec{
	DeploymentPlan: v1alpha1.DeploymentPlanType{
		Size:      1,
		Image:     "quay.io/interconnectedcloud/qdrouterd:latest",
		Role:      "interior",
		Placement: "Any",
	},
}

var (
	travisFramework *framework.Framework
	ConfigMap       *v1.ConfigMap
)

var _ = BeforeEach(func() {
	// Initializes using only Qdr Operator
	travisFramework = framework.NewFrameworkBuilder("travis").WithBuilders(operators.SupportedOperators[operators.OperatorTypeQdr]).Build()

	messaging.GenerateSmallMediumLargeMessagesConfigMap(travisFramework, ConfigMapName)

	deployInterconnect()
	By("Validating router created")
	time.Sleep(5 * time.Second) //wait for router to start
	validateNetwork()
})

func validateNetwork() {
	var delayS int = 10
	var timeoutS int = 60
	ctx := travisFramework.GetFirstContext()

	podList, err := ctx.ListPodsForDeploymentName(IcInteriorRouterName)
	Expect(err).To(BeNil())
	Expect(len(podList.Items)).To(Equal(1))
	pod := podList.Items[0]

	nodes, err := qdrmanagement.QdmanageQueryWithRetries(*ctx, pod.Name, entities.Node{}, delayS,
		timeoutS, nil, func(es []entities.Entity, err error) bool {
			return err != nil || len(es) == 1 //what does this mean?
		})

	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(1))
}

func deployInterconnect() {
	// Deploying Interconnect using provided context
	ctx := travisFramework.GetFirstContext()

	ic, err := deployment.CreateInterconnectFromSpec(*ctx, IcInteriorRouterSpec.DeploymentPlan.Size,
		IcInteriorRouterName, *IcInteriorRouterSpec)

	Expect(err).To(BeNil())
	Expect(ic).NotTo(BeNil())

	// Wait for Interconnect deployment
	err = framework.WaitForDeployment(ctx.Clients.KubeClient, ctx.Namespace,
		IcInteriorRouterName, int(IcInteriorRouterSpec.DeploymentPlan.Size),
		framework.RetryInterval, framework.Timeout)
	Expect(err).To(BeNil())
}
