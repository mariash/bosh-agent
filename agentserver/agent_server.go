package agentserver

import (
	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
)

type AgentServer interface {
	Start(boshhandler.Func) error
	Stop() error
}
