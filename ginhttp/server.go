package ginhttp

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/hera/bootstrap"
	"github.com/tal-tech/xtools/confutil"
	"github.com/gin-gonic/gin"
)

type Server struct {
	server      *http.Server
	beforeFuncs []bootstrap.BeforeServerStartFunc
	afterFuncs  []bootstrap.AfterServerStopFunc
	opts        ServerOptions
	exit        chan os.Signal
}

func NewServer() *Server {
	opts := DefaultOptions()
	s := new(Server)
	s.opts = opts
	handler := gin.New()
	server := &http.Server{
		Addr:         opts.Addr,
		Handler:      handler,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.WriteTimeout,
	}
	s.exit = make(chan os.Signal, 2)
	s.server = server
	return s
}

//Serve serve http request
func (s *Server) Serve() error {
	var err error
	for _, fn := range s.beforeFuncs {
		err = fn()
		if err != nil {
			return err
		}
	}

	if s.opts.Grace {
		graceStart(s.opts.Addr, s.server)
	} else {
		signal.Notify(s.exit, os.Interrupt, syscall.SIGTERM)
		go s.waitShutdown()
		logger.I("GIN-Server", "http server start and serve:%s\n", server.Addr)
		err = s.server.ListenAndServe()
	}

	for _, fn := range s.afterFuncs {
		fn()
	}
	return err
}

//Shutdown close http server
func (s *Server) waitShutdown() {
	<-s.exit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.I("GIN-Server", "shutdown http server ...")

	err := s.server.Shutdown(ctx)
	if err != nil {
		logger.E("GIN-Server", "shutdown http server error:%s", err)
	}
	return
}

func (s *Server) GetServer() *http.Server {
	return s.server
}

func (s *Server) GetGinEngine() *gin.Engine {
	return s.server.Handler.(*gin.Engine)
}

//保留兼容
func (s *Server) AddServerBeforeFunc(fns ...bootstrap.BeforeServerStartFunc) {
	for _, fn := range fns {
		s.beforeFuncs = append(s.beforeFuncs, fn)
	}
}
func (s *Server) AddServerAfterFunc(fns ...bootstrap.AfterServerStopFunc) {
	for _, fn := range fns {
		s.afterFuncs = append(s.afterFuncs, fn)
	}
}

func (s *Server) AddBeforeServerStartFunc(fns ...bootstrap.BeforeServerStartFunc) {
	for _, fn := range fns {
		s.beforeFuncs = append(s.beforeFuncs, fn)
	}
}

func (s *Server) AddAfterServerStopFunc(fns ...bootstrap.AfterServerStopFunc) {
	for _, fn := range fns {
		s.afterFuncs = append(s.afterFuncs, fn)
	}
}

func (s *Server) InitConfig() bootstrap.BeforeServerStartFunc {
	return func() error {
		err := confutil.ConfMapToStruct("Server", &s.opts)
		if err != nil {
			return err
		}
		s.server.Addr = s.opts.Addr
		s.server.ReadTimeout = s.opts.ReadTimeout
		s.server.WriteTimeout = s.opts.WriteTimeout
		s.server.IdleTimeout = s.opts.IdleTimeout
		if s.opts.Mode != "" {
			os.Setenv(gin.EnvGinMode, s.opts.Mode)
			gin.SetMode(s.opts.Mode)
		}
		return nil
	}
}
