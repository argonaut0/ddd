package ddd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"reflect"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const DDD_VERSION = "0.0.1"

type Daemon struct {
	AllowIPv4     bool
	AllowIPv6     bool
	AddressSource AddressSource
	// Interval for checking if the IP address has changed, seconds
	RefreshInterval time.Duration
	RecordTTL       time.Duration
	// ie. "example.com."
	Zone         string
	RecordSetter libdns.RecordSetter
	// name of record relative to root zone
	RecordName string
	// name of interface to retrieve address from
	InterfaceName string
	// URL of whoami API to get address from
	WebURL4 string
	// for ipv6
	WebURL6 string
	// cache of current IPs
	CurrentA    string
	CurrentAAAA string
	// ignore cached current IPs after n intervals
	IgnoreCachedAfter int
}

const DEFAULT_RECORD_TTL = 300       // 5 minutes
const DEFAULT_REFRESH_INTERVAL = 120 // 2 minutes

func (dm *Daemon) Serve() {
	slog.Info("Starting Dynamic DNS Daemon", "version", DDD_VERSION)
	switch dm.AddressSource {
	case SOURCE_WEB:
		slog.Info("Using web api", "ipv4", dm.WebURL4, "ipv6", dm.WebURL6)
	case SOURCE_INTERFACE:
		slog.Info("Using interface", "iface", dm.InterfaceName)
	}
	slog.Info("refresh interval", "seconds", dm.RefreshInterval.Seconds())
	slog.Info("provider", "provider", reflect.TypeOf(dm.RecordSetter))

	if !dm.AllowIPv4 {
		slog.Debug("IPv4 disabled")
	}
	if !dm.AllowIPv6 {
		slog.Debug("IPv6 disabled")
	}

	slog.Info("running initial update")
	dm.CheckIP(true)

	refreshTicker := time.NewTicker(dm.RefreshInterval)

	cycles := 0

	for {
		<-refreshTicker.C
		dm.CheckIP(cycles > dm.IgnoreCachedAfter)
		if cycles > dm.IgnoreCachedAfter {
			cycles = 0
		}
		cycles++
	}
}

func (dm *Daemon) CheckIP(force bool) {
	ctx := context.Background()

	slog.Info("checking current address")
	var candidateIPs []netip.Addr
	var err error
	switch dm.AddressSource {
	case SOURCE_INTERFACE:
		candidateIPs, err = dm.GetIPsByInterface()
	case SOURCE_WEB:
		candidateIPs, err = dm.GetIPsByWeb()
	default:
		slog.Error("no address source specified")
	}
	if err != nil {
		slog.Error("failed to get candidate IPs", "err", err)
		return
	}

	ips, err := dm.ChooseIPs(candidateIPs)
	if err != nil {
		slog.Error("failed to choose IP", "err", err)
		return
	}

	for _, ip := range ips {
		if force {
			slog.Info("ignoring whether address has changed, forcing update")
		} else {
			if ip.Is4() {
				if ip.String() == dm.CurrentA {
					slog.Info("no change in IPv4 address, skipping update", "ip", ip.String())
					continue
				}
			} else {
				if ip.String() == dm.CurrentAAAA {
					slog.Info("no change in IPv6 address, skipping update", "ip", ip.String())
					continue
				}
			}
		}
		err = dm.UpdateDNS(ctx, ip)
		if err != nil {
			slog.Error("failed to update record", "err", err)
			continue
		}
		if ip.Is4() {
			dm.CurrentA = ip.String()
		} else {
			dm.CurrentAAAA = ip.String()
		}
		slog.Info("record updated")
	}
}
func (dm *Daemon) GetIPsByInterface() ([]netip.Addr, error) {
	iface, err := net.InterfaceByName(dm.InterfaceName)
	if err != nil {
		slog.Error("failed to get interface", "iface", dm.InterfaceName, "err", err)
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		slog.Error("failed to get iface addrs", "iface", dm.InterfaceName, "err", err)
		return nil, err
	}

	candidateIPs := []netip.Addr{}
	for _, addr := range addrs {
		pfix, err := netip.ParsePrefix(addr.String())
		if err != nil {
			return nil, fmt.Errorf("could not parse prefix %v : %w", addr.String(), err)
		}
		ip := pfix.Addr()
		if !ip.IsGlobalUnicast() {
			slog.Debug("ignoring non-GUA address", "ip", ip, "iface", dm.InterfaceName)
			continue
		}

		if ip.IsPrivate() {
			slog.Debug("ignoring private address", "ip", ip, "iface", dm.InterfaceName)
			continue
		}

		slog.Info("found potential address", "ip", ip, "iface", dm.InterfaceName)
		candidateIPs = append(candidateIPs, ip)
	}

	return candidateIPs, nil
}

func (dm *Daemon) GetIPsByWeb() ([]netip.Addr, error) {
	candidateIPs := make([]netip.Addr, 0)
	// fixme: this breaks if only one of the APIs is available
	urls := []string{dm.WebURL4, dm.WebURL6}
	for _, url := range urls {
		resp, err := http.Get(url)

		if err != nil {
			slog.Error("failed to make request to whoami", "api", url, "err", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Error("non-ok response from whoami", "api", url, "status", resp.Status)
			continue
		}
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("failed to read response body", "api", url, "err", err)
			continue
		}
		ip, err := netip.ParseAddr(strings.TrimSpace(string(response)))

		if err != nil {
			slog.Error("failed to parse IP from api response", "api", url, "response", string(response), "err", err)
			continue
		}

		slog.Info("found potential address", "ip", ip.String(), "api", url)
		candidateIPs = append(candidateIPs, ip)
	}
	return candidateIPs, nil
}

func (dm *Daemon) ChooseIPs(ips []netip.Addr) ([]netip.Addr, error) {
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses available")
	}
	var v4Candidate netip.Addr
	var v6Candidate netip.Addr
	picked := []netip.Addr{}

	if dm.AllowIPv4 {
		for _, candidateIP := range ips {
			if candidateIP.Is4() {
				v4Candidate = candidateIP
				break
			}
		}
	}
	if dm.AllowIPv6 {
		for _, candidateIP := range ips {
			if candidateIP.Is6() {
				v6Candidate = candidateIP
				break
			}
		}
	}
	if v4Candidate.IsValid() {
		slog.Info("using IPv4 address", "ip", v4Candidate.String())
		picked = append(picked, v4Candidate)
	}
	if v6Candidate.IsValid() {
		slog.Info("using IPv6 address", "ip", v6Candidate.String())
		picked = append(picked, v6Candidate)
	}
	if len(picked) < 1 {
		return nil, fmt.Errorf("no enabled address available")
	}
	return picked, nil
}

func (dm *Daemon) UpdateDNS(ctx context.Context, ip netip.Addr) error {

	var recordType string
	if ip.Is4() {
		recordType = "A"
	} else {
		recordType = "AAAA"
	}

	slog.Info(
		"setting record",
		"zone", dm.Zone,
		"name", dm.RecordName,
		"ip", ip.String(),
		"type", recordType,
		"ttl", dm.RecordTTL.Seconds())

	_, err := dm.RecordSetter.SetRecords(ctx, dm.Zone, []libdns.Record{
		libdns.Address{
			Name: dm.RecordName,
			TTL:  dm.RecordTTL,
			IP:   ip,
		},
	})

	if err != nil {
		slog.Error(
			"failed setting record",
			"zone", dm.Zone,
			"name", dm.RecordName,
			"ip", ip.String(),
			"err", err,
			"provider", reflect.TypeOf(dm.RecordSetter))
		return err
	}
	return nil
}
