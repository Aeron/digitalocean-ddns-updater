package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

var args struct {
	Address       *string  `default:":8080" help:"address to listen on"`
	Endpoint      *string  `default:"/ddns" help:"endpoint path to handle updates"`
	DOAPIToken    *string  `name:"digitalocean-api-token" environ:"DIGITALOCEAN_API_TOKEN" help:"DigitalOcean API token"`
	SecurityToken *string  `name:"security-token" environ:"SECURITY_TOKEN" help:"application security token"`
	LimitRPS      *float64 `name:"limit-rps" default:".01" environ:"LIMIT_RPS" help:"limit requests per second"`
	LimitBurst    *int     `name:"limit-burst" default:"1" environ:"LIMIT_BURST" help:"limit a single burst size"`
}

var limiter *rate.Limiter
var digitalOceanClient *godo.Client

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
		return nil, err
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		log.Println("Check failed:", response.Status)
		return nil, errors.New(response.Status)
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

	if token != *args.SecurityToken {
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
	if err := ParseArgs(&args); err != nil {
		log.Fatalln("argument parsing error:", err.Error())
	}

	if *args.DOAPIToken == "" {
		log.Fatalln("DigitalOcean API token is required")
	}

	if *args.SecurityToken == "" {
		hash := sha512.Sum512_256([]byte(*args.DOAPIToken))
		*args.SecurityToken = hex.EncodeToString(hash[:])
	}

	tokenPrepared := &oauth2.Token{AccessToken: *args.DOAPIToken}
	digitalOceanClient = godo.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(tokenPrepared),
		),
	)

	limiter = rate.NewLimiter(rate.Limit(*args.LimitRPS), *args.LimitBurst)

	mux := http.NewServeMux()
	mux.HandleFunc(*args.Endpoint, ddnsHandler)

	go func() {
		log.Println("Starting server on", *args.Address)
		log.Println("Auth token:", *args.SecurityToken)

		if err := http.ListenAndServe(*args.Address, limit(mux)); err != nil {
			log.Fatalln("Server error:", err)
		}
	}()

	select {}
}
