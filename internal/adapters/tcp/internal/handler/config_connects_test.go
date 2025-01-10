package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/number571/go-peer/pkg/logger"
	net_message "github.com/number571/go-peer/pkg/message/layer1"
	"github.com/number571/go-peer/pkg/network"
	"github.com/number571/go-peer/pkg/network/conn"
	"github.com/number571/go-peer/pkg/storage/cache"
	"github.com/number571/hidden-lake/internal/adapters/tcp/pkg/app/config"
	std_logger "github.com/number571/hidden-lake/internal/utils/logger/std"
)

func TestHandleConfigConnectsAPI(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := logger.NewLogger(
		logger.NewSettings(&logger.SSettings{}),
		func(_ logger.ILogArg) string { return "" },
	)

	handler := HandleConfigConnectsAPI(ctx, &tsConfigWrapper{}, log, &tsNetworkNode{})
	if err := configConnectsRequestMethod(handler); err != nil {
		t.Error(err)
		return
	}
	if err := configConnectsRequestGET(handler); err != nil {
		t.Error(err)
		return
	}
	if err := configConnectsRequestURLParse(handler); err != nil {
		t.Error(err)
		return
	}
	if err := configConnectsRequestURLScheme(handler); err != nil {
		t.Error(err)
		return
	}
	if err := configConnectsRequestAddConnection(handler, http.StatusOK); err != nil {
		t.Error(err)
		return
	}
	if err := configConnectsRequestDelConnection(handler, http.StatusOK); err != nil {
		t.Error(err)
		return
	}

	handlerx := HandleConfigConnectsAPI(ctx, &tsConfigWrapper{fWithFail: true}, log, &tsNetworkNode{})
	if err := configConnectsRequestAddConnection(handlerx, http.StatusInternalServerError); err != nil {
		t.Error(err)
		return
	}
	if err := configConnectsRequestDelConnection(handlerx, http.StatusInternalServerError); err != nil {
		t.Error(err)
		return
	}
}

func configConnectsRequestMethod(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", nil)

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusMethodNotAllowed {
		return errors.New("bad status code") // nolint: err113
	}

	return nil
}

func configConnectsRequestGET(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad status code") // nolint: err113
	}

	return nil
}

func configConnectsRequestURLParse(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("94.103.91.81:9581"))

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTeapot {
		return errors.New("bad status code") // nolint: err113
	}

	return nil
}

func configConnectsRequestURLScheme(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://abc"))

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return errors.New("bad status code") // nolint: err113
	}

	return nil
}

func configConnectsRequestAddConnection(handler http.HandlerFunc, code int) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("tcp://abc"))

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != code {
		return errors.New("bad status code") // nolint: err113
	}

	return nil
}

func configConnectsRequestDelConnection(handler http.HandlerFunc, code int) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader("tcp://abc"))

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != code {
		return errors.New("bad status code") // nolint: err113
	}

	return nil
}

var (
	_ network.INode          = &tsNetworkNode{}
	_ config.IConfig         = &tsConfig{}
	_ config.IEditor         = &tsEditor{}
	_ config.IConfigSettings = &tsConfigSettings{}
	_ config.IAddress        = &tsAddress{}
)

type tsConfigWrapper struct {
	fWithFail bool
}

func (p *tsConfigWrapper) GetConfig() config.IConfig { return &tsConfig{} }
func (p *tsConfigWrapper) GetEditor() config.IEditor { return &tsEditor{p.fWithFail} }

type tsConfig struct{}

func (p *tsConfig) GetLogging() std_logger.ILogging     { return nil }
func (p *tsConfig) GetSettings() config.IConfigSettings { return &tsConfigSettings{} }

func (p *tsConfig) GetAddress() config.IAddress { return &tsAddress{} }
func (p *tsConfig) GetEndpoints() []string      { return []string{"bbb"} }
func (p *tsConfig) GetConnections() []string    { return []string{"aaa"} }

type tsAddress struct{}

func (p *tsAddress) GetExternal() string { return "external" }
func (p *tsAddress) GetInternal() string { return "internal" }
func (p *tsAddress) GetPPROF() string    { return "pprof" }

type tsEditor struct {
	fWithFail bool
}

func (p *tsEditor) UpdateConnections([]string) error {
	if p.fWithFail {
		return errors.New("some error") // nolint: err113
	}
	return nil
}

type tsConfigSettings struct{}

func (p *tsConfigSettings) GetWorkSizeBits() uint64     { return 10 }
func (p *tsConfigSettings) GetNetworkKey() string       { return "_" }
func (p *tsConfigSettings) GetMessageSizeBytes() uint64 { return 8192 }
func (p *tsConfigSettings) GetDatabaseEnabled() bool    { return false }

type tsNetworkNode struct {
	fWithFail bool
}

func (p *tsNetworkNode) Close() error                                       { return nil }
func (p *tsNetworkNode) Run(context.Context) error                          { return nil }
func (p *tsNetworkNode) HandleFunc(uint32, network.IHandlerF) network.INode { return nil }
func (p *tsNetworkNode) GetSettings() network.ISettings {
	return network.NewSettings(&network.SSettings{
		FConnSettings: conn.NewSettings(&conn.SSettings{
			FLimitMessageSizeBytes: 1,
			FWaitReadTimeout:       time.Second,
			FDialTimeout:           time.Second,
			FReadTimeout:           time.Second,
			FWriteTimeout:          time.Second,
			FMessageSettings: net_message.NewSettings(&net_message.SSettings{
				FWorkSizeBits: 1,
				FNetworkKey:   "_",
			}),
		}),
		FMaxConnects:  1,
		FReadTimeout:  time.Second,
		FWriteTimeout: time.Second,
	})
}
func (p *tsNetworkNode) GetCacheSetter() cache.ICacheSetter { return nil }
func (p *tsNetworkNode) GetConnections() map[string]conn.IConn {
	return map[string]conn.IConn{
		"127.0.0.1:9999": nil,
	}
}
func (p *tsNetworkNode) AddConnection(context.Context, string) error {
	if p.fWithFail {
		return errors.New("some error") // nolint: err113
	}
	return nil
}
func (p *tsNetworkNode) DelConnection(string) error {
	if p.fWithFail {
		return errors.New("some error") // nolint: err113
	}
	return nil
}
func (p *tsNetworkNode) BroadcastMessage(context.Context, net_message.IMessage) error { return nil }
