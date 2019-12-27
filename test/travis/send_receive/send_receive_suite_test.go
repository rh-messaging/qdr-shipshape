package send_receive_test

import (
	"github.com/rh-messaging/qdr-shipshape/pkg/testcommon"
	"testing"
)

func TestMain(m *testing.M) {
	testcommon.Initialize(m)
}

func TestInterioredge(t *testing.T) {
	testcommon.RunSpecs(t, "interioredge", "Smoke Interior Edge Suite")
}
