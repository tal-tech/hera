package kafkaconsumer

import (
	"strings"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/hera/bootstrap"
	"github.com/tal-tech/xtools/confutil"
	"github.com/tal-tech/xtools/limitutil"
	"github.com/Shopify/sarama"
	"github.com/spf13/cast"
)

//Server struct
type Server struct {
	server      *IKafka
	Opts        Options
	beforeFuncs []bootstrap.BeforeServerStartFunc
	afterFuncs  []bootstrap.AfterServerStopFunc
}

var execfuncs = make(map[string]func([]byte) error)

func init() {
	sarama.Logger = &saramaLog{}
}

//NewServer get server instance
func NewServer(options ...OptionFunc) *Server {
	opts := DefaultOptions()

	for _, o := range options {
		o(&opts)
	}
	if len(options) == 0 {
		autoInitConf(&opts)
	}

	srv := &Server{
		Opts: opts,
	}
	return srv
}
func autoInitConf(opts *Options) {
	configs := confutil.GetConfStringMap("KafkaServer")
	if _, ok := configs["groupName"]; !ok {
		panic("kafka分组名称不存在")
	}
	opts.GroupName = configs["groupName"]
	if _, ok := configs["kafkaHost"]; !ok {
		panic("kafka host不存在")
	}
	opts.KafkaHost = configs["kafkaHost"]
	if _, ok := configs["failTopic"]; !ok {
		panic("kafka 失败topic不存在")
	}
	opts.FailTopic = configs["failTopic"]
	if _, ok := configs["topic"]; !ok {
		panic("kafka的topic不存在")
	}
	opts.Topic = configs["topic"]
	if _, ok := configs["consumerCnt"]; ok {
		opts.ConsumerCnt = cast.ToInt(configs["consumerCnt"])
	}
}

//NewServerWithOptions with options
func NewServerWithOptions(opts Options) *Server {
	srv := &Server{
		Opts: opts,
	}

	return srv
}

//ConfigureOptions 更新配置
func (srv *Server) ConfigureOptions(options ...OptionFunc) {
	for _, o := range options {
		o(&srv.Opts)
	}
}

//Start 初始化各种插件
func (srv *Server) Start() error {
	if err := limitutil.GrowToMaxFdLimit(); err != nil {
		logger.E("Fd Error", "try grow to max limit under normal priviledge, failed")
		return err
	}
	//init rpc server
	srv.server = NewIKafka(new(Handler).Deal)

	//before func
	for _, fn := range srv.beforeFuncs {
		err := fn()
		if err != nil {
			return err
		}
	}

	kafkaHost := strings.Split(srv.Opts.KafkaHost, ",")
	if len(kafkaHost) == 0 {
		logger.E("kafkaServer", "kafkaHost is empty")
	}

	broker := IBroker{
		Topic:     srv.Opts.Topic,
		FailTopic: srv.Opts.FailTopic,
		KafkaHost: kafkaHost,
	}
	consumer := &IConsumer{
		Cnt:     srv.Opts.ConsumerCnt,
		GroupId: srv.Opts.GroupName,
	}

	srv.server.Run([]IBroker{broker}, consumer)
	logger.I("kafkaServer", "server start IBroker %+v",
		broker)
	return nil
}

// Stop 平滑关闭
func (srv *Server) Stop() {

	srv.server.Close()

	logger.I("kafkaServer", "server stoped")

	for _, fn := range srv.afterFuncs {
		fn()
	}
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

//Server 获取rpcx server
func (srv *Server) Server() *IKafka {
	return srv.server
}

//注入消息处理方法
func (srv *Server) InjectHandleFuncs(funcs []HandlerFunc) {
	for _, handlefunc := range funcs {
		if handlefunc.MessageKey == "" || handlefunc.Func == nil {
			continue
		}
		execfuncs[handlefunc.MessageKey] = handlefunc.Func
	}
}
