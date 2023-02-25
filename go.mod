module github.com/rh-messaging/qdr-shipshape

go 1.13

require (
	github.com/go-ini/ini v1.51.1
	github.com/interconnectedcloud/qdr-operator v0.0.0-20200515123116-0468fa0ffb7a
	github.com/onsi/ginkgo v1.12.3
	github.com/onsi/gomega v1.10.1
	github.com/rh-messaging/shipshape v0.2.5
	github.com/smartystreets/goconvey v1.6.4 // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery v0.15.7
	k8s.io/klog v1.0.0
)

replace bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20120707110453-a547fc61f48d
