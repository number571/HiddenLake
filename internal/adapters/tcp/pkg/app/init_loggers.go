package app

import hla_tcp_settings "github.com/number571/hidden-lake/internal/adapters/tcp/pkg/settings"

func (p *sApp) initLoggers() {
	p.fHTTPAdapter.WithLogger(hla_tcp_settings.GServiceName, p.fHTTPLogger)
	p.fTCPAdapter.WithLogger(hla_tcp_settings.GServiceName, p.fAnonLogger)
}
