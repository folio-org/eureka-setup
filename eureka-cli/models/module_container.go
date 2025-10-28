package models

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Container struct {
	Name          string
	Image         string
	RegistryAuth  string
	Config        *container.Config
	HostConfig    *container.HostConfig
	NetworkConfig *network.NetworkingConfig
	Platform      *v1.Platform
	PullImage     bool
}

func NewModuleContainer(name string,
	image string,
	env []string,
	backendModule BackendModule,
	networkConfig *network.NetworkingConfig) *Container {
	return &Container{
		Name:  name,
		Image: image,
		Config: &container.Config{
			Image:        image,
			Hostname:     name,
			Env:          env,
			ExposedPorts: *backendModule.ModuleExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *backendModule.ModulePortBindings,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Resources:     backendModule.ModuleResources,
			Binds:         backendModule.ModuleVolumes,
		},
		NetworkConfig: networkConfig,
		Platform:      &v1.Platform{},
		PullImage:     backendModule.LocalDescriptorPath == "",
	}
}
