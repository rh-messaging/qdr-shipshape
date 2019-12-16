package testing

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"k8s.io/klog"
	"os"
	"path"
)

// generateReporter returns a slice of ginkgo.Reporter if reportDir has been provided
func generateReporter(uniqueId string) []ginkgo.Reporter {
	var ginkgoReporters []ginkgo.Reporter

	// If report dir specified, create it
	if framework.TestContext.ReportDir != "" {
		if err := os.MkdirAll(framework.TestContext.ReportDir, 0755); err != nil {
			klog.Errorf("Failed creating report directory: %v", err)
		} else {
			ginkgoReporters = append(ginkgoReporters, reporters.NewJUnitReporter(
				path.Join(framework.TestContext.ReportDir,
					fmt.Sprintf("junit_%v%s%02d.xml",
						framework.TestContext.ReportPrefix,
						uniqueId,
						config.GinkgoConfig.ParallelNode))))
		}
	}

	return ginkgoReporters
}
