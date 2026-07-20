package helpers

import (
	_ "embed"
	"bytes"
	"fmt"
	"text/template"
)

//go:embed templates/Dockerfile.custom-frontend
var customFrontendTemplateRaw string

// FrontendTemplateParams defines the data contract for our dynamic Dockerfile
type FrontendTemplateParams struct {
	Branch       string
	URL          string
	EurekaConfig string
}

// GenerateFrontendDockerfile compiles the embedded asset into a valid string payload
func GenerateFrontendDockerfile(branch, url, eurekaConfig string) (string, error) {
	if branch == "" {
		branch = "main"
	}

	if eurekaConfig == "" {
		eurekaConfig = "stripes.config.js"
	}

	tmpl, err := template.New("custom-frontend").Parse(customFrontendTemplateRaw)
	if err != nil {
		return "", fmt.Errorf("failed to parse embedded frontend template: %w", err)
	}

	params := FrontendTemplateParams{
		Branch:       branch,
		URL:          url,
		EurekaConfig: eurekaConfig,
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, params); err != nil {
		return "", fmt.Errorf("failed to execute frontend template: %w", err)
	}

	return buffer.String(), nil
}