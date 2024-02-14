package main

import (
	"errors"
	"net/url"
	"strings"
)

type params struct {
	name   string
	domain string
	token  string
	addr   string
}

func parseParams(v *url.Values) (*params, error) {
	domain, token, ip := v.Get("domain"), v.Get("token"), v.Get("ip")

	if domain == "" || token == "" || ip == "" {
		return nil, errors.New("Empty domain, token, or IP value")
	}

	if parts := strings.SplitN(domain, ".", 2); parts[0] != "" && parts[1] != "" {
		return &params{parts[0], parts[1], token, ip}, nil
	}
	return nil, errors.New("Invalid record name")
}
