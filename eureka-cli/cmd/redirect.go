/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"github.com/spf13/cobra"
)

// redirectCmd represents the redirect command
var redirectCmd = &cobra.Command{
	Use:   "redirect",
	Short: "Redirect modules",
	Long:  `Redirect multiple modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		RedirectModules()
	},
}

// TODO
func RedirectModules() {

}

func init() {
	rootCmd.AddCommand(redirectCmd)
}
