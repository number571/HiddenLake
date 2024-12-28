package http

import (
	"net/http"

	"github.com/number571/go-peer/pkg/logger"
	"github.com/number571/hidden-lake/internal/utils/name"
	"github.com/number571/hidden-lake/pkg/adapters"
)

type IHTTPAdapter interface {
	adapters.IRunnerAdapter

	WithLogger(name.IServiceName, logger.ILogger) IHTTPAdapter
	WithHandlers(map[string]http.HandlerFunc) IHTTPAdapter
	GetOnlines() []string
}

type ISettings interface {
	GetAdapterSettings() adapters.ISettings
	GetAddress() string
}
