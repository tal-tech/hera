package ginhttp

import (
	"net/http"
	"runtime/debug"

	logger "github.com/tal-tech/loggerX"
	"github.com/jpillora/overseer"
	"github.com/spf13/cast"
	"github.com/tal-tech/xtools/pprofutil"
	"github.com/tal-tech/xtools/expvarutil"
)

var server = &http.Server{}
var addresses = make([]string, 0)

func graceStart(addr string, s *http.Server) {
	addresses = append(addresses, addr)
	server = s
	if pprofutil.PprofPort != "" {
		addresses = append(addresses, pprofutil.PprofPort)
	}
	if expvarutil.ExpvarPort != "" {
		addresses = append(addresses, expvarutil.ExpvarPort)
	}

	oversee()
}

func oversee() {
	overseer.Run(overseer.Config{
		Program:   prog,
		Addresses: addresses,
		//Fetcher: &fetcher.File{Path: "my_app_next"},
		Debug: true, //display log of overseer actions
	})

}

func prog(state overseer.State) {
	logger.I("Program", "app (%s) listening...\n", state.ID)
	if len(addresses) > 1 {
		for k, v := range addresses {
			if v == pprofutil.PprofPort {
				go pprofutil.Start(state.Listeners[k])
			}
			if v == expvarutil.ExpvarPort {
				go expvarutil.Start(state.Listeners[k])
			}
		}

	}
	err := server.Serve(state.Listener)
	if err != nil {
		logger.E("ServerError", "Unhandled error: %v\n stack:%v", err.Error(), cast.ToString(debug.Stack()))
	}
	logger.I("Program", "app (%s) exiting...\n", state.ID)
}
