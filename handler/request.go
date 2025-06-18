package handler

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ProtocolVersion int

type Request interface {
	GetReplyTo() string
	GetMethod() string
	GetPayload() []byte
	GetProtocolVersion() ProtocolVersion
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

type directorRequest struct {
	ReplyTo         string `json:"reply_to"`
	Method          string
	Payload         []byte
	ProtocolVersion ProtocolVersion `json:"protocol"`
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

func (r directorRequest) GetProtocolVersion() ProtocolVersion {
	return r.ProtocolVersion
}

func NewAgentRequest(method string, payload []byte) agentRequest {
	return agentRequest{
		Method:  method,
		Payload: payload,
	}
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
