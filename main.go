package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

const tokenEnvVar = "DIGITALOCEAN_API_TOKEN"
const endpoint = "/ddns"

var digitalOceanClient *godo.Client
var tokenHashHex string

func checkRecord(domain string, name string) *godo.DomainRecord {
	options := &godo.ListOptions{
		Page:    0,
		PerPage: 0,
	}
	records, response, err := digitalOceanClient.Domains.Records(context.TODO(), domain, options)

	if err != nil {
		log.Print(err)
		return &godo.DomainRecord{}
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		log.Printf("Check failed: %s", response.Status)
		return &godo.DomainRecord{}
	}

	for _, r := range records {
		if r.Name == name {
			log.Printf("Record %s.%s found", name, domain)
			return &r
		}
	}

	log.Printf("Record %s.%s not found", name, domain)
	return &godo.DomainRecord{}
}

func updateDNS(domain string, id int, ip string) bool {
	updateRequest := &godo.DomainRecordEditRequest{
		Data: ip,
	}

	_, response, err := digitalOceanClient.Domains.EditRecord(context.TODO(), domain, id, updateRequest)

	if err != nil {
		log.Print(err)
		return false
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		log.Printf("Update failed: %s", response.Status)
		return false
	}

	log.Print("Updated successfully")
	return true
}

func ddns(res http.ResponseWriter, req *http.Request) {
	domain := req.URL.Query().Get("domain")
	token := req.URL.Query().Get("token")

	// ip := strings.Split(req.RemoteAddr, ":")[0]
	ip := req.URL.Query().Get("ip")

	res.Header().Set("Content-Type", "text/plain")

	if domain == "" || token == "" || ip == "" {
		http.Error(res, "Invalid domain, token or IP.", http.StatusBadRequest)
		return
	}

	if token != tokenHashHex {
		http.Error(res, "Authentication failed.", http.StatusForbidden)
		return
	}

	log.Printf("Got new IP (%s) for %s", ip, domain)

	parts := strings.SplitN(domain, ".", 2)

	name := parts[0]
	domain = parts[1]

	record := checkRecord(domain, name)

	if record.ID < 0 {
		http.Error(res, "No record found.", http.StatusNotFound)
		return
	}

	if !updateDNS(domain, record.ID, ip) {
		http.Error(res, "Failed.", http.StatusNoContent)
		return
	}

	res.Write([]byte("Done."))
}

func main() {
	var tlsFlag = flag.Bool("tls", false, "use TLS server")
	var tlsCert = flag.String("cert", "cert.pem", "path to certificate")
	var tlsKey = flag.String("key", "key.pem", "path to private key")
	var token = flag.String("token", os.Getenv(tokenEnvVar), "DigitalOcean API token")

	flag.Parse()

	if *token == "" {
		log.Fatalf("Invalid token (%s)", tokenEnvVar)
	}

	tokenHash := sha512.Sum512_256([]byte(*token))
	tokenHashHex = hex.EncodeToString(tokenHash[:])

	log.Printf("Auth token: %s", tokenHashHex)

	tokenPrepared := &oauth2.Token{
		AccessToken: *token,
	}

	tokenSource := oauth2.StaticTokenSource(tokenPrepared)

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	digitalOceanClient = godo.NewClient(oauthClient)

	http.HandleFunc(endpoint, ddns)

	var err error

	if *tlsFlag {
		log.Print("Running secure server on port 443")
		err = http.ListenAndServeTLS(":443", *tlsCert, *tlsKey, nil)
	} else {
		log.Print("Running insecure server on port 80")
		err = http.ListenAndServe(":80", nil)
	}

	if err != nil {
		log.Fatal(err)
	}
}
