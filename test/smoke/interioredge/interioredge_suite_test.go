package interioredge_test

import (
	"github.com/rh-messaging/qdr-shipshape/pkg/testcommon"
    "os"
    "github.com/rh-messaging/shipshape/pkg/framework/log"
	"testing"
)

func TestMain(m *testing.M) {
	testcommon.Initialize(m)
}

func TestInterioredge(t *testing.T) {
    if os.Getenv("IMAGE_QDROUTERD_INTEROP") != "" {
        log.Logf("Interoperability mode is enabled")
    }
	testcommon.RunSpecs(t, "interioredge", "Smoke Interior Edge Suite")
}
