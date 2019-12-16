package interconnect

// GetDefaultServiceName returns the "default" generated service name
// for the given Interconnect Spec deployment
func GetDefaultServiceName(icName string, ns string) string {
	return icName + "." + ns + ".svc.cluster.local"
}
