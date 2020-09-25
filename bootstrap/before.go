package bootstrap

import (
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/loggerX/builders"
	"github.com/tal-tech/xtools/confutil"
	"github.com/tal-tech/xtools/limitutil"
	"github.com/tal-tech/xtools/perfutil"
	"github.com/tal-tech/xtools/pprofutil"
)

type BeforeServerStartFunc func() error

func InitLogger() BeforeServerStartFunc {
	return func() error {
		logger.InitLogger("")
		return nil
	}
}

func InitLoggerWithConf(sections ...string) BeforeServerStartFunc {
	return func() error {
		section := "Log"
		if len(sections) > 0 {
			section = sections[0]
		}
		logMap := confutil.GetConfStringMap(section) //通过配置文件转为map[string]string
		config := logger.NewLogConfig()
		config.SetConfigMap(logMap)
		logger.InitLogWithConfig(config)
		return nil
	}
}

func InitTraceLogger(department, version string) BeforeServerStartFunc {
	return func() error {
		builder := new(builders.TraceBuilder)
		builder.SetTraceDepartment(department)
		builder.SetTraceVersion(version)
		logger.SetBuilder(builder)
		return nil
	}
}

func InitPerfutil() BeforeServerStartFunc {
	return func() error {
		perfutil.InitPerfWithConfig(perfutil.NewperfConfig())
		return nil
	}
}

func GrowMaxFd() BeforeServerStartFunc {
	return func() error {
		return limitutil.GrowToMaxFdLimit()
	}
}

func InitPprof() BeforeServerStartFunc {
	return func() error {
		go pprofutil.Pprof()
		return nil
	}
}
