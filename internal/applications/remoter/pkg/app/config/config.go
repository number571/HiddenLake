package config

import (
	"errors"
	"os"
	"time"

	"github.com/number571/go-peer/pkg/encoding"
	logger "github.com/number571/hidden-lake/internal/utils/logger/std"
)

var (
	_ IConfig  = &SConfig{}
	_ IAddress = &SAddress{}
)

type SConfigSettings struct {
	FExecTimeoutMS uint64 `json:"exec_timeout_ms" yaml:"exec_timeout_ms"`
	FPassword      string `json:"password,omitempty" yaml:"password,omitempty"`
}

type SConfig struct {
	fLogging logger.ILogging

	FSettings *SConfigSettings `yaml:"settings"`
	FLogging  []string         `yaml:"logging,omitempty"`
	FAddress  *SAddress        `yaml:"address"`
}

type SAddress struct {
	FExternal string `yaml:"external,omitempty"`
}

func BuildConfig(pFilepath string, pCfg *SConfig) (IConfig, error) {
	if _, err := os.Stat(pFilepath); !os.IsNotExist(err) {
		return nil, errors.Join(ErrConfigAlreadyExist, err)
	}

	if err := pCfg.initConfig(); err != nil {
		return nil, errors.Join(ErrInitConfig, err)
	}

	if err := os.WriteFile(pFilepath, encoding.SerializeYAML(pCfg), 0o600); err != nil {
		return nil, errors.Join(ErrWriteConfig, err)
	}

	return pCfg, nil
}

func LoadConfig(pFilepath string) (IConfig, error) {
	if _, err := os.Stat(pFilepath); os.IsNotExist(err) {
		return nil, errors.Join(ErrConfigNotExist, err)
	}

	bytes, err := os.ReadFile(pFilepath)
	if err != nil {
		return nil, errors.Join(ErrReadConfig, err)
	}

	cfg := new(SConfig)
	if err := encoding.DeserializeYAML(bytes, cfg); err != nil {
		return nil, errors.Join(ErrDeserializeConfig, err)
	}

	if err := cfg.initConfig(); err != nil {
		return nil, errors.Join(ErrInitConfig, err)
	}

	return cfg, nil
}

func (p *SConfig) isValid() bool {
	return true &&
		p.FAddress.FExternal != "" &&
		p.FSettings.FExecTimeoutMS != 0
}

func (p *SConfig) initConfig() error {
	if p.FSettings == nil {
		p.FSettings = new(SConfigSettings)
	}

	if p.FAddress == nil {
		p.FAddress = new(SAddress)
	}

	if !p.isValid() {
		return ErrInvalidConfig
	}

	if err := p.loadLogging(); err != nil {
		return errors.Join(ErrLoadLogging, err)
	}

	return nil
}

func (p *SConfig) loadLogging() error {
	result, err := logger.LoadLogging(p.FLogging)
	if err != nil {
		return errors.Join(ErrInvalidLogging, err)
	}
	p.fLogging = result
	return nil
}

func (p *SConfig) GetSettings() IConfigSettings {
	return p.FSettings
}

func (p *SConfigSettings) GetExecTimeout() time.Duration {
	return time.Duration(p.FExecTimeoutMS) * time.Millisecond
}

func (p *SConfigSettings) GetPassword() string {
	return p.FPassword
}

func (p *SConfig) GetAddress() IAddress {
	return p.FAddress
}

func (p *SAddress) GetExternal() string {
	return p.FExternal
}

func (p *SConfig) GetLogging() logger.ILogging {
	return p.fLogging
}
