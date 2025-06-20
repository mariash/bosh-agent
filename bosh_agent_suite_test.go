package bosh_agent_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBoshAgent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BoshAgent Suite")
}
