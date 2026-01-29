package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortPath(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{
			name:    "Test get short value of URL",
			value:   "https://vk.com",
			want:    "XxLlqM",
			wantErr: false,
		},
		{
			name:    "Test get short value of empty string",
			value:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := GetShortPath(test.value)
			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.want, value)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		want     bool
	}{
		{
			name:     "Valid URL",
			inputURL: "https://vk.ccom",
			want:     true,
		},
		{
			name:     "Invalid URL",
			inputURL: "123",
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsValidURL(test.inputURL)
			assert.Equal(t, result, test.want)
		})
	}
}
