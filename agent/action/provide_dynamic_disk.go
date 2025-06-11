package action

import (
	"errors"

	boshas "github.com/cloudfoundry/bosh-agent/v2/agent/applier/applyspec"
	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ProvideDynamicDiskAction struct {
	directorClient boshagentserver.DirectorClient
	specService    boshas.V1Service
}

func NewProvideDynamicDiskAction(directorClient agentserver.DirectorClient, specService boshas.V1Service) ProvideDynamicDiskAction {
	return ProvideDynamicDiskAction{
		directorClient: directorClient,
		specService:    specService,
	}
}

func (a ProvideDynamicDiskAction) Run() (interface{}, error) {
	// TODO: add some validation
	spec, err := a.specService.Get()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting job spec")
	}

	resp, err := a.directorClient.ProvideDisk(boshagentserver.ProvideDiskRequest{
		Deployment:   spec.Deployment,
		DiskSizeInMb: 1024,
		DiskName:     "some-disk",
		DiskPoolName: "1024",
	})
	if err != nil {
		return nil, bosherr.WrapError(err, "Sending provide disk request to director")
	}
	return resp, nil
}

func (a ProvideDynamicDiskAction) IsAsynchronous(_ ProtocolVersion) bool {
	return true
}

func (a ProvideDynamicDiskAction) IsPersistent() bool {
	return false
}

func (a ProvideDynamicDiskAction) IsLoggable() bool {
	return true
}

func (a ProvideDynamicDiskAction) Resume() (interface{}, error) {
	return nil, errors.New("not supported")
}

func (a ProvideDynamicDiskAction) Cancel() error {
	return errors.New("not supported")
}
