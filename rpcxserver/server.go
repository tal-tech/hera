package rpcxserver

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/tal-tech/hera/bootstrap"
	logger "github.com/tal-tech/loggerX"
	rpcxplugin "github.com/tal-tech/odinPlugin"
	"github.com/tal-tech/xtools/confutil"
)

//Server struct
type Server struct {
	server      *server.Server
	Opts        Options
	beforeFuncs []bootstrap.BeforeServerStartFunc
	afterFuncs  []bootstrap.AfterServerStopFunc
	exit        chan os.Signal
}

//NewServer get server instance
func NewServer(options ...OptionFunc) *Server {
	opts := DefaultOptions()

	for _, o := range options {
		o(&opts)
	}

	srv := &Server{
		Opts: opts,
	}
	srv.exit = make(chan os.Signal, 2)
	return srv
}

//NewServerWithOptions with options
func NewServerWithOptions(opts Options) *Server {
	srv := &Server{
		Opts: opts,
	}
	srv.exit = make(chan os.Signal, 2)
	return srv
}

//ConfigureOptions 更新配置
func (srv *Server) ConfigureOptions(options ...OptionFunc) {
	for _, o := range options {
		o(&srv.Opts)
	}
}

//Start 初始化各种插件
func (srv *Server) Serve() error {
	//init rpc server
	srv.server = server.NewServer()

	//before func
	for _, fn := range srv.beforeFuncs {
		err := fn()
		if err != nil {
			return err
		}
	}
	signal.Notify(srv.exit, os.Interrupt, syscall.SIGTERM)
	go srv.waitShutdown()

	logger.I("RpcxServer", "server start listen on %s@%s:%s",
		srv.Opts.Network, srv.Opts.Addr, srv.Opts.Port)

	server.WithReadTimeout(srv.Opts.ReadTimeout)(srv.server)
	server.WithWriteTimeout(srv.Opts.ReadTimeout)(srv.server)
	err := srv.server.Serve(srv.Opts.Network, srv.Opts.Addr+":"+srv.Opts.Port)

	if err != nil && err != server.ErrServerClosed {
		logger.E("RpcxServer", "server error:%v", err)
	} else {
		err = nil
	}
	for _, fn := range srv.afterFuncs {
		fn()
	}
	return err
}

// Stop 平滑关闭
func (srv *Server) waitShutdown() {
	<-srv.exit

	srv.server.UnregisterAll()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	srv.server.Shutdown(ctx)

}

//AddBeforeServerStartFunc add before function
func (srv *Server) AddBeforeServerStartFunc(fns ...bootstrap.BeforeServerStartFunc) {

	for _, fn := range fns {
		srv.beforeFuncs = append(srv.beforeFuncs, fn)
	}
}

//AddAfterServerStopFunc add after function
func (srv *Server) AddAfterServerStopFunc(fns ...bootstrap.AfterServerStopFunc) {

	for _, fn := range fns {
		srv.afterFuncs = append(srv.afterFuncs, fn)
	}
}

//RegisterServiceWithName 用于注册自己的服务
func (srv *Server) RegisterServiceWithName(name string, recv interface{}, metadata string) bootstrap.BeforeServerStartFunc {
	return func() error {
		return srv.server.RegisterName(name, recv, metadata)
	}
}

//RegisterServiceWithName 用于注册带插件功能的服务
func (srv *Server) RegisterServiceWithPlugin(name string, recv interface{}, metadata string) bootstrap.BeforeServerStartFunc {
	return func() error {
		plugin := rpcxplugin.NewRpcxPlugin(name)
		err := setplugin(recv, plugin)
		if err != nil {
			return err
		}
		if metadata == "" {
			metadata = plugin.GetMetadata()
			metadata = "group=" + srv.Opts.RegistryOpts.Group + "&" + metadata
		}

		err = srv.server.RegisterName(name, recv, metadata)
		if err != nil {
			return err
		}
		fn := func() error {
			if srv.server.Plugins != nil {
				srv.server.Plugins.DoUnregister(name)
			}
			time.Sleep(time.Second * 5)
			//newmetadata := plugin.GetMetadata()
			//newmetadata = "group=" + srv.Opts.RegistryOpts.Group + "&" + newmetadata
			return srv.server.RegisterName(name, recv, metadata)
		}
		plugin.SetReregister(fn)
		return nil
	}
}

//setplugin to service struct
func setplugin(v interface{}, plugin *rpcxplugin.RpcxPlugin) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
	} else {
		return errors.New("cannot set plugin to non-pointer struct")
	}
	set(val, plugin)
	return nil
}
func set(val reflect.Value, plugin *rpcxplugin.RpcxPlugin) {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		tpField := typ.Field(i)
		isAnonymous := tpField.Type.Kind() == reflect.Ptr && tpField.Anonymous
		if isAnonymous {
			field.Set(reflect.ValueOf(plugin))
		}
	}
	return
}

//DisableHTTPGateway 禁用本地网关模式
func (srv *Server) DisableHTTPGateway() bootstrap.BeforeServerStartFunc {
	return func() error {
		srv.server.DisableHTTPGateway = true
		return nil
	}
}

//AddPlugins 添加rpcx plugin
func (srv *Server) AddPlugins(plugins ...server.Plugin) {
	for _, plugin := range plugins {
		srv.server.Plugins.Add(plugin)
	}
}

//InitRegistry 初始化注册中心
func (srv *Server) InitConfig() bootstrap.BeforeServerStartFunc {
	return func() error {
		err := confutil.ConfMapToStruct("Server", &srv.Opts)
		return err
	}
}

//InitRegistry 初始化注册中心
func (srv *Server) InitRegistry() bootstrap.BeforeServerStartFunc {
	return func() error {
		regOpts := RegistryOptions{}
		err := confutil.ConfMapToStruct("Registry", &regOpts)

		if err != nil {
			return err
		}
		//如果不启用注册中心，则不初始化
		if regOpts.Status != StatusOn {
			return nil
		}

		if len(regOpts.Addrs) == 0 {
			return errors.New("can not found registry config")
		}

		srv.ConfigureOptions(WithRegistryOptions(regOpts))

		return AddRegistryPlugin(srv)
	}
}

//InitRpcxPlugin 初始化rpcx插件
func (srv *Server) InitRpcxPlugin(plugins ...rpcxplugin.Options) bootstrap.BeforeServerStartFunc {
	return func() error {
		rpcxplugin.Init(srv.Opts.Port, plugins...)
		return nil
	}
}

//RegisterPlugin 添加rcpx plugin
func (srv *Server) RegisterPlugin() bootstrap.BeforeServerStartFunc {

	return func() error {

		return nil
	}
}

//Server 获取rpcx server
func (srv *Server) Server() *server.Server {
	return srv.server
}

//InitRpcxAuth 初始化rpcx鉴权
func (srv *Server) InitRpcxAuth(fns ...ValidAccess) bootstrap.BeforeServerStartFunc {

	return func() error {
		validAuth = confutil.GetConfStringMap("ValidRpcxAuth")
		srv.server.AuthFunc = auth
		if len(fns) > 0 {
			accessFn = fns[0]
		} else {
			accessFn = validAuthAccess
		}
		return nil
	}
}

var validAuth map[string]string

type ValidAccess func(string) (string, bool)

var accessFn ValidAccess

func validAuthAccess(appId string) (string, bool) {
	appKey, ok := validAuth[appId]
	return appKey, ok
}

func auth(ctx context.Context, req *protocol.Message, token string) error {
	key, ok := accessFn(token)
	if !ok {
		return logger.NewError("invalid appId")
	}
	timestamp, ok := req.Metadata["X-Auth-TimeStamp"]
	if !ok {
		return logger.NewError("invalid timestamp")
	}
	sign, ok := req.Metadata["X-Auth-Sign"]
	if !ok {
		return logger.NewError("invalid sign")
	}
	if !check(token, key, timestamp, sign) {
		return logger.NewError("check sign fail")
	}
	return nil
}

func check(appId, key, timestamp, sign string) bool {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(appId + "&" + timestamp + key))
	cipherStr := md5Ctx.Sum(nil)
	signstr := hex.EncodeToString(cipherStr)
	if signstr != sign {
		return false
	}
	return true
}
