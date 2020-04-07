package interioredge

import (
	"fmt"
	. "github.com/onsi/ginkgo"
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

	It(fmt.Sprintf("exchanges %d small messages with 1kb using senders and receivers across all router nodes", totalSmall), func() {
		runSmokeTest("multicast/smoke/interioredge", totalSmall, 1024, allRouterNames)
	})

	It(fmt.Sprintf("exchanges %d medium messages with 100kb using senders and receivers", totalMedium), func() {
		runSmokeTest("multicast/smoke/interioredge", totalMedium, 1024*100, allRouterNames)
	})

	It(fmt.Sprintf("exchanges %d large messages with 500kb using senders and receivers", totalLarge), func() {
		runSmokeTest("multicast/smoke/interioredge", totalLarge, 1024*500, allRouterNames)
	})

})
