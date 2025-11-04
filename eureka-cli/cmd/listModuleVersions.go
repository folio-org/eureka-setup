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
	"sort"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// listModuleVersionsCmd represents the listModuleVersions command
var listModuleVersionsCmd = &cobra.Command{
	Use:   "listModuleVersions",
	Short: "List module versions",
	Long:  `List module versions using the registry URL from config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.ListModuleVersions)
		if err != nil {
			return err
		}

		return run.ListModuleVersions()
	},
}

func (run *Run) ListModuleVersions() error {
	if actionParams.ID != "" {
		return run.getModuleDescriptorByID()
	}

	return run.listModuleVersionsSortedDescendingOrder()
}

func (run *Run) getModuleDescriptorByID() error {
	requestURL := fmt.Sprintf("%s/_/proxy/modules/%s", run.Config.Action.ConfigRegistryURL, actionParams.ID)
	httpResponse, err := run.Config.HTTPClient.GetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return err
	}
	defer httpclient.CloseResponse(httpResponse)

	if !actionParams.EnableDebug {
		respBytes, err := httputil.DumpResponse(httpResponse, true)
		if err != nil {
			return err
		}
		fmt.Println(string(respBytes))
	}

	return nil
}

func (run *Run) listModuleVersionsSortedDescendingOrder() error {
	requestURL := fmt.Sprintf("%s/_/proxy/modules", run.Config.Action.ConfigRegistryURL)
	modulesData, err := run.Config.HTTPClient.GetRetryDecodeReturnAny(requestURL, map[string]string{})
	if err != nil {
		return err
	}

	var versions []string
	for _, value := range modulesData.([]any) {
		entry := value.(map[string]any)
		if helpers.MatchesModuleName(entry["id"].(string), actionParams.ModuleName) {
			versions = append(versions, entry["id"].(string))
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		vi := "v" + strings.TrimPrefix(versions[i], actionParams.ModuleName+"-")
		vj := "v" + strings.TrimPrefix(versions[j], actionParams.ModuleName+"-")
		return semver.Compare(vi, vj) > 0
	})

	for idx, version := range versions {
		if idx >= actionParams.Versions {
			break
		}
		fmt.Println(version)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(listModuleVersionsCmd)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&actionParams.ModuleName, "moduleName", "n", "", "Module name, e.g. mod-orders")
	listModuleVersionsCmd.PersistentFlags().StringVarP(&actionParams.ID, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021")
	listModuleVersionsCmd.PersistentFlags().IntVarP(&actionParams.Versions, "versions", "v", 5, "Number of versions, e.g. 5")
	if err := listModuleVersionsCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error("failed to mark moduleName flag as required", "error", err)
		os.Exit(1)
	}
}
