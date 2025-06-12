package handler

import (
	"github.com/cloudfoundry/bosh-agent/v2/agent/task"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Response interface {
	// Shorten attempts to return a response
	// that can be serialized in a smaller size.
	Shorten() Response
}

type valueResponse struct {
	Value interface{} `json:"value"`
}

func NewValueResponse(value interface{}) Response {
	return valueResponse{Value: value}
}

func (r valueResponse) Shorten() Response {
	return r
}

type exceptionResponse struct {
	Exception struct {
		Message string `json:"message,omitempty"`
	} `json:"exception"`

	err error
}

func NewExceptionResponse(err error) (resp Response) {
	r := exceptionResponse{}
	r.Exception.Message = err.Error()
	r.err = err
	return r
}

func (r exceptionResponse) Shorten() Response {
	if typedErr, ok := r.err.(bosherr.ShortenableError); ok {
		sr := exceptionResponse{}
		sr.Exception.Message = typedErr.ShortError()
		sr.err = typedErr
		return sr
	}

	return r
}

type taskResponse struct {
	TaskID string     `json:"task_id"`
	State  task.State `json:"state"`
}

func NewTaskResponse(taskID string, state task.State) Response {
	return taskResponse{TaskID: taskID, State: state}
}

func (r taskResponse) Shorten() Response {
	return r
}
