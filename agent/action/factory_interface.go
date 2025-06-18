package action

import boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"

type Factory interface {
	Create(requestSource boshhandler.RequestSource, method string) (action Action, err error)
}
