package agentserver_test

import (
	"path/filepath"

	faketask "github.com/cloudfoundry/bosh-agent/v2/agent/task/fakes"
	boshdir "github.com/cloudfoundry/bosh-agent/v2/settings/directories"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cloudfoundry/bosh-agent/v2/agentserver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider", func() {
	var (
		provider    agentserver.Provider
		taskService *faketask.FakeService
		logger      boshlog.Logger
		dirProvider boshdir.Provider
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		dirProvider = boshdir.NewProvider("/fake-base-dir")
		taskService = faketask.NewFakeService()

		provider = agentserver.NewProvider(
			logger,
			dirProvider,
			taskService,
		)
	})

	It("provides socket server for centos platform", func() {
		actualAgentServer, err := provider.Get("centos")
		Expect(err).ToNot(HaveOccurred())
		Expect(actualAgentServer).To(Equal(agentserver.NewSocketServer(logger, filepath.Join(dirProvider.BoshDir(), "agent.sock"), taskService)))
	})

	It("provides socket server for ubuntu platform", func() {
		actualAgentServer, err := provider.Get("ubuntu")
		Expect(err).ToNot(HaveOccurred())
		Expect(actualAgentServer).To(Equal(agentserver.NewSocketServer(logger, filepath.Join(dirProvider.BoshDir(), "agent.sock"), taskService)))
	})

	It("provides noop server for dummy platform", func() {
		actualAgentServer, err := provider.Get("dummy")
		Expect(err).ToNot(HaveOccurred())
		Expect(actualAgentServer).To(Equal(agentserver.NewNoopServer(logger)))
	})

	It("provides noop server for windows platform", func() {
		actualAgentServer, err := provider.Get("windows")
		Expect(err).ToNot(HaveOccurred())
		Expect(actualAgentServer).To(Equal(agentserver.NewNoopServer(logger)))
	})

	It("returns an error when the supervisor is not found", func() {
		_, err := provider.Get("does-not-exist")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does-not-exist could not be found"))
	})
})
