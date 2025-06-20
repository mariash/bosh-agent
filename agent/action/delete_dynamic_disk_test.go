package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-agent/v2/agent/action"
	"github.com/cloudfoundry/bosh-agent/v2/agentserver"
	fakeagentserver "github.com/cloudfoundry/bosh-agent/v2/agentserver/agentserverfakes"
)

var _ = Describe("DeleteDynamicDisk", func() {
	var (
		deleteDynamicDiskAction action.DeleteDynamicDiskAction
		directorClient          *fakeagentserver.FakeDirectorClient
	)

	BeforeEach(func() {
		directorClient = &fakeagentserver.FakeDirectorClient{}
		deleteDynamicDiskAction = action.NewDeleteDynamicDiskAction(directorClient)
		directorClient.DeleteDiskReturns(agentserver.DeleteDiskDirectorResponse{}, nil)
	})

	AssertActionIsAsynchronous(deleteDynamicDiskAction)
	AssertActionIsNotPersistent(deleteDynamicDiskAction)
	AssertActionIsLoggable(deleteDynamicDiskAction)

	AssertActionIsNotResumable(deleteDynamicDiskAction)
	AssertActionIsNotCancelable(deleteDynamicDiskAction)

	It("calls director to delete the disk and sets up the disk", func() {
		result, err := deleteDynamicDiskAction.Run("some-disk-name")
		Expect(err).ToNot(HaveOccurred())

		Expect(result).To(Equal(action.DeleteDynamicDiskTaskResult{}))
	})

	Context("when director delete disk call fails", func() {
		BeforeEach(func() {
			directorClient.DeleteDiskReturns(agentserver.DeleteDiskDirectorResponse{}, errors.New("some-delete-disk-error"))
		})

		It("returns an error", func() {
			_, err := deleteDynamicDiskAction.Run("some-disk-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("some-delete-disk-error"))
		})
	})
})
