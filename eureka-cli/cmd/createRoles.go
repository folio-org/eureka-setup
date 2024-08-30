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

// createRolesCmd represents the createRoles command
var createRolesCmd = &cobra.Command{
	Use:   "createRoles",
	Short: "Create roles",
	Long:  `Create all roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateRoles()
	},
}

// TODO
func CreateRoles() {

}

func init() {
	rootCmd.AddCommand(createRolesCmd)
}