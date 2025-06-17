package action

import (
	"errors"

	boshas "github.com/cloudfoundry/bosh-agent/v2/agent/applier/applyspec"
	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshplatform "github.com/cloudfoundry/bosh-agent/v2/platform"
	boshsettings "github.com/cloudfoundry/bosh-agent/v2/settings"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ProvideDynamicDiskAction struct {
	directorClient boshagentserver.DirectorClient
	specService    boshas.V1Service
	settings       boshsettings.Settings
	platform       boshplatform.Platform
}

func NewProvideDynamicDiskAction(directorClient agentserver.DirectorClient, specService boshas.V1Service, settings boshsettings.Settings, platform boshplatform.Platform) ProvideDynamicDiskAction {
	return ProvideDynamicDiskAction{
		directorClient: directorClient,
		specService:    specService,
		settings:       settings,
		platform:       platform,
	}
}

func (a ProvideDynamicDiskAction) Run(diskName string, diskPoolName string, diskSizeInMb uint) (interface{}, error) {
	spec, err := a.specService.Get()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting job spec")
	}

	resp, err := a.directorClient.ProvideDisk(boshagentserver.ProvideDiskRequest{
		Deployment:   spec.Deployment,
		DiskSizeInMb: diskSizeInMb,
		DiskName:     diskName,
		DiskPoolName: diskPoolName,
	})
	if err != nil {
		return nil, bosherr.WrapError(err, "Sending provide disk request to director")
	}

	diskSettings := a.settings.DynamicDiskSettings(resp.DiskName, resp.DiskHint)
	err = a.platform.SetupDynamicDisk(diskSettings)
	if err != nil {
		return nil, bosherr.WrapError(err, "Setting up dynamic disk")
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
