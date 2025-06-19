package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-agent/v2/agent/action"
	fakeapplyspec "github.com/cloudfoundry/bosh-agent/v2/agent/applier/applyspec/fakes"
	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	fakeagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver/agentserverfakes"
	"github.com/cloudfoundry/bosh-agent/v2/platform/platformfakes"
	boshsettings "github.com/cloudfoundry/bosh-agent/v2/settings"
	fakesettings "github.com/cloudfoundry/bosh-agent/v2/settings/fakes"
)

var _ = Describe("ProvideDynamicDisk", func() {
	var (
		platform                 *platformfakes.FakePlatform
		provideDynamicDiskAction action.ProvideDynamicDiskAction

		settingsService *fakesettings.FakeSettingsService
		directorClient  *fakeagentserver.FakeDirectorClient
		specService     *fakeapplyspec.FakeV1Service
	)

	BeforeEach(func() {
		platform = &platformfakes.FakePlatform{}
		directorClient = &fakeagentserver.FakeDirectorClient{}
		specService = fakeapplyspec.NewFakeV1Service()

		settingsService = &fakesettings.FakeSettingsService{
			Settings: boshsettings.Settings{
				Env: boshsettings.Env{
					PersistentDiskFS:           "some-persistent-disk-fs",
					PersistentDiskMountOptions: []string{"opt1", "opt2"},
					PersistentDiskPartitioner:  "some-persistent-disk-partitioner",
				},
			},
		}

		provideDynamicDiskAction = action.NewProvideDynamicDiskAction(directorClient, specService, settingsService, platform)

		directorClient.ProvideDiskReturns(agentserver.ProvideDiskResponse{DiskName: "some-disk-name", DiskHint: "some-disk-hint"}, nil)
		platform.SetupDynamicDiskReturns("/some/device/path", nil)
	})

	AssertActionIsAsynchronous(provideDynamicDiskAction)
	AssertActionIsNotPersistent(provideDynamicDiskAction)
	AssertActionIsLoggable(provideDynamicDiskAction)

	AssertActionIsNotResumable(provideDynamicDiskAction)
	AssertActionIsNotCancelable(provideDynamicDiskAction)

	It("calls director to provide the disk and sets up the disk", func() {
		result, err := provideDynamicDiskAction.Run("some-disk-name", "some-disk-pool-name", 1024)
		Expect(err).ToNot(HaveOccurred())

		Expect(result).To(Equal(action.ProvideDynamicDiskTaskResult{DevicePath: "/some/device/path"}))
	})

	Context("when director provide disk call fails", func() {
		BeforeEach(func() {
			directorClient.ProvideDiskReturns(agentserver.ProvideDiskResponse{}, errors.New("some-provide-disk-error"))
		})

		It("returns an error", func() {
			_, err := provideDynamicDiskAction.Run("some-disk-name", "some-disk-pool-name", 1024)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("some-provide-disk-error"))
		})
	})

	Context("when setting up dynamic disk fails", func() {
		BeforeEach(func() {
			platform.SetupDynamicDiskReturns("", errors.New("some-setup-disk-error"))
		})

		It("returns an error", func() {
			_, err := provideDynamicDiskAction.Run("some-disk-name", "some-disk-pool-name", 1024)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("some-setup-disk-error"))
		})
	})
})
