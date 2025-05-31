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
package main

import (
	"embed"

	"github.com/folio-org/eureka-cli/cmd"
)

//go:embed misc/edge-modules misc/folio-keycloak-nginx misc/folio-vault misc/postgres misc/docker-compose.yaml *.yaml
var mainEmbeddedFs embed.FS

func main() {
	cmd.Execute(mainEmbeddedFs)
}
