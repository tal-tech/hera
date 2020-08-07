package clijob

import (
	"github.com/tal-tech/hera/bootstrap"
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/confutil"
	"github.com/spf13/cast"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

type JobServer struct {
	*bootstrap.FuncSetter
	Opts Options

	Jobs map[string]Job

	//退出
	exit      chan struct{}
	cmdParser CmdParser
}

func init() {
	confutil.InitConfig()
}

func NewJobServer(options ...OptionFunc) *JobServer {
	opts := DefaultOptions()

	for _, o := range options {
		o(&opts)
	}

	srv := &JobServer{
		Opts:       opts,
		FuncSetter: bootstrap.NewFuncSetter(),
		Jobs:       make(map[string]Job),
		exit:       make(chan struct{}),
		cmdParser:  opts.cmdParser,
	}

	return srv
}

func (js *JobServer) Start() (err error) {
	defer recoverProc()
	if err = js.RunBeforeServerStartFunc(); err != nil {
		return nil
	}

	go js.dealExitSignal()
	go js.doJob()

	<-js.exit
	return nil
}

func (js *JobServer) Stop() {
	logger.I("Stop", "正在退出...")
	js.RunAfterServerStopFunc()
	time.Sleep(1 * time.Second)

	js.exit <- struct{}{}
}

func (js *JobServer) dealExitSignal() {
	sg := make(chan os.Signal, 2)
	signal.Notify(sg, os.Interrupt, syscall.SIGTERM)
	<-sg

	js.Stop()
}

func (js *JobServer) AddJobs(jobs map[string]Job) error {
	if len(jobs) == 0 {
		return logger.NewError("请注入任务")
	}
	for key, j := range jobs {
		js.Jobs[key] = j
	}

	return nil
}

func (js *JobServer) doJob() {
	tag := "doJob"

	defer js.Stop()
	defer processMark(tag, "任务主入口")()

	jobSelected, err := js.cmdParser.JobArgParse(js.Jobs)
	if err != nil {
		logger.E(tag, "解析命令错误:%v", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, myjob := range jobSelected {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			defer processMark(tag, "任务:"+job.Name)()
			if err := job.Do(); err != nil {
				logger.E(tag, "[%v] error: %v", job.Name, err)
			}
		}(myjob)
	}
	wg.Wait()
}

func processMark(tag string, msg string) func() {
	logger.I(tag, "[start],%v", msg)
	return func() {
		logger.I(tag, "[end],%v", msg)
	}
}

func recoverProc() {
	if rec := recover(); rec != nil {
		if err, ok := rec.(error); ok {
			logger.E("PanicRecover", "Unhandled error: %v\n stack:%v", err.Error(), cast.ToString(debug.Stack()))
		} else {
			logger.E("PanicRecover", "Panic: %v\n stack:%v", rec, cast.ToString(debug.Stack()))
		}
		time.Sleep(1 * time.Second)
	}
}
