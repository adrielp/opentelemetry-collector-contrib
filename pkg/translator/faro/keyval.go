// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package faro // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/faro"

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	faroTypes "github.com/grafana/faro/pkg/go"
	om "github.com/wk8/go-ordered-map/v2"
)

// keyVal is an ordered map of string to interface
type keyVal = om.OrderedMap[string, any]

// newKeyVal creates new empty keyVal
func newKeyVal() *keyVal {
	return om.New[string, any]()
}

// keyValFromMap will instantiate keyVal from a map[string]string
func keyValFromMap(m map[string]string) *keyVal {
	kv := newKeyVal()
	for _, k := range slices.Sorted(maps.Keys(m)) {
		keyValAdd(kv, k, m[k])
	}
	return kv
}

// keyValFromFloatMap will instantiate keyVal from a map[string]float64
func keyValFromFloatMap(m map[string]float64) *keyVal {
	kv := newKeyVal()
	for _, k := range slices.Sorted(maps.Keys(m)) {
		kv.Set(k, m[k])
	}
	return kv
}

// mergeKeyVal will merge source in target
func mergeKeyVal(target, source *keyVal) {
	for el := source.Oldest(); el != nil; el = el.Next() {
		target.Set(el.Key, el.Value)
	}
}

// mergeKeyValWithPrefix will merge source in target, adding a prefix to each key being merged in
func mergeKeyValWithPrefix(target, source *keyVal, prefix string) {
	for el := source.Oldest(); el != nil; el = el.Next() {
		target.Set(fmt.Sprintf("%s%s", prefix, el.Key), el.Value)
	}
}

// keyValAdd adds a key + value string pair to kv
func keyValAdd(kv *keyVal, key, value string) {
	if value != "" {
		kv.Set(key, value)
	}
}

// keyValToInterfaceSlice converts keyVal to []interface{}, typically used for logging
func keyValToInterfaceSlice(kv *keyVal) []any {
	slice := make([]any, kv.Len()*2)
	idx := 0
	for el := kv.Oldest(); el != nil; el = el.Next() {
		slice[idx] = el.Key
		idx++
		slice[idx] = el.Value
		idx++
	}
	return slice
}

// logToKeyVal represents a Log object as keyVal
func logToKeyVal(l faroTypes.Log) *keyVal {
	kv := newKeyVal()

	// default to info level, prioritize log level if set
	level := string(faroTypes.LogLevelInfo)
	if l.LogLevel != "" {
		level = string(l.LogLevel)
	}

	keyValAdd(kv, faroTimestamp, l.Timestamp.Format(string(faroTypes.TimeFormatRFC3339Milli)))
	keyValAdd(kv, faroKind, string(faroTypes.KindLog))
	keyValAdd(kv, faroLogLevel, level)
	keyValAdd(kv, faroLogMessage, l.Message)
	mergeKeyValWithPrefix(kv, keyValFromMap(l.Context), faroContextPrefix)
	mergeKeyVal(kv, traceToKeyVal(l.Trace))
	mergeKeyVal(kv, actionToKeyVal(l.Action))
	return kv
}

// exceptionToKeyVal represents an Exception object as keyVal
func exceptionToKeyVal(e faroTypes.Exception) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroTimestamp, e.Timestamp.Format(string(faroTypes.TimeFormatRFC3339Milli)))
	keyValAdd(kv, faroKind, string(faroTypes.KindException))
	keyValAdd(kv, faroLogLevel, string(faroTypes.LogLevelError))
	keyValAdd(kv, faroExceptionType, e.Type)
	keyValAdd(kv, faroExceptionValue, e.Value)
	keyValAdd(kv, faroExceptionStacktrace, exceptionToString(e))
	mergeKeyVal(kv, traceToKeyVal(e.Trace))
	mergeKeyValWithPrefix(kv, keyValFromMap(e.Context), faroContextPrefix)
	mergeKeyVal(kv, actionToKeyVal(e.Action))
	return kv
}

// exceptionMessage string is concatenating of the Exception.Type and Exception.Value
func exceptionMessage(e faroTypes.Exception) string {
	return fmt.Sprintf("%s: %s", e.Type, e.Value)
}

// exceptionToString is the string representation of an Exception
func exceptionToString(e faroTypes.Exception) string {
	stacktrace := exceptionMessage(e)
	if e.Stacktrace != nil {
		for _, frame := range e.Stacktrace.Frames {
			stacktrace += frameToString(frame)
		}
	}
	return stacktrace
}

// frameToString function converts a Frame into a human readable string
func frameToString(frame faroTypes.Frame) string {
	module := ""
	if frame.Module != "" {
		module = frame.Module + "|"
	}
	return fmt.Sprintf("\n  at %s (%s%s:%v:%v)", frame.Function, module, frame.Filename, frame.Lineno, frame.Colno)
}

// measurementToKeyVal representation of the measurement object
func measurementToKeyVal(m faroTypes.Measurement) *keyVal {
	kv := newKeyVal()

	keyValAdd(kv, faroTimestamp, m.Timestamp.Format(string(faroTypes.TimeFormatRFC3339Milli)))
	keyValAdd(kv, faroKind, string(faroTypes.KindMeasurement))
	keyValAdd(kv, faroLogLevel, string(faroTypes.LogLevelInfo))
	keyValAdd(kv, faroMeasurementType, m.Type)
	mergeKeyValWithPrefix(kv, keyValFromMap(m.Context), faroContextPrefix)

	for _, k := range slices.Sorted(maps.Keys(m.Values)) {
		keyValAdd(kv, k, fmt.Sprintf("%f", m.Values[k]))
	}
	mergeKeyVal(kv, traceToKeyVal(m.Trace))

	values := make(map[string]float64, len(m.Values))
	for key, value := range m.Values {
		values[key] = value
	}

	mergeKeyValWithPrefix(kv, keyValFromFloatMap(values), faroMeasurementValuePrefix)
	mergeKeyVal(kv, actionToKeyVal(m.Action))
	return kv
}

// eventToKeyVal produces key -> value representation of Event metadata
func eventToKeyVal(e faroTypes.Event) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroTimestamp, e.Timestamp.Format(string(faroTypes.TimeFormatRFC3339Milli)))
	keyValAdd(kv, faroKind, string(faroTypes.KindEvent))
	keyValAdd(kv, faroLogLevel, string(faroTypes.LogLevelInfo))
	keyValAdd(kv, faroEventName, e.Name)
	keyValAdd(kv, faroEventDomain, e.Domain)
	if e.Attributes != nil {
		mergeKeyValWithPrefix(kv, keyValFromMap(e.Attributes), faroEventDataPrefix)
	}
	mergeKeyVal(kv, actionToKeyVal(e.Action))
	mergeKeyVal(kv, traceToKeyVal(e.Trace))
	return kv
}

// actionToKeyVal produces key->value representation of the Action metadata
func actionToKeyVal(a faroTypes.Action) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroActionID, a.ID)
	keyValAdd(kv, faroActionName, a.Name)
	keyValAdd(kv, faroActionParentID, a.ParentID)
	return kv
}

// metaToKeyVal produces key->value representation of the metadata
func metaToKeyVal(m faroTypes.Meta) *keyVal {
	kv := newKeyVal()
	mergeKeyVal(kv, sdkToKeyVal(m.SDK))
	mergeKeyVal(kv, appToKeyVal(m.App))
	mergeKeyVal(kv, userToKeyVal(m.User))
	mergeKeyVal(kv, sessionToKeyVal(m.Session))
	mergeKeyVal(kv, pageToKeyVal(m.Page))
	mergeKeyVal(kv, browserToKeyVal(m.Browser))
	mergeKeyVal(kv, k6ToKeyVal(m.K6))
	mergeKeyVal(kv, viewToKeyVal(m.View))
	mergeKeyVal(kv, geoToKeyVal(m.Geo))
	return kv
}

// sdkToKeyVal produces key->value representation of Sdk metadata
func sdkToKeyVal(sdk faroTypes.SDK) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroSDKName, sdk.Name)
	keyValAdd(kv, faroSDKVersion, sdk.Version)

	if len(sdk.Integrations) > 0 {
		integrations := make([]string, len(sdk.Integrations))

		for i, integration := range sdk.Integrations {
			integrations[i] = sdkIntegrationToString(integration)
		}

		keyValAdd(kv, faroSDKIntegrations, strings.Join(integrations, ","))
	}

	return kv
}

// sdkIntegrationToString is the string representation of an SDKIntegration
func sdkIntegrationToString(i faroTypes.SDKIntegration) string {
	return fmt.Sprintf("%s:%s", i.Name, i.Version)
}

// appToKeyVal produces key-> value representation of App metadata
func appToKeyVal(a faroTypes.App) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroAppName, a.Name)
	keyValAdd(kv, faroAppNamespace, a.Namespace)
	keyValAdd(kv, faroAppRelease, a.Release)
	keyValAdd(kv, faroAppVersion, a.Version)
	keyValAdd(kv, faroAppEnvironment, a.Environment)
	return kv
}

// userToKeyVal produces a key->value representation User metadata
func userToKeyVal(u faroTypes.User) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroUserEmail, u.Email)
	keyValAdd(kv, faroUserID, u.ID)
	keyValAdd(kv, faroUsername, u.Username)
	mergeKeyValWithPrefix(kv, keyValFromMap(u.Attributes), faroUserAttrPrefix)
	return kv
}

// sessionToKeyVal produces key->value representation of the Session metadata
func sessionToKeyVal(s faroTypes.Session) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroSessionID, s.ID)
	mergeKeyValWithPrefix(kv, keyValFromMap(s.Attributes), faroSessionAttrPrefix)
	return kv
}

// pageToKeyVal produces key->val representation of Page metadata
func pageToKeyVal(p faroTypes.Page) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroPageID, p.ID)
	keyValAdd(kv, faroPageURL, p.URL)
	mergeKeyValWithPrefix(kv, keyValFromMap(p.Attributes), faroPageAttrPrefix)

	return kv
}

// browserToKeyVal produces key->value representation of the Browser metadata
func browserToKeyVal(b faroTypes.Browser) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroBrowserName, b.Name)
	keyValAdd(kv, faroBrowserVersion, b.Version)
	keyValAdd(kv, faroBrowserOS, b.OS)
	keyValAdd(kv, faroBrowserMobile, fmt.Sprintf("%v", b.Mobile))
	keyValAdd(kv, faroBrowserUserAgent, b.UserAgent)
	keyValAdd(kv, faroBrowserLanguage, b.Language)
	keyValAdd(kv, faroBrowserViewportWidth, b.ViewportWidth)
	keyValAdd(kv, faroBrowserViewportHeight, b.ViewportHeight)

	if brandsArray, err := b.Brands.AsBrandsArray(); err == nil {
		for i, brand := range brandsArray {
			keyValAdd(kv, fmt.Sprintf("%s%d_%s", faroBrowserBrandPrefix, i, faroBrand), brand.Brand)
			keyValAdd(kv, fmt.Sprintf("%s%d_%s", faroBrowserBrandPrefix, i, faroBrandVersion), brand.Version)
		}
		return kv
	}

	if brandsString, err := b.Brands.AsBrandsString(); err == nil {
		keyValAdd(kv, faroBrowserBrands, brandsString)
		return kv
	}

	return kv
}

// k6ToKeyVal produces a key->value representation K6 metadata
func k6ToKeyVal(k faroTypes.K6) *keyVal {
	kv := newKeyVal()
	if k.IsK6Browser {
		keyValAdd(kv, faroIsK6Browser, strconv.FormatBool(k.IsK6Browser))
	}
	return kv
}

// viewToKeyVal produces a key->value representation View metadata
func viewToKeyVal(v faroTypes.View) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroViewName, v.Name)
	return kv
}

// geoToKeyVal produces a key->value representation Geo metadata
func geoToKeyVal(g faroTypes.Geo) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroGeoContinentIso, g.ContinentISOCode)
	keyValAdd(kv, faroGeoCountryIso, g.CountryISOCode)
	keyValAdd(kv, faroGeoSubdivisionIso, g.SubdivisionISO)
	keyValAdd(kv, faroGeoCity, g.City)
	keyValAdd(kv, faroGeoASNOrg, g.ASNOrg)
	keyValAdd(kv, faroGeoASNID, g.ASNID)
	return kv
}

// traceToKeyVal produces a key->value representation of the trace context object
func traceToKeyVal(tc faroTypes.TraceContext) *keyVal {
	kv := newKeyVal()
	keyValAdd(kv, faroTraceID, tc.TraceID)
	keyValAdd(kv, faroSpanID, tc.SpanID)
	return kv
}
