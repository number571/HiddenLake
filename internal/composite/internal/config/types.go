package config

import (
	logger "github.com/number571/hidden-lake/internal/modules/logger/std"
)

type IConfig interface {
	GetLogging() logger.ILogging
	GetServices() []string
}
