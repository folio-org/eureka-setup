package models

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func NewSidecarContainer(name string,
	image string,
	env []string,
	backendModule BackendModule,
	networkConfig *network.NetworkingConfig,
	resources *container.Resources) *Container {
	return &Container{
		Name:   name,
		Image:  image,
		Config: &container.Config{Image: image, Hostname: name, Env: env, ExposedPorts: *backendModule.SidecarExposedPorts},
		HostConfig: &container.HostConfig{
			PortBindings:  *backendModule.SidecarPortBindings,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Resources:     *resources,
		},
		NetworkConfig: networkConfig,
		Platform:      &v1.Platform{},
		PullImage:     false,
	}
}
