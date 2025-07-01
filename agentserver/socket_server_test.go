package agentserver_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	faketask "github.com/cloudfoundry/bosh-agent/v2/agent/task/fakes"
	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	boshhandler "github.com/cloudfoundry/bosh-agent/v2/handler"
	"github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SocketServer", func() {
	var (
		tmpDir          string
		socketPath      string
		taskService     *faketask.FakeService
		socketServer    agentserver.AgentServer
		socketClient    *http.Client
		receivedRequest boshhandler.Request
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "bosh-tests-")
		Expect(err).ToNot(HaveOccurred())
		logger := &loggerfakes.FakeLogger{}
		socketPath = filepath.Join(tmpDir, "agent.sock")
		taskService = faketask.NewFakeService()
		socketServer = agentserver.NewSocketServer(logger, socketPath, taskService)

		receivedRequest = nil
		go func() {
			defer GinkgoRecover()
			err := socketServer.Start(func(req boshhandler.Request) (resp boshhandler.Response) {
				receivedRequest = req
				return boshhandler.NewTaskResponse("1234", "running")
			})
			Expect(err).NotTo(HaveOccurred())
		}()

		socketClient = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		}
		waitForServerToStart(socketClient)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir) //nolint:errcheck
	})

	Describe("POST /disks", func() {
		It("calls provided handler", func() {
			payload := []byte(`{"disk_name":"some-disk-name","disk_pool_name":"some-disk-pool-name","disk_size":1024}`)

			httpResponse, err := socketClient.Post("http://unix/disks", "application/json", bytes.NewBuffer(payload))
			Expect(err).NotTo(HaveOccurred())

			expectedPayload := []byte(`{"arguments":["some-disk-name","some-disk-pool-name",1024]}`)
			Expect(receivedRequest).To(Equal(boshhandler.NewAgentRequest("provide_dynamic_disk", expectedPayload)))

			httpBody, readErr := io.ReadAll(httpResponse.Body)
			Expect(readErr).ToNot(HaveOccurred())
			defer httpResponse.Body.Close() //nolint:errcheck

			Expect(httpBody).To(Equal([]byte(`{"task_id":"1234","state":"running"}`)))
		})

		Context("when request is invalid", func() {
			It("returns an error", func() {
				payload := []byte(`{"invalid-key":"invalid-value"}`)

				httpResponse, err := socketClient.Post("http://unix/disks", "application/json", bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("POST /disks/{id}/detach", func() {
		It("calls provided handler", func() {
			payload := []byte(`{}`)

			httpResponse, err := socketClient.Post("http://unix/disks/some-disk-name/detach", "application/json", bytes.NewBuffer(payload))
			Expect(err).NotTo(HaveOccurred())

			expectedPayload := []byte(`{"arguments":["some-disk-name"]}`)
			Expect(receivedRequest).To(Equal(boshhandler.NewAgentRequest("detach_dynamic_disk", expectedPayload)))

			httpBody, readErr := io.ReadAll(httpResponse.Body)
			Expect(readErr).ToNot(HaveOccurred())
			defer httpResponse.Body.Close() //nolint:errcheck

			Expect(httpBody).To(Equal([]byte(`{"task_id":"1234","state":"running"}`)))
		})

		Context("when request is invalid", func() {
			It("returns an error", func() {
				httpResponse, err := socketClient.Post(`http://unix/disks//detach`, "application/json", bytes.NewBuffer([]byte("{}")))
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusMethodNotAllowed))
			})
		})
	})

	Describe("DELETE /disks/{id}", func() {
		It("calls provided handler", func() {
			req, err := http.NewRequest(http.MethodDelete, "http://unix/disks/some-disk-name", nil)
			Expect(err).NotTo(HaveOccurred())

			httpResponse, err := socketClient.Do(req)
			Expect(err).NotTo(HaveOccurred())

			expectedPayload := []byte(`{"arguments":["some-disk-name"]}`)
			Expect(receivedRequest).To(Equal(boshhandler.NewAgentRequest("delete_dynamic_disk", expectedPayload)))

			httpBody, readErr := io.ReadAll(httpResponse.Body)
			Expect(readErr).ToNot(HaveOccurred())
			defer httpResponse.Body.Close() //nolint:errcheck

			Expect(httpBody).To(Equal([]byte(`{"task_id":"1234","state":"running"}`)))
		})

		Context("when request is invalid", func() {
			It("returns an error", func() {
				req, err := http.NewRequest(http.MethodDelete, "http://unix/disks", nil)
				Expect(err).NotTo(HaveOccurred())

				httpResponse, err := socketClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusMethodNotAllowed))
			})
		})
	})
})

func waitForServerToStart(httpClient *http.Client) {
	Eventually(func() error {
		httpResponse, err := httpClient.Get("http://unix/tasks") //nolint:noctx
		if err == nil {
			httpResponse.Body.Close() //nolint:errcheck
		}
		return err
	}, time.Second*5).Should(Succeed())
}
