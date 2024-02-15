package clients_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aeron/digitalocean-ddns-updater/app/clients"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDomainsService struct {
	mock.Mock
}

func (m *mockDomainsService) RecordsByTypeAndName(
	ctx context.Context, domain, kind, name string, opts *godo.ListOptions,
) ([]godo.DomainRecord, *godo.Response, error) {
	args := m.Called(ctx, domain, kind, name, opts)
	return args.Get(0).([]godo.DomainRecord), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockDomainsService) EditRecord(
	ctx context.Context, domain string, id int, req *godo.DomainRecordEditRequest,
) (*godo.DomainRecord, *godo.Response, error) {
	args := m.Called(ctx, domain, id, req)
	return args.Get(0).(*godo.DomainRecord), args.Get(1).(*godo.Response), args.Error(2)
}

func TestGetDNSRecordId(t *testing.T) {
	tests := []struct {
		name              string
		serviceOp         *mockDomainsService
		timeout           time.Duration
		callDomain        string
		callType          string
		callName          string
		returnRecord      []godo.DomainRecord
		returnResponse    *godo.Response
		returnError       error
		expectedErrFunc   func(require.TestingT, interface{}, ...interface{})
		expectedValueFunc func(require.TestingT, interface{}, ...interface{})
		expectedValue     int
	}{
		{
			name:              "record-exists",
			serviceOp:         &mockDomainsService{},
			timeout:           5 * time.Second,
			callDomain:        "example.com",
			callType:          "A",
			callName:          "test.example.com",
			returnRecord:      []godo.DomainRecord{{ID: 123}},
			returnResponse:    &godo.Response{Response: &http.Response{StatusCode: 200}},
			returnError:       nil,
			expectedErrFunc:   require.Nil,
			expectedValueFunc: require.NotNil,
			expectedValue:     123,
		},
		{
			name:              "record-does-not",
			serviceOp:         &mockDomainsService{},
			timeout:           5 * time.Second,
			callDomain:        "example.com",
			callType:          "A",
			callName:          "test.example.com",
			returnRecord:      []godo.DomainRecord{},
			returnResponse:    &godo.Response{Response: &http.Response{StatusCode: 200}},
			returnError:       nil,
			expectedErrFunc:   require.NotNil,
			expectedValueFunc: require.NotNil,
			expectedValue:     0,
		},
		{
			name:              "unexpected-response",
			serviceOp:         &mockDomainsService{},
			timeout:           5 * time.Second,
			callDomain:        "example.com",
			callType:          "A",
			callName:          "test.example.com",
			returnRecord:      []godo.DomainRecord{},
			returnResponse:    &godo.Response{Response: &http.Response{StatusCode: 400}},
			returnError:       nil,
			expectedErrFunc:   require.NotNil,
			expectedValueFunc: require.NotNil,
			expectedValue:     0,
		},
		{
			name:              "invalid-type",
			serviceOp:         &mockDomainsService{},
			timeout:           5 * time.Second,
			callDomain:        "example.com",
			callType:          "",
			callName:          "test.example.com",
			returnRecord:      []godo.DomainRecord{},
			returnResponse:    nil,
			returnError:       errors.New("cannot be an empty string"),
			expectedErrFunc:   require.NotNil,
			expectedValueFunc: require.NotNil,
			expectedValue:     0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceOp := tt.serviceOp
			client := clients.DigitalOceanDomains{serviceOp, tt.timeout}

			serviceOp.On(
				"RecordsByTypeAndName",
				mock.Anything,
				tt.callDomain,
				tt.callType,
				tt.callName,
				mock.Anything,
			).Return(tt.returnRecord, tt.returnResponse, tt.returnError)

			id, err := client.GetDNSRecordId(tt.callDomain, tt.callType, tt.callName)

			tt.expectedErrFunc(t, err, id)
			tt.expectedValueFunc(t, id)

			assert.Equal(t, tt.expectedValue, id)
		})
	}
}

func TestUpdateDNSRecord(t *testing.T) {
	tests := []struct {
		name            string
		serviceOp       *mockDomainsService
		timeout         time.Duration
		callDomain      string
		callId          int
		callIP          string
		returnRecord    *godo.DomainRecord
		returnResponse  *godo.Response
		returnError     error
		expectedErrFunc func(require.TestingT, interface{}, ...interface{})
	}{
		{
			name:            "success",
			serviceOp:       &mockDomainsService{},
			timeout:         5 * time.Second,
			callDomain:      "example.com",
			callId:          123,
			callIP:          "192.168.1.1",
			returnRecord:    &godo.DomainRecord{ID: 123},
			returnResponse:  &godo.Response{Response: &http.Response{StatusCode: 200}},
			returnError:     nil,
			expectedErrFunc: require.Nil,
		},
		{
			name:            "unexpected-response",
			serviceOp:       &mockDomainsService{},
			timeout:         5 * time.Second,
			callDomain:      "example.com",
			callId:          123,
			callIP:          "192.168.1.1",
			returnRecord:    &godo.DomainRecord{},
			returnResponse:  &godo.Response{Response: &http.Response{StatusCode: 500}},
			returnError:     nil,
			expectedErrFunc: require.NotNil,
		},
		{
			name:            "invalid-id",
			serviceOp:       &mockDomainsService{},
			timeout:         5 * time.Second,
			callDomain:      "example.com",
			callId:          0,
			callIP:          "192.168.1.1",
			returnRecord:    nil,
			returnResponse:  nil,
			returnError:     errors.New("cannot be less than 1"),
			expectedErrFunc: require.NotNil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceOp := tt.serviceOp
			client := clients.DigitalOceanDomains{serviceOp, tt.timeout}

			serviceOp.On(
				"EditRecord",
				mock.Anything,
				tt.callDomain,
				tt.callId,
				&godo.DomainRecordEditRequest{Data: tt.callIP},
				mock.Anything,
			).Return(tt.returnRecord, tt.returnResponse, tt.returnError)

			err := client.UpdateDNSRecord(tt.callDomain, tt.callId, tt.callIP)

			tt.expectedErrFunc(t, err)
		})
	}
}
