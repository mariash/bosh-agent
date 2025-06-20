package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-agent/v2/agent/action"
	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	fakeagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver/agentserverfakes"
)

var _ = Describe("DetachDynamicDisk", func() {
	var (
		detachDynamicDiskAction action.DetachDynamicDiskAction
		directorClient          *fakeagentserver.FakeDirectorClient
	)

	BeforeEach(func() {
		directorClient = &fakeagentserver.FakeDirectorClient{}
		detachDynamicDiskAction = action.NewDetachDynamicDiskAction(directorClient)
		directorClient.DetachDiskReturns(agentserver.DetachDiskDirectorResponse{}, nil)
	})

	AssertActionIsAsynchronous(detachDynamicDiskAction)
	AssertActionIsNotPersistent(detachDynamicDiskAction)
	AssertActionIsLoggable(detachDynamicDiskAction)

	AssertActionIsNotResumable(detachDynamicDiskAction)
	AssertActionIsNotCancelable(detachDynamicDiskAction)

	It("calls director to detach the disk and sets up the disk", func() {
		result, err := detachDynamicDiskAction.Run("some-disk-name")
		Expect(err).ToNot(HaveOccurred())

		Expect(result).To(Equal(action.DetachDynamicDiskTaskResult{}))
	})

	Context("when director detach disk call fails", func() {
		BeforeEach(func() {
			directorClient.DetachDiskReturns(agentserver.DetachDiskDirectorResponse{}, errors.New("some-detach-disk-error"))
		})

		It("returns an error", func() {
			_, err := detachDynamicDiskAction.Run("some-disk-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("some-detach-disk-error"))
		})
	})
})
