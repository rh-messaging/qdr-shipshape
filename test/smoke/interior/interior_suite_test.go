package interior_test

import (
	"github.com/rh-messaging/qdr-shipshape/pkg/testcommon"
	"github.com/rh-messaging/shipshape/pkg/framework/log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testcommon.Initialize(m)
}

func TestInterior(t *testing.T) {
	if os.Getenv("IMAGE_QDROUTERD_INTEROP") != "" {
		log.Logf("Interoperability mode is enabled")
	}
	testcommon.RunSpecs(t, "interior", "Smoke Interior Only Suite")
}
