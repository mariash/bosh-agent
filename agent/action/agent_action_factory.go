package action

import (
	boshas "github.com/cloudfoundry/bosh-agent/v2/agent/applier/applyspec"
	boshagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshplatform "github.com/cloudfoundry/bosh-agent/v2/platform"
	boshsettings "github.com/cloudfoundry/bosh-agent/v2/settings"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type agentActionFactory struct {
	availableActions map[string]Action
}

func NewAgentActionFactory(
	directorClient boshagentserver.DirectorClient,
	specService boshas.V1Service,
	settings boshsettings.Settings,
	platform boshplatform.Platform,
) Factory {
	return agentActionFactory{
		availableActions: map[string]Action{
			"provide_dynamic_disk": NewProvideDynamicDiskAction(directorClient, specService, settings, platform),
		},
	}
}

func (f agentActionFactory) Create(method string) (Action, error) {
	action, found := f.availableActions[method]
	if !found {
		return nil, bosherr.Errorf("Could not create agent action with method %s", method)
	}

	return action, nil
}
