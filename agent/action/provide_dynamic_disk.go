package action

import (
	"errors"

	boshagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshplatform "github.com/cloudfoundry/bosh-agent/v2/platform"
	boshsettings "github.com/cloudfoundry/bosh-agent/v2/settings"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ProvideDynamicDiskAction struct {
	directorClient  boshagentserver.DirectorClient
	settingsService boshsettings.Service
	platform        boshplatform.Platform
}

type ProvideDynamicDiskTaskResult struct {
	DevicePath string `json:"device_path"`
}

func NewProvideDynamicDiskAction(
	directorClient boshagentserver.DirectorClient,
	settingsService boshsettings.Service,
	platform boshplatform.Platform) ProvideDynamicDiskAction {
	return ProvideDynamicDiskAction{
		directorClient:  directorClient,
		settingsService: settingsService,
		platform:        platform,
	}
}

func (a ProvideDynamicDiskAction) Run(diskName string, diskPoolName string, diskSize uint) (interface{}, error) {
	resp, err := a.directorClient.ProvideDisk(boshagentserver.ProvideDiskDirectorRequest{
		DiskSize:     diskSize,
		DiskName:     diskName,
		DiskPoolName: diskPoolName,
	})
	if err != nil {
		return "", bosherr.WrapError(err, "Sending provide disk request to director")
	}

	diskSettings := a.settingsService.GetSettings().DynamicDiskSettings(resp.DiskName, resp.DiskHint)
	devicePath, err := a.platform.SetupDynamicDisk(diskSettings)
	if err != nil {
		return "", bosherr.WrapError(err, "Setting up dynamic disk")
	}

	return ProvideDynamicDiskTaskResult{DevicePath: devicePath}, nil
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
