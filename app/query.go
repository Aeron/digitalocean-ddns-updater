package main

import (
	"errors"
	"net/url"
	"strings"
)

const defaultType = "A"
const anotherType = "AAAA"

type params struct {
	kind   string
	name   string
	domain string
	token  string
	addr   string
}

func parseParams(v *url.Values) (*params, error) {
	domain, token, addr := v.Get("domain"), v.Get("token"), v.Get("ip")
	kind := strings.ToUpper(v.Get("type"))

	if domain == "" || token == "" || addr == "" {
		return nil, errors.New("Empty domain, token, or IP value")
	} else if kind == "" {
		kind = defaultType
	} else if kind != defaultType && kind != anotherType {
		return nil, errors.New("Invalid type")
	}

	if parts := strings.SplitN(domain, ".", 2); parts[0] != "" && parts[1] != "" {
		return &params{kind, parts[0], parts[1], token, addr}, nil
	}
	return nil, errors.New("Invalid record name")
}
