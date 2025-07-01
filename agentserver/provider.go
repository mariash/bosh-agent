package agentserver

import (
	"path/filepath"

	boshtask "github.com/cloudfoundry/bosh-agent/v2/agent/task"
	boshdirs "github.com/cloudfoundry/bosh-agent/v2/settings/directories"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type provider struct {
	servers map[string]AgentServer
}

type Provider interface {
	Get(name string) (AgentServer, error)
}

func NewProvider(logger boshlog.Logger, dirProvider boshdirs.Provider, taskService boshtask.Service) Provider {
	socketPath := filepath.Join(dirProvider.BoshDir(), "agent.sock")
	socketServer := NewSocketServer(logger, socketPath, taskService)
	noopServer := NewNoopServer(logger)

	return provider{
		servers: map[string]AgentServer{
			"ubuntu":  socketServer,
			"centos":  socketServer,
			"dummy":   noopServer,
			"windows": noopServer,
		},
	}
}

func (p provider) Get(name string) (AgentServer, error) {
	server, found := p.servers[name]
	if !found {
		return nil, bosherr.Errorf("Agent server %s could not be found", name)
	}
	return server, nil
}
