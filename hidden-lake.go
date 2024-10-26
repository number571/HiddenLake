package hiddenlake

import (
	_ "embed"

	"github.com/number571/go-peer/pkg/encoding"
)

const (
	CVersion = "v1.7.2~"
)

var (
	//go:embed networks.yml
	gNetworks []byte
	GNetworks map[string]SNetwork
)

func init() {
	networksYAML := &SNetworksYAML{}
	if err := encoding.DeserializeYAML(gNetworks, networksYAML); err != nil {
		panic(err)
	}
	GNetworks = networksYAML.FNetworks
}

type SNetworksYAML struct {
	FNetworks map[string]SNetwork `yaml:"networks"`
}

type SNetwork struct {
	FMessageSizeBytes uint64   `yaml:"message_size_bytes"`
	FFetchTimeoutMS   uint64   `yaml:"fetch_timeout_ms"`
	FQueuePeriodMS    uint64   `yaml:"queue_period_ms"`
	FWorkSizeBits     uint64   `yaml:"work_size_bits"`
	FConnections      []string `yaml:"connections"`
}
