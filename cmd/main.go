package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

const PublicIPEchoEndpoint = "https://api.ipify.org"

func retrievePublicIp() (string, error) {
	res, err := http.Get(PublicIPEchoEndpoint)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	ip, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

func main() {
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatalf("Error initializing Cloudflare clint: %s", err)
	}
	targetDomain := os.Getenv("DNS_A_RECORD_FQDN")
	targetIP, err := retrievePublicIp()
	zoneID := os.Getenv("CLOUDFLARE_SITE_ZONE_ID")
	if err != nil {
		log.Fatalf("Could not retrieve public IPv4: %s", err)
	}
	log.Printf("Public IP is: %s", targetIP)

	ctx, cancelCtx := context.WithTimeout(context.Background(), 20*time.Second)

	listParams := cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: targetDomain,
	}

	// Fetch user details on the account
	l, i, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), listParams)
	if err != nil {
		log.Fatalf("Error retrieving existing DNS Records: %s", err)
	}
	if i.Count > 1 {
		log.Fatalf("More than 1 record for %s exists.", targetDomain)
	}
	if i.Count == 0 {
		recordCreate := cloudflare.CreateDNSRecordParams{
			Type:    "A",
			Name:    targetDomain,
			Content: targetIP,
			Comment: "Updated by Dynamic DNS Daemon https://github.com/argonaut0/ddd",
			TTL:     60,
		}
		_, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordCreate)
		if err != nil {
			log.Fatalf("Error creating new DNS record: %s", err)
		}
		fmt.Println("Successfully created new DNS record")
	} else {
		recordUpdate := cloudflare.UpdateDNSRecordParams{
			Type:    "A",
			Name:    targetDomain,
			Content: targetIP,
			Comment: "Updated by Dynamic DNS Daemon https://github.com/argonaut0/ddd",
			TTL:     60,
			ID:      l[0].ID,
		}
		_, err := api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordUpdate)
		if err != nil {
			log.Fatalf("Error updating DNS record: %s", err)
		}
		log.Println("Successfully updated DNS record")
	}
	cancelCtx()
}
