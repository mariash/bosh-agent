package action

import (
	"errors"

	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeleteDynamicDiskAction struct {
	directorClient boshagentserver.DirectorClient
}

type DeleteDynamicDiskTaskResult struct{}

func NewDeleteDynamicDiskAction(directorClient agentserver.DirectorClient) DeleteDynamicDiskAction {
	return DeleteDynamicDiskAction{
		directorClient: directorClient,
	}
}

func (a DeleteDynamicDiskAction) Run(diskName string) (interface{}, error) {
	_, err := a.directorClient.DeleteDisk(boshagentserver.DeleteDiskDirectorRequest{
		DiskName: diskName,
	})
	if err != nil {
		return "", bosherr.WrapError(err, "Sending detach disk request to director")
	}

	return DeleteDynamicDiskTaskResult{}, nil
}

func (a DeleteDynamicDiskAction) IsAsynchronous(_ ProtocolVersion) bool {
	return true
}

func (a DeleteDynamicDiskAction) IsPersistent() bool {
	return false
}

func (a DeleteDynamicDiskAction) IsLoggable() bool {
	return true
}

func (a DeleteDynamicDiskAction) Resume() (interface{}, error) {
	return nil, errors.New("not supported")
}

func (a DeleteDynamicDiskAction) Cancel() error {
	return errors.New("not supported")
}
