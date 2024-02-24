package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

const PublicIPEchoEndpoint = "https://api.ipify.org./"

func retrievePublicIp() (string, error) {

	req, err := http.NewRequest("GET", PublicIPEchoEndpoint, nil)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)
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

type Daemon struct {
	Api *cloudflare.API
}

func (d *Daemon) createOrUpdateIP(targetDomain, targetIP, zoneID string) error {
	api := d.Api
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	listParams := cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: targetDomain,
	}

	// Fetch user details on the account
	l, i, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), listParams)
	if err != nil {
		return fmt.Errorf("error retrieving existing DNS Records: %s", err)
	}
	if i.Count > 1 {
		return fmt.Errorf("more than 1 record for %s exists", targetDomain)
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
			return fmt.Errorf("error creating new DNS record: %s", err)
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
			return fmt.Errorf("error updating DNS record: %s", err)
		}
		log.Println("Successfully updated DNS record")
	}
	return nil
}

func (d *Daemon) retrieveRecord(targetDomain, zoneID string) (string, error) {
	api := d.Api
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	listParams := cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: targetDomain,
	}

	// Fetch user details on the account
	l, i, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), listParams)
	if err != nil {
		return "", fmt.Errorf("error retrieving existing DNS Records: %s", err)
	}
	if i.Count > 1 {
		return "", fmt.Errorf("more than 1 record for %s exists", targetDomain)
	}
	return l[0].Content, nil
}

func main() {

	targetDomain := os.Getenv("DNS_A_RECORD_FQDN")
	zoneID := os.Getenv("CLOUDFLARE_SITE_ZONE_ID")
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Printf("error initializing Cloudflare client: %s", err)
		os.Exit(1)
	}
	d := Daemon{Api: api}
	oldIP, err := d.retrieveRecord(targetDomain, zoneID)
	if err != nil {
		log.Printf("could not retrieve existing record: %s", err)
	}
	targetIP, err := retrievePublicIp()
	if err != nil {
		log.Printf("Could not retrieve public IPv4: %s", err)
	}
	log.Printf("Public IP is: %s", targetIP)
	err = d.createOrUpdateIP(targetDomain, targetIP, zoneID)
	if err != nil {
		log.Println(err)
	} else {
		oldIP = targetIP
	}

	poll := os.Getenv("POLLING")
	if poll != "true" {
		os.Exit(0)
	}
	pollVar, intervalSpecified := os.LookupEnv("POLL_INTERVAL_SECONDS")
	var pollInterval = 120
	if intervalSpecified {
		var err error
		pollInterval, err = strconv.Atoi(pollVar)
		if err != nil {
			log.Println("Error parsing poll interval, ignoring")
		}
	}

	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		newIP, err := retrievePublicIp()
		if err != nil {
			log.Printf("Could not retrieve public IPv4: %s", err)
			continue
		}
		if newIP == oldIP {
			log.Println("Retrieved new IP, no change detected")
			continue
		}
		log.Printf("IP changed from %s to %s, updating...", targetIP, newIP)
		err = d.createOrUpdateIP(targetDomain, targetIP, zoneID)
		if err != nil {
			log.Println(err)
			continue
		}
		oldIP = newIP
	}
}
