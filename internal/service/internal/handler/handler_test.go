package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/number571/go-peer/pkg/client"
	"github.com/number571/go-peer/pkg/crypto/asymmetric"
	"github.com/number571/go-peer/pkg/logger"
	"github.com/number571/go-peer/pkg/network"
	"github.com/number571/go-peer/pkg/network/anonymity"
	"github.com/number571/go-peer/pkg/network/anonymity/queue"
	"github.com/number571/go-peer/pkg/network/conn"
	net_message "github.com/number571/go-peer/pkg/network/message"
	"github.com/number571/go-peer/pkg/storage/cache"
	"github.com/number571/go-peer/pkg/storage/database"
	"github.com/number571/go-peer/pkg/types"
	"github.com/number571/hidden-lake/internal/service/internal/config"
	pkg_settings "github.com/number571/hidden-lake/internal/service/pkg/settings"
	"github.com/number571/hidden-lake/internal/utils/closer"
)

const (
	tcMessageSize   = (8 << 10)
	tcWorkSize      = 10
	tcQueuePeriod   = 5_000
	tcFetchTimeout  = 30_000
	tcQueueCapacity = 32
	tcMaxConnects   = 16
	tcCapacity      = 1024
)

var (
	tgPrivKey1 = asymmetric.NewPrivKey()
	tgPrivKey2 = asymmetric.NewPrivKey()
	tgPrivKey3 = asymmetric.NewPrivKey()
)

const (
	tcServiceAddressInHLS = "hidden-echo-service"
	tcPathDBTemplate      = "database_test_%d.db"
	tcPathConfigTemplate  = "config_test_%d.yml"
)

var (
	tcConfig = fmt.Sprintf(`settings:
  message_size_bytes: 8192
  work_size_bits: 22
  fetch_timeout_ms: 60000
  queue_period_ms: 1000
  rand_message_size_bytes: 4096
  network_key: test
address:
  tcp: test_address_tcp
  http: test_address_http
connections:
  - test_connect1
  - test_connect2
  - test_connect3
friends:
  test_recvr: %s
  test_name1: %s
services:
  test_service1: 
    host: test_address1
  test_service2: 
    host: test_address2
  test_service3: 
    host: test_address3
`,
		tgPrivKey1.GetPubKey().ToString(),
		tgPrivKey2.GetPubKey().ToString(),
	)
)

func TestError(t *testing.T) {
	t.Parallel()

	str := "value"
	err := &SHandlerError{str}
	if err.Error() != errPrefix+str {
		t.Error("incorrect err.Error()")
		return
	}
}

func testStartServerHTTP(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", testEchoPage)

	srv := &http.Server{
		Addr:        addr,
		Handler:     http.TimeoutHandler(mux, time.Minute/2, "timeout"),
		ReadTimeout: time.Second,
	}

	go func() { _ = srv.ListenAndServe() }()

	return srv
}

func testEchoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		FMessage string `json:"message"`
	}

	var resp struct {
		FEcho  string `json:"echo"`
		FError int    `json:"error"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp.FError = 1
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp.FEcho = req.FMessage
	_ = json.NewEncoder(w).Encode(resp)
}

func testAllCreate(cfgPath, dbPath, srvAddr string) (config.IWrapper, anonymity.INode, context.Context, context.CancelFunc, *http.Server) {
	wcfg := testNewWrapper(cfgPath)
	node, ctx, cancel := testRunNewNode(dbPath, "")
	srvc := testRunService(ctx, wcfg, node, srvAddr)
	time.Sleep(200 * time.Millisecond)
	return wcfg, node, ctx, cancel, srvc
}

func testAllFree(node anonymity.INode, cancel context.CancelFunc, srv *http.Server, pathCfg, pathDB string) {
	defer func() {
		os.RemoveAll(pathDB)
		os.RemoveAll(pathCfg)
	}()
	cancel()
	_ = closer.CloseAll([]types.ICloser{
		srv,
		node.GetKVDatabase(),
		node.GetNetworkNode(),
	})
}

func testRunService(ctx context.Context, wcfg config.IWrapper, node anonymity.INode, addr string) *http.Server {
	mux := http.NewServeMux()

	logger := logger.NewLogger(
		logger.NewSettings(&logger.SSettings{}),
		func(_ logger.ILogArg) string { return "" },
	)

	cfg := wcfg.GetConfig()

	mux.HandleFunc(pkg_settings.CHandleIndexPath, HandleIndexAPI(logger))
	mux.HandleFunc(pkg_settings.CHandleConfigSettingsPath, HandleConfigSettingsAPI(wcfg, logger, node))
	mux.HandleFunc(pkg_settings.CHandleConfigConnectsPath, HandleConfigConnectsAPI(ctx, wcfg, logger, node))
	mux.HandleFunc(pkg_settings.CHandleConfigFriendsPath, HandleConfigFriendsAPI(wcfg, logger, node))
	mux.HandleFunc(pkg_settings.CHandleNetworkOnlinePath, HandleNetworkOnlineAPI(logger, node))
	mux.HandleFunc(pkg_settings.CHandleNetworkRequestPath, HandleNetworkRequestAPI(ctx, cfg, logger, node))
	mux.HandleFunc(pkg_settings.CHandleNetworkPubKeyPath, HandleNetworkPubKeyAPI(logger, node))

	srv := &http.Server{
		Addr:        addr,
		Handler:     http.TimeoutHandler(mux, time.Minute/2, "timeout"),
		ReadTimeout: time.Second,
	}

	go func() { _ = srv.ListenAndServe() }()
	return srv
}

func testNewWrapper(cfgPath string) config.IWrapper {
	_ = os.WriteFile(cfgPath, []byte(tcConfig), 0o600)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		panic(err)
	}
	return config.NewWrapper(cfg)
}

func testRunNewNode(dbPath, addr string) (anonymity.INode, context.Context, context.CancelFunc) {
	os.RemoveAll(dbPath)
	node := testNewNode(dbPath, addr).HandleFunc(pkg_settings.CServiceMask, nil)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = node.Run(ctx) }()
	return node, ctx, cancel
}

func testNewNode(dbPath, addr string) anonymity.INode {
	db, err := database.NewKVDatabase(dbPath)
	if err != nil {
		panic(err)
	}
	networkMask := uint32(1)
	node := anonymity.NewNode(
		anonymity.NewSettings(&anonymity.SSettings{
			FServiceName:  "TEST",
			FNetworkMask:  networkMask,
			FFetchTimeout: time.Minute,
		}),
		logger.NewLogger(
			logger.NewSettings(&logger.SSettings{}),
			func(_ logger.ILogArg) string { return "" },
		),
		db,
		testNewNetworkNode(addr),
		queue.NewQBProblemProcessor(
			queue.NewSettings(&queue.SSettings{
				FMessageConstructSettings: net_message.NewConstructSettings(&net_message.SConstructSettings{
					FSettings: net_message.NewSettings(&net_message.SSettings{
						FWorkSizeBits: tcWorkSize,
					}),
				}),
				FNetworkMask:      networkMask,
				FMainPoolCapacity: tcQueueCapacity,
				FRandPoolCapacity: tcQueueCapacity,
				FQueuePeriod:      500 * time.Millisecond,
			}),
			client.NewClient(
				tgPrivKey1,
				tcMessageSize,
			),
			asymmetric.NewKEMPrivKey().GetPubKey(),
		),
		asymmetric.NewListPubKeys(),
	)
	return node
}

func testNewNetworkNode(addr string) network.INode {
	return network.NewNode(
		network.NewSettings(&network.SSettings{
			FAddress:      addr,
			FMaxConnects:  tcMaxConnects,
			FReadTimeout:  time.Minute,
			FWriteTimeout: time.Minute,
			FConnSettings: conn.NewSettings(&conn.SSettings{
				FMessageSettings: net_message.NewSettings(&net_message.SSettings{
					FWorkSizeBits: tcWorkSize,
				}),
				FLimitMessageSizeBytes: tcMessageSize,
				FWaitReadTimeout:       time.Hour,
				FDialTimeout:           time.Minute,
				FReadTimeout:           time.Minute,
				FWriteTimeout:          time.Minute,
			}),
		}),
		cache.NewLRUCache(tcCapacity),
	)
}
