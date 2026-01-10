package apihelpers

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// StringMapToAnyMap converts map[string]string to map[string]any for API request bodies.
func StringMapToAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// StringID is a type that can unmarshal both string and numeric JSON values to string.
// This handles API inconsistency where some endpoints return IDs as numbers while
// others return them as strings.
type StringID string

// StringIDFromPointer converts a *string to a *StringID.
func StringIDFromPointer(s *string) *StringID {
	if s == nil {
		return nil
	}
	id := StringID(*s)
	return &id
}

// StringsToStringIDs converts []string to []StringID.
func StringsToStringIDs(strs []string) []StringID {
	if strs == nil {
		return nil
	}
	result := make([]StringID, len(strs))
	for i, s := range strs {
		result[i] = StringID(s)
	}
	return result
}

// StringIDsToStrings converts []StringID to []string.
func StringIDsToStrings(ids []StringID) []string {
	if ids == nil {
		return nil
	}
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// UnmarshalJSON implements json.Unmarshaler for StringID.
// It accepts both JSON string and JSON number values, converting numbers to strings.
func (s *StringID) UnmarshalJSON(data []byte) error {
	// Check if it's a null value
	if string(data) == "null" {
		return nil
	}

	// Try to unmarshal as string first (most common case)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringID(str)
		return nil
	}

	// Try to unmarshal as int64
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*s = StringID(strconv.FormatInt(num, 10))
		return nil
	}

	// Try to unmarshal as float64 (handles scientific notation)
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*s = StringID(strconv.FormatInt(int64(f), 10))
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into StringID", string(data))
}

// String returns the string value of StringID.
func (s StringID) String() string {
	return string(s)
}
