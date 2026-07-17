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
	Branch      string
	URL         string
	StartScript string
}

// GenerateFrontendDockerfile compiles the embedded asset into a valid string payload
func GenerateFrontendDockerfile(branch, url, startScript string) (string, error) {
	if branch == "" {
		branch = "main"
	}
	if startScript == "" {
		startScript = "start"
	}

	tmpl, err := template.New("custom-frontend").Parse(customFrontendTemplateRaw)
	if err != nil {
		return "", fmt.Errorf("failed to parse embedded frontend template: %w", err)
	}

	params := FrontendTemplateParams{
		Branch:      branch,
		URL:         url,
		StartScript: startScript,
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, params); err != nil {
		return "", fmt.Errorf("failed to execute frontend template: %w", err)
	}

	return buffer.String(), nil
}