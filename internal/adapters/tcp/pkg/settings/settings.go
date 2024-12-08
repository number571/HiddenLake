package settings

import "github.com/number571/hidden-lake/internal/utils/name"

var (
	GServiceName = name.LoadServiceName(CServiceFullName)
)

const (
	CServiceFullName    = "hidden-lake-adapter=tcp"
	CServiceDescription = "adapts HL traffic to a custom TCP connection"
)

const (
	CPathYML = "hla_tcp.yml"
)

const (
	CHandleIndexPath          = "/api/index"
	CHandleConfigSettingsPath = "/api/config/settings"
	CHandleConfigConnectsPath = "/api/config/connects"
	CHandleNetworkOnlinePath  = "/api/network/online"
)

const (
	CHandleNetworkAdapterPath = "/api/network/adapter"
)

const (
	CDefaultHTTPAddress = "127.0.0.1:9521"
	CDefaultTCPAddress  = "127.0.0.1:9522"
)
