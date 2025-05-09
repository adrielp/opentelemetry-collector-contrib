// Go API over pdh syscalls
//go:build windows

package win_perf_counters // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters/internal/third_party/telegraf/win_perf_counters"

import (
	"errors"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

// CounterValue is abstraction for PDH_FMT_COUNTERVALUE_ITEM_DOUBLE
type CounterValue struct {
	InstanceName string
	Value        float64
}

// RawCounterValue is abstraction for PDH_RAW_COUNTER_ITEM
type RawCounterValue struct {
	InstanceName string
	RawValue     int64
}

// PerformanceQuery provides wrappers around Windows performance counters API for easy usage in GO
type PerformanceQuery interface {
	Open() error
	Close() error
	AddCounterToQuery(counterPath string) (PDH_HCOUNTER, error)
	AddEnglishCounterToQuery(counterPath string) (PDH_HCOUNTER, error)
	GetCounterPath(counterHandle PDH_HCOUNTER) (string, error)
	GetFormattedCounterValueDouble(hCounter PDH_HCOUNTER) (float64, error)
	GetFormattedCounterArrayDouble(hCounter PDH_HCOUNTER) ([]CounterValue, error)
	GetRawCounterValue(hCounter PDH_HCOUNTER) (int64, error)
	GetRawCounterArray(hCounter PDH_HCOUNTER) ([]RawCounterValue, error)
	CollectData() error
	CollectDataWithTime() (time.Time, error)
	IsVistaOrNewer() bool
}

// PdhError represents error returned from Performance Counters API
type PdhError struct {
	ErrorCode uint32
	errorText string
}

func (m *PdhError) Error() string {
	return m.errorText
}

func NewPdhError(code uint32) error {
	return &PdhError{
		ErrorCode: code,
		errorText: PdhFormatError(code),
	}
}

// PerformanceQueryImpl is implementation of PerformanceQuery interface, which calls phd.dll functions
type PerformanceQueryImpl struct {
	query PDH_HQUERY
}

// Open creates a new counterPath that is used to manage the collection of performance data.
// It returns counterPath handle used for subsequent calls for adding counters and querying data
func (m *PerformanceQueryImpl) Open() error {
	if m.query != 0 {
		err := m.Close()
		if err != nil {
			return err
		}
	}
	var handle PDH_HQUERY

	if ret := PdhOpenQuery(0, 0, &handle); ret != ERROR_SUCCESS {
		return NewPdhError(ret)
	}
	m.query = handle
	return nil
}

// Close closes the counterPath, releases associated counter handles and frees resources
func (m *PerformanceQueryImpl) Close() error {
	if m.query == 0 {
		return errors.New("uninitialised query")
	}

	if ret := PdhCloseQuery(m.query); ret != ERROR_SUCCESS {
		return NewPdhError(ret)
	}
	m.query = 0
	return nil
}

func (m *PerformanceQueryImpl) AddCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	var counterHandle PDH_HCOUNTER
	if m.query == 0 {
		return 0, errors.New("uninitialised query")
	}

	if ret := PdhAddCounter(m.query, counterPath, 0, &counterHandle); ret != ERROR_SUCCESS {
		return 0, NewPdhError(ret)
	}
	return counterHandle, nil
}

func (m *PerformanceQueryImpl) AddEnglishCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	var counterHandle PDH_HCOUNTER
	if m.query == 0 {
		return 0, errors.New("uninitialised query")
	}
	if ret := PdhAddEnglishCounter(m.query, counterPath, 0, &counterHandle); ret != ERROR_SUCCESS {
		return 0, NewPdhError(ret)
	}
	return counterHandle, nil
}

// GetCounterPath return counter information for given handle
func (m *PerformanceQueryImpl) GetCounterPath(counterHandle PDH_HCOUNTER) (string, error) {
	var bufSize uint32
	var buff []byte
	var ret uint32
	if ret = PdhGetCounterInfo(counterHandle, 0, &bufSize, nil); ret == PDH_MORE_DATA {
		buff = make([]byte, bufSize)
		bufSize = uint32(len(buff))
		if ret = PdhGetCounterInfo(counterHandle, 0, &bufSize, &buff[0]); ret == ERROR_SUCCESS {
			ci := (*PDH_COUNTER_INFO)(unsafe.Pointer(&buff[0]))
			return UTF16PtrToString(ci.SzFullPath), nil
		}
	}
	return "", NewPdhError(ret)
}

// GetFormattedCounterValueDouble computes a displayable value for the specified counter
func (m *PerformanceQueryImpl) GetFormattedCounterValueDouble(hCounter PDH_HCOUNTER) (float64, error) {
	var counterType uint32
	var value PDH_FMT_COUNTERVALUE_DOUBLE
	var ret uint32

	if ret = PdhGetFormattedCounterValueDouble(hCounter, &counterType, &value); ret == ERROR_SUCCESS {
		if value.CStatus == PDH_CSTATUS_VALID_DATA || value.CStatus == PDH_CSTATUS_NEW_DATA {
			return value.DoubleValue, nil
		} else {
			return 0, NewPdhError(value.CStatus)
		}
	} else {
		return 0, NewPdhError(ret)
	}
}

func (m *PerformanceQueryImpl) GetFormattedCounterArrayDouble(hCounter PDH_HCOUNTER) ([]CounterValue, error) {
	var buffSize uint32
	var itemCount uint32
	var ret uint32

	if ret = PdhGetFormattedCounterArrayDouble(hCounter, &buffSize, &itemCount, nil); ret == PDH_MORE_DATA {
		buff := make([]byte, buffSize)

		if ret = PdhGetFormattedCounterArrayDouble(hCounter, &buffSize, &itemCount, &buff[0]); ret == ERROR_SUCCESS {
			items := unsafe.Slice((*PDH_FMT_COUNTERVALUE_ITEM_DOUBLE)(unsafe.Pointer(&buff[0])), itemCount)
			values := make([]CounterValue, 0, itemCount)
			for _, item := range items {
				if item.FmtValue.CStatus == PDH_CSTATUS_VALID_DATA || item.FmtValue.CStatus == PDH_CSTATUS_NEW_DATA {
					val := CounterValue{UTF16PtrToString(item.SzName), item.FmtValue.DoubleValue}
					values = append(values, val)
				}
			}
			return values, nil
		}
	}
	return nil, NewPdhError(ret)
}

func (m *PerformanceQueryImpl) GetRawCounterValue(hCounter PDH_HCOUNTER) (int64, error) {
	var counterType uint32
	var rawValue PDH_RAW_COUNTER
	var ret uint32

	if ret = PdhGetRawCounterValue(hCounter, &counterType, &rawValue); ret == ERROR_SUCCESS {
		if rawValue.CStatus == PDH_CSTATUS_VALID_DATA {
			return rawValue.FirstValue, nil
		} else {
			return 0, NewPdhError(rawValue.CStatus)
		}
	} else {
		return 0, NewPdhError(ret)
	}
}

func (m *PerformanceQueryImpl) GetRawCounterArray(hCounter PDH_HCOUNTER) ([]RawCounterValue, error) {
	var buffSize uint32
	var itemCount uint32
	var ret uint32

	if ret = PdhGetRawCounterArrayW(hCounter, &buffSize, &itemCount, nil); ret == PDH_MORE_DATA {
		buff := make([]byte, buffSize)

		if ret = PdhGetRawCounterArrayW(hCounter, &buffSize, &itemCount, &buff[0]); ret == ERROR_SUCCESS {
			items := unsafe.Slice((*PDH_RAW_COUNTER_ITEM)(unsafe.Pointer(&buff[0])), itemCount)
			values := make([]RawCounterValue, 0, itemCount)
			for _, item := range items {
				if item.RawValue.CStatus == PDH_CSTATUS_VALID_DATA {
					val := RawCounterValue{UTF16PtrToString(item.SzName), item.RawValue.FirstValue}
					values = append(values, val)
				}
			}
			return values, nil
		}
	}
	return nil, NewPdhError(ret)
}

func (m *PerformanceQueryImpl) CollectData() error {
	var ret uint32
	if m.query == 0 {
		return errors.New("uninitialised query")
	}

	if ret = PdhCollectQueryData(m.query); ret != ERROR_SUCCESS {
		return NewPdhError(ret)
	}
	return nil
}

func (m *PerformanceQueryImpl) CollectDataWithTime() (time.Time, error) {
	if m.query == 0 {
		return time.Now(), errors.New("uninitialised query")
	}
	ret, mtime := PdhCollectQueryDataWithTime(m.query)
	if ret != ERROR_SUCCESS {
		return time.Now(), NewPdhError(ret)
	}
	return mtime, nil
}

func (m *PerformanceQueryImpl) IsVistaOrNewer() bool {
	return PdhAddEnglishCounterSupported()
}

// ExpandWildCardPath examines local computer and returns those counter paths that match the given counter path which contains wildcard characters.
func ExpandWildCardPath(counterPath string) ([]string, error) {
	var bufSize uint32
	var buff []uint16
	var ret uint32

	if ret = PdhExpandWildCardPath(counterPath, nil, &bufSize); ret == PDH_MORE_DATA {
		buff = make([]uint16, bufSize)
		bufSize = uint32(len(buff))
		ret = PdhExpandWildCardPath(counterPath, &buff[0], &bufSize)
		if ret == ERROR_SUCCESS {
			list := UTF16ToStringArray(buff)
			return list, nil
		}
	}
	return nil, NewPdhError(ret)
}

// UTF16PtrToString converts Windows API LPTSTR (pointer to string) to go string
func UTF16PtrToString(s *uint16) string {
	if s == nil {
		return ""
	}

	len := 0
	curPtr := unsafe.Pointer(s)
	for *(*uint16)(curPtr) != 0 {
		curPtr = unsafe.Pointer(uintptr(curPtr) + unsafe.Sizeof(*s))
		len++
	}

	slice := unsafe.Slice(s, len)
	return string(utf16.Decode(slice))
}

// UTF16ToStringArray converts list of Windows API NULL terminated strings to go string array
func UTF16ToStringArray(buf []uint16) []string {
	var strings []string
	nextLineStart := 0
	stringLine := syscall.UTF16ToString(buf)
	for stringLine != "" {
		strings = append(strings, stringLine)
		nextLineStart += len([]rune(stringLine)) + 1
		remainingBuf := buf[nextLineStart:]
		stringLine = syscall.UTF16ToString(remainingBuf)
	}
	return strings
}
