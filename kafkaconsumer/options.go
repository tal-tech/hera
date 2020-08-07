package kafkaconsumer

const (
	//StatusOn 启用
	StatusOn = "on"
	//StatusOff 关闭
	StatusOff = "off"
)

////RegistryOptions 服务注册中心配置
//type RegistryOptions struct {
//	Status         string        `ini:"status"`
//	Addrs          []string      `ini:"addrs"`
//	BasePath       string        `ini:"basePath"`
//	UpdateInterval time.Duration `ini:"updateInterval"`
//	UserName       string        `ini:"username"`
//	Password       string        `ini:"password"`
//}

type OptionFunc func(*Options)

//Options server options
type Options struct {
	KafkaHost   string `ini:"kafkahost"`
	Topic       string `ini:"topic"`
	FailTopic   string `ini:"failtopic"`
	ConsumerCnt int    `ini:"consumerCnt"`
	GroupName   string `ini:"groupName"`
}

//DefaultOptions default config
func DefaultOptions() Options {
	return Options{
		KafkaHost:   "127.0.0.1:9092",
		Topic:       "",
		FailTopic:   "fail_raw",
		ConsumerCnt: 25,
		GroupName:   "no_name",
	}
}

func KafkaHost(h string) OptionFunc {
	return func(o *Options) {
		o.KafkaHost = h
	}
}

func Topic(t string) OptionFunc {
	return func(o *Options) {
		o.Topic = t
	}
}
func FailTopic(f string) OptionFunc {
	return func(o *Options) {
		o.FailTopic = f
	}
}
func ConsumerCnt(c int) OptionFunc {
	return func(o *Options) {
		o.ConsumerCnt = c
	}
}
func GroupName(n string) OptionFunc {
	return func(o *Options) {
		o.GroupName = n
	}
}
