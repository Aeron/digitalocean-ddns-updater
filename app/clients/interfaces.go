package clients

import (
	"context"

	"github.com/digitalocean/godo"
)

// Represents a simplified interface for managing DNS with the DigitalOcean API.
type DigitalOceanDomainsService interface {
	RecordsByTypeAndName(
		context.Context, string, string, string, *godo.ListOptions,
	) ([]godo.DomainRecord, *godo.Response, error)
	EditRecord(
		context.Context, string, int, *godo.DomainRecordEditRequest,
	) (*godo.DomainRecord, *godo.Response, error)
}
