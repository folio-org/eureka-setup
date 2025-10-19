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
	"os"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listModuleVersionsCmd represents the listModuleVersions command
var listModuleVersionsCmd = &cobra.Command{
	Use:   "listModuleVersions",
	Short: "List module versions",
	Long:  `List module versions using the registry URL from config.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.ListModuleVersions).ListModuleVersions()
	},
}

func (r *Run) ListModuleVersions() {
	registryURL := viper.GetString(field.RegistryURL)
	if rp.ID != "" {
		r.getModuleDescriptorById(registryURL)
		return
	}
	r.listModuleVersions(registryURL)
}

func (r *Run) getModuleDescriptorById(registryURL string) {
	resp := r.Config.HTTPClient.DoGetReturnResponse(fmt.Sprintf("%s/_/proxy/modules/%s", registryURL, rp.ID), true, map[string]string{})
	defer func() {
		_ = resp.Body.Close()
	}()

	if !rp.EnableDebug {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			slog.Error(r.Config.Action.Name, "error", err)
			return
		}

		idx := strings.Index(string(respBytes), constant.NewLinePattern)
		if idx == -1 {
			slog.Error(r.Config.Action.Name, "error", "response does not contain an empty line")
			return
		}

		fmt.Println(string([]byte(respBytes[idx:])))
	}
}

func (r *Run) listModuleVersions(registryURL string) {
	resp := r.Config.HTTPClient.DoGetDecodeReturnAny(fmt.Sprintf("%s/_/proxy/modules", registryURL), true, map[string]string{})

	for _, value := range resp.([]any) {
		mapEntry := value.(map[string]any)

		if strings.Contains(mapEntry["id"].(string), rp.ModuleName) {
			fmt.Println(mapEntry["id"])
		}
	}
}

func init() {
	rootCmd.AddCommand(listModuleVersionsCmd)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&rp.ModuleName, "moduleName", "m", "", "Module name, e.g. mod-orders (required)")
	listModuleVersionsCmd.PersistentFlags().StringVarP(&rp.ID, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021")
	if err := listModuleVersionsCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
