package network

import (
	"time"

	gopeer_logger "github.com/number571/go-peer/pkg/logger"
	gopeer_message "github.com/number571/go-peer/pkg/network/message"
	hiddenlake "github.com/number571/hidden-lake"
)

var (
	_ ISettings = &sSettings{}
)

type SSettings sSettings
type sSettings struct {
	SSubSettings
	FMessageSettings  gopeer_message.ISettings
	FQueuePeriod      time.Duration
	FFetchTimeout     time.Duration
	FMessageSizeBytes uint64
}

type SSubSettings struct {
	FLogger      gopeer_logger.ILogger
	FParallel    uint64
	FTCPAddress  string
	FServiceName string
}

func NewSettingsByNetworkKey(pNetworkKey string, pSubSettings *SSubSettings) ISettings {
	network, ok := hiddenlake.GNetworks[pNetworkKey]
	if !ok {
		panic("network not found")
	}
	if pSubSettings == nil {
		pSubSettings = &SSubSettings{}
	}
	return (&sSettings{
		FMessageSettings: gopeer_message.NewSettings(&gopeer_message.SSettings{
			FWorkSizeBits: network.FWorkSizeBits,
			FNetworkKey:   pNetworkKey,
		}),
		FMessageSizeBytes: network.FMessageSizeBytes,
		FQueuePeriod:      time.Duration(network.FQueuePeriodMS) * time.Millisecond,
		FFetchTimeout:     time.Duration(network.FFetchTimeoutMS) * time.Millisecond,
		SSubSettings: SSubSettings{
			FParallel:    pSubSettings.FParallel,
			FLogger:      pSubSettings.FLogger,
			FTCPAddress:  pSubSettings.FTCPAddress,
			FServiceName: pSubSettings.FServiceName,
		},
	}).defaultValue().mustNotNull()
}

func NewSettings(pSett *SSettings) ISettings {
	return (&sSettings{
		FMessageSettings:  pSett.FMessageSettings,
		FMessageSizeBytes: pSett.FMessageSizeBytes,
		FQueuePeriod:      pSett.FQueuePeriod,
		FFetchTimeout:     pSett.FFetchTimeout,
		SSubSettings: SSubSettings{
			FParallel:    pSett.FParallel,
			FLogger:      pSett.FLogger,
			FTCPAddress:  pSett.FTCPAddress,
			FServiceName: pSett.FServiceName,
		},
	}).defaultValue().mustNotNull()
}

func (p *sSettings) defaultValue() *sSettings {
	if p.FMessageSettings == nil {
		p.FMessageSettings = gopeer_message.NewSettings(&gopeer_message.SSettings{})
	}
	if p.FLogger == nil {
		p.FLogger = gopeer_logger.NewLogger(
			gopeer_logger.NewSettings(&gopeer_logger.SSettings{}),
			func(_ gopeer_logger.ILogArg) string { return "" },
		)
	}
	if p.FServiceName == "" {
		p.FServiceName = "_"
	}
	return p
}

func (p *sSettings) mustNotNull() *sSettings {
	if p.FMessageSizeBytes == 0 {
		panic(`p.FMessageSizeBytes == 0`)
	}
	if p.FQueuePeriod == 0 {
		panic(`p.FQueuePeriod == 0`)
	}
	if p.FFetchTimeout == 0 {
		panic(`p.FFetchTimeout == 0`)
	}
	return p
}

func (p *sSettings) GetMessageSettings() gopeer_message.ISettings {
	return p.FMessageSettings
}

func (p *sSettings) GetMessageSizeBytes() uint64 {
	return p.FMessageSizeBytes
}

func (p *sSettings) GetParallel() uint64 {
	return p.FParallel
}

func (p *sSettings) GetServiceName() string {
	return p.FServiceName
}

func (p *sSettings) GetTCPAddress() string {
	return p.FTCPAddress
}

func (p *sSettings) GetLogger() gopeer_logger.ILogger {
	return p.FLogger
}

func (p *sSettings) GetQueuePeriod() time.Duration {
	return p.FQueuePeriod
}

func (p *sSettings) GetFetchTimeout() time.Duration {
	return p.FFetchTimeout
}
