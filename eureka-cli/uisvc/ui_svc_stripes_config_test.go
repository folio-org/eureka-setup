package uisvc

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	apperrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

// The unspaced ${name} spelling, the missing-file and the no-placeholder cases are covered in ui_svc_test.go.
// These cases pin the whitespace-tolerant matching and the fail-fast on placeholders the CLI does not know.
func TestPrepareStripesConfigJS_PlaceholderSpelling(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantContains  []string
		wantErrNaming []string
	}{
		{
			name: "SpacedPlaceholders",
			content: `module.exports = {
  okapi: { url: '${kongUrl}', uri: '${tenantUrl}', authnUrl: '${keycloakUrl}' },
  config: {
    hasAllPerms: ${ hasAllPerms },
    tenantOptions: ${ tenantOptions },
    aboutInstallDate: ${ aboutInstallDate },
    aboutInstallMessage: ${ aboutInstallMsg },
    isSingleTenant: ${ isSingleTenant },
    enableEcsRequests: ${ enableEcsRequests },
  },
};`,
			wantContains: []string{
				"url: '" + constant.KongExternalHTTP + "'",
				"uri: 'http://lsp.test'",
				"authnUrl: '" + constant.KeycloakExternalHTTP + "'",
				"hasAllPerms: false",
				`tenantOptions: {diku: {name: "diku", displayName: "diku", clientId: "diku"}}`,
				"aboutInstallMessage: 'Local build'",
				"isSingleTenant: true",
				"enableEcsRequests: false",
			},
		},
		{
			name:         "MixedSpacing",
			content:      "a: ${hasAllPerms}, b: ${  hasAllPerms  }, c: ${\thasAllPerms}",
			wantContains: []string{"a: false, b: false, c: false"},
		},
		{
			name:          "UnknownPlaceholderFailsBeforeWriting",
			content:       "a: ${hasAllPerms}, b: ${ newThing }, c: ${other}",
			wantErrNaming: []string{"${ newThing }", "${other}"},
		},
		{
			name:          "JavaScriptInterpolationCountsAsPlaceholder",
			content:       "const base = 'http://x';\nurl: `${base}/api`, a: ${hasAllPerms}",
			wantErrNaming: []string{"${base}"},
		},
		{
			name:          "UnknownPlaceholderWithNonWordCharactersFails",
			content:       "a: ${hasAllPerms}, b: ${new-feature}, c: ${ config.value }, d: ${}",
			wantErrNaming: []string{"${new-feature}", "${ config.value }", "${}"},
		},
	}

	for _, tt := range tests {
		t.Run("TestPrepareStripesConfigJS_"+tt.name, func(t *testing.T) {
			// Arrange
			act := testhelpers.NewMockAction()
			act.Param = &action.Param{PlatformLspURL: "http://lsp.test", SingleTenant: true}
			svc := New(act, nil, nil, nil, nil)
			dir := t.TempDir()
			filePath := filepath.Join(dir, "stripes.config.js")
			assert.NoError(t, os.WriteFile(filePath, []byte(tt.content), 0644))

			// Act
			err := svc.PrepareStripesConfigJS("diku", dir)

			// Assert
			written, readErr := os.ReadFile(filePath)
			assert.NoError(t, readErr)
			if len(tt.wantErrNaming) > 0 {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, apperrors.ErrDeploymentFailed))
				for _, name := range tt.wantErrNaming {
					assert.Contains(t, err.Error(), name)
				}
				assert.Equal(t, tt.content, string(written), "file must not be written when placeholders remain")
				return
			}
			assert.NoError(t, err)
			for _, want := range tt.wantContains {
				assert.Contains(t, string(written), want)
			}
			assert.NotContains(t, string(written), "${")
		})
	}
}
