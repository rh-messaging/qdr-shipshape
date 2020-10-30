package interior

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"os"
)

var _ = Describe("Exchange MultiCast messages across all nodes", func() {

	const (
		totalSmall  = 100000
		totalMedium = 10000
		totalLarge  = 1000
	)

	var (
		allRouterNames = TopologySmoke.AllRouterNames()
	)

	var testSufix string

	if os.Getenv("IMAGE_QDROUTERD_INTEROP") != "" {
		testSufix = " - Using Interoperability mode"
	}

	It(fmt.Sprintf("exchanges %d small messages with 1kb using senders and receivers across all router nodes%s", totalSmall, testSufix), func() {
		runSmokeTest("multicast/smoke/interior", totalSmall, 1024, allRouterNames)
	})

	It(fmt.Sprintf("exchanges %d medium messages with 100kb using senders and receivers across all router nodes%s", totalMedium, testSufix), func() {
		runSmokeTest("multicast/smoke/interior", totalMedium, 1024*100, allRouterNames)
	})

	It(fmt.Sprintf("exchanges %d large messages with 500kb using senders and receivers across all router nodes%s", totalLarge, testSufix), func() {
		runSmokeTest("multicast/smoke/interior", totalLarge, 1024*500, allRouterNames)
	})

})
