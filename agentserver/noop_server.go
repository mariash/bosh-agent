package agentserver

import (
	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type noopServer struct {
	logger boshlog.Logger
}

func NewNoopServer(logger boshlog.Logger) AgentServer {
	return &noopServer{
		logger: logger,
	}
}

func (s *noopServer) Start(handlerFunc boshhandler.Func) error {
	return nil
}

func (s *noopServer) Stop() error {
	return nil
}
