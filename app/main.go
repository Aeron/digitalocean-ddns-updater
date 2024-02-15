package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	arguments "github.com/aeron/digitalocean-ddns-updater/app/args"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

var args struct {
	Address       *string  `default:":8080" help:"address to listen on"`
	Endpoint      *string  `default:"/ddns" help:"endpoint path to handle updates"`
	DOAPIToken    *string  `name:"digitalocean-api-token" environ:"DIGITALOCEAN_API_TOKEN" help:"DigitalOcean API token"`
	SecurityToken *string  `name:"security-token" environ:"SECURITY_TOKEN" help:"application security token"`
	LimitRPS      *float64 `name:"limit-rps" default:".01" environ:"LIMIT_RPS" help:"limit requests per second"`
	LimitBurst    *int     `name:"limit-burst" default:"1" environ:"LIMIT_BURST" help:"limit a single burst size"`
}

func main() {
	if err := arguments.Parse(&args); err != nil {
		log.Fatalln("Argument parsing error:", err.Error())
	}

	if *args.DOAPIToken == "" {
		log.Fatalln("DigitalOcean API token is required")
	}

	if *args.SecurityToken == "" {
		hash := sha512.Sum512_256([]byte(*args.DOAPIToken))
		*args.SecurityToken = hex.EncodeToString(hash[:])

		log.Println("New auth token:", *args.SecurityToken)
	}

	client := digitalOceanClient{
		client: godo.NewClient(
			oauth2.NewClient(
				context.Background(),
				oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *args.DOAPIToken}),
			),
		),
		timeout: 5 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(*args.Endpoint, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		par, err := ParseParams(&query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if par.Token != *args.SecurityToken {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		domain, err := par.Domain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("New IP for [%s] %s: %s", par.Kind, par.Name, par.Addr)

		if id, err := client.getDNSRecordId(domain, par.Kind, par.Name); err == nil {
			if err := client.updateDNSRecord(domain, *id, par.Addr); err != nil {
				http.Error(w, err.Error(), http.StatusFailedDependency)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(200)
		fmt.Fprintln(w, "Done")
	})

	go func() {
		log.Println("Starting server on", *args.Address)

		if err := http.ListenAndServe(*args.Address, limit(mux)); err != nil {
			log.Fatalln("Server error:", err)
		}
	}()

	select {}
}
