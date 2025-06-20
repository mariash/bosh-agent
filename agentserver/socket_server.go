package agentserver

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	boshtask "github.com/cloudfoundry/bosh-agent/v2/agent/task"
	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const socketServerLogTag = "SocketServer"

type ErrorResponse struct {
	Error string `json:"error"`
}

type ProvideDiskAgentRequest struct {
	DiskName     string `json:"disk_name"`
	DiskSize     uint   `json:"disk_size"`
	DiskPoolName string `json:"disk_pool_name"`
}

func (r ProvideDiskAgentRequest) Validate() error {
	if r.DiskName == "" {
		return bosherr.Error("missing disk_name")
	}
	if r.DiskPoolName == "" {
		return bosherr.Error("missing disk_pool_name")
	}
	if r.DiskSize == 0 {
		return bosherr.Error("missing disk_size")
	}

	return nil
}

func (r ProvideDiskAgentRequest) AgentRequest() (boshhandler.Request, error) {
	handlerRequest := payloadForHandler{Arguments: []interface{}{r.DiskName, r.DiskPoolName, r.DiskSize}}
	payload, err := json.Marshal(handlerRequest)
	if err != nil {
		return nil, err
	}
	return boshhandler.NewAgentRequest("provide_dynamic_disk", payload), nil
}

type DetachDiskAgentRequest struct {
	DiskName string `json:"disk_name"`
}

func (r DetachDiskAgentRequest) Validate() error {
	if r.DiskName == "" {
		return bosherr.Error("missing disk_name")
	}

	return nil
}

func (r DetachDiskAgentRequest) AgentRequest() (boshhandler.Request, error) {
	handlerRequest := payloadForHandler{Arguments: []interface{}{r.DiskName}}
	payload, err := json.Marshal(handlerRequest)
	if err != nil {
		return nil, err
	}
	return boshhandler.NewAgentRequest("detach_dynamic_disk", payload), nil
}

type DeleteDiskAgentRequest struct {
	DiskName string `json:"disk_name"`
}

func (r DeleteDiskAgentRequest) Validate() error {
	if r.DiskName == "" {
		return bosherr.Error("missing disk_name")
	}

	return nil
}

func (r DeleteDiskAgentRequest) AgentRequest() (boshhandler.Request, error) {
	handlerRequest := payloadForHandler{Arguments: []interface{}{r.DiskName}}
	payload, err := json.Marshal(handlerRequest)
	if err != nil {
		return nil, err
	}
	return boshhandler.NewAgentRequest("delete_dynamic_disk", payload), nil
}

type TaskStatusResponse struct {
	State  string      `json:"state"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type payloadForHandler struct {
	Arguments []interface{} `json:"arguments"`
}

type socketServer struct {
	logger      boshlog.Logger
	socketPath  string
	taskService boshtask.Service
	listener    net.Listener
}

func NewSocketServer(logger boshlog.Logger, socketPath string, taskService boshtask.Service) AgentServer {
	return &socketServer{
		logger:      logger,
		socketPath:  socketPath,
		taskService: taskService,
	}
}

func (s *socketServer) Start(handlerFunc boshhandler.Func) error {
	err := os.Remove(s.socketPath)
	if err != nil && !os.IsNotExist(err) {
		return bosherr.WrapError(err, "Deleting existing socket")
	}

	s.listener, err = net.Listen("unix", s.socketPath)
	if err != nil {
		return bosherr.WrapError(err, "Listening on socket")
	}
	err = os.Chmod(s.socketPath, 0600)
	if err != nil {
		return bosherr.WrapError(err, "Setting socket permissions")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /disks", func(w http.ResponseWriter, r *http.Request) {
		s.provideDisk(w, r, handlerFunc)
	})

	mux.HandleFunc("POST /disks/{id}/detach", func(w http.ResponseWriter, r *http.Request) {
		s.detachDisk(w, r, handlerFunc)
	})

	mux.HandleFunc("DELETE /disks/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.deleteDisk(w, r, handlerFunc)
	})

	mux.HandleFunc("GET /tasks/{id}", s.taskStatus)

	server := &http.Server{
		Handler: mux,
	}

	return server.Serve(s.listener)
}

func (s *socketServer) Stop() error {
	err := s.listener.Close()
	if err != nil {
		return bosherr.WrapError(err, "Stopping socket server")
	}

	return nil
}

func (s *socketServer) provideDisk(w http.ResponseWriter, r *http.Request, handlerFunc boshhandler.Func) {
	var provideDiskRequest ProvideDiskAgentRequest
	err := s.parseRequest(r, &provideDiskRequest)
	if err != nil {
		s.respond(w, http.StatusBadRequest, boshhandler.NewExceptionResponse(err))
		return
	}

	err = provideDiskRequest.Validate()
	if err != nil {
		s.respond(w, http.StatusBadRequest, boshhandler.NewExceptionResponse(err))
		return
	}

	agentRequest, err := provideDiskRequest.AgentRequest()
	if err != nil {
		s.respond(w, http.StatusInternalServerError, boshhandler.NewExceptionResponse(err))
		return
	}

	resp := handlerFunc(agentRequest)
	s.respond(w, http.StatusOK, resp)
}

func (s *socketServer) detachDisk(w http.ResponseWriter, r *http.Request, handlerFunc boshhandler.Func) {
	detachDiskRequest := DetachDiskAgentRequest{DiskName: r.PathValue("id")}
	err := detachDiskRequest.Validate()
	if err != nil {
		s.respond(w, http.StatusBadRequest, boshhandler.NewExceptionResponse(err))
		return
	}

	agentRequest, err := detachDiskRequest.AgentRequest()
	if err != nil {
		s.respond(w, http.StatusInternalServerError, boshhandler.NewExceptionResponse(err))
		return
	}

	resp := handlerFunc(agentRequest)
	s.respond(w, http.StatusOK, resp)
}

func (s *socketServer) deleteDisk(w http.ResponseWriter, r *http.Request, handlerFunc boshhandler.Func) {
	deleteDiskRequest := DeleteDiskAgentRequest{DiskName: r.PathValue("id")}
	err := deleteDiskRequest.Validate()
	if err != nil {
		s.respond(w, http.StatusBadRequest, boshhandler.NewExceptionResponse(err))
		return
	}

	agentRequest, err := deleteDiskRequest.AgentRequest()
	if err != nil {
		s.respond(w, http.StatusInternalServerError, boshhandler.NewExceptionResponse(err))
		return
	}

	resp := handlerFunc(agentRequest)
	s.respond(w, http.StatusOK, resp)
}

func (s *socketServer) taskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	task, found := s.taskService.FindTaskWithID(taskID)
	if !found {
		s.respond(w, http.StatusNotFound, boshhandler.NewExceptionResponse(fmt.Errorf("task %s not found", taskID)))
		return
	}

	resp := TaskStatusResponse{State: string(task.State)}

	if task.Error != nil {
		resp.Error = task.Error.Error()
	} else {
		resp.Result = task.Value
	}

	s.respond(w, http.StatusOK, resp)
}

func (s *socketServer) respond(w http.ResponseWriter, statusCode int, resp interface{}) {
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error(socketServerLogTag, "Failed marshalling response: %s", err.Error())
		fmt.Fprintf(w, "Failed marshalling response: %s", err.Error())
		return
	}

	w.WriteHeader(statusCode)
	_, err = w.Write(respBytes)
	if err != nil {
		s.logger.Error(socketServerLogTag, "Failed sending response: %s", err.Error())
		fmt.Fprintf(w, "Failed sending response: %s", err.Error())
	}
}

func (s *socketServer) parseRequest(r *http.Request, object interface{}) error {
	err := json.NewDecoder(r.Body).Decode(&object)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}
