package platform_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedpresolv "github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver/fakes"
	. "github.com/cloudfoundry/bosh-agent/platform"
	fakecert "github.com/cloudfoundry/bosh-agent/platform/cert/fakes"
	fakenet "github.com/cloudfoundry/bosh-agent/platform/net/fakes"
	fakestats "github.com/cloudfoundry/bosh-agent/platform/stats/fakes"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	boshdirs "github.com/cloudfoundry/bosh-agent/settings/directories"
	"github.com/cloudfoundry/bosh-agent/syslog"
	fakesyslog "github.com/cloudfoundry/bosh-agent/syslog/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("WindowsPlatform", func() {
	var (
		collector                  *fakestats.FakeCollector
		fs                         *fakesys.FakeFileSystem
		cmdRunner                  *fakesys.FakeCmdRunner
		dirProvider                boshdirs.Provider
		netManager                 *fakenet.FakeManager
		devicePathResolver         *fakedpresolv.FakeDevicePathResolver
		platform                   Platform
		fakeDefaultNetworkResolver *fakenet.FakeDefaultNetworkResolver
		certManager                *fakecert.FakeManager
		syslogger                  syslog.Logger

		logger boshlog.Logger
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		collector = &fakestats.FakeCollector{}
		fs = fakesys.NewFakeFileSystem()
		cmdRunner = fakesys.NewFakeCmdRunner()
		dirProvider = boshdirs.NewProvider("/fake-dir")
		netManager = &fakenet.FakeManager{}
		devicePathResolver = fakedpresolv.NewFakeDevicePathResolver()
		fakeDefaultNetworkResolver = &fakenet.FakeDefaultNetworkResolver{}
		certManager = new(fakecert.FakeManager)
		syslogger = fakesyslog.NewFakeSyslogger()
	})

	JustBeforeEach(func() {
		platform = NewWindowsPlatform(
			collector,
			fs,
			cmdRunner,
			dirProvider,
			netManager,
			certManager,
			devicePathResolver,
			logger,
			fakeDefaultNetworkResolver,
			syslogger,
		)
	})

	Describe("GetFileContentsFromCDROM", func() {
		It("reads file from D drive", func() {
			fs.WriteFileString("D:/env", "fake-contents")
			contents, err := platform.GetFileContentsFromCDROM("env")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal([]byte("fake-contents")))
		})
	})

	Describe("SetupTmpDir", func() {
		act := func() error {
			return platform.SetupTmpDir()
		}

		It("creates new temp dir", func() {
			err := act()
			Expect(err).NotTo(HaveOccurred())

			fileStats := fs.GetFileTestStat("/fake-dir/data/tmp")
			Expect(fileStats).NotTo(BeNil())
			Expect(fileStats.FileType).To(Equal(fakesys.FakeFileType(fakesys.FakeFileTypeDir)))
		})

		It("returns error if creating new temp dir errs", func() {
			fs.MkdirAllError = errors.New("fake-mkdir-error")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-mkdir-error"))
		})

		It("sets TMP and TEMP environment variable so that children of this process will use new temp dir", func() {
			err := act()
			Expect(err).NotTo(HaveOccurred())

			fakeTmpDir := filepath.FromSlash("/fake-dir/data/tmp")
			Expect(os.Getenv("TMP")).To(Equal(fakeTmpDir))
			Expect(os.Getenv("TEMP")).To(Equal(fakeTmpDir))
		})

		It("returns error if setting TMPDIR errs", func() {
			// uses os package; no way to trigger err
		})
	})

	Describe("SetupNetworking", func() {
		It("delegates to the NetManager", func() {
			networks := boshsettings.Networks{}

			err := platform.SetupNetworking(networks)
			Expect(err).ToNot(HaveOccurred())

			Expect(netManager.SetupNetworkingNetworks).To(Equal(networks))
		})
	})

	Describe("GetDefaultNetwork", func() {
		It("delegates to the defaultNetworkResolver", func() {
			defaultNetwork := boshsettings.Network{IP: "1.2.3.4"}
			fakeDefaultNetworkResolver.GetDefaultNetworkNetwork = defaultNetwork

			network, err := platform.GetDefaultNetwork()
			Expect(err).ToNot(HaveOccurred())

			Expect(network).To(Equal(defaultNetwork))
		})
	})

	Describe("SetTimeWithNtpServers", func() {
		It("sets time with ntp servers", func() {
			servers := []string{"0.north-america.pool.ntp.org", "1.north-america.pool.ntp.org"}
			platform.SetTimeWithNtpServers(servers)

			Expect(len(cmdRunner.RunCommands)).To(Equal(6))
			Expect(cmdRunner.RunCommands[0]).To(ContainElement(ContainSubstring("new-netfirewallrule")))
			Expect(cmdRunner.RunCommands[1]).To(ContainElement(ContainSubstring("stop")))
			ntpServers := strings.Join(servers, " ")
			Expect(cmdRunner.RunCommands[2]).To(ContainElement(ContainSubstring(ntpServers)))
			Expect(cmdRunner.RunCommands[3]).To(ContainElement(ContainSubstring("start")))
			Expect(cmdRunner.RunCommands[4]).To(ContainElement(ContainSubstring("/update")))
			Expect(cmdRunner.RunCommands[5]).To(ContainElement(ContainSubstring("/resync")))
		})

		It("sets time with ntp servers is noop when no ntp server provided", func() {
			platform.SetTimeWithNtpServers([]string{})
			Expect(len(cmdRunner.RunCommands)).To(Equal(0))
		})
	})

	Describe("DeleteARPEntryWithIP", func() {
		It("cleans the arp entry for the given ip", func() {
			err := platform.DeleteARPEntryWithIP("1.2.3.4")
			deleteArpEntry := []string{"arp", "-d", "1.2.3.4"}
			Expect(cmdRunner.RunCommands[0]).To(Equal(deleteArpEntry))
			Expect(err).ToNot(HaveOccurred())
		})

		It("fails if arp command fails", func() {
			result := fakesys.FakeCmdResult{
				Error:      errors.New("failure"),
				ExitStatus: 1,
				Stderr:     "",
				Stdout:     "",
			}
			cmdRunner.AddCmdResult("arp -d 1.2.3.4", result)

			err := platform.DeleteARPEntryWithIP("1.2.3.4")

			Expect(err).To(HaveOccurred())
		})
	})
})
