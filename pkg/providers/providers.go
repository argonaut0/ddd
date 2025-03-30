package providers

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/libdns/alidns"
	"github.com/libdns/autodns"
	"github.com/libdns/azure"
	"github.com/libdns/bunny"
	"github.com/libdns/cloudflare"
	"github.com/libdns/cloudns"
	"github.com/libdns/ddnss"
	"github.com/libdns/desec"
	"github.com/libdns/digitalocean"
	"github.com/libdns/dinahosting"
	"github.com/libdns/directadmin"
	"github.com/libdns/dnsexit"
	"github.com/libdns/dnsimple"
	"github.com/libdns/dnspod"
	"github.com/libdns/dnsupdate"
	"github.com/libdns/domainnameshop"
	"github.com/libdns/libdns"
)

const ENV_PREFIX = "DDD_"

func GetProvider() (libdns.RecordSetter, error) {
	// alidns
	{
		accessKeyId := os.Getenv(ENV_PREFIX + "ALIDNS_ACCESS_KEY_ID")
		accessKeySecret := os.Getenv(ENV_PREFIX + "ALIDNS_ACCESS_KEY_SECRET")
		regionId := os.Getenv(ENV_PREFIX + "ALIDNS_REGION_ID")
		if accessKeyId != "" && accessKeySecret != "" {
			slog.Debug("found env vars for alidns")
			return &alidns.Provider{
				AccKeyID:     accessKeyId,
				AccKeySecret: accessKeySecret,
				RegionID:     regionId,
			}, nil
		}
	}

	// autodns
	{
		username := os.Getenv(ENV_PREFIX + "AUTODNS_USERNAME")
		password := os.Getenv(ENV_PREFIX + "AUTODNS_PASSWORD")
		endpoint := os.Getenv(ENV_PREFIX + "AUTODNS_ENDPOINT")
		context := os.Getenv(ENV_PREFIX + "AUTODNS_CONTEXT")
		if username != "" && password != "" {
			slog.Debug("found env vars for autodns")
			return &autodns.Provider{
				Username: username,
				Password: password,
				Endpoint: endpoint,
				Context:  context,
			}, nil
		}
	}

	// azure
	{
		subscriptionId := os.Getenv(ENV_PREFIX + "AZURE_SUBSCRIPTION_ID")
		resourceGroup := os.Getenv(ENV_PREFIX + "AZURE_RESOURCE_GROUP_NAME")
		tenantId := os.Getenv(ENV_PREFIX + "AZURE_TENANT_ID")
		clientId := os.Getenv(ENV_PREFIX + "AZURE_CLIENT_ID")
		clientSecret := os.Getenv(ENV_PREFIX + "AZURE_CLIENT_SECRET")
		if subscriptionId != "" && resourceGroup != "" {
			slog.Debug("found env vars for azure")
			return &azure.Provider{
				SubscriptionId:    subscriptionId,
				ResourceGroupName: resourceGroup,
				TenantId:          tenantId,
				ClientId:          clientId,
				ClientSecret:      clientSecret,
			}, nil
		}
	}

	// bunny
	{
		apiKey := os.Getenv(ENV_PREFIX + "BUNNY_API_KEY")
		if apiKey != "" {
			slog.Debug("found env vars for bunny")
			return &bunny.Provider{
				AccessKey: apiKey,
			}, nil
		}
	}

	// cloudflare
	{
		apiToken := os.Getenv(ENV_PREFIX + "CLOUDFLARE_API_TOKEN")
		zoneToken := os.Getenv(ENV_PREFIX + "CLOUDFLARE_ZONE_TOKEN")
		if apiToken != "" {
			slog.Debug("found env vars for cloudflare")
			return &cloudflare.Provider{
				APIToken:  apiToken,
				ZoneToken: zoneToken,
			}, nil
		}
	}

	// cloudns
	{
		authId := os.Getenv(ENV_PREFIX + "CLOUDNS_AUTH_ID")
		authPassword := os.Getenv(ENV_PREFIX + "CLOUDNS_AUTH_PASSWORD")
		subAuthId := os.Getenv(ENV_PREFIX + "CLOUDNS_SUB_AUTH_ID")
		if authId != "" && authPassword != "" {
			slog.Debug("found env vars for cloudns")
			return &cloudns.Provider{
				AuthId:       authId,
				AuthPassword: authPassword,
				SubAuthId:    subAuthId,
			}, nil
		}
	}

	// ddnss
	{
		apiToken := os.Getenv(ENV_PREFIX + "DDNSS_API_TOKEN")
		username := os.Getenv(ENV_PREFIX + "DDNSS_USERNAME")
		password := os.Getenv(ENV_PREFIX + "DDNSS_PASSWORD")
		if apiToken != "" || (username != "" && password != "") {
			slog.Debug("found env vars for ddnss")
			return &ddnss.Provider{
				APIToken: apiToken,
				Username: username,
				Password: password,
			}, nil
		}
	}

	// desec
	{
		token := os.Getenv(ENV_PREFIX + "DESEC_TOKEN")
		if token != "" {
			slog.Debug("found env vars for desec")
			return &desec.Provider{
				Token: token,
			}, nil
		}
	}

	// digitalocean
	{
		apiToken := os.Getenv(ENV_PREFIX + "DIGITALOCEAN_API_TOKEN")
		if apiToken != "" {
			slog.Debug("found env vars for digitalocean")
			return &digitalocean.Provider{
				APIToken: apiToken,
			}, nil
		}

	}

	// dinahosting
	{
		username := os.Getenv(ENV_PREFIX + "DINAHOSTING_USERNAME")
		password := os.Getenv(ENV_PREFIX + "DINAHOSTING_PASSWORD")
		if username != "" && password != "" {
			slog.Debug("found env vars for dinahosting")
			return &dinahosting.Provider{
				Username: username,
				Password: password,
			}, nil
		}
	}

	// directadmin
	{
		serverUrl := os.Getenv(ENV_PREFIX + "DIRECTADMIN_SERVER_URL")
		user := os.Getenv(ENV_PREFIX + "DIRECTADMIN_USER")
		loginKey := os.Getenv(ENV_PREFIX + "DIRECTADMIN_LOGIN_KEY")
		insecureRequests := os.Getenv(ENV_PREFIX + "DIRECTADMIN_INSECURE_REQUESTS")
		if loginKey != "" {
			slog.Debug("found env vars for directadmin")
			return &directadmin.Provider{
				ServerURL:        serverUrl,
				User:             user,
				LoginKey:         loginKey,
				InsecureRequests: insecureRequests == "true",
			}, nil
		}
	}

	// dnsexit
	{
		apiKey := os.Getenv(ENV_PREFIX + "DNSEXIT_API_KEY")
		if apiKey != "" {
			slog.Debug("found env vars for dnsexit")
			return &dnsexit.Provider{
				APIKey: apiKey,
			}, nil
		}
	}

	// dnsimple
	{
		apiAccessToken := os.Getenv(ENV_PREFIX + "DNSSIMPLE_API_ACCESS_TOKEN")
		accountId := os.Getenv(ENV_PREFIX + "DNSSIMPLE_ACCOUNT_ID")
		apiUrl := os.Getenv(ENV_PREFIX + "DNSSIMPLE_API_URL")
		if apiAccessToken != "" {
			slog.Debug("found env vars for dnssimple")
			return &dnsimple.Provider{
				APIAccessToken: apiAccessToken,
				AccountID:      accountId,
				APIURL:         apiUrl,
			}, nil
		}
	}

	// dnspod
	{
		apiToken := os.Getenv(ENV_PREFIX + "DNSPOD_API_TOKEN")
		if apiToken != "" {
			slog.Debug("found env vars for dnspod")
			return &dnspod.Provider{
				APIToken: apiToken,
			}, nil
		}
	}

	// dnsupdate
	{
		addr := os.Getenv(ENV_PREFIX + "DNSUPDATE_ADDR")
		if addr != "" {
			slog.Debug("found env vars for dnsupdate")
			return &dnsupdate.Provider{
				Addr: addr,
			}, nil
		}
	}

	// domainnameshop
	{
		apiToken := os.Getenv(ENV_PREFIX + "DOMAINNAMESHOP_API_TOKEN")
		apiSecret := os.Getenv(ENV_PREFIX + "DOMAINNAMESHOP_API_SECRET")
		if apiToken != "" && apiSecret != "" {
			slog.Debug("found env vars for domainnameshop")
			return &domainnameshop.Provider{
				APIToken:  apiToken,
				APISecret: apiSecret,
			}, nil
		}
	}

	return nil, fmt.Errorf("no valid set of env vars found")
}
