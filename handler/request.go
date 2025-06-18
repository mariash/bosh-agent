package handler

import (
	"encoding/json"

	boshtask "github.com/cloudfoundry/bosh-agent/v2/agent/task"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ProtocolVersion int
type RequestSource string

const (
	RequestSourceDirector RequestSource = "director"
	RequestSourceAgent    RequestSource = "agent"
)

type Request interface {
	Source() RequestSource
	GetReplyTo() string
	GetMethod() string
	GetPayload() []byte
	GetProtocolVersion() ProtocolVersion
	TaskStateResponse(taskID string, state boshtask.State) Response
}

type directorRequest struct {
	ReplyTo         string `json:"reply_to"`
	Method          string
	Payload         []byte
	ProtocolVersion ProtocolVersion `json:"protocol"`
}

func NewDirectorRequest(replyTo, method string, payload []byte, protocolVersion ProtocolVersion) directorRequest {
	return directorRequest{
		ReplyTo:         replyTo,
		Method:          method,
		Payload:         payload,
		ProtocolVersion: protocolVersion,
	}
}

func NewDirectorRequestFromJSON(rawJSON []byte) (directorRequest, error) {
	var request directorRequest

	err := json.Unmarshal(rawJSON, &request)
	if err != nil {
		return request, bosherr.WrapError(err, "Unmarshalling JSON payload")
	}
	request.Payload = rawJSON

	return request, nil
}

func (r directorRequest) Source() RequestSource {
	return RequestSourceDirector
}

func (r directorRequest) GetReplyTo() string {
	return r.ReplyTo
}

func (r directorRequest) GetMethod() string {
	return r.Method
}

func (r directorRequest) GetPayload() []byte {
	return r.Payload
}

func (r directorRequest) TaskStateResponse(taskID string, state boshtask.State) Response {
	return NewValueResponse(boshtask.StateValue{
		AgentTaskID: taskID,
		State:       state,
	})
}

func (r directorRequest) GetProtocolVersion() ProtocolVersion {
	return r.ProtocolVersion
}

func NewAgentRequest(method string, payload []byte) agentRequest {
	return agentRequest{
		Method:  method,
		Payload: payload,
	}
}

func (r agentRequest) Source() RequestSource {
	return RequestSourceAgent
}

type agentRequest struct {
	Method  string
	Payload []byte
}

func (r agentRequest) GetReplyTo() string {
	return ""
}

func (r agentRequest) GetMethod() string {
	return r.Method
}

func (r agentRequest) GetPayload() []byte {
	return r.Payload
}

func (r agentRequest) GetProtocolVersion() ProtocolVersion {
	return 0
}

func (r agentRequest) TaskStateResponse(taskID string, state boshtask.State) Response {
	return NewTaskResponse(taskID, state)
}
