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
	"os"
	"sort"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	if params.ID != "" {
		return run.getModuleDescriptorByID()
	}

	return run.listModuleVersionsSortedDescendingOrder()
}

func (run *Run) getModuleDescriptorByID() error {
	requestURL := fmt.Sprintf("%s/_/proxy/modules/%s", run.Config.Action.ConfigRegistryURL, params.ID)
	respBytes, err := run.Config.HTTPClient.GetReturnRawBytes(requestURL, map[string]string{})
	if err != nil {
		return err
	}

	if !params.EnableDebug {
		fmt.Println(string(respBytes))
	}

	return nil
}

func (run *Run) listModuleVersionsSortedDescendingOrder() error {
	requestURL := fmt.Sprintf("%s/_/proxy/modules", run.Config.Action.ConfigRegistryURL)

	var decodedResponse models.ProxyModulesResponse
	if err := run.Config.HTTPClient.GetRetryReturnStruct(requestURL, map[string]string{}, &decodedResponse); err != nil {
		return err
	}

	var versions []string
	for _, module := range decodedResponse {
		if helpers.MatchesModuleName(module.ID, params.ModuleName) {
			versions = append(versions, module.ID)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		vi := strings.TrimPrefix(versions[i], params.ModuleName+"-")
		vj := strings.TrimPrefix(versions[j], params.ModuleName+"-")
		return helpers.IsVersionGreater(vi, vj)
	})

	for idx, version := range versions {
		if idx >= params.Versions {
			break
		}
		fmt.Println(version)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(listModuleVersionsCmd)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&params.ID, action.ID.Long, action.ID.Short, "", action.ID.Description)
	listModuleVersionsCmd.PersistentFlags().IntVarP(&params.Versions, action.Versions.Long, action.Versions.Short, 5, action.Versions.Description)

	if err := listModuleVersionsCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}

	if err := listModuleVersionsCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
