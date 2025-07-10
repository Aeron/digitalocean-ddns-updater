package main

import (
	"fmt"
	"log"
	"net/http"

	clients "github.com/aeron/digitalocean-ddns-updater/app/clients"
)

// Represents the default route controller.
type Controller struct {
	Client *clients.DigitalOceanDomains
}

// Handles all the incoming requests.
func (c *Controller) Handle(w http.ResponseWriter, r *http.Request) {
	// Parsing query parameters
	query := r.URL.Query()
	params, err := ParseParams(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validating the security token
	if params.Token != *args.SecurityToken {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Validating the domain name
	domain, err := params.Domain()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Updating [%s] %s to %s", params.Kind, params.Name, params.Addr)

	// Retrieving the existing record identifier
	recId, err := c.Client.GetDNSRecordId(domain, params.Kind, params.Name)
	if err != nil {
		log.Println("Cannot get a DNS record identifier:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Updating the record by its identifier
	if err := c.Client.UpdateDNSRecord(domain, recId, params.Addr); err != nil {
		log.Println("Cannot update a DNS record:", err)
		http.Error(w, err.Error(), http.StatusFailedDependency)
		return
	}

	// Writing a response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(200)
	fmt.Fprintln(w, "Done")
}
