package main_test

import (
	"net/url"
	"testing"

	query "github.com/aeron/digitalocean-ddns-updater/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParams(t *testing.T) {
	tests := []struct {
		name              string
		query             url.Values
		expectedErrFunc   func(t require.TestingT, object interface{}, msgAndArgs ...interface{})
		expectedParamFunc func(t require.TestingT, object interface{}, msgAndArgs ...interface{})
		expectedKind      string
		expectedName      string
		expectedDomain    string
		expectedToken     string
		expectedAddr      string
	}{
		{
			name: "normal-case",
			query: url.Values{
				"type":   []string{"A"},
				"domain": []string{"test.example.com"},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.Nil,
			expectedParamFunc: require.NotNil,
			expectedKind:      "A",
			expectedName:      "test",
			expectedDomain:    "example.com",
			expectedToken:     "1234567890",
			expectedAddr:      "192.168.1.1",
		},
		{
			name: "missing-params",
			query: url.Values{
				"type":   []string{""},
				"domain": []string{"test.example.com"},
				"token":  []string{""},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.NotNil,
			expectedParamFunc: require.Nil,
		},
		{
			name: "invalid-type",
			query: url.Values{
				"type":   []string{"AAAAA"},
				"domain": []string{"test.example.com"},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.NotNil,
			expectedParamFunc: require.Nil,
		},
		{
			name: "invalid-name",
			query: url.Values{
				"type":   []string{"AAAAA"},
				"domain": []string{"test"},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.NotNil,
			expectedParamFunc: require.Nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := query.ParseParams(&tt.query)

			tt.expectedErrFunc(t, err)
			tt.expectedParamFunc(t, params)

			if params != nil {
				assert.Equal(t, tt.expectedKind, params.Kind)
				assert.Equal(t, tt.expectedName, params.Name)
				assert.Equal(t, tt.expectedDomain, params.Domain)
				assert.Equal(t, tt.expectedToken, params.Token)
				assert.Equal(t, tt.expectedAddr, params.Addr)
			}
		})
	}
}
