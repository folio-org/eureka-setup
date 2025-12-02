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
package main

import (
	"embed"

	"github.com/folio-org/eureka-cli/cmd"
)

var (
	//go:embed misc/edge-modules
	//go:embed misc/folio-kafka-tools
	//go:embed misc/folio-keycloak-nginx
	//go:embed misc/folio-netcat
	//go:embed misc/folio-vault
	//go:embed misc/postgres
	//go:embed misc/.env
	//go:embed misc/*.ps1
	//go:embed misc/*.sh
	//go:embed misc/docker-compose.yaml
	//go:embed *.yaml
	fs embed.FS
)

func main() {
	cmd.Execute(&fs)
}
