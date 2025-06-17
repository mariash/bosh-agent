package agentserver

import (
	"fmt"

	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
	boshmbus "github.com/cloudfoundry/bosh-agent/v2/mbus"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

const ProvideDiskTopic = "disk.provide"

type ProvideDiskRequest struct {
	DiskName     string `json:"disk_name"`
	DiskSizeInMb uint    `json:"disk_size"`
	DiskPoolName string `json:"disk_pool_name"`

	Deployment string `json:"deployment"`

	Metadata map[string]interface{} `json:"metadata"`
}

type ProvideDiskResponse struct {
	Error    string `json:"error"`
	DiskName string `json:"disk_name"`
	DiskHint string `json:"disk_hint"`
}

type DirectorClient interface {
	ProvideDisk(ProvideDiskRequest) (ProvideDiskResponse, error)
}

type directorClient struct {
	mbusHandler boshmbus.Handler
}

func NewDirectorClient(mbusHandler boshmbus.Handler) DirectorClient {
	return directorClient{
		mbusHandler: mbusHandler,
	}
}

func (c directorClient) ProvideDisk(req ProvideDiskRequest) (ProvideDiskResponse, error) {
	var resp ProvideDiskResponse
	err := c.mbusHandler.Request(boshhandler.Director, ProvideDiskTopic, req, &resp)
	if err != nil {
		return ProvideDiskResponse{}, bosherr.WrapError(err, "Sending provide disk request")
	}
	if resp.Error != "" {
		return ProvideDiskResponse{}, bosherr.WrapError(fmt.Errorf(resp.Error), "Provide disk request failed")
	}
	return resp, nil
}
