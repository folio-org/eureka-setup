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
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
)

// listModuleVersionsCmd represents the listModuleVersions command
var listModuleVersionsCmd = &cobra.Command{
	Use:   "listModuleVersions",
	Short: "List module versions",
	Long:  `List module versions using the registry URL from config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.ListModuleVersions)
		if err != nil {
			return err
		}

		err = r.ListModuleVersions()
		if err != nil {
			return err
		}

		return nil
	},
}

func (r *Run) ListModuleVersions() error {
	registryURL := viper.GetString(field.RegistryURL)
	if rp.ID != "" {
		return r.getModuleDescriptorById(registryURL)
	}

	return r.listModuleVersionsSortedDescendingOrder(registryURL)
}

func (r *Run) getModuleDescriptorById(registryURL string) error {
	resp, err := r.Config.HTTPClient.DoGetReturnResponse(fmt.Sprintf("%s/_/proxy/modules/%s", registryURL, rp.ID), map[string]string{})
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if !rp.EnableDebug {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}

		idx := strings.Index(string(respBytes), constant.NewLinePattern)
		if idx == -1 {
			return fmt.Errorf("response does not contain an empty line using %s id", rp.ID)
		}

		fmt.Println(string([]byte(respBytes[idx:])))
	}

	return nil
}

func (r *Run) listModuleVersionsSortedDescendingOrder(registryURL string) error {
	resp, err := r.Config.HTTPClient.DoGetDecodeReturnAny(fmt.Sprintf("%s/_/proxy/modules", registryURL), map[string]string{})
	if err != nil {
		return err
	}

	var versions []string

	for _, value := range resp.([]any) {
		mapEntry := value.(map[string]any)

		if helpers.MatchesModuleName(mapEntry["id"].(string), rp.ModuleName) {
			versions = append(versions, mapEntry["id"].(string))
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		vi := "v" + strings.TrimPrefix(versions[i], rp.ModuleName+"-")
		vj := "v" + strings.TrimPrefix(versions[j], rp.ModuleName+"-")
		return semver.Compare(vi, vj) > 0
	})

	for idx, version := range versions {
		if idx >= rp.Lines {
			break
		}
		fmt.Println(version)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(listModuleVersionsCmd)
	listModuleVersionsCmd.PersistentFlags().StringVarP(&rp.ModuleName, "moduleName", "m", "", "Module name, e.g. mod-orders (required)")
	listModuleVersionsCmd.PersistentFlags().StringVarP(&rp.ID, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021")
	listModuleVersionsCmd.PersistentFlags().IntVarP(&rp.Lines, "lines", "L", 5, "Number of lines, e.g. 5")
	if err := listModuleVersionsCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
