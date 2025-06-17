package agentserver

import (
	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . AgentServer
type AgentServer interface {
	Start(boshhandler.Func) error
	Stop() error
}
