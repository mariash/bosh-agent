package agentserver

import (
	"encoding/json"
	"fmt"
	"io"
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

type TaskStatusResponse struct {
	State  string      `json:"state"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
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

	mux.HandleFunc("POST /disks/{id}/detach", s.detachDisk)
	mux.HandleFunc("DELETE /disks/{id}", s.deleteDisk)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.respond(w, http.StatusBadRequest, boshhandler.NewExceptionResponse(err))
		return
	}
	defer r.Body.Close()

	resp := handlerFunc(boshhandler.NewAgentRequest("provide_dynamic_disk", body))
	s.respond(w, http.StatusOK, resp)
}

func (s *socketServer) detachDisk(w http.ResponseWriter, r *http.Request) {
}

func (s *socketServer) deleteDisk(w http.ResponseWriter, r *http.Request) {
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
