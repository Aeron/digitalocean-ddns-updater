package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"

	arguments "github.com/aeron/digitalocean-ddns-updater/app/args"
	clients "github.com/aeron/digitalocean-ddns-updater/app/clients"
)

const shutdownTimeout = 5 * time.Second
const requestTimeout = 5 * time.Second

var args struct {
	Address       *string  `default:":8080" help:"address to listen on"`
	Endpoint      *string  `default:"/ddns" help:"endpoint path to handle updates"`
	DOAPIToken    *string  `name:"digitalocean-api-token" environ:"DIGITALOCEAN_API_TOKEN" help:"DigitalOcean API token"`
	SecurityToken *string  `name:"security-token" environ:"SECURITY_TOKEN" help:"application security token"`
	LimitRPS      *float64 `name:"limit-rps" default:".01" environ:"LIMIT_RPS" help:"limit requests per second"`
	LimitBurst    *int     `name:"limit-burst" default:"1" environ:"LIMIT_BURST" help:"limit a single burst size"`
}

func main() {
	// Parsing and validating arguments
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

	// Setting up the DigitalOcean client
	client := clients.DigitalOceanDomains{
		Op: godo.NewClient(
			oauth2.NewClient(
				context.Background(),
				oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *args.DOAPIToken}),
			),
		).Domains,
		Timeout: requestTimeout,
	}

	// Setting up the controller
	controller := Controller{&client}

	// Setting up the router
	mux := http.NewServeMux()
	mux.HandleFunc(*args.Endpoint, controller.Handle)

	// Setting up the server
	server := http.Server{
		Addr:    *args.Address,
		Handler: limit(mux),
	}

	// Setting up the signal channel
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	// Binding the shutdown signals
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// Running the server
	go func() {
		log.Println("Starting server on", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln("Fatal server error:", err)
		}
	}()

	// Listening for the channel/signal
	sig := <-sigChan

	// Shutting down
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("Shutting down server on", sig.String())

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown gracefully")
}
