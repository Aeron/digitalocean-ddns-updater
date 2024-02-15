package main_test

import (
	"net/url"
	"testing"

	query "github.com/aeron/digitalocean-ddns-updater/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsDomain(t *testing.T) {
	tests := []struct {
		name              string
		params            query.Params
		expectedErrFunc   func(t require.TestingT, object interface{}, msgAndArgs ...interface{})
		expectedValueFunc func(t require.TestingT, object interface{}, msgAndArgs ...interface{})
		expectedValue     string
	}{
		{
			name:              "normal-domain",
			params:            query.Params{Name: "example.com"},
			expectedErrFunc:   require.Nil,
			expectedValueFunc: require.NotNil,
			expectedValue:     "example.com",
		},
		{
			name:              "normal-sub",
			params:            query.Params{Name: "test.example.com"},
			expectedErrFunc:   require.Nil,
			expectedValueFunc: require.NotNil,
			expectedValue:     "example.com",
		},
		{
			name:              "normal-subsub",
			params:            query.Params{Name: "test.app.example.com"},
			expectedErrFunc:   require.Nil,
			expectedValueFunc: require.NotNil,
			expectedValue:     "example.com",
		},
		{
			name:              "invalid-name",
			params:            query.Params{Name: "example"},
			expectedErrFunc:   require.NotNil,
			expectedValueFunc: require.NotNil,
			expectedValue:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := tt.params.Domain()

			tt.expectedErrFunc(t, err)
			tt.expectedValueFunc(t, domain)

			assert.Equal(t, tt.expectedValue, domain)
		})
	}
}

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
			name: "normal-ipv4",
			query: url.Values{
				"type":   []string{"A"},
				"domain": []string{"example.com"},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.Nil,
			expectedParamFunc: require.NotNil,
			expectedKind:      "A",
			expectedName:      "example.com",
			expectedToken:     "1234567890",
			expectedAddr:      "192.168.1.1",
		},
		{
			name: "normal-ipv6",
			query: url.Values{
				"type":   []string{"AAAA"},
				"domain": []string{"test.app.example.com"},
				"token":  []string{"1234567890"},
				"ip":     []string{"::ffff:c0a8:101"},
			},
			expectedErrFunc:   require.Nil,
			expectedParamFunc: require.NotNil,
			expectedKind:      "AAAA",
			expectedName:      "test.app.example.com",
			expectedToken:     "1234567890",
			expectedAddr:      "::ffff:c0a8:101",
		},
		{
			name: "normal-extra-space",
			query: url.Values{
				"type":   []string{" AAAA "},
				"domain": []string{"example.com "},
				"token":  []string{"1234567890"},
				"ip":     []string{" ::ffff:c0a8:101 "},
			},
			expectedErrFunc:   require.Nil,
			expectedParamFunc: require.NotNil,
			expectedKind:      "AAAA",
			expectedName:      "example.com",
			expectedToken:     "1234567890",
			expectedAddr:      "::ffff:c0a8:101",
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
			name: "invalid-type-addr",
			query: url.Values{
				"type":   []string{"A"},
				"domain": []string{"test.example.com"},
				"token":  []string{"1234567890"},
				"ip":     []string{"::ffff:c0a8:101"},
			},
			expectedErrFunc:   require.NotNil,
			expectedParamFunc: require.Nil,
		},
		{
			name: "invalid-name-one-word",
			query: url.Values{
				"type":   []string{"AAAAA"},
				"domain": []string{"example"},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.NotNil,
			expectedParamFunc: require.Nil,
		},
		{
			name: "invalid-name-end-dot",
			query: url.Values{
				"type":   []string{"AAAAA"},
				"domain": []string{"example.com."},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1"},
			},
			expectedErrFunc:   require.NotNil,
			expectedParamFunc: require.Nil,
		},
		{
			name: "invalid-addr",
			query: url.Values{
				"type":   []string{"AAAA"},
				"domain": []string{"example.com"},
				"token":  []string{"1234567890"},
				"ip":     []string{"192.168.1.1.1"},
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
				assert.Equal(t, tt.expectedToken, params.Token)
				assert.Equal(t, tt.expectedAddr, params.Addr)
			}
		})
	}
}
