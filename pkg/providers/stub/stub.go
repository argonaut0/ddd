package stub

import (
	"context"

	"github.com/libdns/libdns"
)

type Provider struct{}

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return recs, nil
}
