//+build zookeeper

package rpcxserver

import (
	"github.com/smallnest/rpcx/serverplugin"
)

func AddRegistryPlugin(s *Server) error {
	plugin := &serverplugin.ZooKeeperRegisterPlugin{
		ServiceAddress:   s.Opts.Network + "@" + s.Opts.Addr + ":" + s.Opts.Port,
		ZooKeeperServers: s.Opts.RegistryOpts.Addrs,
		BasePath:         s.Opts.RegistryOpts.BasePath,
		UpdateInterval:   s.Opts.RegistryOpts.UpdateInterval,
	}

	err := plugin.Start()

	if err != nil {
		return err
	}

	s.server.Plugins.Add(plugin)

	return nil
}
