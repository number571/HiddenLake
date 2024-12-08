package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/number571/go-peer/pkg/encoding"
	hla_settings "github.com/number571/hidden-lake/internal/adapters/tcp/pkg/settings"
	"github.com/number571/hidden-lake/internal/service/pkg/config"
	"github.com/number571/hidden-lake/internal/utils/api"
)

var (
	_ IRequester = &sRequester{}
)

const (
	cHandleIndexTemplate          = "%s" + hla_settings.CHandleIndexPath
	cHandleConfigSettingsTemplate = "%s" + hla_settings.CHandleConfigSettingsPath
	cHandleConfigConnectsTemplate = "%s" + hla_settings.CHandleConfigConnectsPath
	cHandleNetworkOnlineTemplate  = "%s" + hla_settings.CHandleNetworkOnlinePath
)

type sRequester struct {
	fHost   string
	fClient *http.Client
}

func NewRequester(pHost string, pClient *http.Client) IRequester {
	return &sRequester{
		fHost:   pHost,
		fClient: pClient,
	}
}

func (p *sRequester) GetIndex(pCtx context.Context) (string, error) {
	res, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodGet,
		fmt.Sprintf(cHandleIndexTemplate, p.fHost),
		nil,
	)
	if err != nil {
		return "", errors.Join(ErrBadRequest, err)
	}

	result := string(res)
	if result != hla_settings.CServiceFullName {
		return "", ErrInvalidTitle
	}

	return result, nil
}

func (p *sRequester) GetSettings(pCtx context.Context) (config.IConfigSettings, error) {
	res, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodGet,
		fmt.Sprintf(cHandleConfigSettingsTemplate, p.fHost),
		nil,
	)
	if err != nil {
		return nil, errors.Join(ErrBadRequest, err)
	}

	cfgSettings := new(config.SConfigSettings)
	if err := encoding.DeserializeJSON(res, cfgSettings); err != nil {
		return nil, errors.Join(ErrDecodeResponse, err)
	}

	return cfgSettings, nil
}

func (p *sRequester) GetOnlines(pCtx context.Context) ([]string, error) {
	res, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodGet,
		fmt.Sprintf(cHandleNetworkOnlineTemplate, p.fHost),
		nil,
	)
	if err != nil {
		return nil, errors.Join(ErrBadRequest, err)
	}

	var onlines []string
	if err := encoding.DeserializeJSON(res, &onlines); err != nil {
		return nil, errors.Join(ErrDecodeResponse, err)
	}

	return onlines, nil
}

func (p *sRequester) DelOnline(pCtx context.Context, pConnect string) error {
	_, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodDelete,
		fmt.Sprintf(cHandleNetworkOnlineTemplate, p.fHost),
		pConnect,
	)
	if err != nil {
		return errors.Join(ErrBadRequest, err)
	}
	return nil
}

func (p *sRequester) GetConnections(pCtx context.Context) ([]string, error) {
	res, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodGet,
		fmt.Sprintf(cHandleConfigConnectsTemplate, p.fHost),
		nil,
	)
	if err != nil {
		return nil, errors.Join(ErrBadRequest, err)
	}

	var connects []string
	if err := encoding.DeserializeJSON(res, &connects); err != nil {
		return nil, errors.Join(ErrDecodeResponse, err)
	}

	return connects, nil
}

func (p *sRequester) AddConnection(pCtx context.Context, pConnect string) error {
	_, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodPost,
		fmt.Sprintf(cHandleConfigConnectsTemplate, p.fHost),
		pConnect,
	)
	if err != nil {
		return errors.Join(ErrBadRequest, err)
	}
	return nil
}

func (p *sRequester) DelConnection(pCtx context.Context, pConnect string) error {
	_, err := api.Request(
		pCtx,
		p.fClient,
		http.MethodDelete,
		fmt.Sprintf(cHandleConfigConnectsTemplate, p.fHost),
		pConnect,
	)
	if err != nil {
		return errors.Join(ErrBadRequest, err)
	}
	return nil
}
