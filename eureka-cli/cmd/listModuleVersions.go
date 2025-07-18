/*
Copyright © 2025 Open Library Foundation

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
	"log/slog"
	"net/http/httputil"
	"strings"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	listModuleVersionsCommand string = "List Module Versions"

	emptyLinePattern = "\r\n\r\n"
)

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
	if withId != "" {
		getModuleDescritorById(registryUrl)
		return
	}
	listModuleVersions(registryUrl)
}

func getModuleDescritorById(registryUrl string) {
	resp := internal.DoGetReturnResponse(listModuleVersionsCommand, fmt.Sprintf("%s/_/proxy/modules/%s", registryUrl, withId), withEnableDebug, true, map[string]string{})
	defer func() {
		_ = resp.Body.Close()
	}()

	if !withEnableDebug {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			slog.Error(listModuleVersionsCommand, internal.GetFuncName(), "httputil.DumpResponse error")
			panic(err)
		}

		idx := strings.Index(string(respBytes), emptyLinePattern)
		if idx == -1 {
			slog.Error(listModuleVersionsCommand, internal.GetFuncName(), "strings.Index() warning - response from %s does not contain an empty line")
			return
		}

		fmt.Println(string([]byte(respBytes[idx:])))
	}
}

func listModuleVersions(registryUrl string) {
	resp := internal.DoGetDecodeReturnAny(listModuleVersionsCommand, fmt.Sprintf("%s/_/proxy/modules", registryUrl), withEnableDebug, true, map[string]string{})

	for _, value := range resp.([]any) {
		mapEntry := value.(map[string]any)

		if strings.Contains(mapEntry["id"].(string), withModuleName) {
			fmt.Println(mapEntry["id"])
		}
	}
}

func init() {
	rootCmd.AddCommand(listModuleVersionsCmd)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&withModuleName, "moduleName", "m", "", "Module name, e.g. mod-orders (required)")
	listModuleVersionsCmd.PersistentFlags().StringVarP(&withId, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021")
	if err := listModuleVersionsCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error(listModuleVersionsCommand, internal.GetFuncName(), "listModuleVersionsCmd.MarkPersistentFlagRequired error")
		panic(err)
	}
}
