package action_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-agent/v2/agent/script/scriptfakes"
	"github.com/cloudfoundry/bosh-agent/v2/platform/platformfakes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	boshaction "github.com/cloudfoundry/bosh-agent/v2/agent/action"
	boshscript "github.com/cloudfoundry/bosh-agent/v2/agent/script"
	boshdir "github.com/cloudfoundry/bosh-agent/v2/settings/directories"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	fakeas "github.com/cloudfoundry/bosh-agent/v2/agent/applier/applyspec/fakes"
	fakeappl "github.com/cloudfoundry/bosh-agent/v2/agent/applier/fakes"
	fakeagentblobstore "github.com/cloudfoundry/bosh-agent/v2/agent/blobstore/blobstorefakes"
	fakecomp "github.com/cloudfoundry/bosh-agent/v2/agent/compiler/fakes"
	fakeblobdelegator "github.com/cloudfoundry/bosh-agent/v2/agent/httpblobprovider/blobstore_delegator/blobstore_delegatorfakes"
	faketask "github.com/cloudfoundry/bosh-agent/v2/agent/task/fakes"
	fakeagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver/agentserverfakes"
	fakejobsuper "github.com/cloudfoundry/bosh-agent/v2/jobsupervisor/fakes"
	fakenotif "github.com/cloudfoundry/bosh-agent/v2/notification/fakes"
	fakesettings "github.com/cloudfoundry/bosh-agent/v2/settings/fakes"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/fake_clock.go code.cloudfoundry.org/clock.Clock

var _ = Describe("directorActionFactory", func() {
	var (
		settingsService   *fakesettings.FakeSettingsService
		platform          *platformfakes.FakePlatform
		blobManager       *fakeagentblobstore.FakeBlobManagerInterface
		taskService       *faketask.FakeService
		notifier          *fakenotif.FakeNotifier
		applier           *fakeappl.FakeApplier
		compiler          *fakecomp.FakeCompiler
		jobSupervisor     *fakejobsuper.FakeJobSupervisor
		specService       *fakeas.FakeV1Service
		jobScriptProvider boshscript.JobScriptProvider
		factory           boshaction.Factory
		logger            boshlog.Logger
		fileSystem        *fakesys.FakeFileSystem
		blobDelegator     *fakeblobdelegator.FakeBlobstoreDelegator
		directorClient    *fakeagentserver.FakeDirectorClient
	)

	BeforeEach(func() {
		settingsService = &fakesettings.FakeSettingsService{}

		platform = &platformfakes.FakePlatform{}
		fileSystem = fakesys.NewFakeFileSystem()
		platform.GetFsReturns(fileSystem)
		platform.GetDirProviderReturns(boshdir.NewProvider("/var/vcap"))

		blobManager = &fakeagentblobstore.FakeBlobManagerInterface{}
		taskService = &faketask.FakeService{}
		notifier = fakenotif.NewFakeNotifier()
		applier = fakeappl.NewFakeApplier()
		compiler = fakecomp.NewFakeCompiler()
		jobSupervisor = fakejobsuper.NewFakeJobSupervisor()
		specService = fakeas.NewFakeV1Service()
		jobScriptProvider = &scriptfakes.FakeJobScriptProvider{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		blobDelegator = &fakeblobdelegator.FakeBlobstoreDelegator{}
		directorClient = &fakeagentserver.FakeDirectorClient{}

		factory = boshaction.NewFactory(
			settingsService,
			platform,
			blobManager,
			taskService,
			notifier,
			applier,
			compiler,
			jobSupervisor,
			specService,
			jobScriptProvider,
			logger,
			blobDelegator,
			directorClient,
		)
	})

	It("returns error if boshaction cannot be created for action type", func() {
		action, err := factory.Create("fake-unknown-action-type", "apply")
		Expect(err).To(HaveOccurred())
		Expect(action).To(BeNil())
	})

	It("returns error if boshaction cannot be created for action method", func() {
		action, err := factory.Create("director", "fake-unknown-boshaction")
		Expect(err).To(HaveOccurred())
		Expect(action).To(BeNil())
	})

	It("apply", func() {
		action, err := factory.Create("director", "apply")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(BeEquivalentTo(boshaction.NewApply(
			applier,
			specService,
			settingsService,
			boshdir.NewProvider("/var/vcap"),
			fileSystem,
		)))
	})

	It("drain", func() {
		action, err := factory.Create("director", "drain")
		Expect(err).ToNot(HaveOccurred())
		// Cannot do equality check since channel is used in initializer
		Expect(action).To(BeAssignableToTypeOf(boshaction.DrainAction{}))
	})

	It("fetch_logs", func() {
		action, err := factory.Create("director", "fetch_logs")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewFetchLogs(platform.GetLogsTarProvider(), blobDelegator)))
	})

	It("fetch_logs_with_signed_url", func() {
		action, err := factory.Create("director", "fetch_logs_with_signed_url")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewFetchLogsWithSignedURLAction(platform.GetLogsTarProvider(), blobDelegator)))
	})

	It("bundle_logs", func() {
		action, err := factory.Create("director", "bundle_logs")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewBundleLogs(platform.GetLogsTarProvider(), platform.GetFs())))
	})

	It("remove_file", func() {
		action, err := factory.Create("director", "remove_file")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewRemoveFile(platform.GetFs())))
	})

	It("get_task", func() {
		action, err := factory.Create("director", "get_task")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewGetTask(taskService)))
	})

	It("cancel_task", func() {
		action, err := factory.Create("director", "cancel_task")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewCancelTask(taskService)))
	})

	It("get_state", func() {
		action, err := factory.Create("director", "get_state")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewGetState(settingsService, specService, jobSupervisor, platform.GetVitalsService())))
	})

	It("list_disk", func() {
		action, err := factory.Create("director", "list_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewListDisk(settingsService, platform, logger)))
	})

	It("migrate_disk", func() {
		action, err := factory.Create("director", "migrate_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewMigrateDisk(platform, platform.GetDirProvider())))
	})

	It("mount_disk", func() {
		action, err := factory.Create("director", "mount_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewMountDisk(settingsService, platform, platform.GetDirProvider(), logger)))
	})

	It("ping", func() {
		action, err := factory.Create("director", "ping")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewPing()))
	})

	It("info", func() {
		action, err := factory.Create("director", "info")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewInfo()))
	})

	It("ssh", func() {
		action, err := factory.Create("director", "ssh")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewSSH(settingsService, platform, platform.GetDirProvider(), logger)))
	})

	It("start", func() {
		action, err := factory.Create("director", "start")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewStart(jobSupervisor, applier, specService)))
	})

	It("stop", func() {
		action, err := factory.Create("director", "stop")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewStop(jobSupervisor)))
	})

	It("remove_persistent_disk", func() {
		action, err := factory.Create("director", "remove_persistent_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewRemovePersistentDiskAction(settingsService)))
	})

	It("unmount_disk", func() {
		action, err := factory.Create("director", "unmount_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewUnmountDisk(settingsService, platform)))
	})

	It("compile_package", func() {
		action, err := factory.Create("director", "compile_package")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewCompilePackage(compiler)))
	})

	It("compile_package_with_signed_url", func() {
		action, err := factory.Create("director", "compile_package_with_signed_url")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewCompilePackageWithSignedURL(compiler)))
	})

	It("run_errand", func() {
		action, err := factory.Create("director", "run_errand")
		Expect(err).ToNot(HaveOccurred())

		// Cannot do equality check since channel is used in initializer
		Expect(action).To(BeAssignableToTypeOf(boshaction.RunErrandAction{}))
	})

	It("run_script", func() {
		action, err := factory.Create("director", "run_script")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewRunScript(jobScriptProvider, specService, logger)))
	})

	It("prepare", func() {
		action, err := factory.Create("director", "prepare")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewPrepare(applier)))
	})

	It("delete_arp_entries", func() {
		action, err := factory.Create("director", "delete_arp_entries")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewDeleteARPEntries(platform)))
	})

	It("shutdown", func() {
		action, err := factory.Create("director", "shutdown")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewShutdown(platform)))
	})

	It("sync_dns", func() {
		action, err := factory.Create("director", "sync_dns")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(boshaction.NewSyncDNS(blobDelegator, settingsService, platform, logger)))
	})

	It("upload_blob", func() {
		action, err := factory.Create("director", "upload_blob")
		Expect(err).ToNot(HaveOccurred())

		Expect(action).To(Equal(boshaction.NewUploadBlobAction(blobManager)))
	})

	It("provide_dynamic_disk", func() {
		action, err := factory.Create("agent", "provide_dynamic_disk")
		Expect(err).ToNot(HaveOccurred())

		Expect(action).To(Equal(boshaction.NewProvideDynamicDiskAction(directorClient, specService, settingsService, platform)))
	})

	It("detach_dynamic_disk", func() {
		action, err := factory.Create("agent", "detach_dynamic_disk")
		Expect(err).ToNot(HaveOccurred())

		Expect(action).To(Equal(boshaction.NewDetachDynamicDiskAction(directorClient)))
	})

	It("delete_dynamic_disk", func() {
		action, err := factory.Create("agent", "delete_dynamic_disk")
		Expect(err).ToNot(HaveOccurred())

		Expect(action).To(Equal(boshaction.NewDeleteDynamicDiskAction(directorClient)))
	})

})
