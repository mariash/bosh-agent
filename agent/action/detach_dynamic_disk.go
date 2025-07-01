package action

import (
	"errors"

	boshagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DetachDynamicDiskAction struct {
	directorClient boshagentserver.DirectorClient
}

type DetachDynamicDiskTaskResult struct{}

func NewDetachDynamicDiskAction(directorClient boshagentserver.DirectorClient) DetachDynamicDiskAction {
	return DetachDynamicDiskAction{
		directorClient: directorClient,
	}
}

func (a DetachDynamicDiskAction) Run(diskName string) (interface{}, error) {
	_, err := a.directorClient.DetachDisk(boshagentserver.DetachDiskDirectorRequest{
		DiskName: diskName,
	})
	if err != nil {
		return "", bosherr.WrapError(err, "Sending detach disk request to director")
	}

	return DetachDynamicDiskTaskResult{}, nil
}

func (a DetachDynamicDiskAction) IsAsynchronous(_ ProtocolVersion) bool {
	return true
}

func (a DetachDynamicDiskAction) IsPersistent() bool {
	return false
}

func (a DetachDynamicDiskAction) IsLoggable() bool {
	return true
}

func (a DetachDynamicDiskAction) Resume() (interface{}, error) {
	return nil, errors.New("not supported")
}

func (a DetachDynamicDiskAction) Cancel() error {
	return errors.New("not supported")
}
