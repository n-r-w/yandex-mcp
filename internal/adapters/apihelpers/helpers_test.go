package apihelpers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringID_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		json     string
		expected StringID
		wantErr  bool
	}{
		{
			name:     "string value",
			json:     `"abc123"`,
			expected: "abc123",
			wantErr:  false,
		},
		{
			name:     "integer value",
			json:     `12345`,
			expected: "12345",
			wantErr:  false,
		},
		{
			name:     "large integer value",
			json:     `8000000000000029`,
			expected: "8000000000000029",
			wantErr:  false,
		},
		{
			name:     "zero integer",
			json:     `0`,
			expected: "0",
			wantErr:  false,
		},
		{
			name:     "negative integer",
			json:     `-123`,
			expected: "-123",
			wantErr:  false,
		},
		{
			name:     "null value",
			json:     `null`,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "empty string",
			json:     `""`,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "invalid json",
			json:     `{invalid}`,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "array value",
			json:     `[1,2,3]`,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "object value",
			json:     `{"id": 1}`,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s StringID
			err := json.Unmarshal([]byte(tt.json), &s)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, s)
		})
	}
}

func TestStringID_UnmarshalJSON_InStruct(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		ID   StringID `json:"id"`
		Name string   `json:"name"`
	}

	tests := []struct {
		name     string
		json     string
		expected testStruct
	}{
		{
			name:     "string ID in struct",
			json:     `{"id": "abc123", "name": "test"}`,
			expected: testStruct{ID: "abc123", Name: "test"},
		},
		{
			name:     "integer ID in struct",
			json:     `{"id": 4381, "name": "test"}`,
			expected: testStruct{ID: "4381", Name: "test"},
		},
		{
			name:     "large integer ID in struct",
			json:     `{"id": 8000000000000029, "name": "test"}`,
			expected: testStruct{ID: "8000000000000029", Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s testStruct
			err := json.Unmarshal([]byte(tt.json), &s)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, s)
		})
	}
}

func TestStringID_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       StringID
		expected string
	}{
		{
			name:     "non-empty value",
			id:       StringID("abc123"),
			expected: "abc123",
		},
		{
			name:     "empty value",
			id:       StringID(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.id.String())
		})
	}
}
