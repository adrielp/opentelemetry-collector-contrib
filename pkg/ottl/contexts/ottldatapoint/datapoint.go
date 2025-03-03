// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottldatapoint // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottldatapoint"

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap/zapcore"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/ctxdatapoint"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/ctxerror"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/ctxmetric"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/ctxresource"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/ctxscope"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/ctxutil"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/internal/logging"
)

// Experimental: *NOTE* this constant is subject to change or removal in the future.
const ContextName = ctxdatapoint.Name

var (
	_ internal.ResourceContext             = (*TransformContext)(nil)
	_ internal.InstrumentationScopeContext = (*TransformContext)(nil)
	_ ctxmetric.Context                    = (*TransformContext)(nil)
	_ zapcore.ObjectMarshaler              = (*TransformContext)(nil)
)

type TransformContext struct {
	dataPoint            any
	metric               pmetric.Metric
	metrics              pmetric.MetricSlice
	instrumentationScope pcommon.InstrumentationScope
	resource             pcommon.Resource
	cache                pcommon.Map
	scopeMetrics         pmetric.ScopeMetrics
	resourceMetrics      pmetric.ResourceMetrics
}

func (tCtx TransformContext) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	err := encoder.AddObject("resource", logging.Resource(tCtx.resource))
	err = errors.Join(err, encoder.AddObject("scope", logging.InstrumentationScope(tCtx.instrumentationScope)))
	err = errors.Join(err, encoder.AddObject("metric", logging.Metric(tCtx.metric)))

	switch dp := tCtx.dataPoint.(type) {
	case pmetric.NumberDataPoint:
		err = encoder.AddObject("datapoint", logging.NumberDataPoint(dp))
	case pmetric.HistogramDataPoint:
		err = encoder.AddObject("datapoint", logging.HistogramDataPoint(dp))
	case pmetric.ExponentialHistogramDataPoint:
		err = encoder.AddObject("datapoint", logging.ExponentialHistogramDataPoint(dp))
	case pmetric.SummaryDataPoint:
		err = encoder.AddObject("datapoint", logging.SummaryDataPoint(dp))
	}

	err = errors.Join(err, encoder.AddObject("cache", logging.Map(tCtx.cache)))
	return err
}

type Option func(*ottl.Parser[TransformContext])

type TransformContextOption func(*TransformContext)

func NewTransformContext(dataPoint any, metric pmetric.Metric, metrics pmetric.MetricSlice, instrumentationScope pcommon.InstrumentationScope, resource pcommon.Resource, scopeMetrics pmetric.ScopeMetrics, resourceMetrics pmetric.ResourceMetrics, options ...TransformContextOption) TransformContext {
	tc := TransformContext{
		dataPoint:            dataPoint,
		metric:               metric,
		metrics:              metrics,
		instrumentationScope: instrumentationScope,
		resource:             resource,
		cache:                pcommon.NewMap(),
		scopeMetrics:         scopeMetrics,
		resourceMetrics:      resourceMetrics,
	}
	for _, opt := range options {
		opt(&tc)
	}
	return tc
}

// Experimental: *NOTE* this option is subject to change or removal in the future.
func WithCache(cache *pcommon.Map) TransformContextOption {
	return func(p *TransformContext) {
		if cache != nil {
			p.cache = *cache
		}
	}
}

func (tCtx TransformContext) GetDataPoint() any {
	return tCtx.dataPoint
}

func (tCtx TransformContext) GetInstrumentationScope() pcommon.InstrumentationScope {
	return tCtx.instrumentationScope
}

func (tCtx TransformContext) GetResource() pcommon.Resource {
	return tCtx.resource
}

func (tCtx TransformContext) GetMetric() pmetric.Metric {
	return tCtx.metric
}

func (tCtx TransformContext) GetMetrics() pmetric.MetricSlice {
	return tCtx.metrics
}

func (tCtx TransformContext) getCache() pcommon.Map {
	return tCtx.cache
}

func (tCtx TransformContext) GetScopeSchemaURLItem() internal.SchemaURLItem {
	return tCtx.scopeMetrics
}

func (tCtx TransformContext) GetResourceSchemaURLItem() internal.SchemaURLItem {
	return tCtx.resourceMetrics
}

func NewParser(functions map[string]ottl.Factory[TransformContext], telemetrySettings component.TelemetrySettings, options ...Option) (ottl.Parser[TransformContext], error) {
	pep := pathExpressionParser{telemetrySettings}
	p, err := ottl.NewParser[TransformContext](
		functions,
		pep.parsePath,
		telemetrySettings,
		ottl.WithEnumParser[TransformContext](parseEnum),
	)
	if err != nil {
		return ottl.Parser[TransformContext]{}, err
	}
	for _, opt := range options {
		opt(&p)
	}
	return p, nil
}

// EnablePathContextNames enables the support to path's context names on statements.
// When this option is configured, all statement's paths must have a valid context prefix,
// otherwise an error is reported.
//
// Experimental: *NOTE* this option is subject to change or removal in the future.
func EnablePathContextNames() Option {
	return func(p *ottl.Parser[TransformContext]) {
		ottl.WithPathContextNames[TransformContext]([]string{
			ctxdatapoint.Name,
			ctxresource.Name,
			ctxscope.LegacyName,
			ctxmetric.Name,
		})(p)
	}
}

type StatementSequenceOption func(*ottl.StatementSequence[TransformContext])

func WithStatementSequenceErrorMode(errorMode ottl.ErrorMode) StatementSequenceOption {
	return func(s *ottl.StatementSequence[TransformContext]) {
		ottl.WithStatementSequenceErrorMode[TransformContext](errorMode)(s)
	}
}

func NewStatementSequence(statements []*ottl.Statement[TransformContext], telemetrySettings component.TelemetrySettings, options ...StatementSequenceOption) ottl.StatementSequence[TransformContext] {
	s := ottl.NewStatementSequence(statements, telemetrySettings)
	for _, op := range options {
		op(&s)
	}
	return s
}

type ConditionSequenceOption func(*ottl.ConditionSequence[TransformContext])

func WithConditionSequenceErrorMode(errorMode ottl.ErrorMode) ConditionSequenceOption {
	return func(c *ottl.ConditionSequence[TransformContext]) {
		ottl.WithConditionSequenceErrorMode[TransformContext](errorMode)(c)
	}
}

func NewConditionSequence(conditions []*ottl.Condition[TransformContext], telemetrySettings component.TelemetrySettings, options ...ConditionSequenceOption) ottl.ConditionSequence[TransformContext] {
	c := ottl.NewConditionSequence(conditions, telemetrySettings)
	for _, op := range options {
		op(&c)
	}
	return c
}

func parseEnum(val *ottl.EnumSymbol) (*ottl.Enum, error) {
	if val != nil {
		if enum, ok := ctxdatapoint.SymbolTable[*val]; ok {
			return &enum, nil
		}
		return nil, fmt.Errorf("enum symbol, %s, not found", *val)
	}
	return nil, fmt.Errorf("enum symbol not provided")
}

type pathExpressionParser struct {
	telemetrySettings component.TelemetrySettings
}

func (pep *pathExpressionParser) parsePath(path ottl.Path[TransformContext]) (ottl.GetSetter[TransformContext], error) {
	if path == nil {
		return nil, ctxerror.New("nil", "nil", ctxdatapoint.Name, ctxdatapoint.DocRef)
	}
	// Higher contexts parsing
	if path.Context() != "" && path.Context() != ctxdatapoint.Name {
		return pep.parseHigherContextPath(path.Context(), path)
	}
	// Backward compatibility with paths without context
	if path.Context() == "" && (path.Name() == ctxresource.Name ||
		path.Name() == ctxscope.LegacyName ||
		path.Name() == ctxmetric.Name) {
		return pep.parseHigherContextPath(path.Name(), path.Next())
	}

	switch path.Name() {
	case "cache":
		if path.Keys() == nil {
			return accessCache(), nil
		}
		return accessCacheKey(path.Keys()), nil
	case "attributes":
		if path.Keys() == nil {
			return accessAttributes(), nil
		}
		return accessAttributesKey(path.Keys()), nil
	case "start_time_unix_nano":
		return accessStartTimeUnixNano(), nil
	case "time_unix_nano":
		return accessTimeUnixNano(), nil
	case "start_time":
		return accessStartTime(), nil
	case "time":
		return accessTime(), nil
	case "value_double":
		return accessDoubleValue(), nil
	case "value_int":
		return accessIntValue(), nil
	case "exemplars":
		return accessExemplars(), nil
	case "flags":
		return accessFlags(), nil
	case "count":
		return accessCount(), nil
	case "sum":
		return accessSum(), nil
	case "bucket_counts":
		return accessBucketCounts(), nil
	case "explicit_bounds":
		return accessExplicitBounds(), nil
	case "scale":
		return accessScale(), nil
	case "zero_count":
		return accessZeroCount(), nil
	case "positive":
		nextPath := path.Next()
		if nextPath != nil {
			switch nextPath.Name() {
			case "offset":
				return accessPositiveOffset(), nil
			case "bucket_counts":
				return accessPositiveBucketCounts(), nil
			default:
				return nil, ctxerror.New(nextPath.Name(), path.String(), ctxdatapoint.Name, ctxdatapoint.DocRef)
			}
		}
		return accessPositive(), nil
	case "negative":
		nextPath := path.Next()
		if nextPath != nil {
			switch nextPath.Name() {
			case "offset":
				return accessNegativeOffset(), nil
			case "bucket_counts":
				return accessNegativeBucketCounts(), nil
			default:
				return nil, ctxerror.New(nextPath.Name(), path.String(), ctxdatapoint.Name, ctxdatapoint.DocRef)
			}
		}
		return accessNegative(), nil
	case "quantile_values":
		return accessQuantileValues(), nil
	default:
		return nil, ctxerror.New(path.Name(), path.String(), ctxdatapoint.Name, ctxdatapoint.DocRef)
	}
}

func (pep *pathExpressionParser) parseHigherContextPath(context string, path ottl.Path[TransformContext]) (ottl.GetSetter[TransformContext], error) {
	switch context {
	case ctxresource.Name:
		return internal.ResourcePathGetSetter(ctxdatapoint.Name, path)
	case ctxscope.LegacyName:
		return internal.ScopePathGetSetter(ctxdatapoint.Name, path)
	case ctxmetric.Name:
		return ctxmetric.PathGetSetter(path)
	default:
		var fullPath string
		if path != nil {
			fullPath = path.String()
		}
		return nil, ctxerror.New(context, fullPath, ctxdatapoint.Name, ctxdatapoint.DocRef)
	}
}

func accessCache() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			return tCtx.getCache(), nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if m, ok := val.(pcommon.Map); ok {
				m.CopyTo(tCtx.getCache())
			}
			return nil
		},
	}
}

func accessCacheKey(key []ottl.Key[TransformContext]) ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(ctx context.Context, tCtx TransformContext) (any, error) {
			return ctxutil.GetMapValue[TransformContext](ctx, tCtx, tCtx.getCache(), key)
		},
		Setter: func(ctx context.Context, tCtx TransformContext, val any) error {
			return ctxutil.SetMapValue[TransformContext](ctx, tCtx, tCtx.getCache(), key, val)
		},
	}
}

func accessAttributes() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return tCtx.GetDataPoint().(pmetric.NumberDataPoint).Attributes(), nil
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Attributes(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Attributes(), nil
			case pmetric.SummaryDataPoint:
				return tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Attributes(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				if attrs, ok := val.(pcommon.Map); ok {
					attrs.CopyTo(tCtx.GetDataPoint().(pmetric.NumberDataPoint).Attributes())
				}
			case pmetric.HistogramDataPoint:
				if attrs, ok := val.(pcommon.Map); ok {
					attrs.CopyTo(tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Attributes())
				}
			case pmetric.ExponentialHistogramDataPoint:
				if attrs, ok := val.(pcommon.Map); ok {
					attrs.CopyTo(tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Attributes())
				}
			case pmetric.SummaryDataPoint:
				if attrs, ok := val.(pcommon.Map); ok {
					attrs.CopyTo(tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Attributes())
				}
			}
			return nil
		},
	}
}

func accessAttributesKey(key []ottl.Key[TransformContext]) ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(ctx context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return ctxutil.GetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.NumberDataPoint).Attributes(), key)
			case pmetric.HistogramDataPoint:
				return ctxutil.GetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Attributes(), key)
			case pmetric.ExponentialHistogramDataPoint:
				return ctxutil.GetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Attributes(), key)
			case pmetric.SummaryDataPoint:
				return ctxutil.GetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Attributes(), key)
			}
			return nil, nil
		},
		Setter: func(ctx context.Context, tCtx TransformContext, val any) error {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return ctxutil.SetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.NumberDataPoint).Attributes(), key, val)
			case pmetric.HistogramDataPoint:
				return ctxutil.SetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Attributes(), key, val)
			case pmetric.ExponentialHistogramDataPoint:
				return ctxutil.SetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Attributes(), key, val)
			case pmetric.SummaryDataPoint:
				return ctxutil.SetMapValue[TransformContext](ctx, tCtx, tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Attributes(), key, val)
			}
			return nil
		},
	}
}

func accessStartTimeUnixNano() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return tCtx.GetDataPoint().(pmetric.NumberDataPoint).StartTimestamp().AsTime().UnixNano(), nil
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).StartTimestamp().AsTime().UnixNano(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).StartTimestamp().AsTime().UnixNano(), nil
			case pmetric.SummaryDataPoint:
				return tCtx.GetDataPoint().(pmetric.SummaryDataPoint).StartTimestamp().AsTime().UnixNano(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newTime, ok := val.(int64); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.NumberDataPoint:
					tCtx.GetDataPoint().(pmetric.NumberDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				}
			}
			return nil
		},
	}
}

func accessStartTime() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return tCtx.GetDataPoint().(pmetric.NumberDataPoint).StartTimestamp().AsTime(), nil
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).StartTimestamp().AsTime(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).StartTimestamp().AsTime(), nil
			case pmetric.SummaryDataPoint:
				return tCtx.GetDataPoint().(pmetric.SummaryDataPoint).StartTimestamp().AsTime(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newTime, ok := val.(time.Time); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.NumberDataPoint:
					tCtx.GetDataPoint().(pmetric.NumberDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(newTime))
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(newTime))
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(newTime))
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetStartTimestamp(pcommon.NewTimestampFromTime(newTime))
				}
			}
			return nil
		},
	}
}

func accessTimeUnixNano() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return tCtx.GetDataPoint().(pmetric.NumberDataPoint).Timestamp().AsTime().UnixNano(), nil
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Timestamp().AsTime().UnixNano(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Timestamp().AsTime().UnixNano(), nil
			case pmetric.SummaryDataPoint:
				return tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Timestamp().AsTime().UnixNano(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newTime, ok := val.(int64); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.NumberDataPoint:
					tCtx.GetDataPoint().(pmetric.NumberDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(0, newTime)))
				}
			}
			return nil
		},
	}
}

func accessTime() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return tCtx.GetDataPoint().(pmetric.NumberDataPoint).Timestamp().AsTime(), nil
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Timestamp().AsTime(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Timestamp().AsTime(), nil
			case pmetric.SummaryDataPoint:
				return tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Timestamp().AsTime(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newTime, ok := val.(time.Time); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.NumberDataPoint:
					tCtx.GetDataPoint().(pmetric.NumberDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(newTime))
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(newTime))
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(newTime))
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetTimestamp(pcommon.NewTimestampFromTime(newTime))
				}
			}
			return nil
		},
	}
}

func accessDoubleValue() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if numberDataPoint, ok := tCtx.GetDataPoint().(pmetric.NumberDataPoint); ok {
				return numberDataPoint.DoubleValue(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newDouble, ok := val.(float64); ok {
				if numberDataPoint, ok := tCtx.GetDataPoint().(pmetric.NumberDataPoint); ok {
					numberDataPoint.SetDoubleValue(newDouble)
				}
			}
			return nil
		},
	}
}

func accessIntValue() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if numberDataPoint, ok := tCtx.GetDataPoint().(pmetric.NumberDataPoint); ok {
				return numberDataPoint.IntValue(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newInt, ok := val.(int64); ok {
				if numberDataPoint, ok := tCtx.GetDataPoint().(pmetric.NumberDataPoint); ok {
					numberDataPoint.SetIntValue(newInt)
				}
			}
			return nil
		},
	}
}

func accessExemplars() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return tCtx.GetDataPoint().(pmetric.NumberDataPoint).Exemplars(), nil
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Exemplars(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Exemplars(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newExemplars, ok := val.(pmetric.ExemplarSlice); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.NumberDataPoint:
					newExemplars.CopyTo(tCtx.GetDataPoint().(pmetric.NumberDataPoint).Exemplars())
				case pmetric.HistogramDataPoint:
					newExemplars.CopyTo(tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Exemplars())
				case pmetric.ExponentialHistogramDataPoint:
					newExemplars.CopyTo(tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Exemplars())
				}
			}
			return nil
		},
	}
}

func accessFlags() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.NumberDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.NumberDataPoint).Flags()), nil
			case pmetric.HistogramDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Flags()), nil
			case pmetric.ExponentialHistogramDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Flags()), nil
			case pmetric.SummaryDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Flags()), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newFlags, ok := val.(int64); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.NumberDataPoint:
					tCtx.GetDataPoint().(pmetric.NumberDataPoint).SetFlags(pmetric.DataPointFlags(newFlags))
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetFlags(pmetric.DataPointFlags(newFlags))
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetFlags(pmetric.DataPointFlags(newFlags))
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetFlags(pmetric.DataPointFlags(newFlags))
				}
			}
			return nil
		},
	}
}

func accessCount() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.HistogramDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Count()), nil
			case pmetric.ExponentialHistogramDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Count()), nil
			case pmetric.SummaryDataPoint:
				return int64(tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Count()), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newCount, ok := val.(int64); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetCount(uint64(newCount))
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetCount(uint64(newCount))
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetCount(uint64(newCount))
				}
			}
			return nil
		},
	}
}

func accessSum() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			switch tCtx.GetDataPoint().(type) {
			case pmetric.HistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.HistogramDataPoint).Sum(), nil
			case pmetric.ExponentialHistogramDataPoint:
				return tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).Sum(), nil
			case pmetric.SummaryDataPoint:
				return tCtx.GetDataPoint().(pmetric.SummaryDataPoint).Sum(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newSum, ok := val.(float64); ok {
				switch tCtx.GetDataPoint().(type) {
				case pmetric.HistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.HistogramDataPoint).SetSum(newSum)
				case pmetric.ExponentialHistogramDataPoint:
					tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint).SetSum(newSum)
				case pmetric.SummaryDataPoint:
					tCtx.GetDataPoint().(pmetric.SummaryDataPoint).SetSum(newSum)
				}
			}
			return nil
		},
	}
}

func accessExplicitBounds() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if histogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.HistogramDataPoint); ok {
				return histogramDataPoint.ExplicitBounds().AsRaw(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newExplicitBounds, ok := val.([]float64); ok {
				if histogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.HistogramDataPoint); ok {
					histogramDataPoint.ExplicitBounds().FromRaw(newExplicitBounds)
				}
			}
			return nil
		},
	}
}

func accessBucketCounts() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if histogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.HistogramDataPoint); ok {
				return histogramDataPoint.BucketCounts().AsRaw(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newBucketCount, ok := val.([]uint64); ok {
				if histogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.HistogramDataPoint); ok {
					histogramDataPoint.BucketCounts().FromRaw(newBucketCount)
				}
			}
			return nil
		},
	}
}

func accessScale() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return int64(expoHistogramDataPoint.Scale()), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newScale, ok := val.(int64); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					expoHistogramDataPoint.SetScale(int32(newScale))
				}
			}
			return nil
		},
	}
}

func accessZeroCount() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return int64(expoHistogramDataPoint.ZeroCount()), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newZeroCount, ok := val.(int64); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					expoHistogramDataPoint.SetZeroCount(uint64(newZeroCount))
				}
			}
			return nil
		},
	}
}

func accessPositive() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return expoHistogramDataPoint.Positive(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newPositive, ok := val.(pmetric.ExponentialHistogramDataPointBuckets); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					newPositive.CopyTo(expoHistogramDataPoint.Positive())
				}
			}
			return nil
		},
	}
}

func accessPositiveOffset() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return int64(expoHistogramDataPoint.Positive().Offset()), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newPositiveOffset, ok := val.(int64); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					expoHistogramDataPoint.Positive().SetOffset(int32(newPositiveOffset))
				}
			}
			return nil
		},
	}
}

func accessPositiveBucketCounts() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return expoHistogramDataPoint.Positive().BucketCounts().AsRaw(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newPositiveBucketCounts, ok := val.([]uint64); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					expoHistogramDataPoint.Positive().BucketCounts().FromRaw(newPositiveBucketCounts)
				}
			}
			return nil
		},
	}
}

func accessNegative() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return expoHistogramDataPoint.Negative(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newNegative, ok := val.(pmetric.ExponentialHistogramDataPointBuckets); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					newNegative.CopyTo(expoHistogramDataPoint.Negative())
				}
			}
			return nil
		},
	}
}

func accessNegativeOffset() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return int64(expoHistogramDataPoint.Negative().Offset()), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newNegativeOffset, ok := val.(int64); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					expoHistogramDataPoint.Negative().SetOffset(int32(newNegativeOffset))
				}
			}
			return nil
		},
	}
}

func accessNegativeBucketCounts() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
				return expoHistogramDataPoint.Negative().BucketCounts().AsRaw(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newNegativeBucketCounts, ok := val.([]uint64); ok {
				if expoHistogramDataPoint, ok := tCtx.GetDataPoint().(pmetric.ExponentialHistogramDataPoint); ok {
					expoHistogramDataPoint.Negative().BucketCounts().FromRaw(newNegativeBucketCounts)
				}
			}
			return nil
		},
	}
}

func accessQuantileValues() ottl.StandardGetSetter[TransformContext] {
	return ottl.StandardGetSetter[TransformContext]{
		Getter: func(_ context.Context, tCtx TransformContext) (any, error) {
			if summaryDataPoint, ok := tCtx.GetDataPoint().(pmetric.SummaryDataPoint); ok {
				return summaryDataPoint.QuantileValues(), nil
			}
			return nil, nil
		},
		Setter: func(_ context.Context, tCtx TransformContext, val any) error {
			if newQuantileValues, ok := val.(pmetric.SummaryDataPointValueAtQuantileSlice); ok {
				if summaryDataPoint, ok := tCtx.GetDataPoint().(pmetric.SummaryDataPoint); ok {
					newQuantileValues.CopyTo(summaryDataPoint.QuantileValues())
				}
			}
			return nil
		},
	}
}
