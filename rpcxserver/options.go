package rpcxserver

import (
	"time"
)

const (
	//StatusOn 启用
	StatusOn = "on"
	//StatusOff 关闭
	StatusOff = "off"
)

//RegistryOptions 服务注册中心配置
type RegistryOptions struct {
	Status         string        `ini:"status"`
	Addrs          []string      `ini:"addrs"`
	BasePath       string        `ini:"basePath"`
	UpdateInterval time.Duration `ini:"updateInterval"`
	UserName       string        `ini:"username"`
	Password       string        `ini:"password"`
	Group          string        `ini:"group"`
}

type OptionFunc func(*Options)

//Options server options
type Options struct {
	Network      string        `ini:"network"`
	Addr         string        `ini:"addr"`
	Port         string        `ini:"port"`
	WriteTimeout time.Duration `ini:"writeTimeout"`
	ReadTimeout  time.Duration `ini:"readTimeout"`
	RegistryOpts RegistryOptions
}

//DefaultOptions default config
func DefaultOptions() Options {
	return Options{
		Network: "tcp",
		Addr:    "127.0.0.1",
		Port:    "18900",
	}
}

func Network(n string) OptionFunc {
	return func(o *Options) {
		o.Network = n
	}
}

func Addr(a string) OptionFunc {
	return func(o *Options) {
		o.Addr = a
	}
}

func Port(p string) OptionFunc {
	return func(o *Options) {
		o.Port = p
	}
}

func ReadTimeout(t time.Duration) OptionFunc {
	return func(o *Options) {
		o.ReadTimeout = t
	}
}

func WriteTimeout(t time.Duration) OptionFunc {
	return func(o *Options) {
		o.WriteTimeout = t
	}
}

func WithRegistryOptions(registryOpts RegistryOptions) OptionFunc {
	return func(o *Options) {
		o.RegistryOpts = registryOpts
	}
}
