package helpers

import (
	"fmt"
	"log/slog"
	"net/netip"
	"strconv"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func GetModuleNetworkConfig() *network.NetworkingConfig {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{constant.NetworkID: {
			NetworkID: constant.NetworkID,
			Aliases:   []string{constant.NetworkAlias},
		}},
	}
}

func GetPlatform() *v1.Platform {
	return &v1.Platform{}
}

func GetRestartPolicy() *container.RestartPolicy {
	return &container.RestartPolicy{
		Name: container.RestartPolicyAlways,
	}
}

func GetConfigSidecarCmd(cmd []string) []string {
	if len(cmd) > 0 {
		return cmd
	}

	return nil
}

func GetSidecarName(moduleName string) string {
	return fmt.Sprintf("%s-sc", moduleName)
}

// TCPPort converts a port number to a network.Port with the default TCP protocol
func TCPPort(port int) network.Port {
	return network.MustParsePort(strconv.Itoa(port))
}

func CreateExposedPorts(privateServerPort int) *network.PortSet {
	portSet := network.PortSet{
		TCPPort(privateServerPort):                       {},
		network.MustParsePort(constant.PrivateDebugPort): {},
	}

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int, privateServerPort int) *network.PortMap {
	serverPortBinding := []network.PortBinding{{
		HostIP:   netip.MustParseAddr(constant.HostIP),
		HostPort: strconv.Itoa(hostServerPort),
	}}
	debugPortBinding := []network.PortBinding{{
		HostIP:   netip.MustParseAddr(constant.HostIP),
		HostPort: strconv.Itoa(hostServerDebugPort),
	}}
	portMap := network.PortMap{
		TCPPort(privateServerPort):                       serverPortBinding,
		network.MustParsePort(constant.PrivateDebugPort): debugPortBinding,
	}

	return &portMap
}

func CreateResources(isModule bool, r map[string]any) *container.Resources {
	if len(r) == 0 {
		return createDefaultResources(isModule)
	}

	return &container.Resources{
		CPUCount:          GetIntOrDefault(r, field.ModuleResourceCpuCountEntry, constant.ModuleCPU),
		MemoryReservation: ConvertMemory(MibToBytes, GetIntOrDefault(r, field.ModuleResourceMemoryReservationEntry, constant.ModuleMemoryReservation)),
		Memory:            ConvertMemory(MibToBytes, GetIntOrDefault(r, field.ModuleResourceMemoryEntry, constant.ModuleMemory)),
		MemorySwap:        ConvertMemory(MibToBytes, GetIntOrDefault(r, field.ModuleResourceMemorySwapEntry, constant.ModuleSwap)),
		OomKillDisable:    BoolPtr(GetBoolOrDefault(r, field.ModuleResourceOomKillDisableEntry, false)),
	}
}

func createDefaultResources(isModule bool) *container.Resources {
	if isModule {
		return &container.Resources{
			CPUCount:          constant.ModuleCPU,
			MemoryReservation: ConvertMemory(MibToBytes, constant.ModuleMemoryReservation),
			Memory:            ConvertMemory(MibToBytes, constant.ModuleMemory),
			MemorySwap:        ConvertMemory(MibToBytes, constant.ModuleSwap),
			OomKillDisable:    BoolPtr(false),
		}
	}

	return &container.Resources{
		CPUCount:          constant.SidecarCPU,
		MemoryReservation: ConvertMemory(MibToBytes, constant.SidecarMemoryReservation),
		Memory:            ConvertMemory(MibToBytes, constant.SidecarMemory),
		MemorySwap:        ConvertMemory(MibToBytes, constant.SidecarSwap),
		OomKillDisable:    BoolPtr(false),
	}
}

func AppendRequiredContainers(actionName string, containers []string, backendModules map[string]any) []string {
	if IsModuleEnabled(constant.ModSearchModule, backendModules) {
		containers = append(containers, constant.OpenSearchContainer)
	}
	if IsModuleEnabled(constant.ModDataExportWorkerModule, backendModules) {
		extraContainers := []string{constant.MinIOContainer, constant.CreateBucketsContainer, constant.FTPServerContainer}
		containers = append(containers, extraContainers...)
	}
	if len(containers) > 0 {
		slog.Info(actionName, "text", "Retrieved required containers", "containers", containers)
	}

	return containers
}
