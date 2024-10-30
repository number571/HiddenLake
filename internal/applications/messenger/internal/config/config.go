package config

import (
	"errors"
	"os"
	"sync"

	"github.com/number571/go-peer/pkg/encoding"
	"github.com/number571/hidden-lake/internal/utils/language"
	logger "github.com/number571/hidden-lake/internal/utils/logger/std"
)

var (
	_ IConfigSettings = &SConfigSettings{}
	_ IConfig         = &SConfig{}
	_ IAddress        = &SAddress{}
)

type SConfigSettings struct {
	fMutex    sync.RWMutex
	fLanguage language.ILanguage

	FMessagesCapacity uint64 `json:"messages_capacity" yaml:"messages_capacity"`
	FLanguage         string `json:"language,omitempty" yaml:"language,omitempty"`
}

type SConfig struct {
	fFilepath string
	fLogging  logger.ILogging

	FSettings   *SConfigSettings `yaml:"settings"`
	FLogging    []string         `yaml:"logging,omitempty"`
	FAddress    *SAddress        `yaml:"address"`
	FConnection string           `yaml:"connection"`
}

type SAddress struct {
	FInterface string `yaml:"interface"`
	FIncoming  string `yaml:"incoming"`
	FPPROF     string `yaml:"pprof,omitempty"`
}

func BuildConfig(pFilepath string, pCfg *SConfig) (IConfig, error) {
	if _, err := os.Stat(pFilepath); !os.IsNotExist(err) {
		return nil, errors.Join(ErrConfigAlreadyExist, err)
	}

	pCfg.fFilepath = pFilepath
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

	cfg.fFilepath = pFilepath
	if err := cfg.initConfig(); err != nil {
		return nil, errors.Join(ErrInitConfig, err)
	}

	return cfg, nil
}

func (p *SConfig) GetSettings() IConfigSettings {
	return p.FSettings
}

func (p *SConfigSettings) GetMessagesCapacity() uint64 {
	return p.FMessagesCapacity
}

func (p *SConfigSettings) GetLanguage() language.ILanguage {
	p.fMutex.RLock()
	defer p.fMutex.RUnlock()

	return p.fLanguage
}

func (p *SConfigSettings) loadLanguage() error {
	res, err := language.ToILanguage(p.FLanguage)
	if err != nil {
		return errors.Join(ErrToLanguage, err)
	}
	p.fLanguage = res
	return nil
}

func (p *SConfig) isValid() bool {
	return true &&
		p.FConnection != "" &&
		p.FAddress.FInterface != "" &&
		p.FAddress.FIncoming != "" &&
		p.FSettings.FMessagesCapacity != 0
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

	if err := p.FSettings.loadLanguage(); err != nil {
		return errors.Join(ErrLoadLanguage, err)
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

func (p *SConfig) GetAddress() IAddress {
	return p.FAddress
}

func (p *SConfig) GetConnection() string {
	return p.FConnection
}

func (p *SAddress) GetInterface() string {
	return p.FInterface
}

func (p *SAddress) GetIncoming() string {
	return p.FIncoming
}

func (p *SAddress) GetPPROF() string {
	return p.FPPROF
}

func (p *SConfig) GetLogging() logger.ILogging {
	return p.fLogging
}
