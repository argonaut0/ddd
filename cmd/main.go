package cmd

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/argonaut0/ddd/pkg/ddd"
	"github.com/argonaut0/ddd/pkg/providers"
)

const LOG_LEVEL = "DEBUG"

var programLogLevel = new(slog.LevelVar)

func main() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLogLevel})
	slog.SetDefault(slog.New(h))
	programLogLevel.Set(slog.LevelDebug)

	usev4 := os.Getenv("DDD_USE_IPV4")
	usev6 := os.Getenv("DDD_USE_IPV6")
	if usev4 == "" && usev6 == "" {
		slog.Error("fatal: DDD_USE_IPV4 or DDD_USE_IPV6 must be set")
		os.Exit(1)
	}

	var addressSource ddd.AddressSource
	addressSourceEnvVar := os.Getenv("DDD_ADDRESS_SOURCE")
	if addressSourceEnvVar == "web" {
		addressSource = ddd.SOURCE_WEB
	} else if addressSourceEnvVar == "interface" {
		addressSource = ddd.SOURCE_INTERFACE
	} else {
		slog.Error("fatal: DDD_ADDRESS_SOURCE must be one of [web, interface]")
		os.Exit(1)
	}

	var recordTTL time.Duration
	recordTTLEnvVar := os.Getenv("DDD_RECORD_TTL")
	if recordTTLEnvVar == "" {
		recordTTL = time.Second * ddd.DEFAULT_RECORD_TTL
	} else {
		ttl, err := strconv.Atoi(recordTTLEnvVar)
		if err != nil {
			slog.Error("could not parse DDD_RECORD_TTL", "error", err, "var", recordTTLEnvVar)
			recordTTL = time.Second * ddd.DEFAULT_RECORD_TTL
		} else {
			recordTTL = time.Second * time.Duration(ttl)
		}
	}

	var refreshInterval time.Duration
	refreshIntervalEnvVar := os.Getenv("DDD_REFRESH_INTERVAL")
	if refreshIntervalEnvVar == "" {
		refreshInterval = ddd.DEFAULT_REFRESH_INTERVAL * time.Second
	} else {
		ttl, err := strconv.Atoi(refreshIntervalEnvVar)
		if err != nil {
			slog.Error("could not parse DDD_REFRESH_INTERVAL", "error", err, "var", refreshIntervalEnvVar)
			refreshInterval = ddd.DEFAULT_REFRESH_INTERVAL * time.Second
		} else {
			refreshInterval = time.Second * time.Duration(ttl)
		}
	}

	zone := os.Getenv("DDD_ZONE")
	if zone == "" {
		slog.Error("fatal: DDD_ZONE must be set")
		os.Exit(1)
	}
	zone = strings.TrimSuffix(zone, ".")
	zone = zone + "."

	recordName := os.Getenv("DDD_RECORD_NAME")
	if recordName == "" {
		slog.Error("fatal: DDD_RECORD_NAME must be set")
		os.Exit(1)
	}

	interfaceName := os.Getenv("DDD_INTERFACE_NAME")

	if addressSource == ddd.SOURCE_INTERFACE && interfaceName == "" {
		slog.Error("fatal: DDD_INTERFACE_NAME must be set when using interface address source")
		os.Exit(1)
	}

	dnsProvider, err := providers.GetProvider()
	if err != nil {
		slog.Error("fatal: no dns provider configured", "error", err)
		os.Exit(1)
	}

	url4 := os.Getenv("DDD_WEB_URLV4")
	url6 := os.Getenv("DDD_WEB_URLV6")
	if addressSource == ddd.SOURCE_WEB {
		if usev4 == "true" && url4 == "" {
			slog.Error("fatal: DDD_WEB_URLV4 must be set when using web address source")
			os.Exit(1)
		}
		if usev6 == "true" && url6 == "" {
			slog.Error("fatal: DDD_WEB_URLV6 must be set when using web address source")
			os.Exit(1)
		}
	}

	ignoreCachedAfter := os.Getenv("DDD_IGNORE_CACHED_AFTER")
	ignoreCachedAfterInt, err := strconv.Atoi(ignoreCachedAfter)
	if err != nil {
		ignoreCachedAfterInt = 5
	}

	ddd := &ddd.Daemon{
		AllowIPv4:         usev4 == "true",
		AllowIPv6:         usev6 == "true",
		AddressSource:     addressSource,
		RefreshInterval:   refreshInterval,
		RecordTTL:         recordTTL,
		Zone:              zone,
		RecordName:        recordName,
		InterfaceName:     interfaceName,
		RecordSetter:      dnsProvider,
		WebURL4:           url4,
		WebURL6:           url6,
		IgnoreCachedAfter: ignoreCachedAfterInt,
	}
	ddd.Serve()
}
