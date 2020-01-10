package topologies

type Topology interface {
	Deploy() error
	ValidateDeployment() error
	AllowedProperties() []string
	AllRouterNames() []string
	InteriorRouterNames() []string
	EdgeRouterNames() []string
}
