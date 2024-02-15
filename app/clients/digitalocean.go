package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// Represents a DigitalOcean domain service client.
type DigitalOceanDomains struct {
	Op      DigitalOceanDomainsService
	Timeout time.Duration
}

// Retrieves a record identifier for a given domain and record name pair.
func (c *DigitalOceanDomains) GetDNSRecordId(domain, kind, name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	if rec, res, err := c.Op.RecordsByTypeAndName(
		ctx, domain, kind, name, &godo.ListOptions{Page: 0, PerPage: 0},
	); err != nil {
		return 0, err
	} else if res.StatusCode != 200 {
		return 0, errors.New(fmt.Sprintf("Unexpected response: %s", res.Status))
	} else if len(rec) >= 1 && rec[0].ID >= 0 {
		return rec[0].ID, nil
	}

	return 0, errors.New("Record not found")
}

// Updates a record value for a given domain and record identifier pair.
func (c *DigitalOceanDomains) UpdateDNSRecord(domain string, id int, ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	if _, res, err := c.Op.EditRecord(
		ctx, domain, id, &godo.DomainRecordEditRequest{Data: ip},
	); err != nil {
		return err
	} else if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Unexpected response: %s", res.Status))
	}

	return nil
}
