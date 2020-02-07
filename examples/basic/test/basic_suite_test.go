package test_test

import (
	"github.com/onsi/gomega"
	"github.com/rh-messaging/qdr-shipshape/pkg/testcommon"
	"github.com/rh-messaging/shipshape/pkg/framework/ginkgowrapper"
	"testing"
)

func TestMain(m *testing.M) {
	testcommon.Initialize(m)
}

func TestBasic(t *testing.T) {
	gomega.RegisterFailHandler(ginkgowrapper.Fail)
	testcommon.RunSpecs(t, "basic","Basic Suite")
}
