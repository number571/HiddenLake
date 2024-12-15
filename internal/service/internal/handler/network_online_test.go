package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/number571/go-peer/pkg/logger"
	hls_client "github.com/number571/hidden-lake/internal/service/pkg/client"
	std_logger "github.com/number571/hidden-lake/internal/utils/logger/std"
	"github.com/number571/hidden-lake/pkg/adapters/http/client"
	testutils "github.com/number571/hidden-lake/test/utils"
)

func TestHandleOnlineAPI2(t *testing.T) {
	t.Parallel()

	httpLogger := std_logger.NewStdLogger(
		func() std_logger.ILogging {
			logging, err := std_logger.LoadLogging([]string{})
			if err != nil {
				panic(err)
			}
			return logging
		}(),
		func(_ logger.ILogArg) string {
			return ""
		},
	)

	ctx := context.Background()

	epClients := []client.IClient{
		client.NewClient(&tsRequester{}),
	}

	handler := HandleNetworkOnlineAPI(ctx, httpLogger, epClients)
	if err := onlineAPIRequestOK(handler); err != nil {
		t.Error(err)
		return
	}
	if err := onlineAPIRequestDeleteOK(handler); err != nil {
		t.Error(err)
		return
	}

	if err := onlineAPIRequestMethod(handler); err == nil {
		t.Error("request success with invalid method")
		return
	}

	epClientsx := []client.IClient{
		client.NewClient(&tsRequester{fWithFail: true}),
	}

	handlerx := HandleNetworkOnlineAPI(ctx, httpLogger, epClientsx)
	if err := onlineAPIRequestDeleteOK(handlerx); err == nil {
		t.Error("request success with delete error")
		return
	}
}

func onlineAPIRequestOK(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad status code") // nolint: err113
	}

	if _, err := io.ReadAll(res.Body); err != nil {
		return err
	}

	return nil
}

func onlineAPIRequestDeleteOK(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader("127.0.0.1:9999"))

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad status code") // nolint: err113
	}

	if _, err := io.ReadAll(res.Body); err != nil {
		return err
	}

	return nil
}

func onlineAPIRequestMethod(handler http.HandlerFunc) error {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad status code") // nolint: err113
	}

	if _, err := io.ReadAll(res.Body); err != nil {
		return err
	}

	return nil
}

func TestHandleOnlineAPI(t *testing.T) {
	t.Parallel()

	pathCfg := fmt.Sprintf(tcPathConfigTemplate, 6)
	pathDB := fmt.Sprintf(tcPathDBTemplate, 6)

	_, node, _, cancel, srv := testAllCreate(pathCfg, pathDB, testutils.TgAddrs[12])
	defer testAllFree(node, cancel, srv, pathCfg, pathDB)

	client := hls_client.NewClient(
		hls_client.NewBuilder(),
		hls_client.NewRequester(
			testutils.TgAddrs[12],
			&http.Client{Timeout: time.Minute},
		),
	)

	testGetOnlines(t, client)
	testDelOnline(t, client)
}

func testGetOnlines(t *testing.T, client hls_client.IClient) {
	onlines, err := client.GetOnlines(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	if len(onlines) != 1 || onlines[0] != tcConnection {
		t.Error("len(onlines) != 1 || onlines[0] != tcConnection")
		return
	}
}

func testDelOnline(t *testing.T, client hls_client.IClient) {
	if err := client.DelOnline(context.Background(), "tcp://bbb"); err != nil {
		t.Error(err)
		return
	}
}
