package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const defaultType string = "A"
const anotherType string = "AAAA"

// Represents a usual domain name.
//
// Basically, it is a valid DNS record name, minus the ending dot option,
// minus the single word option.
const domainName string = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})+$`
const ipV4Addr string = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`
const ipV6Addr string = `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`

var rxDomainName = regexp.MustCompile(domainName)
var rxIPV4Addr = regexp.MustCompile(ipV4Addr)
var rxIPV6Addr = regexp.MustCompile(ipV6Addr)

// Represents app supported query parameters.
type Params struct {
	Kind  string
	Name  string
	Token string
	Addr  string
}

// Returns a domain name if the name contains one.
func (p *Params) Domain() (string, error) {
	parts := strings.Split(p.Name, ".")
	length := len(parts)

	if length >= 2 {
		return fmt.Sprintf("%s.%s", parts[length-2], parts[length-1]), nil
	}
	return "", errors.New("Invalid domain name")
}

// Parses app supported query parameters.
func ParseParams(v *url.Values) (*Params, error) {
	kind := strings.ToUpper(strings.TrimSpace(v.Get("type")))
	name := strings.TrimSpace(v.Get("domain"))
	token := strings.TrimSpace(v.Get("token"))
	addr := strings.TrimSpace(v.Get("ip"))

	if name == "" || token == "" || addr == "" {
		return nil, errors.New("Empty domain, token, or IP value")
	} else if kind == "" {
		kind = defaultType
	} else if kind != defaultType && kind != anotherType {
		return nil, errors.New("Invalid type")
	}

	if kind == defaultType && !rxIPV4Addr.MatchString(addr) {
		return nil, errors.New("Invalid IPv4 address")
	} else if kind == anotherType && !rxIPV6Addr.MatchString(addr) {
		return nil, errors.New("Invalid IPv6 address")
	}

	// TODO: we can resolve domain name if we want to be 100% sure it exists. Should we?
	if !rxDomainName.MatchString(name) {
		return nil, errors.New("Invalid record name")
	}

	return &Params{kind, name, token, addr}, nil
}
