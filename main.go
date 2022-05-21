package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

const (
	defaultAddress      = ":8080"
	defaultEndpoint     = "/ddns"
	defaultLimitRPS     = .01
	defaultLimitBurst   = 1
	envVarDOAPIToken    = "DIGITALOCEAN_API_TOKEN"
	envVarSecurityToken = "SECURITY_TOKEN"
	envVarLimitRPS      = "LIMIT_RPS"
	envVarLimitBurst    = "LIMIT_BURST"
)

var (
	address    = flag.String("address", defaultAddress, "address to listen on")
	endpoint   = flag.String("endpoint", defaultEndpoint, "endpoint to handle updates")
	doAPIToken = flag.String(
		"digitalocean-api-token",
		getEnvString(envVarDOAPIToken, ""),
		"DigitalOcean API token",
	)
	securityToken = flag.String(
		"security-token",
		getEnvString(envVarSecurityToken, ""),
		"security token",
	)
	limitRPS = flag.Float64(
		"limit-rps",
		getEnvFloat64(envVarLimitRPS, defaultLimitRPS),
		"limit requests per second",
	)
	limitBurst = flag.Int(
		"limit-burst",
		getEnvInt(envVarLimitBurst, defaultLimitBurst),
		"limit a single burst size",
	)
)

var limiter = rate.NewLimiter(rate.Limit(*limitRPS), *limitBurst)
var digitalOceanClient *godo.Client

func getEnvString(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnvString(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func getEnvFloat64(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(getEnvString(key, ""), 64)
	if err != nil {
		return fallback
	}
	return value
}

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			delay := limiter.Reserve().Delay().Seconds()
			w.Header().Add("Retry-After", fmt.Sprintf("%.0f", math.Ceil(delay)))
			http.Error(
				w,
				http.StatusText(http.StatusTooManyRequests),
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func checkRecord(domain string, name string) (*godo.DomainRecord, error) {
	records, response, err := digitalOceanClient.Domains.Records(
		context.TODO(),
		domain,
		&godo.ListOptions{Page: 0, PerPage: 0},
	)

	if err != nil {
		log.Print(err)
		return &godo.DomainRecord{}, err
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		log.Println("Check failed:", response.Status)
		return &godo.DomainRecord{}, errors.New(response.Status)
	}

	for _, r := range records {
		if r.Name == name {
			log.Printf("Record %s.%s found", name, domain)
			return &r, nil
		}
	}

	log.Printf("Record %s.%s not found", name, domain)
	return nil, errors.New("Record not found")
}

func updateDNS(domain string, id int, ip string) error {
	_, response, err := digitalOceanClient.Domains.EditRecord(
		context.TODO(),
		domain,
		id,
		&godo.DomainRecordEditRequest{Data: ip},
	)

	if err != nil {
		log.Println("Request failed:", err)
		return err
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		log.Println("Update failed:", response.Status)
		return errors.New(response.Status)
	}

	log.Print("Updated successfully")
	return nil
}

func ddnsHandler(res http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	domain := query.Get("domain")
	token := query.Get("token")
	ip := query.Get("ip")

	res.Header().Set("Content-Type", "text/plain")

	if domain == "" || token == "" || ip == "" {
		http.Error(res, "Invalid domain, token or IP", http.StatusBadRequest)
		return
	}

	if token != *securityToken {
		http.Error(res, "Authentication failed", http.StatusUnauthorized)
		return
	}

	log.Printf("Got new IP (%s) for %s", ip, domain)

	parts := strings.SplitN(domain, ".", 2)

	name := parts[0]
	domain = parts[1]

	var record *godo.DomainRecord

	if record, err := checkRecord(domain, name); err != nil || record.ID < 0 {
		http.Error(res, "No record found", http.StatusNotFound)
		return
	}

	if err := updateDNS(domain, record.ID, ip); err != nil {
		http.Error(res, "Failed", http.StatusNoContent)
		return
	}

	res.Write([]byte("Done"))
}

func main() {
	flag.Parse()

	if *doAPIToken == "" {
		log.Fatalln("DigitalOcean API token is required")
	}

	if *securityToken == "" {
		hash := sha512.Sum512_256([]byte(*doAPIToken))
		hashHex := hex.EncodeToString(hash[:])
		securityToken = &hashHex
	}

	tokenPrepared := &oauth2.Token{AccessToken: *doAPIToken}
	digitalOceanClient = godo.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(tokenPrepared),
		),
	)

	mux := http.NewServeMux()
	mux.HandleFunc(defaultEndpoint, ddnsHandler)

	go func() {
		log.Println("Starting server on", *address)
		log.Println("Auth token:", *securityToken)

		if err := http.ListenAndServe(*address, limit(mux)); err != nil {
			log.Fatalln("Server error:", err)
		}
	}()

	select {}
}
