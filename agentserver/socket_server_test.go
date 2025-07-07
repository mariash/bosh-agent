//go:build !windows
// +build !windows

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
	boshsettings "github.com/cloudfoundry/bosh-agent/v2/settings"
	fakesettings "github.com/cloudfoundry/bosh-agent/v2/settings/fakes"
	"github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SocketServer", func() {
	var (
		tmpDir          string
		socketPath      string
		logger          *loggerfakes.FakeLogger
		taskService     *faketask.FakeService
		settingsService *fakesettings.FakeSettingsService
		socketServer    agentserver.AgentServer
		socketClient    *http.Client
		receivedRequest boshhandler.Request
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "bosh-tests-")
		Expect(err).ToNot(HaveOccurred())
		logger = &loggerfakes.FakeLogger{}
		socketPath = filepath.Join(tmpDir, "agent.sock")
		taskService = faketask.NewFakeService()
		settingsService = &fakesettings.FakeSettingsService{}
		settingsService.Settings.Env.Bosh.Agent.Settings.DiskManagementPrivileges = []boshsettings.DiskManagementPrivilege{
			boshsettings.ProvideDiskManagementPrivilege,
			boshsettings.DetachDiskManagementPrivilege,
			boshsettings.DeleteDiskManagementPrivilege,
		}
	})

	JustBeforeEach(func() {
		socketServer = agentserver.NewSocketServer(logger, socketPath, taskService, settingsService)

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

	Describe("POST /v1/disks", func() {
		It("calls provided handler", func() {
			payload := []byte(`{"disk_name":"some-disk-name","disk_pool_name":"some-disk-pool-name","disk_size":1024}`)

			httpResponse, err := socketClient.Post("http://unix/v1/disks", "application/json", bytes.NewBuffer(payload))
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

				httpResponse, err := socketClient.Post("http://unix/v1/disks", "application/json", bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when disk privileges don't allow to provide disks", func() {
			BeforeEach(func() {
				settingsService.Settings.Env.Bosh.Agent.Settings.DiskManagementPrivileges = []boshsettings.DiskManagementPrivilege{
					boshsettings.DetachDiskManagementPrivilege,
					boshsettings.DeleteDiskManagementPrivilege,
				}
			})

			It("returns an error", func() {
				payload := []byte(`{"disk_name":"some-disk-name","disk_pool_name":"some-disk-pool-name","disk_size":1024}`)

				httpResponse, err := socketClient.Post("http://unix/v1/disks", "application/json", bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("POST /v1/disks/{id}/detach", func() {
		It("calls provided handler", func() {
			payload := []byte(`{}`)

			httpResponse, err := socketClient.Post("http://unix/v1/disks/some-disk-name/detach", "application/json", bytes.NewBuffer(payload))
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
				httpResponse, err := socketClient.Post(`http://unix/v1/disks//detach`, "application/json", bytes.NewBuffer([]byte("{}")))
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusMethodNotAllowed))
			})
		})

		Context("when disk privileges don't allow to detach disks", func() {
			BeforeEach(func() {
				settingsService.Settings.Env.Bosh.Agent.Settings.DiskManagementPrivileges = []boshsettings.DiskManagementPrivilege{
					boshsettings.ProvideDiskManagementPrivilege,
					boshsettings.DeleteDiskManagementPrivilege,
				}
			})

			It("returns an error", func() {
				payload := []byte(`{}`)

				httpResponse, err := socketClient.Post("http://unix/v1/disks/some-disk-name/detach", "application/json", bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("DELETE /v1/disks/{id}", func() {
		It("calls provided handler", func() {
			req, err := http.NewRequest(http.MethodDelete, "http://unix/v1/disks/some-disk-name", nil)
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
				req, err := http.NewRequest(http.MethodDelete, "http://unix/v1/disks", nil)
				Expect(err).NotTo(HaveOccurred())

				httpResponse, err := socketClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusMethodNotAllowed))
			})
		})

		Context("when disk privileges don't allow to delete disks", func() {
			BeforeEach(func() {
				settingsService.Settings.Env.Bosh.Agent.Settings.DiskManagementPrivileges = []boshsettings.DiskManagementPrivilege{
					boshsettings.ProvideDiskManagementPrivilege,
					boshsettings.DetachDiskManagementPrivilege,
				}
			})

			It("returns an error", func() {
				req, err := http.NewRequest(http.MethodDelete, "http://unix/v1/disks/some-disk-name", nil)
				Expect(err).NotTo(HaveOccurred())

				httpResponse, err := socketClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(httpResponse.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})
})

func waitForServerToStart(httpClient *http.Client) {
	Eventually(func() error {
		httpResponse, err := httpClient.Get("http://unix/v1/tasks") //nolint:noctx
		if err == nil {
			httpResponse.Body.Close() //nolint:errcheck
		}
		return err
	}, time.Second*5).Should(Succeed())
}
