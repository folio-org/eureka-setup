package action

// Action is a container that holds the state of the deployment
type Action struct {
	Name          string
	GatewayURL    string
	StartPort     int
	EndPort       int
	ReservedPorts []int
}

func New(name, gatewayURL string) *Action {
	return newGeneric(name, gatewayURL, 30000, 30999)
}

func NewCustom(name, gatewayURL string, startPort, endPort int) *Action {
	return newGeneric(name, gatewayURL, startPort, endPort)
}

func newGeneric(name, gatewayURL string, startPort, endPort int) *Action {
	return &Action{
		Name:          name,
		GatewayURL:    gatewayURL,
		StartPort:     startPort,
		EndPort:       endPort,
		ReservedPorts: []int{},
	}
}
