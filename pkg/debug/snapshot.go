package debug

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement"
	"github.com/rh-messaging/shipshape/pkg/apps/qdrouterd/qdrmanagement/entities"
	"github.com/rh-messaging/shipshape/pkg/framework"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"os"
	"sync"
	"time"
)

const (
	SnapshotDelay = 30
)

// snapshotRouters executes qdmanage query at a regular internal against all router deployment names
//                 provided, retrieving the provided entity and using the filter function
//				   (use in debug mode only)
func SnapshotRouters(allRouterNames []string, ctx *framework.ContextData, entity entities.Entity, filter func(entities.Entity) bool, wg *sync.WaitGroup, done *bool) {
	for _, routerName := range allRouterNames {
		pods, err := ctx.ListPodsForDeploymentName(routerName)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		for _, pod := range pods.Items {
			wg.Add(1)
			go func(podName string) {
				defer wg.Done()
				logFile := fmt.Sprintf("/tmp/ns_%s_pod_%s_linkStats.log", ctx.Namespace, podName)
				if len(os.Getenv("LOGFOLDER")) > 0 {
					logFile = fmt.Sprintf("%s/ns_%s_pod_%s_linkStats.log", os.Getenv("LOGFOLDER"), ctx.Namespace, podName)
				}
				log.Logf("Saving linkStatus to: %s", logFile)
				f, _ := os.Create(logFile)
				defer f.Close()
				w := bufio.NewWriter(f)

				// Once all senders/receivers stopped, this should become false
				for !*done {
					// Retrieve links from router for the related address
					links, err := qdrmanagement.QdmanageQuery(*ctx, podName, entity, filter)
					if err != nil {
						log.Logf("Error querying router: %v", err)
					}

					if len(links) > 0 {
						for _, e := range links {
							stat, err := json.Marshal(e)
							if err != nil {
								log.Logf("Error marshalling link: %v", err)
								continue
							}
							w.Write(stat)
							w.WriteString("\n")
							w.Flush()
						}
					}

					time.Sleep(SnapshotDelay)
				}

				log.Logf("Finished snapshoting router - Logfile: %s", logFile)
			}(pod.Name)
		}
	}
}
