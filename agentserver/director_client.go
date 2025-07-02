package agentserver

import (
	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

const (
	ProvideDiskTopic = "disk.provide"
	DetachDiskTopic  = "disk.detach"
	DeleteDiskTopic  = "disk.delete"
)

type ProvideDiskDirectorRequest struct {
	DiskName     string `json:"disk_name"`
	DiskSize     uint   `json:"disk_size"`
	DiskPoolName string `json:"disk_pool_name"`

	Deployment string `json:"deployment"`

	Metadata map[string]interface{} `json:"metadata"`
}

type ProvideDiskDirectorResponse struct {
	Error    string `json:"error"`
	DiskName string `json:"disk_name"`
	DiskHint string `json:"disk_hint"`
}

type DetachDiskDirectorRequest struct {
	DiskName string `json:"disk_name"`
}

type DetachDiskDirectorResponse struct {
	Error string `json:"error"`
}

type DeleteDiskDirectorRequest struct {
	DiskName string `json:"disk_name"`
}

type DeleteDiskDirectorResponse struct {
	Error string `json:"error"`
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . DirectorClient
type DirectorClient interface {
	ProvideDisk(ProvideDiskDirectorRequest) (ProvideDiskDirectorResponse, error)
	DetachDisk(DetachDiskDirectorRequest) (DetachDiskDirectorResponse, error)
	DeleteDisk(req DeleteDiskDirectorRequest) (DeleteDiskDirectorResponse, error)
}

type Handler interface {
	Run(boshhandler.Func) error
	Start(boshhandler.Func) error
	RegisterAdditionalFunc(boshhandler.Func)
	Send(target boshhandler.Target, topic boshhandler.Topic, message interface{}) error
	Request(target boshhandler.Target, topic boshhandler.Topic, message interface{}, response interface{}) error
	Stop()
}

type directorClient struct {
	mbusHandler Handler
}

func NewDirectorClient(mbusHandler Handler) DirectorClient {
	return directorClient{
		mbusHandler: mbusHandler,
	}
}

func (c directorClient) ProvideDisk(req ProvideDiskDirectorRequest) (ProvideDiskDirectorResponse, error) {
	var resp ProvideDiskDirectorResponse
	err := c.mbusHandler.Request(boshhandler.Director, ProvideDiskTopic, req, &resp)
	if err != nil {
		return ProvideDiskDirectorResponse{}, bosherr.WrapError(err, "Sending provide disk request")
	}
	if resp.Error != "" {
		return ProvideDiskDirectorResponse{}, bosherr.Errorf("Provide disk request failed: %s", resp.Error)
	}
	return resp, nil
}

func (c directorClient) DetachDisk(req DetachDiskDirectorRequest) (DetachDiskDirectorResponse, error) {
	var resp DetachDiskDirectorResponse
	err := c.mbusHandler.Request(boshhandler.Director, DetachDiskTopic, req, &resp)
	if err != nil {
		return DetachDiskDirectorResponse{}, bosherr.WrapError(err, "Sending detach disk request")
	}
	if resp.Error != "" {
		return DetachDiskDirectorResponse{}, bosherr.Errorf("Detach disk request failed: %s", resp.Error)
	}
	return resp, nil
}

func (c directorClient) DeleteDisk(req DeleteDiskDirectorRequest) (DeleteDiskDirectorResponse, error) {
	var resp DeleteDiskDirectorResponse
	err := c.mbusHandler.Request(boshhandler.Director, DeleteDiskTopic, req, &resp)
	if err != nil {
		return DeleteDiskDirectorResponse{}, bosherr.WrapError(err, "Sending delete disk request")
	}
	if resp.Error != "" {
		return DeleteDiskDirectorResponse{}, bosherr.Errorf("Delete disk request failed: %s", resp.Error)
	}
	return resp, nil
}
