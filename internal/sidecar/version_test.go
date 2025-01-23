package sidecar

import (
	"errors"
	"testing"

	servicemock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCheckVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name                string
		configVersion       string
		setupMocks          func(*mock.MockDockerSidecar, *servicemock.MockGitHubService)
		expectedCurrent     string
		expectedLatest      string
		expectedNeedsUpdate bool
		expectedError       string
	}{
		{
			name:          "latest tag - up to date",
			configVersion: "latest",
			setupMocks: func(r *mock.MockDockerSidecar, g *servicemock.MockGitHubService) {
				g.EXPECT().GetLatestVersion().Return("1.0.0", nil)
				r.EXPECT().Version().Return("1.0.0", nil)
			},
			expectedCurrent:     "1.0.0",
			expectedLatest:      "1.0.0",
			expectedNeedsUpdate: false,
		},
		{
			name:          "latest tag - needs update",
			configVersion: "latest",
			setupMocks: func(r *mock.MockDockerSidecar, g *servicemock.MockGitHubService) {
				g.EXPECT().GetLatestVersion().Return("2.0.0", nil)
				r.EXPECT().Version().Return("1.0.0", nil)
			},
			expectedCurrent:     "1.0.0",
			expectedLatest:      "2.0.0",
			expectedNeedsUpdate: true,
		},
		{
			name:          "specific version - up to date",
			configVersion: "1.0.0",
			setupMocks: func(r *mock.MockDockerSidecar, g *servicemock.MockGitHubService) {
				g.EXPECT().GetLatestVersion().Return("1.0.0", nil)
			},
			expectedCurrent:     "1.0.0",
			expectedLatest:      "1.0.0",
			expectedNeedsUpdate: false,
		},
		{
			name:          "specific version - needs update",
			configVersion: "1.0.0",
			setupMocks: func(r *mock.MockDockerSidecar, g *servicemock.MockGitHubService) {
				g.EXPECT().GetLatestVersion().Return("2.0.0", nil)
			},
			expectedCurrent:     "1.0.0",
			expectedLatest:      "2.0.0",
			expectedNeedsUpdate: true,
		},
		{
			name:          "github error",
			configVersion: "latest",
			setupMocks: func(r *mock.MockDockerSidecar, g *servicemock.MockGitHubService) {
				g.EXPECT().GetLatestVersion().Return("", errors.New("github error"))
			},
			expectedError: "failed to get latest version",
		},
		{
			name:          "version error with latest tag",
			configVersion: "latest",
			setupMocks: func(r *mock.MockDockerSidecar, g *servicemock.MockGitHubService) {
				g.EXPECT().GetLatestVersion().Return("1.0.0", nil)
				r.EXPECT().Version().Return("", errors.New("version error"))
			},
			expectedError: "failed to get running version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGitHub := servicemock.NewMockGitHubService(ctrl)
			mockRunner := mock.NewMockDockerSidecar(ctrl)

			tt.setupMocks(mockRunner, mockGitHub)

			current, latest, needsUpdate, err := CheckVersion(mockRunner, mockGitHub, tt.configVersion)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCurrent, current)
			assert.Equal(t, tt.expectedLatest, latest)
			assert.Equal(t, tt.expectedNeedsUpdate, needsUpdate)
		})
	}
}
