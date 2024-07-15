package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

const (
	ApplicationsPort       = 9901
	TenantsPort            = 9902
	TenantEntitlementsPort = 9903
)

type RegistryModule struct {
	Id     string `json:"id"`
	Action string `json:"action"`

	Name    string
	Version string
}

type RegistryModules []RegistryModule

type Applications struct {
	ApplicationDescriptors []map[string]interface{} `json:"applicationDescriptors"`
	TotalRecords           int                      `json:"totalRecords"`
}

func DeregisterModules(commandName string, moduleName string, enableDebug bool) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), nil)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
		panic(err)
	}

	if enableDebug {
		DumpHttpRequest(commandName, req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer resp.Body.Close()

	if enableDebug {
		DumpHttpResponse(commandName, resp)
	}

	var apps Applications

	err = json.NewDecoder(resp.Body).Decode(&apps)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
		panic(err)
	}

	if apps.TotalRecords == 0 {
		slog.Info(commandName, SecondaryMessageKey, "No deployed module applications were found")
	}

	if apps.TotalRecords > 0 {
		for _, v := range apps.ApplicationDescriptors {
			id := v["id"].(string)

			if moduleName != "" {
				moduleNameFiltered := ModuleIdRegexp.ReplaceAllString(id, `$1`)

				if moduleNameFiltered[strings.LastIndex(moduleNameFiltered, "-")] == 45 {
					moduleNameFiltered = moduleNameFiltered[:strings.LastIndex(moduleNameFiltered, "-")]
				}

				if moduleNameFiltered != moduleName {
					continue
				}
			}

			slog.Info(commandName, "Deregistering application", id)

			delAppReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, fmt.Sprintf("/applications/%s", id)), nil)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			if enableDebug {
				DumpHttpRequest(commandName, delAppReq)
			}

			delAppResp, err := http.DefaultClient.Do(delAppReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer delAppResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, delAppResp)
			}
		}
	}
}

func RegisterModules(commandName string, enableDebug bool, dto *RegisterModuleDto) {
	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, SecondaryMessageKey, fmt.Sprintf("Registering %s registry modules", registryName))

		for moduleIndex, module := range registryModules {
			module.Name = ModuleIdRegexp.ReplaceAllString(module.Id, `$1`)
			module.Version = ModuleIdRegexp.ReplaceAllString(module.Id, `$2$3`)

			if module.Name[strings.LastIndex(module.Name, "-")] == 45 {
				module.Name = module.Name[:strings.LastIndex(module.Name, "-")]
			}

			registryModules[moduleIndex] = module

			_, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			moduleNameEnv := EnvNameRegexp.ReplaceAllString(strings.ToUpper(module.Name), `_`)

			_, err := dto.CacheFileModuleEnvPointer.WriteString(fmt.Sprintf("export %s_VERSION=%s\r\n", moduleNameEnv, module.Version))
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "moduleEnvVarsFile.WriteString error")
				panic(err)
			}

			var moduleDescriptorsUrl string

			if registryName == "folio" {
				moduleDescriptorsUrl = fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls["folio"], module.Id)
			} else {
				moduleDescriptorsUrl = fmt.Sprintf("%s/descriptors/%s.json", dto.RegistryUrls["eureka"], module.Id)
			}

			moduleDescriptorsReq, err := http.NewRequest(http.MethodGet, moduleDescriptorsUrl, nil)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			if enableDebug {
				DumpHttpRequest(commandName, moduleDescriptorsReq)
			}

			moduleDescriptorsResp, err := http.DefaultClient.Do(moduleDescriptorsReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer moduleDescriptorsResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, moduleDescriptorsResp)
			}

			var moduleDescriptors interface{}

			err = json.NewDecoder(moduleDescriptorsResp.Body).Decode(&moduleDescriptors)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
				panic(err)
			}

			dto.ModuleDescriptorsMap[module.Id] = moduleDescriptors

			var appModules []map[string]string
			appModules = append(appModules, map[string]string{"name": module.Name, "version": module.Version})

			appBytes, err := json.Marshal(map[string]interface{}{
				"id":                module.Id,
				"version":           module.Version,
				"name":              module.Name,
				"description":       "Deployed by Eureka CLI",
				"modules":           appModules,
				"moduleDescriptors": moduleDescriptors,
			})
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
				panic(err)
			}

			if enableDebug {
				fmt.Println("### Dumping HTTP Request Body")
				fmt.Println(string(appBytes))
			}

			postAppReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/applications"), bytes.NewBuffer(appBytes))
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			postAppReq.Header.Add("Content-Type", ApplicationJson)

			if enableDebug {
				DumpHttpRequest(commandName, postAppReq)
			}

			postAppResp, err := http.DefaultClient.Do(postAppReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer postAppResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, postAppResp)
			}

			var appDiscModules []map[string]string
			appDiscModules = append(appDiscModules, map[string]string{
				"id":       module.Id,
				"name":     module.Name,
				"version":  module.Version,
				"location": fmt.Sprintf("http://%s.eureka:%s", module.Name, ServerPort),
			})

			appDiscInfoBytes, err := json.Marshal(map[string]interface{}{
				"discovery": appDiscModules,
			})
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
				panic(err)
			}

			if enableDebug {
				fmt.Println("### Dumping HTTP Request Body")
				fmt.Println(string(appDiscInfoBytes))
			}

			postAppDiscReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(DockerInternalUrl, ApplicationsPort, "/modules/discovery"), bytes.NewBuffer(appDiscInfoBytes))
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			postAppDiscReq.Header.Add("Content-Type", ApplicationJson)

			if enableDebug {
				DumpHttpRequest(commandName, postAppDiscReq)
			}

			postAppDiscResp, err := http.DefaultClient.Do(postAppDiscReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer postAppDiscResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, postAppResp)
			}

			slog.Info(commandName, "Registered module", fmt.Sprintf("%s %d %d", module.Id, postAppResp.StatusCode, postAppDiscResp.StatusCode))
		}
	}
}

func CreateTenants(commandName string, enableDebug bool) {
	tenants := viper.GetStringSlice(TenantConfigKey)

	getTenantsReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), nil)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
		panic(err)
	}

	if enableDebug {
		DumpHttpRequest(commandName, getTenantsReq)
	}

	getTenantsResp, err := http.DefaultClient.Do(getTenantsReq)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer getTenantsResp.Body.Close()

	if enableDebug {
		DumpHttpResponse(commandName, getTenantsResp)
	}

	var foundTenantsMap map[string]interface{}

	err = json.NewDecoder(getTenantsResp.Body).Decode(&foundTenantsMap)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
		panic(err)
	}

	foundTenants := foundTenantsMap["tenants"].([]interface{})

	for _, value := range foundTenants {
		mapEntry := value.(map[string]interface{})

		name := mapEntry["name"].(string)

		if slices.Contains(tenants, name) {
			id := mapEntry["id"].(string)

			delTenantReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(DockerInternalUrl, TenantsPort, fmt.Sprintf("/tenants/%s", id)), nil)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			if enableDebug {
				DumpHttpRequest(commandName, delTenantReq)
			}

			delTenantResp, err := http.DefaultClient.Do(delTenantReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer delTenantResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, delTenantResp)
			}

			slog.Info(commandName, SecondaryMessageKey, fmt.Sprintf("Removed tenant %s", name))

			break
		}
	}

	for _, tenant := range tenants {
		tenantBytes, err := json.Marshal(map[string]string{
			"name":        tenant,
			"description": "Default_tenant",
		})
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
			panic(err)
		}

		if enableDebug {
			fmt.Println("### Dumping HTTP Request Body")
			fmt.Println(string(tenantBytes))
		}

		tenantReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(DockerInternalUrl, TenantsPort, "/tenants"), bytes.NewBuffer(tenantBytes))
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
			panic(err)
		}

		tenantReq.Header.Add("Content-Type", ApplicationJson)

		if enableDebug {
			DumpHttpRequest(commandName, tenantReq)
		}

		tenantResp, err := http.DefaultClient.Do(tenantReq)
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
			panic(err)
		}
		defer tenantResp.Body.Close()

		if enableDebug {
			DumpHttpResponse(commandName, tenantResp)
		}

		slog.Info(commandName, SecondaryMessageKey, fmt.Sprintf("Created tenant %s", tenant))
	}
}
