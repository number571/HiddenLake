package config

import (
	"errors"
	"os"

	"github.com/number571/go-peer/pkg/encoding"
	"github.com/number571/hidden-lake/build"
	hla_tcp_settings "github.com/number571/hidden-lake/internal/adapters/tcp/pkg/settings"
	hlm_settings "github.com/number571/hidden-lake/internal/applications/messenger/pkg/settings"
	hlp_settings "github.com/number571/hidden-lake/internal/applications/pinger/pkg/settings"
	hls_settings "github.com/number571/hidden-lake/internal/service/pkg/settings"
	logger "github.com/number571/hidden-lake/internal/utils/logger/std"
)

func InitConfig(cfgPath string, initCfg *SConfig, useNetwork string) (IConfig, error) {
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		cfg, err := LoadConfig(cfgPath)
		if err != nil {
			return nil, errors.Join(ErrLoadConfig, err)
		}
		return rebuildConfig(cfg, useNetwork)
	}
	if initCfg == nil {
		initCfg = initConfig()
	}
	cfg, err := BuildConfig(cfgPath, initCfg)
	if err != nil {
		return nil, errors.Join(ErrBuildConfig, err)
	}
	return rebuildConfig(cfg, useNetwork)
}

func rebuildConfig(pCfg IConfig, pUseNetwork string) (IConfig, error) {
	if pUseNetwork == "" {
		return pCfg, nil
	}

	cfg := pCfg.(*SConfig)
	network, ok := build.GNetworks[pUseNetwork]
	if !ok {
		return nil, errors.Join(ErrRebuildConfig, ErrNetworkNotFound)
	}

	cfg.FSettings.FFetchTimeoutMS = network.FFetchTimeoutMS
	cfg.FSettings.FMessageSizeBytes = network.FMessageSizeBytes
	cfg.FSettings.FQueuePeriodMS = network.FQueuePeriodMS
	cfg.FSettings.FWorkSizeBits = network.FWorkSizeBits
	cfg.FSettings.FNetworkKey = pUseNetwork

	if err := os.WriteFile(cfg.fFilepath, encoding.SerializeYAML(cfg), 0o600); err != nil {
		return nil, errors.Join(ErrRebuildConfig, ErrWriteConfig, err)
	}

	rCfg, err := LoadConfig(cfg.fFilepath)
	if err != nil {
		return nil, errors.Join(ErrRebuildConfig, ErrLoadConfig, err)
	}

	return rCfg, nil
}

func initConfig() *SConfig {
	defaultNetwork := build.GNetworks[build.CDefaultNetwork]
	return &SConfig{
		FSettings: &SConfigSettings{
			FMessageSizeBytes: defaultNetwork.FMessageSizeBytes,
			FWorkSizeBits:     defaultNetwork.FWorkSizeBits,
			FFetchTimeoutMS:   defaultNetwork.FFetchTimeoutMS,
			FQueuePeriodMS:    defaultNetwork.FQueuePeriodMS,
		},
		FLogging: []string{logger.CLogInfo, logger.CLogWarn, logger.CLogErro},
		FAddress: &SAddress{
			FExternal: hls_settings.CDefaultExternalAddress,
			FInternal: hls_settings.CDefaultInternalAddress,
		},
		FServices: map[string]string{
			hlp_settings.CServiceFullName: hlp_settings.CDefaultExternalAddress,
			hlm_settings.CServiceFullName: hlm_settings.CDefaultExternalAddress,
		},
		FEndpoints: []string{
			hla_tcp_settings.CDefaultInternalAddress,
		},
		FFriends: map[string]string{},
	}
}
