package models

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	License         string            `json:"license"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Resolutions     map[string]string `json:"resolutions"`
}
