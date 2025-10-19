package action

// Action is a container that holds the state of the deployment
type Action struct {
	Name          string
	StartPort     int
	EndPort       int
	ReservedPorts []int
}

func New(name string) *Action {
	return &Action{
		Name:          name,
		StartPort:     30000,
		EndPort:       30999,
		ReservedPorts: []int{},
	}
}

func NewCustom(name string, startPort, endPort int) *Action {
	return &Action{
		Name:          name,
		StartPort:     startPort,
		EndPort:       endPort,
		ReservedPorts: []int{},
	}
}
