/*
Copyright © 2025 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const listModuleVersionsCommand string = "List Module Versions"

// listModuleVersionsCmd represents the listModuleVersions command
var listModuleVersionsCmd = &cobra.Command{
	Use:   "listModuleVersions",
	Short: "List module versions",
	Long:  `List module versions using the registry URL from config.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListModuleVersions()
	},
}

func ListModuleVersions() {
	registryUrl := viper.GetString(internal.RegistryUrlKey)
	if id != "" {
		getModuleDescritorById(registryUrl)
		return
	}
	listModuleVersions(registryUrl)
}

func getModuleDescritorById(registryUrl string) {
	requestUrl := fmt.Sprintf("%s/_/proxy/modules/%s", registryUrl, id)
	internal.DoGetDecodeReturnAny(listModuleVersionsCommand, requestUrl, true, true, map[string]string{})
}

func listModuleVersions(registryUrl string) {
	requestUrl := fmt.Sprintf("%s/_/proxy/modules", registryUrl)
	response := internal.DoGetDecodeReturnAny(listModuleVersionsCommand, requestUrl, enableDebug, true, map[string]string{})
	for _, value := range response.([]any) {
		mapEntry := value.(map[string]any)
		if strings.Contains(mapEntry["id"].(string), moduleName) {
			fmt.Println(mapEntry["id"])
		}
	}
}

func init() {
	rootCmd.AddCommand(listModuleVersionsCmd)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&moduleName, "moduleName", "m", "", "Module name, e.g. mod-users (required)")
	listModuleVersionsCmd.PersistentFlags().StringVarP(&id, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021")
	listModuleVersionsCmd.MarkPersistentFlagRequired("moduleName")
}
