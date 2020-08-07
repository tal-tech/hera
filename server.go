package hera

import "github.com/tal-tech/hera/bootstrap"

//Server接口，未直接使用，ADD新类型参考
type Hera interface {
	AddBeforeServerStartFunc(bootstrap.BeforeServerStartFunc)
	AddAfterServerStopFunc(bootstrap.AfterServerStopFunc)
	Serve() error
}
