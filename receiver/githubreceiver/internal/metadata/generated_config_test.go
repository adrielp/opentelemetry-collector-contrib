// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestMetricsBuilderConfig(t *testing.T) {
	tests := []struct {
		name string
		want MetricsBuilderConfig
	}{
		{
			name: "default",
			want: DefaultMetricsBuilderConfig(),
		},
		{
			name: "all_set",
			want: MetricsBuilderConfig{
				Metrics: MetricsConfig{
					VcsContributorCount:               MetricConfig{Enabled: true},
					VcsRefCount:                       MetricConfig{Enabled: true},
					VcsRefLinesDelta:                  MetricConfig{Enabled: true},
					VcsRefRevisionsDelta:              MetricConfig{Enabled: true},
					VcsRefTime:                        MetricConfig{Enabled: true},
					VcsRepositoryChangeCount:          MetricConfig{Enabled: true},
					VcsRepositoryChangeTimeOpen:       MetricConfig{Enabled: true},
					VcsRepositoryChangeTimeToApproval: MetricConfig{Enabled: true},
					VcsRepositoryChangeTimeToMerge:    MetricConfig{Enabled: true},
					VcsRepositoryCount:                MetricConfig{Enabled: true},
				},
				ResourceAttributes: ResourceAttributesConfig{
					OrganizationName: ResourceAttributeConfig{Enabled: true},
					VcsVendorName:    ResourceAttributeConfig{Enabled: true},
				},
			},
		},
		{
			name: "none_set",
			want: MetricsBuilderConfig{
				Metrics: MetricsConfig{
					VcsContributorCount:               MetricConfig{Enabled: false},
					VcsRefCount:                       MetricConfig{Enabled: false},
					VcsRefLinesDelta:                  MetricConfig{Enabled: false},
					VcsRefRevisionsDelta:              MetricConfig{Enabled: false},
					VcsRefTime:                        MetricConfig{Enabled: false},
					VcsRepositoryChangeCount:          MetricConfig{Enabled: false},
					VcsRepositoryChangeTimeOpen:       MetricConfig{Enabled: false},
					VcsRepositoryChangeTimeToApproval: MetricConfig{Enabled: false},
					VcsRepositoryChangeTimeToMerge:    MetricConfig{Enabled: false},
					VcsRepositoryCount:                MetricConfig{Enabled: false},
				},
				ResourceAttributes: ResourceAttributesConfig{
					OrganizationName: ResourceAttributeConfig{Enabled: false},
					VcsVendorName:    ResourceAttributeConfig{Enabled: false},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadMetricsBuilderConfig(t, tt.name)
			diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(MetricConfig{}, ResourceAttributeConfig{}))
			require.Emptyf(t, diff, "Config mismatch (-expected +actual):\n%s", diff)
		})
	}
}

func loadMetricsBuilderConfig(t *testing.T, name string) MetricsBuilderConfig {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	sub, err := cm.Sub(name)
	require.NoError(t, err)
	cfg := DefaultMetricsBuilderConfig()
	require.NoError(t, sub.Unmarshal(&cfg))
	return cfg
}

func TestResourceAttributesConfig(t *testing.T) {
	tests := []struct {
		name string
		want ResourceAttributesConfig
	}{
		{
			name: "default",
			want: DefaultResourceAttributesConfig(),
		},
		{
			name: "all_set",
			want: ResourceAttributesConfig{
				OrganizationName: ResourceAttributeConfig{Enabled: true},
				VcsVendorName:    ResourceAttributeConfig{Enabled: true},
			},
		},
		{
			name: "none_set",
			want: ResourceAttributesConfig{
				OrganizationName: ResourceAttributeConfig{Enabled: false},
				VcsVendorName:    ResourceAttributeConfig{Enabled: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadResourceAttributesConfig(t, tt.name)
			diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(ResourceAttributeConfig{}))
			require.Emptyf(t, diff, "Config mismatch (-expected +actual):\n%s", diff)
		})
	}
}

func loadResourceAttributesConfig(t *testing.T, name string) ResourceAttributesConfig {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	sub, err := cm.Sub(name)
	require.NoError(t, err)
	sub, err = sub.Sub("resource_attributes")
	require.NoError(t, err)
	cfg := DefaultResourceAttributesConfig()
	require.NoError(t, sub.Unmarshal(&cfg))
	return cfg
}
