package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// Represents a minimal DigitalOcean client.
type digitalOceanClient struct {
	client  *godo.Client
	timeout time.Duration
}

// Retrieves a record identifier for a given domain and record name pair.
func (c *digitalOceanClient) getDNSRecordId(domain, kind, name string) (*int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var err error
	opts := godo.ListOptions{Page: 0, PerPage: 0}

	if rec, res, err := c.client.Domains.RecordsByTypeAndName(
		ctx, domain, kind, name, &opts,
	); err == nil {
		if res.StatusCode != 200 {
			return nil, errors.New(fmt.Sprintf("Unexpected response: %s", res.Status))
		} else if len(rec) >= 1 && rec[0].ID >= 0 {
			return &rec[0].ID, nil
		}
		return nil, errors.New("Record not found")
	}
	return nil, err
}

// Updates a record value for a given domain and record identifier pair.
func (c *digitalOceanClient) updateDNSRecord(domain string, id int, ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var err error
	req := godo.DomainRecordEditRequest{Data: ip}

	if _, res, err := c.client.Domains.EditRecord(ctx, domain, id, &req); err == nil {
		if res.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Unexpected response: %s", res.Status))
		}
		return nil
	}
	return err
}
