// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type testDataSet int

const (
	testDataSetDefault testDataSet = iota
	testDataSetAll
	testDataSetNone
)

func TestMetricsBuilder(t *testing.T) {
	tests := []struct {
		name        string
		metricsSet  testDataSet
		resAttrsSet testDataSet
		expectEmpty bool
	}{
		{
			name: "default",
		},
		{
			name:        "all_set",
			metricsSet:  testDataSetAll,
			resAttrsSet: testDataSetAll,
		},
		{
			name:        "none_set",
			metricsSet:  testDataSetNone,
			resAttrsSet: testDataSetNone,
			expectEmpty: true,
		},
		{
			name:        "filter_set_include",
			resAttrsSet: testDataSetAll,
		},
		{
			name:        "filter_set_exclude",
			resAttrsSet: testDataSetAll,
			expectEmpty: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := pcommon.Timestamp(1_000_000_000)
			ts := pcommon.Timestamp(1_000_001_000)
			observedZapCore, observedLogs := observer.New(zap.WarnLevel)
			settings := receivertest.NewNopSettings()
			settings.Logger = zap.New(observedZapCore)
			mb := NewMetricsBuilder(loadMetricsBuilderConfig(t, test.name), settings, WithStartTime(start))

			expectedWarnings := 0

			assert.Equal(t, expectedWarnings, observedLogs.Len())

			defaultMetricsCount := 0
			allMetricsCount := 0

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryBranchCommitAheadbyCountDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryBranchCommitBehindbyCountDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryBranchCountDataPoint(ts, 1, "repository.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryBranchLineAdditionCountDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryBranchLineDeletionCountDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryBranchTimeDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			allMetricsCount++
			mb.RecordVcsRepositoryContributorCountDataPoint(ts, 1, "repository.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryCountDataPoint(ts, 1)

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryPullRequestCountDataPoint(ts, 1, AttributePullRequestStateOpen, "repository.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryPullRequestTimeOpenDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryPullRequestTimeToApprovalDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordVcsRepositoryPullRequestTimeToMergeDataPoint(ts, 1, "repository.name-val", "branch.name-val")

			rb := mb.NewResourceBuilder()
			rb.SetOrganizationName("organization.name-val")
			rb.SetVcsVendorName("vcs.vendor.name-val")
			res := rb.Emit()
			metrics := mb.Emit(WithResource(res))

			if test.expectEmpty {
				assert.Equal(t, 0, metrics.ResourceMetrics().Len())
				return
			}

			assert.Equal(t, 1, metrics.ResourceMetrics().Len())
			rm := metrics.ResourceMetrics().At(0)
			assert.Equal(t, res, rm.Resource())
			assert.Equal(t, 1, rm.ScopeMetrics().Len())
			ms := rm.ScopeMetrics().At(0).Metrics()
			if test.metricsSet == testDataSetDefault {
				assert.Equal(t, defaultMetricsCount, ms.Len())
			}
			if test.metricsSet == testDataSetAll {
				assert.Equal(t, allMetricsCount, ms.Len())
			}
			validatedMetrics := make(map[string]bool)
			for i := 0; i < ms.Len(); i++ {
				switch ms.At(i).Name() {
				case "vcs.repository.branch.commit.aheadby.count":
					assert.False(t, validatedMetrics["vcs.repository.branch.commit.aheadby.count"], "Found a duplicate in the metrics slice: vcs.repository.branch.commit.aheadby.count")
					validatedMetrics["vcs.repository.branch.commit.aheadby.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of commits a branch is ahead of the default branch (trunk).", ms.At(i).Description())
					assert.Equal(t, "{commit}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.branch.commit.behindby.count":
					assert.False(t, validatedMetrics["vcs.repository.branch.commit.behindby.count"], "Found a duplicate in the metrics slice: vcs.repository.branch.commit.behindby.count")
					validatedMetrics["vcs.repository.branch.commit.behindby.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of commits a branch is behind the default branch (trunk).", ms.At(i).Description())
					assert.Equal(t, "{commit}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.branch.count":
					assert.False(t, validatedMetrics["vcs.repository.branch.count"], "Found a duplicate in the metrics slice: vcs.repository.branch.count")
					validatedMetrics["vcs.repository.branch.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of branches in a repository.", ms.At(i).Description())
					assert.Equal(t, "{branch}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
				case "vcs.repository.branch.line.addition.count":
					assert.False(t, validatedMetrics["vcs.repository.branch.line.addition.count"], "Found a duplicate in the metrics slice: vcs.repository.branch.line.addition.count")
					validatedMetrics["vcs.repository.branch.line.addition.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of lines added in a branch relative to the default branch (trunk).", ms.At(i).Description())
					assert.Equal(t, "{line}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.branch.line.deletion.count":
					assert.False(t, validatedMetrics["vcs.repository.branch.line.deletion.count"], "Found a duplicate in the metrics slice: vcs.repository.branch.line.deletion.count")
					validatedMetrics["vcs.repository.branch.line.deletion.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of lines deleted in a branch relative to the default branch (trunk).", ms.At(i).Description())
					assert.Equal(t, "{line}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.branch.time":
					assert.False(t, validatedMetrics["vcs.repository.branch.time"], "Found a duplicate in the metrics slice: vcs.repository.branch.time")
					validatedMetrics["vcs.repository.branch.time"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "Time a branch created from the default branch (trunk) has existed.", ms.At(i).Description())
					assert.Equal(t, "s", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.contributor.count":
					assert.False(t, validatedMetrics["vcs.repository.contributor.count"], "Found a duplicate in the metrics slice: vcs.repository.contributor.count")
					validatedMetrics["vcs.repository.contributor.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of unique contributors to a repository.", ms.At(i).Description())
					assert.Equal(t, "{contributor}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
				case "vcs.repository.count":
					assert.False(t, validatedMetrics["vcs.repository.count"], "Found a duplicate in the metrics slice: vcs.repository.count")
					validatedMetrics["vcs.repository.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of repositories in an organization.", ms.At(i).Description())
					assert.Equal(t, "{repository}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
				case "vcs.repository.pull_request.count":
					assert.False(t, validatedMetrics["vcs.repository.pull_request.count"], "Found a duplicate in the metrics slice: vcs.repository.pull_request.count")
					validatedMetrics["vcs.repository.pull_request.count"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The number of pull requests in a repository, categorized by their state (either open or merged).", ms.At(i).Description())
					assert.Equal(t, "{pull_request}", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("pull_request.state")
					assert.True(t, ok)
					assert.EqualValues(t, "open", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
				case "vcs.repository.pull_request.time_open":
					assert.False(t, validatedMetrics["vcs.repository.pull_request.time_open"], "Found a duplicate in the metrics slice: vcs.repository.pull_request.time_open")
					validatedMetrics["vcs.repository.pull_request.time_open"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The amount of time a pull request has been open.", ms.At(i).Description())
					assert.Equal(t, "s", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.pull_request.time_to_approval":
					assert.False(t, validatedMetrics["vcs.repository.pull_request.time_to_approval"], "Found a duplicate in the metrics slice: vcs.repository.pull_request.time_to_approval")
					validatedMetrics["vcs.repository.pull_request.time_to_approval"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The amount of time it took a pull request to go from open to approved.", ms.At(i).Description())
					assert.Equal(t, "s", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				case "vcs.repository.pull_request.time_to_merge":
					assert.False(t, validatedMetrics["vcs.repository.pull_request.time_to_merge"], "Found a duplicate in the metrics slice: vcs.repository.pull_request.time_to_merge")
					validatedMetrics["vcs.repository.pull_request.time_to_merge"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The amount of time it took a pull request to go from open to merged.", ms.At(i).Description())
					assert.Equal(t, "s", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("repository.name")
					assert.True(t, ok)
					assert.EqualValues(t, "repository.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("branch.name")
					assert.True(t, ok)
					assert.EqualValues(t, "branch.name-val", attrVal.Str())
				}
			}
		})
	}
}
